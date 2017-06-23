package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/tangbinbin/tlog"
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

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

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

type Server struct{}

func newServer() error {
	s = new(Server)
	return nil
}

func (s *Server) start() {}

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
