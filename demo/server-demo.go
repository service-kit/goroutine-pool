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
	"net"
)

const (
	GP_INIT       uint32        = 10
	GP_MAX_IDLE   uint32        = 100
	GP_MAX_ACTIVE uint32        = 200
	TASK_TCP      task.TaskType = 1
)

type HttpParam struct {
	w http.ResponseWriter
	r *http.Request
}

type Demo_TCPServer struct {
	gp pool.IPool
}

var m *Demo_TCPServer
var once sync.Once
var logger *zap.Logger

func GetInstance() *Demo_TCPServer {
	once.Do(func() {
		m = &Demo_TCPServer{}
	})
	return m
}

func (dhs *Demo_TCPServer) Init() (err error) {
	dhs.gp, err = pool.NewPool(GP_INIT, GP_MAX_IDLE, GP_MAX_ACTIVE, 5*time.Second)
	if nil != err {
		return
	}
	dhs.gp.RegisterTask(TASK_TCP, func(wp interface{}) error {
		con := wp.(*net.TCPConn)
		readBuf := make([]byte, 1024)
		con.Read(readBuf)
		con.CloseRead()
		time.Sleep(6 * time.Second)
		con.Write([]byte("pong"))
		con.CloseWrite()
		return nil
	})
	return
}

func (dhs *Demo_TCPServer) Start() {
	defer func() {
		fmt.Println("server shutdown!!!")
	}()
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ":8001")
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		return
	}
	defer lis.Close()
	for {
		con, err := lis.AcceptTCP()
		if nil != err {
			continue
		}
		dhs.handleRequest(con)
	}
}

func (dhs *Demo_TCPServer) AddTask(t task.ITask) {
	dhs.gp.AddTask(t)
}

func (dhs *Demo_TCPServer) handleRequest(con net.Conn) {
	t := task.NewTask(TASK_TCP, con)
	dhs.gp.AddTask(t)
}

func main() {
	ds := new(Demo_TCPServer)
	err := ds.Init()
	if nil != err {
		fmt.Println("init err ", err.Error())
		return
	}
	ds.Start()
}
