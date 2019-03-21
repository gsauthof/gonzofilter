// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "io"
    "regexp"
)


type header_filter_writer struct {
    out io.WriteCloser
    state int
}
func new_header_filter_writer(out io.WriteCloser) *header_filter_writer {
    w := new(header_filter_writer)
    w.out = out
    return w
}
// expects full words - i.e. to be chained after the word_split_writer
func (w *header_filter_writer) Write(word []byte) (int, error) {
    const ( AT_NAME = iota
            IGNORE_VALUE
            WRITE_LINE
            FILTER_RECEIVED
            FILTER_CT
        )
    nl := []byte("\n")
    ignore_hdrs := [][]byte {
        []byte("date:"), []byte("message-id:"), []byte("references:"),
        []byte("x-bogosity:"), []byte("message-id-hash:"),
        []byte("x-message-id-hash:"),
        []byte("x-gonzo:"), []byte("x-old-gonzo"),
    }
    n := len(word)
    switch w.state {
    case AT_NAME:
        if iequal_any(word, ignore_hdrs) {
            w.state = IGNORE_VALUE
        } else if ihas_prefix(word, []byte("x-spam-")) {
            w.state = IGNORE_VALUE
        } else if ihas_suffix(word, []byte("dkim-signature:")) {
            w.state = IGNORE_VALUE
        } else if iequal(word, []byte("received:")) {
            w.state = FILTER_RECEIVED
            if _, err := w.out.Write(word); err != nil {
                return 0, err
            }
        } else if iequal(word, []byte("content-type:")) {
            w.state = FILTER_CT
            if _, err := w.out.Write(word); err != nil {
                return 0, err
            }
        } else {
            w.state = WRITE_LINE
            if _, err := w.out.Write(word); err != nil {
                return 0, err
            }
        }
    case IGNORE_VALUE:
        if n == 1 && word[0] == byte('\n') {
            w.state = AT_NAME
        }
    case WRITE_LINE:
        if n == 1 && word[0] == byte('\n') {
            w.state = AT_NAME
        }
        if _, err := w.out.Write(word); err != nil {
            return 0, err
        }
    case FILTER_RECEIVED:
        if _, err := w.out.Write(word); err != nil {
            return 0, err
        }
        if bytes.HasSuffix(word, []byte("id")) {
            w.state = IGNORE_VALUE
            if _, err := w.out.Write(nl); err != nil {
                return 0, err
            }
        } else if n == 1 && word[0] == byte('\n') {
            w.state = AT_NAME
        }
    case FILTER_CT:
        if _, err := w.out.Write(word); err != nil {
            return 0, err
        }
        if ihas_prefix(word, []byte("boundary")) {
            w.state = IGNORE_VALUE
            if _, err := w.out.Write(nl); err != nil {
                return 0, err
            }
        } else if n == 1 && word[0] == byte('\n') {
            w.state = AT_NAME
        }
    }
    return n, nil
}
func (w *header_filter_writer) Close() error {
    return w.out.Close()
}

type xgonzo_filter_writer struct {
    out io.WriteCloser
    state int
    partial []byte
    off int
    xgonzo_re *regexp.Regexp
}
func new_xgonzo_filter_writer(out io.WriteCloser) *xgonzo_filter_writer {
    w := new(xgonzo_filter_writer)
    w.out = out
    w.partial = make([]byte, 0, 9)
    w.off = 1
    w.xgonzo_re = regexp.MustCompile("(?i)\nx-gonzo: ")
    return w
}
func (w *xgonzo_filter_writer) Write(block []byte) (int, error) {
    const ( IN_GONZO = iota
            AFTER_GONZO
        )
    n := len(block)
    xgonzo := []byte("\nx-gonzo: ")
    old_xgonzo := []byte("X-old-gonzo: ")
    for len(block) != 0 {
        switch w.state {
        case IN_GONZO:
            i := imatch_prefix(block, xgonzo[w.off:])
            switch {
            case i + w.off == len(xgonzo):
                w.out.Write(old_xgonzo)
                block = block[i:]
                w.state = AFTER_GONZO
            case i > 0 && i == len(block):
                w.partial = append(w.partial, block[:i]...)
                w.off += i
                block = block[i:]
            default:
                if len(w.partial) > 0 {
                    w.out.Write(w.partial)
                }
                w.state = AFTER_GONZO
            }
        case AFTER_GONZO:
            i := -1
            if loc := w.xgonzo_re.FindIndex(block); loc != nil {
                i = loc[0] // til we have a case-insenstive bytes.Index version
            }
            if i == -1 {
                i := bytes.LastIndexByte(block, byte('\n'))
                if i == -1 {
                    w.out.Write(block)
                    block = block[:0]
                } else {
                    l := len(block) - i
                    if l < len(xgonzo) && imatch_prefix(block[i:], xgonzo) == l {
                        w.partial = w.partial[:0]
                        w.partial = append(w.partial, block[i+1:]...)
                        w.out.Write(block[:i+1])
                        w.off = l
                        w.state = IN_GONZO
                    } else {
                        w.out.Write(block)
                    }
                    block = block[:0]
                }
            } else {
                w.out.Write(block[:i+1])
                block = block[i+len(xgonzo):]
                w.out.Write(old_xgonzo)
            }
        }
    }
    return n, nil
}
func (w *xgonzo_filter_writer) Close() error {
    return w.out.Close()
}
