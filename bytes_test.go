// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "testing"

func Test_imatch_prefix(t *testing.T) {
   if imatch_prefix([]byte(""), []byte("")) != 0 {
       t.Errorf("empty of empty should be 0")
   }
   if imatch_prefix([]byte("hello"), []byte("hello")) != 5 {
       t.Errorf("all prefix should be all")
   }
   if imatch_prefix([]byte("HeLlo"), []byte("hello world")) != 5 {
       t.Errorf("should match case insensitively")
   }
}

func Test_match_prefix(t *testing.T) {
   if match_prefix([]byte(""), []byte("")) != 0 {
       t.Errorf("empty of empty should be 0")
   }
   if match_prefix([]byte("hello"), []byte("hello")) != 5 {
       t.Errorf("all prefix should be all")
   }
   if match_prefix([]byte("heLlo"), []byte("hello world")) != 2 {
       t.Errorf("should match case sensitively")
   }
}
