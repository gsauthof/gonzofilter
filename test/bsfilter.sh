#!/usr/bin/bash

set -eu

action=$1
filename=$3

if [ "$action" = -check ]; then
    set +e
    bsfilter "$filename"
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
    exec bsfilter -u -s "$filename"
fi

if [ "$action" = -ham ]; then
    exec bsfilter -u -c "$filename"
fi
