all: install

GOPATH:=$(CURDIR)
export GOPATH

dep:
	go get github.com/tangbinbin/tlog

install:dep
	go install im 
