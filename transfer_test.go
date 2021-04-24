// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "io"
    "testing"
)

func Test_extra_pipe_input(t *testing.T) {

    var bw bytes.Buffer

    dw := new_transfer_decode_writer([]byte("quoted-printable"), new_close_writer(&bw))


    dw.Write([]byte("Hello"))
    if _, err := dw.Write([]byte("GCg==\n--=_ab8")); err != nil {
	t.Errorf("Unexpected error in second write: %v", err)
    }

    // => yields io.ErrUnexpectedEOF in readHexByte() in quotedprintable/reader.go
    n, err := dw.Write([]byte("World"))

    if err != io.ErrUnexpectedEOF {
	t.Errorf("Expected the io.ErrUnexpectedEOF - got: %v", err)
    }
    if n != 0 {
	t.Errorf("Expected n=0 - got: %d", n)
    }

    cerr := dw.Close()

    if cerr != nil {
	t.Errorf("Unexpected close error: %v", err)
    }

    if !bytes.Equal(bw.Bytes(), []byte("HelloGCg")) {
	t.Errorf("unexpected decode: %s", bw.Bytes());
    }


}
