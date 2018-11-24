package pool

import (
	"time"
	"github.com/service-kit/goroutine-pool/task"
)

type IPool interface {
	SetMaxIdle(mi uint32)
	SetMaxActive(ma uint32)
	SetTimeout(t time.Duration)
	AddTask(t task.Task) error
}

type Pool struct {
	maxIdle   uint32
	maxActive uint32
	timeout time.Duration
}

func (p *Pool) SetMaxIdle(mi uint32) {
	p.maxIdle = mi
}

func (p *Pool) SetMaxActive(ma uint32) {
	p.maxActive = ma
}

func(p *Pool) SetTimeout(t time.Duration) {
	p.timeout = t
}

func (p *Pool) AddTask(t task.Task) (err error) {
	return
}

func NewPool(maxIdle,maxActive uint32,timeout time.Duration) IPool {
	p := new(Pool)
	p.maxActive = maxActive
	p.maxIdle = maxIdle
	p.timeout = timeout
	return p
}