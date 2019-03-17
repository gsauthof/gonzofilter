#!/usr/bin/bash

set -eu

action=$1
filename=$3

if [ "$action" = -check ]; then
    set +e
    qsf --test < "$filename"
    r=$?
    set -e
    if [ "$r" -eq 0 ]; then
        ## HAM
        exit 10
    else
        ## SPAM
        exit 11
    fi
fi

if [ $(stat -c '%s' "$filename") -gt 524288 ]; then
    exit
fi

if [ "$action" = -spam ]; then
    exec qsf -m < "$filename"
fi

if [ "$action" = -ham ]; then
    exec qsf -M < "$filename"
fi
