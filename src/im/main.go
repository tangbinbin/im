package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/tangbinbin/tlog"
	"os"
	"os/signal"
	"runtime"
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
