// SPDX-FileCopyrightText: © 2020 Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

// yes, the below is a go build magic pragma for conditional builds

// +build !linux

package main

func sandbox_me() {
    // no-op on non-Linux OS
}

func blacklist_opens() {
    // no-op on non-Linux OS
}
