// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "io"
)

type shrink_space_writer struct {
    out          io.WriteCloser
    state        int
    saw_newline  bool
}
func new_shrink_space_writer(out io.WriteCloser) *shrink_space_writer {
    w     := new(shrink_space_writer)
    w.out  = out
    return w
}
func (w *shrink_space_writer) Write(block []byte) (int, error) {
    const ( OUTSIDE = iota
            SKIP_SOME
          )
    n     := len(block)
    space := []byte(" \n\t\r")
    end   := 0
    for len(block) != 0 {
        switch w.state {
        case OUTSIDE:
            end  = 0
            t   := block
            for {
                i := index_any(t, space)
                if i == -1 {
                    end += len(t)
                    break
                } else {
                    if i + 1 < len(t) {
                        if is_space(t[i+1]) {
                            end           += i
                            w.saw_newline  = (t[i] == byte('\n'))
                            w.state        = SKIP_SOME
                            break
                        } else {
                            t    = t[i+2:]
                            end += i + 2
                        }
                    } else {
                        end           += i
                        w.saw_newline  = (t[i] == byte('\n'))
                        w.state        = SKIP_SOME
                        break
                    }
                }
            }
            if _, err := w.out.Write(block[:end]); err != nil {
                return 0, err
            }
            block = block[end:]
        case SKIP_SOME:
            var i int
            for i = 0; i < len(block); i++ {
                if is_space(block[i]) {
                    w.saw_newline = w.saw_newline || (block[i] == byte('\n'))
                } else {
                    w.state       = OUTSIDE
                    break
                }
            }
            block = block[i:]
            if w.state == OUTSIDE {
                if w.saw_newline {
                    if _, err := w.out.Write([]byte("\n")); err != nil {
                        return 0, err
                    }
                } else {
                    if _, err :=  w.out.Write([]byte(" ")); err != nil {
                        return 0, err
                    }
                }
            }
        }
    }
    return n, nil
}
func (w *shrink_space_writer) Close() error {
    if _, err := w.out.Write([]byte("\n")); err != nil {
        return err
    }
    return w.out.Close()
}



