// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "encoding/binary"
    "github.com/coreos/bbolt"
    "log"
)



func increment_key(b *bolt.Bucket, key []byte, i uint32) (bool, error) {
    fresh := false
    n := i
    v := b.Get(key)
    if v == nil {
        fresh = true
    } else {
        n += binary.LittleEndian.Uint32(v)
    }
    t := make([]byte, 4)
    binary.LittleEndian.PutUint32(t, n)
    err := b.Put(key, t)
    if err != nil {
        return fresh, err
    }
    return fresh, nil
}

func store_words(db *bolt.DB, class_name []byte, ch chan []byte, done chan error) {
    err := db.Update(func(tx *bolt.Tx) error {
        b, err := tx.CreateBucketIfNotExists(class_name)
        if err != nil {
            return err
        }
        word_bag, err := b.CreateBucketIfNotExists([]byte("word_bag"))
        if err != nil {
            return err
        }
        stat, err := b.CreateBucketIfNotExists([]byte("stat"))
        if err != nil {
            return err
        }
        var vocabulary uint32
        var i uint32
        for word := range ch {
            i++
            fresh, err := increment_key(word_bag, word, 1)
            if err != nil {
                return err
            }
            if fresh {
                vocabulary++
            }
        }
        _, err = increment_key(stat, []byte("vocabulary"), vocabulary)
        _, err = increment_key(stat, []byte("messages"), 1)
        _, err = increment_key(stat, []byte("bag_size"), i)
        return nil
    })
    done <- err
}
func learn_message(args *args) {
    in := open_input(args.in_filename)

    db, err := bolt.Open(args.db_filename, 0666, nil)
    if err != nil {
        log.Fatal(err)
    }
    ch := make(chan []byte, 100)
    cw := new_channel_writer(ch)
    cwo := new_keep_open_writer(cw)

    h := new_word_split_writer(2, new_header_filter_writer(
            new_replace_chars_writer(
            new_mark_copy_header_writer(byte('h'), cwo))))
    b := new_word_split_writer(-1,
            new_replace_chars_writer(
            new_mark_copy_body_writer(byte('b'), cwo)))
    m := new_word_split_writer(-1, new_header_filter_writer(
            new_replace_chars_writer(
            new_mark_copy_header_writer(byte('m'), cwo))))

    class_name := []byte("ham")
    if args.spam {
        class_name = []byte("spam")
    }

    done := make(chan error)

    go store_words(db, class_name, ch,  done)

    if err := write_message(in, h, b, m); err != nil {
        log.Fatal(err)
    }

    if err := h.Close(); err != nil {
        log.Fatal(err)
    }
    if err := b.Close(); err != nil {
        log.Fatal(err)
    }
    if err := m.Close(); err != nil {
        log.Fatal(err)
    }
    if err := cw.Close(); err != nil {
        log.Fatal(err)
    }

    if err := <- done; err != nil {
        log.Fatal(err)
    }

    if err := db.Close(); err != nil {
        log.Fatal(err)
    }
}

