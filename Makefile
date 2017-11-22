all: install

GOPATH:=$(CURDIR)
export GOPATH

dep:
	go get github.com/tangbinbin/tlog
	go get github.com/BurntSushi/toml
	go get github.com/golang/protobuf/proto
	protoc --go_out=./src/proto im.proto

install:dep
	go install im 

clean:
	-rm -fr src/github.com
	-rm -fr bin/*
	-rm -fr pkg
