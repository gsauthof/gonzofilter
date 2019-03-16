// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "html"
    "io"
)

var entity_trans_map = map[string][]byte {
    // soft hyphen
    "&#173;" : []byte(""),
    "&shy;"  : []byte(""),
    // zero-width non-joiner
    "&#8204;": []byte(""),
    "&zwnj;" : []byte(""),
    // non-breaking space
    "&#160;" : []byte(" "),
    "&nbsp;" : []byte(" "),
}

func unescape_entity(s []byte) []byte {
    // Go map don't support []byte slice keys because Go doesn't
    // define equal on them ...
    if v := entity_trans_map[string(s)]; v == nil {
        r := html.UnescapeString(string(s))
        return []byte(r)
    } else {
        return v
    }
}

type replace_entities_writer struct {
    out io.WriteCloser
    state int
    partial_entity []byte
}
func new_replace_entities_writer(out io.WriteCloser) *replace_entities_writer {
    w := new(replace_entities_writer)
    w.out = out
    w.partial_entity = make([]byte, 1, max_entity_len)
    w.partial_entity[0] = byte('&')
    return w
}
func (w *replace_entities_writer) Write(block []byte) (int, error) {
    const ( OUTSIDE = iota
            AFTER_ET
            IN_ENTITY
        )
    n := len(block)
    for {
        if len(block) == 0 {
            break
        }
        switch w.state {
        case OUTSIDE:
            i := bytes.IndexByte(block, byte('&'))
            if i == -1 {
                if _, err := w.out.Write(block); err != nil {
                    return 0, err
                }
                block = block[:0]
            } else {
                if _, err := w.out.Write(block[:i]); err != nil {
                    return 0, err
                }

                j := index_any(block[i+1:], []byte(";<"))
                if j != -1 && block[i+1+j] == byte(';') && !is_space(block[i+1]) && j + 1 < max_entity_len {
                    // fast path
                    s := unescape_entity(block[i:i+1+j+1])
                    if _, err := w.out.Write([]byte(s)); err != nil {
                        return 0, err
                    }
                    block = block[i+1+j+1:]
                } else {
                    block = block[i+1:]
                    w.state = AFTER_ET
                    w.partial_entity = w.partial_entity[:1]
                }
            }
        case AFTER_ET:
            // violates the spec, but & may show up as-is in some HTML
            if is_space(block[0]) {
                if _, err := w.out.Write(w.partial_entity); err != nil {
                    return 0, err
                }
                w.state = OUTSIDE
            } else {
                w.state = IN_ENTITY
            }
        case IN_ENTITY:
            i := index_any(block, []byte(";<"))
            if i == -1 {
                if len(w.partial_entity) + len(block) > max_entity_len {
                    if _, err := w.out.Write(w.partial_entity); err != nil {
                        return 0, err
                    }
                    w.state = OUTSIDE
                } else {
                    w.partial_entity = append(w.partial_entity, block...)
                    block = block[:0]
                }
            } else {
                if block[i] == byte('<') {
                    if _, err := w.out.Write(w.partial_entity); err != nil {
                        return 0, err
                    }
                    w.state = OUTSIDE
                } else {
                    if len(w.partial_entity) + i > max_entity_len {
                        if _, err := w.out.Write(w.partial_entity); err != nil {
                            return 0, err
                        }
                        w.state = OUTSIDE
                    } else {
                        w.partial_entity = append(w.partial_entity, block[:i+1]...)
                        s := unescape_entity(w.partial_entity)
                        if _, err := w.out.Write([]byte(s)); err != nil {
                            return 0, err
                        }
                        block = block[i+1:]
                        w.state = OUTSIDE
                    }
                }
            }
        }
    }
    return n, nil
}
func (w *replace_entities_writer) Close() error {
    return w.out.Close()
}

