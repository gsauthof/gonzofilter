

.PHONY: all
all:
	GOPATH=$$HOME/go:/usr/share/gocode go build


.PHONY: check
check:
	GOPATH=$$HOME/go:/usr/share/gocode go test -v

