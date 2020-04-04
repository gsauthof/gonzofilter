// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
    "encoding/binary"
    "errors"
    "fmt"
    bolt "go.etcd.io/bbolt"
    "log"
)

func dump_db_class(db *bolt.DB, class_name []byte) error {
    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(class_name)
        if b == nil  {
            fmt.Printf("%s --- is empty!\n", class_name)
            return nil
        }
        stat := b.Bucket([]byte("stat"))
        if stat == nil {
            return errors.New("stat bucket is missing")
        }
        fmt.Printf("%s --- word bag size: %v, vocabulary size: %v, messages: %v\n", class_name, get_uint32(stat, []byte("bag_size")), get_uint32(stat, []byte("vocabulary")), get_uint32(stat, []byte("messages")))
        word_bag := b.Bucket([]byte("word_bag"))
        if word_bag == nil {
            return errors.New("word_bag bucket is missing")
        }
        c := word_bag.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            fmt.Printf("%s : %v\n", k, binary.LittleEndian.Uint32(v))
	}
        return nil
    })
    return err
}

func dump_db(args *args) {
    db, err := bolt.Open(args.db_filename, 0666, nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, class_name := range [][]byte{[]byte("ham"), []byte("spam")} {
        if err := dump_db_class(db, class_name); err != nil {
            log.Fatal(err)
        }
    }
    if err := db.Close(); err != nil {
        log.Fatal(err)
    }
}

