// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
)

func min(a, b int) int {
    if a < b {
        return a
    } else {
        return b
    }
}

func imatch_prefix(text, pattern []byte) int {
    l := min(len(text), len(pattern))
    r := 0
    for i := 0; i < l;  i++ {
        if text[i] >= byte(' ') {
            // lower-case the text, assumning that pattern already is lowercase
            if (text[i] | 0x20) != pattern[i] { // i.e. or with 0b100000
                break
            }
        } else {
            // lower-case newlines etc.
            if text[i] != pattern[i] {
                break
            }
        }
        r++;
    }
    return r
}

func match_prefix(text, pattern []byte) int {
    l := min(len(text), len(pattern))
    r := 0
    for i := 0; i < l;  i++ {
        if text[i] != pattern[i] {
            break
        }
        r++;
    }
    return r
}

func is_space(c byte) bool {
    return c == byte(' ') || c == byte('\n') || c == byte('\t') ||
           c == byte('\r')
}

func index_any(t []byte, qs []byte) int {
    i := -1
    for _, q := range qs {
        if i == -1 {
            i = bytes.IndexByte(t, q)
        } else {
            j := bytes.IndexByte(t[:i], q)
            if j != -1 && j < i {
                i = j
            }
        }
    }
    return i
}


func iequal(word []byte, pattern []byte) bool {
    n := len(word)
    if n !=  len(pattern) {
        return false
    }
    for i := 0; i < n;  i++ {
        // lower-case the text, assumning that pattern already is lowercase
        if (word[i] | 0x20) != pattern[i] { // i.e. or with 0b100000
            return false
        }
    }
    return true
}

func iequal_any(word []byte, ps [][]byte) bool {
    for _, p := range ps {
        if iequal(word, p) {
            return true
        }
    }
    return false
}

func ihas_prefix(word []byte, pattern []byte) bool {
    n := len(pattern)
    if n > len(word) {
        return false
    }
    return iequal(word[:n], pattern)
}

func ihas_suffix(word []byte, pattern []byte) bool {
    n := len(pattern)
    m := len(word)
    if n > m {
        return false
    }
    return iequal(word[m-n:], pattern)
}


