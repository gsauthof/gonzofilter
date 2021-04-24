module github.com/gsauthof/gonzofilter

go 1.15

require (
	go.etcd.io/bbolt v1.3.5
	golang.org/x/sys v0.0.0-20210423185535-09eb48e85fd7
	golang.org/x/text v0.3.6
)

// when compiling under Fedora and
// golang-etcd-bbolt-devel golang-x-sys-devel golang-x-text-devel
// are installed:
replace (
	go.etcd.io/bbolt => /usr/share/gocode/src/go.etcd.io/bbolt
	golang.org/x/sys => /usr/share/gocode/src/golang.org/x/sys
	golang.org/x/text => /usr/share/gocode/src/golang.org/x/text
// indirect dependencies:
	golang.org/x/tools => /usr/share/gocode/src/golang.org/x/tools
	github.com/yuin/goldmark => /usr/share/gocode/src/github.com/yuin/goldmark
	golang.org/x/mod => /usr/share/gocode/src/golang.org/x/mod
	golang.org/x/net => /usr/share/gocode/src/golang.org/x/net
	golang.org/x/sync => /usr/share/gocode/src/golang.org/x/sync
	golang.org/x/xerrors => /usr/share/gocode/src/golang.org/x/xerrors
	golang.org/x/crypto => /usr/share/gocode/src/golang.org/x/crypto
	golang.org/x/term => /usr/share/gocode/src/golang.org/x/term
//replace-this
)
