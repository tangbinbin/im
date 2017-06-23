package main

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/tangbinbin/tlog"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

var (
	configFile = flag.String("C", "./conf/im.conf", "config file")
	config     *Config
	s          *Server
)

func init() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	config = new(Config)
	if _, err := toml.DecodeFile(*configFile, config); err != nil {
		return
	}
	tlog.Init(config.Log)
	tlog.Infof("%+v", config)

	if err := newServer(); err != nil {
		tlog.Info(err.Error())
		return
	}

	s.start()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit
	s.Close()
	tlog.Close()
}

type Config struct {
	Addr string      `toml:"addr"`
	Log  tlog.Config `toml:"log"`
}

type Server struct {
	stop bool
}

func newServer() error {
	s = new(Server)
	return nil
}

func (s *Server) start() {
	go s.runTcp()
}

func (s *Server) runTcp() {
	ln, err := net.Listen("tcp", config.Addr)
	if err != nil {
		tlog.Infof("errorListen||errmsg=%s", err.Error())
		return
	}
	defer ln.Close()
	for {
		if s.stop {
			return
		}
		conn, err := ln.Accept()
		if err != nil {
			tlog.Infof("errorAccept||errmsg=%s", err.Error())
			break
		}
		tlog.Infof("accept||local=%s||remote=%s", conn.LocalAddr(), conn.RemoteAddr())
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {}

func (s *Server) Close() {}

type customer struct{}

type Mapper struct {
	mu sync.RWMutex
	ms map[uint64]*customer
}

func newMapper() *Mapper {
	return &Mapper{ms: make(map[uint64]*customer, 10240)}
}

func (m *Mapper) set(c *customer) error {
	return nil
}

func (m *Mapper) del(c *customer) {}

var (
	HEADER   []byte     = []byte{110, 119, 120}
	bytePool *sync.Pool = &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
)

type CP interface {
	Encode() []byte
	String() string
	Decode(*bufio.Reader) error
}

const (
	UNKNOWN    uint16 = 0
	CONNECT    uint16 = 1
	CONNACK    uint16 = 2
	DISCONNECT uint16 = 3
)
