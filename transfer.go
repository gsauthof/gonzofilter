// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "encoding/base64"
    "io"
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
                n, decode_err := w.decoder.Read(block)
                // i.e. even if decoder encounters an error it might return some partial decode, as well!
                // we thus check for it after writing the preceding decoder result ...
                if n > 0 {
                    block = block[:n]
                    if _, err := w.out.Write(block); err != nil {
                        debugf("failed to write pipe result: %v", err)
                        w.pipe_out.CloseWithError(err)
                        break;
                    }
                }
                // signal write-on-closed-pipe in case writer tries to write
                // some trailing bytes - otherwise it could deadlock
                if decode_err == io.EOF || n == 0 {
                    w.pipe_out.Close()
                    break;
                }
                if decode_err != nil {
                    debugf("decode_writer decoder read error: %v (n=%d)", decode_err, n)
                    w.pipe_out.CloseWithError(decode_err)
                    break;
                }
            }
            w.pipe_out_done <- struct{}{} // i.e. an empty struct
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
