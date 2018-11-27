package pool

import (
	"sync"
	"time"

	"github.com/service-kit/goroutine-pool/task"
	"github.com/service-kit/goroutine-pool/worker"
)

type WorkerMap map[uint32]worker.IWorker

type IPool interface {
	SetMaxIdle(mi uint32)
	SetMaxActive(ma uint32)
	SetInit(i uint32)
	SetTimeout(t time.Duration)
	AddTask(t task.ITask) error
	InitPool() error
	RegisterTask(wt task.TaskType, tf worker.WorkerFunc)
}

type Pool struct {
	maxIdle    uint32
	maxActive  uint32
	timeout    time.Duration
	wm         WorkerMap
	lb         *LoadBalancer
	capability uint32
	size       uint32
	mutex      sync.Mutex
	wtm        worker.WorkerTaskMap
	pm         map[uint32]uint64
}

func (p *Pool) SetInit(i uint32) {
	p.size = i
	p.capability = i
}

func (p *Pool) SetMaxIdle(mi uint32) {
	p.maxIdle = mi
}

func (p *Pool) SetMaxActive(ma uint32) {
	p.maxActive = ma
}

func (p *Pool) SetTimeout(t time.Duration) {
	p.timeout = t
}

func (p *Pool) AddTask(t task.ITask) (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	id := p.lb.GetMinLoadID()
	p.pm[id]++
	return p.wm[id].AddTask(t)
}

func (p *Pool) ShowPoolLoad() {
	for {
		p.lb.ShowLoad()
		time.Sleep(time.Second)
	}
}

func (p *Pool) InitPool() (err error) {
	p.pm = make(map[uint32]uint64)
	p.wm = make(WorkerMap)
	p.wtm = make(worker.WorkerTaskMap)
	p.lb = NewLoadBalancer()
	var i uint32 = 1
	for ; i <= p.size; i++ {
		w := worker.NewWorker(i, 1024)
		w.SetTimeout(p.timeout)
		p.lb.RegisterLoadCounter(i)
		w.SetLoadCounter(p.lb.GetLoacCounter(i))
		w.Start()
		p.wm[i] = w
	}
	return
}

func (p *Pool) RegisterTask(tt task.TaskType, tf worker.WorkerFunc) {
	p.wtm[tt] = tf
	var i uint32 = 1
	for ; i <= p.size; i++ {
		p.wm[i].RegisterTask(tt, tf)
	}
}

func NewPool(init, maxIdle, maxActive uint32, timeout time.Duration) (po IPool, err error) {
	p := new(Pool)
	p.SetInit(init)
	p.SetMaxActive(maxActive)
	p.SetMaxIdle(maxIdle)
	p.SetTimeout(timeout)
	err = p.InitPool()
	if nil != err {
		return nil, err
	}
	return p, nil
}
