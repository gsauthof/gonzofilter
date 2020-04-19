// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "io"
)

type multipart_split_writer struct {
    boundary        []byte
    out             io.WriteCloser
    new_writer      func() io.WriteCloser

    state           int
    off             int
    partial_marker  []byte
}
func new_multipart_split_writer(boundary []byte,
        new_writer func() io.WriteCloser) *multipart_split_writer {
    w            := new(multipart_split_writer)
    w.boundary    = append(w.boundary, []byte("\n--")...)
    w.boundary    = append(w.boundary, boundary...)
    w.new_writer  = new_writer
    w.out         = new_dev_null_writer()
    return w
}
func (w *multipart_split_writer) Close() error {
    return w.out.Close()
}
func (w *multipart_split_writer) Write(block []byte) (int, error) {
    const ( AFTER_NEWLINE = iota
            INSIDE_BOUNDARY
            AFTER_BOUNDARY
            SKIP_TO_NEWLINE
            INSIDE_EOF
            AFTER_EOF
          )
    n := len(block)
    for len(block) != 0 {
        switch w.state {
        case AFTER_NEWLINE:
            switch {
            case bytes.HasPrefix(block, w.boundary[1:]):
                block            = block[len(w.boundary)-1:]
                w.state          = AFTER_BOUNDARY
            case block[0] == byte('-'):
                w.partial_marker = w.partial_marker[:0]
                w.partial_marker = append(w.partial_marker, byte('-'))
                block            = block[1:]
                w.state          = INSIDE_BOUNDARY
                w.off            = 2
            default:
                w.state          = SKIP_TO_NEWLINE
            }
        case INSIDE_BOUNDARY:
            i := match_prefix(block, w.boundary[w.off:])
            switch {
            case i == 0:
                w.state = SKIP_TO_NEWLINE
                if _, err := w.out.Write(w.partial_marker); err != nil {
                    return 0, err
                }
            case w.off + i == len(w.boundary):
                block   = block[i:]
                w.state = AFTER_BOUNDARY
            default:
                block            = block[i:]
                w.partial_marker = append(w.partial_marker, w.boundary[w.off:w.off+i]...)
                w.off           += i
            }
        case AFTER_BOUNDARY:
            if block[0] == byte('-') {
                block   = block[1:]
                w.state = INSIDE_EOF
                w.out.Close()
            } else {
                i := bytes.IndexByte(block, byte('\n'))
                if i == -1 {
                    block   = block[:0]
                } else {
                    block   = block[i+1:]
                    w.state = AFTER_NEWLINE
                    w.out.Close()
                    w.out   = w.new_writer()
                }
            }
        case SKIP_TO_NEWLINE:
            // take a short cut if possible
            i := bytes.Index(block, w.boundary)
            if i == -1 {
                i := bytes.IndexByte(block, byte('\n'))
                if i == -1  {
                    if _, err := w.out.Write(block); err != nil {
                        return 0, err
                    }
                    block = block[:0]
                } else {
                    if _, err := w.out.Write(block[:i+1]); err != nil {
                        return 0, err
                    }
                    block   = block[i+1:]
                    w.state = AFTER_NEWLINE
                }
            } else {
                if _, err := w.out.Write(block[:i+1]); err != nil {
                    return 0, err
                }
                block   = block[i+len(w.boundary):]
                w.state = AFTER_BOUNDARY
            }
        case INSIDE_EOF:
            if block[0] == byte('-') {
                block   = block[1:]
                w.state = AFTER_EOF
            } else {
                block   = block[:0]
            }
        case AFTER_EOF:
            block = block[:0]
        }
    }
    return n, nil
}

