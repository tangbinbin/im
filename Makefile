all: install

GOPATH:=$(CURDIR)
export GOPATH

dep:
	go get github.com/garyburd/redigo/redis
	go get github.com/golang/protobuf/proto
	go get github.com/go-sql-driver/mysql

install:dep
	go install im 
	go install api
