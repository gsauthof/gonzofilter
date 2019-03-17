#!/usr/bin/bash

set -eu

action=$1
filename=$3

if [ "$action" = -check ]; then
    set +e
    spamprobe score "$filename" | grep '^SPAM'
    r=$?
    set -e
    if [ "$r" -eq 0 ]; then
        ## SPAM
        exit 11
    else
        ## HAM
        exit 10
    fi
fi

if [ "$action" = -spam ]; then
    exec spamprobe -c spam "$filename"
fi

if [ "$action" = -ham ]; then
    exec spamprobe -c good "$filename"
fi
