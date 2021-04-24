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
)
