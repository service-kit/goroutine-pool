package main

import (
	"fmt"
	"io/ioutil"
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
	fmt.Print(rsp.StatusCode, et-bt)
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
				c := new(http.Client)
				ping(c, "ping")
			}()
			time.Sleep(30 * time.Millisecond)
		}
	}()
	wg.Add(1)
	wg.Wait()
}
