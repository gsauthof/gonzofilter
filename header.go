// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "encoding/base64"
    "encoding/hex"
    "io"
)


func mk_base64_decoder() func (block []byte) []byte {
    old := make([]byte, 0, 4)
    return func(s []byte) []byte {
        a := len(old)
        b := len(s)
        n := a + b
        m := n / 4 * 4
        d := n - m
        if m == 0 {
            old = append(old, s...)
            return []byte("")
        } else {
            l := base64.StdEncoding.DecodedLen(m)
            t := make([]byte, l)
            if len(old) > 0 {
                u := make([]byte, n)
                copy(u, old)
                copy(u[a:], s)
                k, _ := base64.StdEncoding.Decode(t, u[:m])
                t = t[:k]
                old = old[:0]
                if d > 0 {
                    old = append(old, u[m:]...)
                }
            } else {
                k, _ := base64.StdEncoding.Decode(t, s[:m])
                t = t[:k]
                if d > 0 {
                    old = append(old, s[m:]...)
                }
            }
            return t
        }
    }
}


type header_decode_writer struct {
    out io.WriteCloser
    encoding int
    state int
    base64_decoder func (block []byte) []byte
    cw *charset_writer
    partial_quoted []byte
    charset []byte
}
func new_header_decode_writer(out io.WriteCloser) *header_decode_writer {
    w := new(header_decode_writer)
    w.out = out
    w.base64_decoder = mk_base64_decoder()
    w.partial_quoted = make([]byte, 0, 1)
    w.charset = make([]byte, 16)
    return w
}

func (w *header_decode_writer) Close() error {
    return w.out.Close()
}

