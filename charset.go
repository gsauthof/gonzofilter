// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "golang.org/x/text/encoding"
    "golang.org/x/text/encoding/charmap"
    "golang.org/x/text/encoding/japanese"
    "golang.org/x/text/encoding/simplifiedchinese"
    "golang.org/x/text/transform"
    "io"
)


type charset_writer struct {
    out          io.WriteCloser
    transformer  *transform.Writer
}

// implements io.Writer interface
func (w *charset_writer) Write(p []byte) (n int, err error) {
    if w.transformer == nil {
        if _, err := w.out.Write(p); err != nil {
            return 0, err
        }
        return len(p), nil
    } else {
        return w.transformer.Write(p)
    }
}
func (w *charset_writer) Close() error {
    if w.transformer != nil {
        return w.transformer.Close()
    } else {
        return w.out.Close()
    }
}

// the lookup function lower-cases the charset input string
var charset_trans_map = map[string](func()*encoding.Decoder){
    "latin1"       : charmap.ISO8859_1.NewDecoder     ,
    "latin2"       : charmap.ISO8859_2.NewDecoder     ,
    "latin9"       : charmap.ISO8859_15.NewDecoder    ,
    "iso-8859-1"   : charmap.ISO8859_1.NewDecoder     ,
    "iso-8859-2"   : charmap.ISO8859_2.NewDecoder     ,
    "iso-8859-15"  : charmap.ISO8859_15.NewDecoder    ,
    "iso8859-15"   : charmap.ISO8859_15.NewDecoder    ,
    "windows-1250" : charmap.Windows1250.NewDecoder   ,
    "windows-1251" : charmap.Windows1251.NewDecoder   ,
    "windows-1252" : charmap.Windows1252.NewDecoder   ,
    "windows-1255" : charmap.Windows1255.NewDecoder   ,
    "windows-1256" : charmap.Windows1256.NewDecoder   ,
    "windows-1258" : charmap.Windows1258.NewDecoder   ,
    "cp1252"       : charmap.Windows1252.NewDecoder   ,
    "ibm852"       : charmap.CodePage852.NewDecoder   ,
    "koi8-r"       : charmap.KOI8R.NewDecoder         ,
    "iso-2022-jp"  : japanese.ISO2022JP.NewDecoder    ,
    "gbk"          : simplifiedchinese.GBK.NewDecoder ,
}


func new_charset_writer(charset []byte, out io.WriteCloser) *charset_writer {
    w     := new(charset_writer)
    w.out  = out
    trans := charset_trans_map[string(bytes.ToLower(charset))]
    if trans != nil {
        //trans.Reset()
        w.transformer = transform.NewWriter(out, trans())
    }

    return w
}
