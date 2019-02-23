// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "testing"
)


func Test_unfold(t *testing.T) {
    var b bytes.Buffer
    w := new_unfold_writer(new_close_writer(&b))
    w.Write([]byte("Subject: Hel\n lo World\nFrom: Nobody\n"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("Subject: Hello World\nFrom: Nobody\n")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}

func Test_unfold_tab(t *testing.T) {
    var b bytes.Buffer
    w := new_unfold_writer(new_close_writer(&b))
    w.Write([]byte("Subject: Hel\n\tlo World\nFrom: Nobody\n"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("Subject: Hello World\nFrom: Nobody\n")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}

func Test_unfold_mult_space(t *testing.T) {
    var b bytes.Buffer
    w := new_unfold_writer(new_close_writer(&b))
    w.Write([]byte("Subject: Hello\n  World\nFrom: Nobody\n"))
    w.Close()
    if !bytes.Equal(b.Bytes(), []byte("Subject: Hello World\nFrom: Nobody\n")) {
        t.Errorf("Unexpected result: |%s|", b.Bytes())
    }
}