func (w *header_decode_writer) Write(block []byte) (int, error) {
    const ( OUTSIDE = iota
        IN_BEGIN
        IN_CHARSET
        IN_ENC
        IN_ENC2
        IN_TEXT
        IN_BASE64
        IN_QUOTED
        IN_END
    )
    const ( QUOTED_ENCODING = iota
        BASE64_ENCODING
    )

    space := []byte(" ")
    equal := []byte("=")
    n := len(block)
    if n == 0 {
        return 0, nil
    }
    for {
        if len(block) == 0 {
            break
        }
        switch w.state {
        case OUTSIDE:
            i := bytes.Index(block, []byte("=?"))
            if i == -1 {
                if bytes.HasSuffix(block, []byte("=")) {
                    if _, err := w.out.Write(block[:len(block)-1]); err != nil {
                        return 0, err
                    }
                    w.state = IN_BEGIN
                } else {
                    if _, err := w.out.Write(block); err != nil {
                        return 0, err
                    }
                }
                block = block[:0]
            } else {
                if _, err := w.out.Write(block[:i]); err != nil {
                    return 0, err
                }
                block = block[i+2:]
                w.state = IN_CHARSET
                w.charset = w.charset[:0]
            }
        case IN_BEGIN:
            if bytes.HasPrefix(block, []byte("?")) {
                w.state = IN_CHARSET
                w.charset = w.charset[:0]
            } else {
                if _, err := w.out.Write(equal); err != nil {
                    return 0, err
                }
                w.state = OUTSIDE
            }
            block = block[1:]
        case IN_CHARSET:
            i := bytes.IndexByte(block, byte('?'))
            if i == -1 {
                w.charset = append(w.charset, block...)
                block = block[:0]
            } else {
                w.charset = append(w.charset, block[:i]...)
                p := bytes.IndexByte(w.charset, byte('*'))
                if p != -1 {
                    w.charset = w.charset[:p]
                }
                w.cw = new_charset_writer(w.charset, w.out)
                block = block[i+1:]
                w.state = IN_ENC
            }
        case IN_ENC:
            switch block[0] {
            case byte('q'):
                w.state = IN_ENC2
                w.encoding = QUOTED_ENCODING
            case byte('Q'):
                w.state = IN_ENC2
                w.encoding = QUOTED_ENCODING
            case byte('b'):
                w.state = IN_ENC2
                w.encoding = BASE64_ENCODING
            case byte('B'):
                w.state = IN_ENC2
                w.encoding = BASE64_ENCODING
            default:
                w.state = OUTSIDE
            }
            block = block[1:]
        case IN_ENC2:
            if block[0] == byte('?') {
                switch w.encoding {
                case QUOTED_ENCODING:
                    w.state = IN_TEXT
                case BASE64_ENCODING:
                    w.state = IN_BASE64
                default:
                    w.state = OUTSIDE
                }
            } else {
                debugf("Malformed w.encoding: %v\n", block[0])
                w.state = OUTSIDE
            }
            block = block[1:]
        case IN_BASE64:
            i := bytes.Index(block, []byte("?="))
            off := 0
            if i == -1 {
                if bytes.HasSuffix(block, []byte("?")) {
                    w.state = IN_END
                    i = len(block)-1
                    off = 1
                } else {
                    i = len(block)
                }
            } else {
                w.state = OUTSIDE
                off = 2
            }
            b64 := block[:i]
            block = block[i+off:]
            t := w.base64_decoder(b64)
            if _, err := w.cw.Write(t); err != nil {
                return 0, err
            }
            if w.state != IN_BASE64 {
                w.base64_decoder = mk_base64_decoder()
                w.cw.Close()
            }
        case IN_TEXT:
            var end int
            var endP int
            i := bytes.Index(block, []byte("?="))
            if i == -1 {
                if bytes.HasSuffix(block, []byte("?")) {
                    w.state = IN_END
                    end = len(block) - 1
                    endP = len(block)
                } else {
                    end = len(block)
                    endP = len(block)
                }
            } else {
                w.state = OUTSIDE
                end = i
                endP = i +2
            }
            quoted := block[:end]
            for {
                if len(quoted) == 0 {
                    break
                }
                i := index_any(quoted, []byte("=_"))
                if i == -1 {
                    if _, err := w.cw.Write(quoted); err != nil {
                        return 0, err
                    }
                    quoted = quoted[:0]
                } else {
                    if i > 0 {
                        w.cw.Write(quoted[:i])
                        quoted = quoted[i:]
                    }
                    switch quoted[0] {
                    case byte('_'):
                        if _, err := w.cw.Write(space); err != nil {
                            return 0, err
                        }
                        quoted = quoted[1:]
                    case byte('='):
                        quoted = quoted[1:]
                        switch {
                        case len(quoted) > 1:
                            t := make([]byte, 1)
                            hex.Decode(t, quoted[:2])
                            if _, err := w.cw.Write(t); err != nil {
                                return 0, err
                            }
                            quoted = quoted[2:]
                        case len(quoted) == 1:
                            if (end != endP) {
                                w.state = OUTSIDE
                            } else {
                                w.partial_quoted = append(w.partial_quoted, quoted...)
                                quoted = quoted[:0]
                                w.state = IN_QUOTED
                            }
                        case len(quoted) == 0:
                            if (end != endP) {
                                w.state = OUTSIDE
                            } else {
                                quoted = quoted[:0]
                                w.state = IN_QUOTED
                            }
                        }
                    }
                }
            }
            block = block[endP:]
            if w.state == OUTSIDE {
                w.cw.Close()
            }
        case IN_QUOTED:
            if len(w.partial_quoted) == 0 && len(block) == 1 {
                w.partial_quoted = append(w.partial_quoted, block[0])
                block = block[:0]
            } else {
                t := make([]byte, 1)
                if len(w.partial_quoted) == 0 {
                    hex.Decode(t, block[:2])
                    block = block[2:]
                } else {
                    x := []byte{w.partial_quoted[0], block[0]}
                    hex.Decode(t, x)
                    block = block[1:]
                }
                w.partial_quoted = w.partial_quoted[:0]
                w.state = IN_TEXT
            }
        case IN_END:
            if block[0] == byte('=') {
                block = block[1:]
            } else {
                debugf("incomplete end marker: %v", block[0])
            }
            w.state = OUTSIDE
        }
    }
    return n, nil
}

type extract_header_writer struct {
    // text/plain, multipart ...
    content_type []byte
    // base64, quoted-printable, 7bit, 8bit, binary ....
    transfer_encoding []byte

    typ []byte
    charset []byte
    boundary []byte

    state int
    off int
}
func new_extract_header_writer() *extract_header_writer {
    w := new(extract_header_writer)
    w.content_type = make([]byte, 0, 16)
    w.transfer_encoding = make([]byte, 0, 16)

    return w
}
func (w *extract_header_writer) parse() {
    w.typ = w.typ[:0]
    w.charset = w.charset[:0]
    w.boundary = w.boundary[:0]
    xs := bytes.Split(w.content_type, []byte(";"))
    if len(xs) > 0 {
        w.typ = append(w.typ, bytes.ToLower(xs[0])...)
        for _, x := range xs[1:] {
            y := bytes.TrimSpace(x)
            if bytes.HasPrefix(y, []byte("charset=")) {
                w.charset = append(w.charset, bytes.Trim(y[8:], `"`)...)
            } else if bytes.HasPrefix(y, []byte("boundary=")) {
                w.boundary = append(w.boundary, bytes.Trim(y[9:], `"`)...)
            }
        }
    }
    if len(w.typ) == 0 {
        w.typ = append(w.typ, []byte("text/plain")...)
    }
}


