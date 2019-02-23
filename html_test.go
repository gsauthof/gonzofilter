// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "testing"
)

func Test_remove_tags(t *testing.T) {
    var b bytes.Buffer
    w := new_remove_tags_writer(new_close_writer(&b))
    w.Write([]byte("<!DOCTYPE html>\n<!--a0a0a0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0_a01-->\n<html>  Hello <head> World </head>"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte(" \n\n   Hello   World  ")) {

        t.Errorf("Unexpected result: |%v|", b.Bytes())
    }
}

func Test_remove_tags_comment_with_gt(t *testing.T) {
    var b bytes.Buffer
    w := new_remove_tags_writer(new_close_writer(&b))
    w.Write([]byte("<!DOCTYPE html>\n<!--       bla bla bal              >   lol -->\n<html>  Hello <head> World </head>"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte(" \n\n   Hello   World  ")) {

        t.Errorf("Unexpected result: |%v|", b.Bytes())
    }
}
