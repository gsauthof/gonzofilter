// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "testing"
)


func Test_xgonzo_filter(t *testing.T) {
    var b bytes.Buffer
    w := new_xgonzo_filter_writer(new_close_writer(&b))
    w.Write([]byte("X-gonzo: spam"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("X-old-gonzo: spam")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}

func Test_xgonzo_filter_partial(t *testing.T) {
    var b bytes.Buffer
    w := new_xgonzo_filter_writer(new_close_writer(&b))
    w.Write([]byte("Subject: blah\nX-gonz"))
    w.Write([]byte("o: spam"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("Subject: blah\nX-old-gonzo: spam")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}

func Test_xgonzo_filter_partial2(t *testing.T) {
    var b bytes.Buffer
    w := new_xgonzo_filter_writer(new_close_writer(&b))
    w.Write([]byte("Subject: blah\nX-gonz"))
    w.Write([]byte(": spam"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("Subject: blah\nX-gonz: spam")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}

func Test_xgonzo_filter_mult(t *testing.T) {
    var b bytes.Buffer
    w := new_xgonzo_filter_writer(new_close_writer(&b))
    w.Write([]byte("X-gonzo: ham\nSubject: blah\nX-gonzo"))
    w.Write([]byte(": spam"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("X-old-gonzo: ham\nSubject: blah\nX-old-gonzo: spam")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}

func Test_xgonzo_filter_insensitive(t *testing.T) {
    var b bytes.Buffer
    w := new_xgonzo_filter_writer(new_close_writer(&b))
    w.Write([]byte("x-gonzo: ham\nSubject: blah\nX-gonzo: spam"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("X-old-gonzo: ham\nSubject: blah\nX-old-gonzo: spam")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}
