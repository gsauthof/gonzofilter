// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "encoding/binary"
    "errors"
    "fmt"
    "github.com/coreos/bbolt"
    "io"
    "log"
    "math"
    "os"
)

func get_uint32(b *bolt.Bucket, key []byte) uint32 {
    v := b.Get(key)
    if v == nil {
        return 0
    }
    return binary.LittleEndian.Uint32(v)
}

type class_result struct {
    prob float64
    err error
}

func classify_words(db *bolt.DB, class_name []byte, ch chan []byte, vocabulary uint32, done chan class_result) {
    var prob float64
    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(class_name)
        if b == nil {
            return errors.New("main bucket not found")
        }
        word_bag := b.Bucket([]byte("word_bag"))
        if word_bag == nil {
            return errors.New("word_bag not found")
        }
        stat := b.Bucket([]byte("stat"))
        if stat == nil {
            return errors.New("stat bag not found")
        }
        bag_size := get_uint32(stat, []byte("bag_size"))
        for word := range ch {
            f := get_uint32(word_bag, word)
            // using pseudo-counts here
            a := float64(f + 1) / float64(bag_size + vocabulary)
            // we do multiplications in log-space
            prob += math.Log(a)
        }
        return nil
    })
    done <- class_result{prob, err}
}

func get_vocabulary(db *bolt.DB) (uint32, error) {
    var vocabulary uint32
    err := db.View(func(tx *bolt.Tx) error {
        for _, class_name := range [][]byte{[]byte("ham"), []byte("spam")} {
            b := tx.Bucket(class_name)
            if b == nil {
                return errors.New("main bucket not found")
            }
            stat := b.Bucket([]byte("stat"))
            if stat == nil {
                return errors.New("stat bag not found")
            }
            v := get_uint32(stat, []byte("vocabulary"))
            if v > vocabulary {
                v = vocabulary
            }
        }
        return nil
    })
    return vocabulary, err
}

func classify_file(in io.Reader, args *args) (bool, error) {

    db, err := bolt.Open(args.db_filename, 0666, nil)
    if err != nil {
        return true, err
    }
    ch := make(chan []byte, 100)
    cw := new_channel_writer(ch)
    cwo := new_keep_open_writer(cw)

    ch1 := make(chan []byte, 100)
    ch2 := make(chan []byte, 100)
    go tee_words(ch, ch1, ch2)

    h := new_word_split_writer(new_header_filter_writer(
            new_replace_chars_writer(
            new_mark_copy_header_writer(byte('h'), cwo))))
    b := new_word_split_writer(
            new_replace_chars_writer(
            new_mark_copy_body_writer(byte('b'), cwo)))
    m := new_word_split_writer(new_header_filter_writer(
            new_replace_chars_writer(
            new_mark_copy_header_writer(byte('m'), cwo))))

    ham_done := make(chan class_result)
    spam_done := make(chan class_result)
    var vocabulary uint32

    vocabulary, err = get_vocabulary(db)
    if err != nil {
        return true, err
    }

    go classify_words(db, []byte("ham"), ch1,  vocabulary, ham_done)
    go classify_words(db, []byte("spam"), ch2,  vocabulary, spam_done)

    if err := write_message(in, h, b, m); err != nil {
        return true, err
    }

    if err := h.Close(); err != nil {
        return true, err
    }
    if err := b.Close(); err != nil {
        return true, err
    }
    if err := m.Close(); err != nil {
        return true, err
    }
    if err := cw.Close(); err != nil {
        return true, err
    }

    ham := <- ham_done
    if ham.err != nil {
        return true, err
    }
    spam := <- spam_done
    if spam.err != nil {
        return true, err
    }
    score := ham.prob/(ham.prob+spam.prob)
    debugf("Ham %v vs. Spam %v score %v\n", ham.prob, spam.prob, score)

    is_ham := score < 0.5 + 0.01

    if err := db.Close(); err != nil {
        return true, err
    }

    return is_ham, nil
}

func classify_message(args *args) {
    in := open_input(args.in_filename)
    is_ham, err := classify_file(in, args)
    if err != nil {
        log.Fatal(err)
    }
    if is_ham {
        fmt.Printf("HAM\n")
        os.Exit(10)
    } else {
        fmt.Printf("SPAM\n")
        os.Exit(11)
    }
}

