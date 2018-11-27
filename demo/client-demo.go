package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func ping(c *http.Client, path string) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8001/"+path, strings.NewReader("ping"))
	if nil != err {
		fmt.Println("new request err ", err.Error())
		return
	}
	bt := time.Now().UnixNano() / int64(time.Microsecond)
	rsp, err := c.Do(req)
	if nil != err {
		fmt.Println("ping err ", err.Error())
		return
	}
	et := time.Now().UnixNano() / int64(time.Microsecond)
	fmt.Printf("status code:%v cost time:%v\n", rsp.StatusCode, et-bt)
	data, err := ioutil.ReadAll(rsp.Body)
	if nil != err {
		fmt.Println("ping err ", err.Error())
		return
	}
	fmt.Println(string(data))
}

func main() {
	wg := sync.WaitGroup{}
	go func() {
		for {
			go func() {
				tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8001")
				conn, err := net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					fmt.Println("server is not starting")
					return
				}
				conn.Write([]byte("ping"))
				conn.CloseWrite()
				fmt.Println(conn.RemoteAddr().String())
				readBuf := new(bytes.Buffer)
				readBuf.ReadFrom(conn)
				conn.CloseRead()
				fmt.Println(readBuf.String())
			}()
			time.Sleep(time.Second)
		}
	}()
	wg.Add(1)
	wg.Wait()
}
