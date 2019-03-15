// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "io"
)

type word_split_writer struct {
    out io.WriteCloser
    state int
    partial_word []byte
    saw_newline bool
}
func new_word_split_writer(out io.WriteCloser) *word_split_writer {
    w := new(word_split_writer)
    w.out = out
    w.partial_word = make([]byte, 0, max_word_len)
    return w
}
func (w *word_split_writer) Write(block []byte) (int, error) {
    const ( SKIP_SPACE = iota
            WORD_START
            IN_WORD
        )
    n := len(block)
    space := []byte(" \n\t\r/'\"")
    nl := []byte("\n")
    for {
        if len(block) == 0 {
            break
        }
        switch w.state {
        case WORD_START:
            i := index_any(block, space)
            if i == -1 {
                if len(block) < max_word_len {
                    w.partial_word = w.partial_word[:0]
                    w.partial_word = append(w.partial_word, block...)
                    w.state = IN_WORD
                }
                block = block[:0]
            } else {
                if i <= max_word_len && i >= min_word_len {
                    if _, err := w.out.Write(block[:i]); err != nil {
                        return 0, err
                    }
                }
                w.saw_newline = (block[i] == byte('\n'))
                block = block[i+1:]
                w.state = SKIP_SPACE
            }
        case IN_WORD:
            i := index_any(block, space)
            if i == -1 {
                if len(w.partial_word) + len(block) < max_word_len {
                    w.partial_word = append(w.partial_word, block...)
                }
                block = block[:0]
            } else {
                l := len(w.partial_word) + i
                if l >= min_word_len && l <= max_word_len {
                    w.partial_word = append(w.partial_word, block[:i]...)
                    if _, err := w.out.Write(w.partial_word); err != nil {
                        return 0, err
                    }
                }
                w.saw_newline = (block[i] == byte('\n'))
                block = block[i+1:]
                w.state = SKIP_SPACE
            }
        case SKIP_SPACE:
            var i int
            if len(block) == 0 {
                w.state = WORD_START
            }
            for i = 0; i < len(block); i++ {
                if is_space(block[i]) {
                    w.saw_newline = w.saw_newline || (block[i] == byte('\n'))
                } else {
                    w.state = WORD_START
                    break
                }
            }
            block = block[i:]
            if w.state == WORD_START {
                if w.saw_newline {
                    if _, err := w.out.Write(nl); err != nil {
                        return 0, err
                    }
                }
            }
        }
    }
    return n, nil
}
func (w *word_split_writer) Close() error {
    return w.out.Close()
}


// This writer deals with special code points that are directly
// encoded in the source charset. See also html.go which deals
// with the ones specified via html entities.
// expects full words, i.e. to be chained after split words writer
type replace_chars_writer struct {
    out io.WriteCloser
}
func new_replace_chars_writer(out io.WriteCloser) *replace_chars_writer {
    return &replace_chars_writer{out}
}
func (w *replace_chars_writer) Write(word []byte) (int, error) {
    n := len(word)
    soft_hyphen := []byte{0xc2, 0xad}
    empty := []byte{}
    var x []byte
    if bytes.Index(word, soft_hyphen) == -1 {
        x = word
    } else {
        x = bytes.Replace(word, soft_hyphen, empty, -1)
    }
    if _, err := w.out.Write(x); err != nil {
        return 0, err
    }
    return n, nil
}
func (w *replace_chars_writer) Close() error {
    return w.out.Close()
}


type word_writer struct {
    out io.WriteCloser
    word []byte
}
func new_word_writer(out io.WriteCloser) *word_writer {
    w := new(word_writer)
    w.out = out
    w.word = make([]byte, 1, max_word_len)
    w.word[0] = byte('{')
    return w
}
func (w *word_writer) Write(block []byte) (int, error) {
    if bytes.Equal(block, []byte("\n")) {
        if _, err := w.out.Write([]byte("NL\n")); err != nil {
            return 0, err
        }
    } else {
        w.word = w.word[:1]
        w.word = append(w.word, block...)
        w.word = append(w.word, []byte("}\n")...)
        if _, err := w.out.Write(w.word); err != nil {
            return 0, err
        }
    }
    return len(block), nil
}
func (w *word_writer) Close() error {
    return w.out.Close()
}

type nl_writer struct {
    out io.WriteCloser
    word []byte
}
func new_nl_writer(out io.WriteCloser) *nl_writer {
    w := new(nl_writer)
    w.out = out
    w.word = make([]byte, 0, max_word_len)
    return w
}
func (w *nl_writer) Write(block []byte) (int, error) {
    if bytes.Equal(block, []byte("\n")) {
        if _, err := w.out.Write([]byte("NL\n")); err != nil {
            return 0, err
        }
    } else {
        w.word = w.word[:0]
        w.word = append(w.word, block...)
        w.word = append(w.word, []byte("\n")...)
        if _, err := w.out.Write(w.word); err != nil {
            return 0, err
        }
    }
    return len(block), nil
}
func (w *nl_writer) Close() error {
    return w.out.Close()
}
