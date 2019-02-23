// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "log"
)

func debugf(format string, v ...interface{}) {
    if verbosity > 0 {
        log.Printf(format, v...)
    }
}

func dump_text(args *args) {
    in := open_input(args.in_filename)
    h  := open_output(args.header_filename)
    b  := open_output(args.body_filename)
    m  := open_output(args.mime_filename)

    if err := write_message(in, h, b, m); err != nil {
        log.Fatal(err)
    }

    if err := h.Close(); err != nil {
        log.Fatal(err)
    }
    if err := b.Close(); err != nil {
        log.Fatal(err)
    }
    if err := m.Close(); err != nil {
        log.Fatal(err)
    }
}

func dump_words(args *args) {
    in := open_input(args.in_filename)
    h  := open_output(args.header_filename)
    b  := open_output(args.body_filename)
    m  := open_output(args.mime_filename)

    //h = new_word_split_writer(new_word_writer(h))
    h = new_word_split_writer(new_header_filter_writer(
            new_word_writer(h)))
    b = new_word_split_writer(new_word_writer(b))
    m = new_word_split_writer(new_word_writer(m))

    if err := write_message(in, h, b, m); err != nil {
        log.Fatal(err)
    }

    if err := h.Close(); err != nil {
        log.Fatal(err)
    }
    if err := b.Close(); err != nil {
        log.Fatal(err)
    }
    if err := m.Close(); err != nil {
        log.Fatal(err)
    }
}

func dump_mark(args *args) {
    in := open_input(args.in_filename)

    o := new_nl_writer(open_output_or_stdout(args.out_filename))
    out := new_keep_open_writer(o)

    h := new_word_split_writer(new_header_filter_writer(
            new_mark_copy_header_writer(byte('h'), out)))
    b := new_word_split_writer(new_mark_copy_body_writer(byte('b'), out))
    m := new_word_split_writer(new_header_filter_writer(
            new_mark_copy_header_writer(byte('m'), out)))

    if err := write_message(in, h, b, m); err != nil {
        log.Fatal(err)
    }

    if err := h.Close(); err != nil {
        log.Fatal(err)
    }
    if err := m.Close(); err != nil {
        log.Fatal(err)
    }
    if err := b.Close(); err != nil {
        log.Fatal(err)
    }
    if err := o.Close(); err != nil {
        log.Fatal(err)
    }
}



