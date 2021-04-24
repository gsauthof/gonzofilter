// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "bytes"
    "io"
)


func write_messageP(in io.Reader, h io.WriteCloser, b io.WriteCloser,
        m io.WriteCloser, pedantic bool) error {
    sw := new_split_message_writer(h, b, m)

    // use small read size for testing
    block := make([]byte, read_size)
    for {
        block  := block[:cap(block)]
        n, err := in.Read(block)
        if err != nil && err != io.EOF {
            return err
        }
        block = block[:n]
        if n == 0 {
            break
        }
        if _, err := sw.Write(block); err != nil {
            debugf("write_message split_message_writer write failed: %v", err)
            if pedantic {
                return err
            } else {
                // when classifying messages it makes sense to ignore write errors
                // since we want to work with what we already have
                // (think: a decoding error in some mime-attachment was preceded
                //         by many spammy words in the header/other parts)
                return nil
            }
        }
    }
    if err := sw.Close(); err != nil {
        return err
    }
    return nil
}

func write_message(in io.Reader, h io.WriteCloser, b io.WriteCloser,
        m io.WriteCloser) error {
    return write_messageP(in, h, b, m, false)
}

func new_content_writer(typ []byte, boundary []byte, houtP io.WriteCloser, boutP io.WriteCloser, moutP io.WriteCloser) io.WriteCloser {
    if bytes.HasPrefix(typ, []byte("multipart")) {
        w := new_multipart_split_writer(boundary, func() io.WriteCloser {
            return new_split_message_writer(moutP, boutP, moutP)
        })
        return w
    } else if bytes.Equal(typ, []byte("text/plain")) {
        return boutP
    } else if bytes.Equal(typ, []byte("text/html")) {
        return new_remove_tags_writer(new_replace_entities_writer(
                    new_shrink_space_writer(boutP)))
        //return new_remove_tags_writer((boutP))
    }
    return new_dev_null_writer()
}

type split_message_writer struct {
    houtP io.WriteCloser
    boutP io.WriteCloser
    moutP io.WriteCloser

    hout io.WriteCloser
    bout io.WriteCloser

    eh *extract_header_writer
    m split_machine
}
func new_split_message_writer(houtP io.WriteCloser, boutP io.WriteCloser,
                              moutP io.WriteCloser) *split_message_writer {
    w       := new(split_message_writer)
    w.houtP  = new_keep_open_writer(houtP)
    w.boutP  = new_keep_open_writer(boutP)
    w.moutP  = new_keep_open_writer(moutP)

    w.eh     = new_extract_header_writer()
    w.hout   = new_unfold_writer(
                 newMultiWriteCloser(w.eh, new_header_decode_writer(w.houtP)))
    return w
}

func (w *split_message_writer) Write(block []byte) (int, error) {
    n := len(block)
    for len(block) != 0 {
        action, bs, bl := w.m.next(block)
        block = bl
        switch action {
        case split_machine_WRITE_HEADER:
            if _, err := w.hout.Write(bs); err != nil {
                return 0, err
            }
        case split_machine_SETUP_BODY:
            if err := w.hout.Close(); err != nil {
                return 0, err
            }
            w.eh.parse()
            debugf("parsed: content type |%s| type |%s| charset |%s|" +
                   " encoding |%s| boundary |%s|\n",
                   w.eh.content_type, w.eh.typ, w.eh.charset,
                   w.eh.transfer_encoding, w.eh.boundary)
            w.bout = new_transfer_decode_writer(w.eh.transfer_encoding,
                          new_charset_writer(w.eh.charset,
                               new_content_writer(w.eh.typ, w.eh.boundary,
                                                  w.houtP, w.boutP, w.moutP)))
        case split_machine_WRITE_BODY:
            if _, err := w.bout.Write(bs); err != nil {
                return 0, err
            }
        case split_machine_MORE:
            ;
        }
    }
    return n, nil
}

func (w *split_message_writer) Close() error {
    if w.bout != nil {
        return w.bout.Close()
    }
    return nil
}

const (
        split_machine_WRITE_HEADER = iota
        split_machine_SETUP_BODY
        split_machine_WRITE_BODY
        split_machine_MORE
      )

type split_machine struct {
    state int
}
func (w *split_machine) next(block []byte) (int, []byte, []byte) {
    const ( IN_HEADER = iota
            IN_DELIM
            IN_BODY_SETUP
            IN_BODY
          )
    switch w.state {
    case IN_HEADER:
        i := bytes.Index(block, []byte("\n\n"))
        if i == -1 {
            if block[len(block)-1] == byte('\n') {
                w.state = IN_DELIM
            }
            r     := block
            block  = block[:0]
            return split_machine_WRITE_HEADER, r, block
        } else {
            r       := block[:i+1]
            block    = block[i+2:]
            w.state  = IN_BODY_SETUP
            return split_machine_WRITE_HEADER, r, block
        }
    case IN_DELIM:
        if block[0] == byte('\n') {
            block   = block[1:]
            w.state = IN_BODY_SETUP
        } else {
            w.state = IN_HEADER
        }
        return split_machine_MORE, nil, block
    case IN_BODY_SETUP:
        w.state = IN_BODY
        return split_machine_SETUP_BODY, nil, block
    case IN_BODY:
        r      := block
        block   = block[:0]
        return split_machine_WRITE_BODY, r, block
    }
    return split_machine_MORE, nil, block
}

