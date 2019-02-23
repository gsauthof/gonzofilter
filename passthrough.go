// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "golang.org/x/sys/unix"
    "io"
    "log"
    "os"
)

func copy_stdin(tmpdir string) (*os.File, error) {
    f, err := os.OpenFile(tmpdir, os.O_RDWR | os.O_EXCL | unix.O_TMPFILE, 0600)
    if err != nil {
        return f, err
    }
    block := make([]byte, read_size)
    for {
        block := block[:cap(block)]
        n, err := os.Stdin.Read(block)
        if err != nil && err != io.EOF {
            return f, err
        }
        block = block[:n]
        if n == 0 {
            break
        }
        if _, err := f.Write(block); err != nil {
            return f, err
        }
    }
    if _, err := f.Seek(0, 0); err != nil {
        return f, err
    }
    return f, nil
}

func passthrough_file(in io.Reader, is_ham bool, out io.WriteCloser) error {
    var m split_machine
    block := make([]byte, read_size)
    hout := new_xgonzo_filter_writer(new_keep_open_writer(out))
    for {
        block := block[:cap(block)]
        n, err := in.Read(block)
        if err != nil && err != io.EOF {
            return err
        }
        block = block[:n]
        if n == 0 {
            break
        }

        for {
            if len(block) == 0 {
                break
            }
            action, bs, bl := m.next(block)
            block = bl
            switch action {
            case split_machine_WRITE_HEADER:
                if _, err := hout.Write(bs); err != nil {
                    return err
                }
            case split_machine_WRITE_BODY:
                if _, err := out.Write(bs); err != nil {
                    return err
                }
            case split_machine_SETUP_BODY:
                if err := hout.Close(); err != nil {
                    return err
                }
                if is_ham {
                    out.Write([]byte("X-gonzo: ham\n\n"))
                } else {
                    out.Write([]byte("X-gonzo: spam\n\n"))
                }
            case split_machine_MORE:
                ;
            }
        }

    }
    return nil
}

func passthrough_message(args *args) {
    var f *os.File
    var err error
    if args.in_filename == "-" {
        f, err = copy_stdin(args.tmpdir)
        if err != nil {
            log.Fatal(err)
        }
    } else {
        f, err = os.OpenFile(args.in_filename, os.O_RDONLY, 0644)
        if err != nil {
            log.Fatal(err)
        }
    }

    var is_ham bool
    is_ham, err = classify_file(f, args)
    if err != nil {
        log.Fatal(err)
    }
    if _, err := f.Seek(0, 0); err != nil {
        log.Fatal(err)
    }

    out := open_output_or_stdout(args.out_filename)

    err = passthrough_file(f, is_ham, out)
    if err != nil {
        log.Fatal(err)
    }
    if err := out.Close(); err != nil {
        log.Fatal(err)
    }
}

