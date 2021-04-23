// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "testing"
)

func Test_return_harmless_transfer_encoding(t *testing.T) {

    eh := new_extract_header_writer()

    eh.Write([]byte("Date: Mon, 23 Jan 2012 18:48:48 +0100\n" +
        "From: juser@example.org\n" +
        "To: shop@example.com\n" +
        "Subject: Order\n" +
        "Message-ID: <20120123174848.GE3373@example.org>\n" +
        "MIME-Version: 1.0\n" +
        "Content-Type: multipart/mixed; boundary=\"xHFwDpU9dbj6ez1V\"\n" +
        "Content-Disposition: inline\n" +
        "Content-Transfer-Encoding: 8bit\n" +
        "User-Agent: Mutt/1.5.21 (2010-09-15)\n"))

    eh.parse()

    eh.Close()

    if !bytes.Equal(eh.transfer_encoding, []byte("8bit")) {
        t.Errorf("Unexpected encoding: |%s|", eh.transfer_encoding)
    }
}

func Test_ignore_actual_transfer_encoding(t *testing.T) {

    eh := new_extract_header_writer()

    eh.Write([]byte("Date: Mon, 23 Jan 2012 18:48:48 +0100\n" +
        "From: juser@example.org\n" +
        "To: shop@example.com\n" +
        "Subject: Order\n" +
        "Message-ID: <20120123174848.GE3373@example.org>\n" +
        "MIME-Version: 1.0\n" +
        "Content-Type: multipart/related;" + // explicitly unfolded to keep the test simple
        " boundary=\"=_deadcafe0123456789012348942892c1\"\n" +
        "Content-Disposition: inline\n" +
        "Content-Transfer-Encoding: quoted-printable\n" +
        "User-Agent: Mutt/1.5.21 (2010-09-15)\n"))

    eh.parse()

    eh.Close()

    if !bytes.Equal(eh.transfer_encoding, []byte("")) {
        t.Errorf("Unexpected encoding: |%s|", eh.transfer_encoding)
    }
}
