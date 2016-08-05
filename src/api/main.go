package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var (
	configFile = flag.String("c", "./conf/api.conf", "config path")
	db         *sql.DB
	redisPool  *redis.Pool
	config     Config
)

const RspErr int = 1

type Config struct {
	Addr  string `json:"addr"`
	Redis string `json:"redis"`
	Mysql string `json:"mysql"`
}

func init() {
	flag.Parse()
	content, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", config)
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
	db, err = sql.Open("mysql", config.Mysql)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	log.Fatal(http.ListenAndServe(config.Addr, nil))
}

/*register*********************************************************************
req: POST
	u string
	p string
	rp string

******************************************************************************/
type RegisterRsp struct {
	Code   int    `json:"code"`
	ErrMsg string `json:"errmsg"`
	Uid    int64  `json:"uid"`
	Token  string `json:"token"`
}

func register(w http.ResponseWriter, r *http.Request) {
	u := r.PostFormValue("u")
	p := r.PostFormValue("p")
	rp := r.PostFormValue("rp")
	w.Header().Set("Content-Type", "application/json;charset=utf8")
	if p != rp {
		b, _ := json.Marshal(RegisterRsp{Code: RspErr, ErrMsg: "密码不一致"})
		w.Write(b)
		return
	}
	if len(p) < 8 || len(rp) < 8 {
		b, _ := json.Marshal(RegisterRsp{Code: RspErr, ErrMsg: "密码长度不能少于8位"})
		w.Write(b)
		return
	}
	ret, err := db.Exec("insert ignore into user_account (username, password, created_time) values (?, ?, ?)", u, fmt.Sprintf("%x", md5.Sum([]byte(p))), time.Now().Unix())
	if err != nil {
		b, _ := json.Marshal(RegisterRsp{Code: RspErr, ErrMsg: "数据库错误"})
		w.Write(b)
		return
	}
	if n, _ := ret.RowsAffected(); n != 1 {
		b, _ := json.Marshal(RegisterRsp{Code: RspErr, ErrMsg: "用户已经存在"})
		w.Write(b)
		return
	}
	uid, _ := ret.LastInsertId()
	token := irand(32)
	db.Exec("update user_account set token = ? where id = ?", token, uid)
	b, _ := json.Marshal(RegisterRsp{Uid: uid, Token: token})
	w.Write(b)
}

/*login************************************************************************
******************************************************************************/
type LoginRsp struct {
	Code   int    `json:"code"`
	ErrMsg string `json:"errmsg"`
	Uid    int64  `json:""errmsg"`
	Token  string `json:"token"`
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=utf8")
	u := r.PostFormValue("u")
	p := r.PostFormValue("p")
	mdp := fmt.Sprintf("%x", md5.Sum([]byte(p)))
	var uid int64
	var token string
	err := db.QueryRow("select id, token from user_account where username = ? and password = ?", u, mdp).Scan(&uid, &token)
	if err == sql.ErrNoRows {
		b, _ := json.Marshal(LoginRsp{Code: RspErr, ErrMsg: "用户名或密码错误"})
		w.Write(b)
		return
	}
	if err != nil {
		log.Println(err)
		b, _ := json.Marshal(RegisterRsp{Code: RspErr, ErrMsg: "数据库错误"})
		w.Write(b)
		return
	}
	b, _ := json.Marshal(LoginRsp{Uid: uid, Token: token})
	w.Write(b)
}

/*interval functions***********************************************************
******************************************************************************/
//生成随机字符串
func irand(size int) string {
	kinds := [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}
	result := make([]uint8, size)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		ikind := rand.Intn(3)
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return string(result)
}