type remove_tags_writer struct {
    out io.WriteCloser
    state int
    off int
}
func new_remove_tags_writer(out io.WriteCloser) *remove_tags_writer {
    w := new(remove_tags_writer)
    w.out = out
    return w
}
func (w *remove_tags_writer) Write(block []byte) (int, error) {
    const ( OUTSIDE = iota
            IN_TAG
            FINISH_TAG
            FINISH_END_STYLE
            FINISH_END_STYLE2
            IN_DECL_COMMENT
            FINISH_COMMENT
            FINISH_COMMENT2
            IN_STYLE_SCRIPT
            IN_STYLE
            IN_SCRIPT
            AFTER_STYLE_SCRIPT
        )
    n := len(block)
    space := []byte(" ")
    yle := []byte("yle")
    ript := []byte("ript")
    end_style := []byte("</style>")
    end_comment := []byte("-->")
    for {
        if len(block) == 0 {
            break
        }
        switch w.state {
        case OUTSIDE:
            i := bytes.IndexByte(block, byte('<'))
            if i == -1 {
                if _, err := w.out.Write(block); err != nil {
                    return 0, err
                }
                block = block[:0]
            } else {
                if _, err := w.out.Write(block[:i]); err != nil {
                    return 0, err
                }
                w.state = IN_TAG
                block = block[i+1:]
            }
        case IN_TAG:
            if block[0] == byte('s') || block[0] == byte('S') {
                block = block[1:]
                w.state = IN_STYLE_SCRIPT
            } else if block[0] == byte('!') {
                block = block[1:]
                w.state = IN_DECL_COMMENT
            } else {
                w.state = FINISH_TAG
            }
        case IN_DECL_COMMENT:
            if block[0] == byte('-') {
                w.state = FINISH_COMMENT
            } else {
                w.state = FINISH_TAG
            }
            block = block[1:]
        case FINISH_COMMENT:
            i := bytes.IndexByte(block, byte('-'))
            if i == -1 {
                block = block[:0]
            } else {
                w.off = 1
                block = block[i+1:]
                w.state = FINISH_COMMENT2
            }
        case FINISH_COMMENT2:
            i := imatch_prefix(block, end_comment[w.off:])
            if i > 0 {
                w.off += i
                block = block[i:]
                if w.off == len(end_comment) {
                    w.state = OUTSIDE
                }
            } else {
                w.state = FINISH_COMMENT
            }
        case FINISH_TAG:
            i := bytes.IndexByte(block, byte('>'))
            if i == -1 {
                block = block[:0]
            } else {
                w.state = OUTSIDE
                block = block[i+1:]
                if _, err := w.out.Write(space); err != nil {
                    return 0, err
                }
            }
        case IN_STYLE_SCRIPT:
            if block[0] == byte('t') || block[0] == byte('T') {
                block = block[1:]
                w.state = IN_STYLE
                w.off = 0
            } else if block[0] == byte('c') || block[0] == byte('C') {
                block = block[1:]
                w.state = IN_SCRIPT
                w.off = 0
            } else {
                w.state = FINISH_TAG
            }
        case IN_STYLE:
            i := imatch_prefix(block, yle[w.off:])
            if i > 0 {
                w.off += i
                block = block[i:]
                if w.off == len(yle) {
                    w.state = FINISH_END_STYLE
                }
            } else {
                w.state = FINISH_TAG
            }
        case FINISH_END_STYLE:
            i := bytes.IndexByte(block, byte('<'))
            if i == -1 {
                block = block[:0]
            } else {
                block = block[i+1:]
                w.state = FINISH_END_STYLE2
                w.off = 1
            }
        case FINISH_END_STYLE2:
            i := imatch_prefix(block, end_style[w.off:])
            if i > 0 {
                w.off += i
                block = block[i:]
                if w.off == len(end_style) {
                    w.state = OUTSIDE
                }
            } else {
                w.state = FINISH_END_STYLE
            }
        case IN_SCRIPT:
            i := imatch_prefix(block, ript[w.off:])
            if i > 0 {
                w.off += i
                block = block[i:]
                if w.off == len(yle) {
                    w.state = AFTER_STYLE_SCRIPT
                }
            } else {
                w.state = FINISH_TAG
            }
        case AFTER_STYLE_SCRIPT:
            i := bytes.IndexByte(block, byte('>'))
            if i == -1 {
                block = block[:0]
            } else {
                block = block[i+1:]
                w.state = FINISH_TAG
            }
        }
    }
    return n, nil
}
func (w *remove_tags_writer) Close() error {
    return w.out.Close()
}

