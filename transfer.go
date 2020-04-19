// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "encoding/base64"
    "io"
    "log"
    "mime/quotedprintable"
)

type decode_writer struct {
    decoder        io.Reader
    out            io.WriteCloser
    pipe_out       *io.PipeReader
    pipe_in        *io.PipeWriter
    pipe_out_done  chan struct{}
}

func (w *decode_writer) Write(p []byte) (n int, err error) {
    if w.decoder == nil {
        return w.out.Write(p)
    } else {
        return w.pipe_in.Write(p)
    }
}
func (w *decode_writer) Close() error {
    if w.decoder == nil {
        return w.out.Close()
    } else {
        w.pipe_in.Close()
        <- w.pipe_out_done
        return w.out.Close()
    }
}

func new_decode_writer(mk_decoder func(io.Reader)io.Reader, out io.WriteCloser) io.WriteCloser {
    w     := new(decode_writer)
    w.out  = out
    if mk_decoder != nil {
        w.pipe_out, w.pipe_in = io.Pipe()
        w.decoder             = mk_decoder(w.pipe_out)
        w.pipe_out_done       = make(chan struct{})
        go func() {
            block := make([]byte, read_size)
            for {
                block  = block[:cap(block)]
                n, _  := w.decoder.Read(block)
                if n == 0 {
                    break
                }
                block = block[:n]
                if _, err := w.out.Write(block); err != nil {
                    log.Fatal(err)
                }
            }
            w.pipe_out_done <- struct{}{}
            close(w.pipe_out_done)
        }()
    }
    return w
}

func new_transfer_decode_writer(encoding []byte,
                                out io.WriteCloser) io.WriteCloser {
    var mk_decoder func(io.Reader) io.Reader
    if bytes.Equal(encoding, []byte("base64")) {
        mk_decoder = func(r io.Reader) io.Reader {
            return base64.NewDecoder(base64.StdEncoding, r)
        }
    } else if bytes.Equal(encoding, []byte("quoted-printable")) {
        mk_decoder = func(r io.Reader) io.Reader {
            return quotedprintable.NewReader(r)
        }
    }
    return new_decode_writer(mk_decoder, out)
}