func (w *extract_header_writer) Close() error {
    return nil
}
func (w *extract_header_writer) Write(block []byte) (int, error) {
    const ( AFTER_NEWLINE = iota
            PARTIAL_PREFIX
            AFTER_PREFIX
            PARTIAL_CONTENT_TYPE
            CONTENT_TYPE_VALUE
            PARTIAL_TRANSFER_ENCODING
            TRANSFER_ENCODING_VALUE
            SKIP_TO_NEWLINE
    )
    n := len(block)
    // the comparison is case-insensitive in the 1st argument
    content_t := []byte("content-t")
    ype := []byte("ype: ")
    ransfer_encoding := []byte("ransfer-encoding: ")
    for {
        if len(block) == 0 {
            break
        }
        switch w.state {
        case AFTER_NEWLINE:
            w.off = 0
            fallthrough
        case PARTIAL_PREFIX:
            i := imatch_prefix(block, content_t[w.off:])
            if i > 0 {
                w.off = w.off + i
                if w.off == len(content_t) {
                    w.state = AFTER_PREFIX
                } else {
                    w.state = PARTIAL_PREFIX
                }
                block = block[i:]
            } else  {
                w.state = SKIP_TO_NEWLINE
            }
        case AFTER_PREFIX:
            w.off = 0
            i := imatch_prefix(block, ype[w.off:])
            if i > 0 {
                w.off = w.off + i
                if w.off == len(ype) {
                    w.state = CONTENT_TYPE_VALUE
                    w.content_type = w.content_type[:0]
                } else {
                    w.state = PARTIAL_CONTENT_TYPE
                }
                block = block[i:]
            } else {
                i := imatch_prefix(block, ransfer_encoding[w.off:])
                if i > 0 {
                    w.off = w.off + i
                    if w.off == len(ransfer_encoding) {
                        w.state = TRANSFER_ENCODING_VALUE
                        w.transfer_encoding = w.transfer_encoding[:0]
                    } else {
                        w.state = PARTIAL_TRANSFER_ENCODING
                    }
                    block = block[i:]
                } else {
                    w.state = SKIP_TO_NEWLINE
                }
            }
        case PARTIAL_CONTENT_TYPE:
            i := imatch_prefix(block, ype[w.off:])
            if i > 0 {
                w.off = w.off + i
                if w.off == len(ype) {
                    w.state = CONTENT_TYPE_VALUE
                    w.content_type = w.content_type[:0]
                }
                block = block[i:]
            } else {
                w.state = SKIP_TO_NEWLINE
            }
        case CONTENT_TYPE_VALUE:
            i := bytes.IndexByte(block, byte('\n'))
            if i == -1 {
                i = len(block)
                w.content_type = append(w.content_type, block[:i]...)
                block = block[:0]
            } else {
                w.content_type = append(w.content_type, block[:i]...)
                block = block[i+1:]
                w.state = AFTER_NEWLINE
            }
        case PARTIAL_TRANSFER_ENCODING:
            i := imatch_prefix(block, ransfer_encoding[w.off:])
            if i > 0 {
                w.off = w.off + i
                if w.off == len(ransfer_encoding) {
                    w.state = TRANSFER_ENCODING_VALUE
                    w.transfer_encoding = w.transfer_encoding[:0]
                }
                block = block[i:]
            } else {
                w.state = SKIP_TO_NEWLINE
            }
        case TRANSFER_ENCODING_VALUE:
            i := bytes.IndexByte(block, byte('\n'))
            if i == -1 {
                i = len(block)
                w.transfer_encoding = append(w.transfer_encoding, block[:i]...)
                block = block[:0]
            } else {
                w.transfer_encoding = append(w.transfer_encoding, block[:i]...)
                block = block[i+1:]
                w.state = AFTER_NEWLINE
            }
        case SKIP_TO_NEWLINE:
            i := bytes.IndexByte(block, byte('\n'))
            if i == -1  {
                block = block[:0]
            } else {
                block = block[i+1:]
                w.state = AFTER_NEWLINE
            }
        }
    }
    return n, nil
}


