// 2019, Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

type channel_writer struct {
    out chan<- []byte
}
func new_channel_writer(out chan<- []byte) *channel_writer {
    return &channel_writer{out}
}

// implements io.Writer interface
// expects that caller doesn't modify the array the slice p points to
func (w *channel_writer) Write(p []byte) (n int, err error) {
    w.out <- p
    return len(p), nil
}
func (w *channel_writer) Close() error {
    close(w.out)
    return nil
}


func tee_words(in <-chan []byte, out1, out2 chan<- []byte) {
    for x := range in {
        o1, o2 := out1, out2
        for i := 0; i < 2; i++ {
            // sending something into a nil channel always blocks
            select {
                case o1 <- x: o1 = nil
                case o2 <- x: o2 = nil
            }
        }
    }
    close(out1)
    close(out2)
}
