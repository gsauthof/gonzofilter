// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "io"
)

type unfold_writer struct {
    out    io.WriteCloser
    state  int
}
func new_unfold_writer(out io.WriteCloser) *unfold_writer {
    w     := new(unfold_writer)
    w.out  = out
    return w
}
func (w *unfold_writer) Close() error {
    newline := []byte("\n")
    if _, err := w.out.Write(newline); err != nil {
        return err
    }
    return w.out.Close()
}
func (w *unfold_writer) Write(block []byte) (int, error) {
    const ( OUTSIDE = iota
            IN_FOLD
          )
    n       := len(block)
    newline := []byte("\n")
    for len(block) != 0 {
        switch w.state {
        case OUTSIDE:
            i := bytes.IndexByte(block, byte('\n'))
            if i == -1 {
                if _, err := w.out.Write(block); err != nil {
                    return 0, err
                }
                block = block[:0]
            } else {
                if i+1 < len(block) {
                    if block[i+1] == byte(' ') || block[i+1] == byte('\t') {
                        if _, err := w.out.Write(block[:i]); err != nil {
                            return 0, err
                        }
                        block = block[i+2:]
                    } else {
                        if _, err := w.out.Write(block[:i+1]); err != nil {
                            return 0, err
                        }
                        block = block[i+1:]
                    }
                } else {
                    if _, err := w.out.Write(block[:i]); err != nil {
                        return 0, err
                    }
                    block   = block[i+1:]
                    w.state = IN_FOLD
                }
            }
        case IN_FOLD:
            if block[0] == byte(' ') || block[0] == byte('\t') {
                //w.out.Write(block[:1])
                block = block[1:]
            } else {
                if _, err := w.out.Write(newline); err != nil {
                    return 0, err
                }
            }
            w.state = OUTSIDE
        }
    }
    return n, nil
}
