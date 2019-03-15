// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "io"
)

type mark_copy_header_writer struct {
    out io.WriteCloser
    prefix byte
    name []byte
    state int
}
func new_mark_copy_header_writer(prefix byte, out io.WriteCloser) *mark_copy_header_writer {
    w := new(mark_copy_header_writer)
    w.out = out
    w.prefix = prefix
    return w
}
// XXX optimize the small allocations? e.g. use large allocations and
// cut slices from those?
// expects full words - i.e. to be chained after the word_split_writer
func (w *mark_copy_header_writer) Write(word []byte) (int, error) {
    const ( AT_NAME = iota
            WRITE_LINE
        )
    too_long := []byte("h:X-gonzo-fake-too-long:")
    n := len(word)
    switch w.state {
    case AT_NAME:
        if n == 1 && word[0] == byte('\n') {
            w.name = w.name[:0]
            w.name = append(w.name, too_long...)
            if _, err := w.out.Write(too_long); err != nil {
                return 0, err
            }
        } else if n > 0 && word[n-1] != byte(':') {
            w.name = w.name[:0]
            w.name = append(w.name, too_long...)
            if _, err := w.out.Write(too_long); err != nil {
                return 0, err
            }

            t := make([]byte, len(w.name) + n)
            copy(t, w.name)
            copy(t[len(w.name):], word)
            if _, err := w.out.Write(t); err != nil {
                return 0, err
            }
            w.state = WRITE_LINE
        } else {
            w.name = w.name[:0]
            // yes, Go doesn't support mixing word... with the other args
            w.name = append(w.name, w.prefix, byte(':'))
            w.name = append(w.name, word...)
            w.state = WRITE_LINE
            if _, err := w.out.Write(w.name); err != nil {
                return 0, err
            }
        }
    case WRITE_LINE:
        if n == 1 && word[0] == byte('\n') {
            w.state = AT_NAME
        } else if n < min_word_len {
            // skip it
        } else {
            t := make([]byte, len(w.name) + n)
            copy(t, w.name)
            copy(t[len(w.name):], word)
            if _, err := w.out.Write(t); err != nil {
                return 0, err
            }
        }
    }
    return n, nil
}
func (w *mark_copy_header_writer) Close() error {
    return w.out.Close()
}

type mark_copy_body_writer struct {
    out io.WriteCloser
    prefix byte
}
func new_mark_copy_body_writer(prefix byte, out io.WriteCloser) *mark_copy_body_writer {
    w := new(mark_copy_body_writer)
    w.out = out
    w.prefix = prefix
    return w
}
// XXX optimize the small allocations? e.g. use large allocations and
// cut slices from those?
// expects full words - i.e. to be chained after the word_split_writer
func (w *mark_copy_body_writer) Write(word []byte) (int, error) {
    n := len(word)
    if n == 1 && word[0] == byte('\n') {
        return n, nil
    }
    t := make([]byte, 2 + len(word))
    t[0] = w.prefix
    t[1] = byte(':')
    copy(t[2:], word)
    if _, err := w.out.Write(t); err != nil {
        return 0, err
    }
    return n, nil
}
func (w *mark_copy_body_writer) Close() error {
    return w.out.Close()
}
