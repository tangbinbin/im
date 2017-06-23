all: install

GOPATH:=$(CURDIR)
export GOPATH

dep:
	go get github.com/tangbinbin/tlog
	go get github.com/BurntSushi/toml

install:dep
	go install im 
