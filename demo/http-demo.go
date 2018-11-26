package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/service-kit/goroutine-pool/pool"
	"github.com/service-kit/goroutine-pool/task"
)

const (
	GP_INIT       uint32        = 500
	GP_MAX_IDLE   uint32        = 100
	GP_MAX_ACTIVE uint32        = 200
	TASK_HTTP     task.TaskType = 1
)

type HttpParam struct {
	w http.ResponseWriter
	r *http.Request
}

type Demo_HttpServer struct {
	gp pool.IPool
}

var m *Demo_HttpServer
var once sync.Once
var logger *zap.Logger

func GetInstance() *Demo_HttpServer {
	once.Do(func() {
		m = &Demo_HttpServer{}
	})
	return m
}

func (dhs *Demo_HttpServer) Init() (err error) {
	dhs.gp, err = pool.NewPool(GP_INIT, GP_MAX_IDLE, GP_MAX_ACTIVE, 5*time.Second)
	if nil != err {
		return
	}
	dhs.gp.RegisterTask(TASK_HTTP, func(wp interface{}) error {
		req := wp.(*HttpParam)
		req.r.ParseForm()
		time.Sleep(time.Duration(2000) * time.Millisecond)
		req.w.Write([]byte("pong"))
		return nil
	})
	http.HandleFunc("/ping", dhs.handleRequest)
	return
}

func (dhs *Demo_HttpServer) Start() {
	defer func() {
		fmt.Println("server shutdown!!!")
	}()
	http.ListenAndServe(":8001", nil)
}

func (dhs *Demo_HttpServer) AddTask(t task.ITask) {
	dhs.gp.AddTask(t)
}

func (dhs *Demo_HttpServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	t := task.NewTask(TASK_HTTP, &HttpParam{w: w, r: r})
	dhs.gp.AddTask(t)
}

func main() {
	ds := new(Demo_HttpServer)
	err := ds.Init()
	if nil != err {
		fmt.Println("init err ", err.Error())
		return
	}
	ds.Start()
}
