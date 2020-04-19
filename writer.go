// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bufio"
    "io"
    "log"
    "os"
)

func open_input(filename string) io.Reader {
    if filename == "-" {
        return os.Stdin
    } else {
        f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
        if err != nil {
            log.Fatal(err)
        }
        return f
        // we don't need to a buffered file because we do the reads
        // with an optimal block size
        // return bufio.NewReader(f)
    }
}

func open_output(filename string) io.WriteCloser {
    h, err := new_buf_file_writer(filename)
    if err != nil {
        log.Fatal(err)
    }
    return h
}

func open_output_or_stdout(filename string) io.WriteCloser {
    var out io.WriteCloser
    if filename == "-" {
        out = new_buf_fd_writer(os.Stdout)
    } else {
        out = open_output(filename)
    }
    return out
}



// i.e. to wrap the last stream - e.g. stdout
type keep_open_writer struct {
    out io.WriteCloser
}

func (w *keep_open_writer) Write(block []byte) (int, error) {
    return w.out.Write(block)
}
func (w *keep_open_writer) Close() error {
    return nil
}

func new_keep_open_writer(out io.WriteCloser) *keep_open_writer {
    return &keep_open_writer{out}
}

type close_writer struct {
    out io.Writer
}
func (w *close_writer) Write(block []byte) (int, error) {
    return w.out.Write(block)
}
func (w *close_writer) Close() error {
    return nil
}
func new_close_writer(out io.Writer) *close_writer {
    return &close_writer{out}
}



type dev_null_writer struct {
}
func new_dev_null_writer() *dev_null_writer {
    return &dev_null_writer{}
}
func (w *dev_null_writer) Close() error {
    return nil
}
func (w *dev_null_writer) Write(block []byte) (int, error) {
    return len(block), nil
}

type buf_file_writer struct {
    f *os.File
    b *bufio.Writer
}
func (w *buf_file_writer) Write(block []byte) (int, error) {
    return w.b.Write(block)
}
func (w *buf_file_writer) Close() error {
    if err := w.b.Flush(); err != nil {
        return err
    }
    return w.f.Close()
}
func new_buf_file_writer(filename string) (*buf_file_writer, error) {
    r       := new(buf_file_writer)
    var err error
    r.f, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
    if err != nil {
        return nil, err
    }
    r.b = bufio.NewWriter(r.f)
    return r, nil
}

func new_buf_fd_writer(f *os.File) *buf_file_writer {
    r   := new(buf_file_writer)
    r.f  = f
    r.b  = bufio.NewWriter(r.f)
    return r
}


type multiWriteCloser struct {
    writers []io.WriteCloser
}

func newMultiWriteCloser(writers ...io.WriteCloser) io.WriteCloser {
    return &multiWriteCloser{writers}
}

func (w *multiWriteCloser) Write(block []byte) (int, error) {
    for _, x := range w.writers {
        n, err := x.Write(block)
        if err != nil {
            return n, err
        }
    }
    return len(block), nil
}
func (w *multiWriteCloser) Close() error {
    for _, x := range w.writers {
        err := x.Close()
        if err != nil {
            return err
        }
    }
    return nil
}
