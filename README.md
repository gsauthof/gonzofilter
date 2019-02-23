This repository contains gonzofilter, a bayes classifying spam
mail filter written in Go.

2019, Georg Sauthoff <mail@gms.tf>, GPLv3+

## Getting started

Build a new database with some already classified messages
(either manually classified or classified with another classifier):

    $ ./toe.py

See the short `toe.py` script for details.

To classify new messages:

    $ ./gonzofilter -check -in path/to/maildir/msg
    $ echo $?

The exit status 10 stands for a 'ham' classification result while
11 stands for 'spam'.

For integration with a mail-delivery-agent it also supports a
pass-through mode (cf. the `-pass option`).

To just use it as tokenizer:

    $ ./gonzofilter -dump-mark -in path/to/maildir/msg

See also `-h` for additional commands and options.

## Build Instructions

Compile it:

    $ GOPATH=$HOME/go:/usr/share/gocode go build

Run the unittests:

    $ GOPATH=$HOME/go:/usr/share/gocode go test -v

Set the GOPATH differently if the dependencies are installed
elsewhere or you want to use another workspace location.

It only needs a few extra dependencies:

    - github.com/coreos/bbolt
    - golang.org/x/text/encoding
    - golang.org/x/sys/unix

They can be installed with `go get` or the distribution's package
manager. For example, on Fedora:

    # dnf install golang-github-coreos-bbolt-devel \
                  golang-golangorg-text-devel \
                  golang-github-golang-sys-devel

## Motivation

- Have an accessible platform to test different text
  classification approaches
- Evaluate the trade-offs when writing something exposed as a
  mail filter in a memory-safe language
- Learn a new programming language (Go) - which has some
  interesting features, arguably is better designed than Java,
  but also has some shortcomings


