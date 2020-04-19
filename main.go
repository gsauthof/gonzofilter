// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "flag"
    "fmt"
    "log"
    "os"
)

type args struct {
    read_size        int
    in_filename      string
    out_filename     string
    header_filename  string
    body_filename    string
    mime_filename    string
    db_filename      string
    dump_text        bool
    dump_words       bool
    dump_mark        bool
    dump_db          bool
    ham              bool
    spam             bool
    check            bool
    passthrough      bool
    tmpdir           string
    sandbox          bool
    sandbox_debug    bool
    h                bool
    help             bool
    verbose          bool
}

func parse_args() *args {
    args := new(args)
    flag.IntVar(&args.read_size, "rsize", 128*1024,
            "read size - use small value for testing")
    flag.StringVar(&args.in_filename, "in", "-", "input message filename")
    flag.StringVar(&args.out_filename, "out", "-",
            "output filename when passing through a message")
    flag.StringVar(&args.header_filename, "header", "h",
            "header debug output filename")
    flag.StringVar(&args.body_filename, "body", "b",
            "body debug output filename")
    flag.StringVar(&args.mime_filename, "mime", "m",
            "mime debug output filename")
    flag.StringVar(&args.db_filename, "db", "hamspam.db", "word database")
    flag.BoolVar(&args.dump_text, "dump-text", false,
            "dump the parsed text for debugging purposes")
    flag.BoolVar(&args.dump_words, "dump-words", false,
            "dump the parsed text after after word splitting for debugging" +
            " purposes")
    flag.BoolVar(&args.dump_mark, "dump-mark", false,
            "dump the parsed text after after word splitting and marking for" +
            " debugging purposes")
    flag.BoolVar(&args.dump_db, "dump-db", false, "dump the database")
    flag.BoolVar(&args.check, "check", false, "classify a message")
    flag.BoolVar(&args.passthrough, "pass", false, "classify a message and" +
            " pass it through with the result stored in a X-gonzo header")
    flag.StringVar(&args.tmpdir, "tmpdir", "/tmp", "directory to store" +
            " message when passing it through from stdin")
    flag.BoolVar(&args.ham, "ham", false, "classify message as spam")
    flag.BoolVar(&args.spam, "spam", false, "classify message as spam")
    flag.BoolVar(&args.sandbox, "sandbox", false, "Sandbox this program" +
            " where supported (e.g. on Linux with seccomp). Only useful in" +
            " combination with -check or -passthrough.")
    flag.BoolVar(&args.sandbox_debug, "sb-debug", false, "if -sandbox is" +
            " given just log violations (to syslog) for testing purposes")
    flag.BoolVar(&args.h, "h", false, "show this help screen")
    flag.BoolVar(&args.help, "help", false, "show this help screen")
    flag.BoolVar(&args.verbose, "v", false, "print some debug messages")
    flag.Parse()
    if args.h || args.help {
        fmt.Printf("Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        os.Exit(0)
    }
    if args.verbose {
        verbosity = 1
    }
    return args
}

func main() {
    args := parse_args()

    if args.sandbox {
        sandbox_me(args.sandbox_debug)
    }

    read_size = args.read_size

    switch {
        case args.dump_text        : dump_text(args)
        case args.dump_words       : dump_words(args)
        case args.dump_mark        : dump_mark(args)
        case args.ham || args.spam : learn_message(args)
        case args.dump_db          : dump_db(args)
        case args.check            : classify_message(args)
        case args.passthrough      : passthrough_message(args)
        default                    : log.Fatal("No action selected")
    }


    os.Exit(0)
}
