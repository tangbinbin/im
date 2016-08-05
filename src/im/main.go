package main

import (
	"bufio"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"improto"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

var (
	configFile = flag.String("c", "./conf/im.conf", "config file")
	config     Config
	db         *sql.DB
	redisPool  *redis.Pool
	readPool   = sync.Pool{
		New: func() interface{} { return bufio.NewReaderSize(nil, 8192) },
	}
)

type Config struct {
	Addr  string `json:"addr"`
	Redis string `json:"redis"`
	Mysql string `json:"mysql"`
}

func init() {
	flag.Parse()
	concent, e := ioutil.ReadFile(*configFile)
	if e != nil {
		log.Fatal(e)
	}
	e = json.Unmarshal(concent, &config)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("%+v\n", config)
	redisPool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		Wait:        true,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			t := 100 * time.Millisecond
			c, err := redis.DialTimeout("tcp", config.Redis, t, t, t)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	db, e = sql.Open("mysql", config.Mysql)
	if e != nil {
		log.Fatal(e)
	}
	e = db.Ping()
	if e != nil {
		log.Fatal(e)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ln, e := net.Listen("tcp", config.Addr)
	if e != nil {
		log.Fatal(e)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error")
			break
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	reader := readPool.Get().(*bufio.Reader)
	reader.Reset(conn)
	packet, err := readPacket(reader)
	if err != nil {
		goto EXIT
	}
	if packet.MsgType != improto.MsgType_LoginReq {
		goto EXIT
	}
EXIT:
	readPool.Put(reader)
	conn.Close()
}

/******************************************************************************
packet
******************************************************************************/
type Packet struct {
	MsgType improto.MsgType
	Body    []byte
}

func readPacket(reader *bufio.Reader) (Packet, error) {
	head := make([]byte, 4)
	err := readFull(reader, head)
	if err != nil {
		return Packet{}, err
	}
	length := binary.LittleEndian.Uint16(head[2:4])
	body := make([]byte, length)
	err = readFull(reader, body)
	if err != nil {
		return Packet{}, err
	}
	return Packet{MsgType: improto.MsgType(binary.LittleEndian.Uint16(head[0:2])), Body: body}, nil
}

func readFull(reader *bufio.Reader, in []byte) error {
	l := len(in)
	cnt := 0
	for {
		n, err := reader.Read(in[cnt:l])
		if err != nil {
			return err
		}
		cnt += n
		if cnt < l {
			continue
		}
		break
	}
	return nil
}
