// SPDX-FileCopyrightText: © 2020 Georg Sauthoff <mail@gms.tf>
// SPDX-License-Identifier: GPL-3.0-or-later

package main

// This sandboxing is meant as a defence-in-depth measure.
// Meaning that since Go is a memory safe language an attacker
// already should have a hard time to get to the point making
// this program execute unexpected syscalls.
// However, if it happens the attacker only has a limited set of
// syscalls available (see below).

import (
    seccomp "github.com/seccomp/libseccomp-golang"
)


func whitelist_syscalls(syscalls []string, debug bool) {
    action := seccomp.ActKillProcess
    if debug {
        action = seccomp.ActLog
    }
    f, err := seccomp.NewFilter(action)
    if err != nil {
        panic(err)
    }
    for _, s := range syscalls {
        id, err := seccomp.GetSyscallFromName(s)
        if err != nil {
            panic(err)
        }
        if err := f.AddRule(id, seccomp.ActAllow); err != nil {
            panic(err)
        }
    }
    if err := f.Load(); err != nil {
        panic(err)
    }
}

func blacklist_syscalls(syscalls []string, debug bool) {
    f, err := seccomp.NewFilter(seccomp.ActAllow)
    if err != nil {
        panic(err)
    }
    for _, s := range syscalls {
        id, err := seccomp.GetSyscallFromName(s)
        if err != nil {
            panic(err)
        }
        action := seccomp.ActKillProcess
        if debug {
            action = seccomp.ActLog
        }
        if err := f.AddRule(id, action); err != nil {
            panic(err)
        }
    }
    if err := f.Load(); err != nil {
        panic(err)
    }
}


func sandbox_me(debug bool) {
    var syscalls = []string {
        "arch_prctl"        ,
        "clone"             ,
        "close"             ,
        "epoll_create1"     ,
        "epoll_ctl"         ,
        "epoll_pwait"       ,
        "exit_group"        ,
        "fcntl"             ,
        "flock"             ,
        "fstat"             ,
        "futex"             ,
        "gettid"            ,
        "lseek"             ,
        "madvise"           ,
        "mmap"              ,
        "mprotect"          ,
        "munmap"            ,
        "nanosleep"         ,
        "openat"            ,
        "pread64"           ,
        "prctl"             ,
        "read"              ,
        "readlinkat"        ,
        "rt_sigaction"      ,
        "rt_sigprocmask"    ,
        "sched_getaffinity" ,
        "sched_yield"       ,
        "seccomp"           , // such that we can downsize this list later
        "set_robust_list"   ,
        "sigaltstack"       ,
        "write"             }
    whitelist_syscalls(syscalls, debug)
}

// This function is intended to be called after sandbox_me()
// and after all necessary files are opened (i.e. database, a temporary file
// and some files from pseudo-filesystems) to downsize the whitelist. 
func blacklist_open(debug bool) {
    var syscalls = []string {
        "seccomp"           ,
	"openat"            }
    blacklist_syscalls(syscalls, debug)
}


