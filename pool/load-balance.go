package pool

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type LoadCounter struct {
	loadCount uint64
	mutex     sync.RWMutex
}

func (lc *LoadCounter) AddLoad() {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	lc.loadCount++
}

func (lc *LoadCounter) ReduceLoad() {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	lc.loadCount--
}

func (lc *LoadCounter) Load() uint64 {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()
	return lc.loadCount
}

type WorkerLoadCounterMap map[uint32]*LoadCounter

type LoadBalancer struct {
	workerLoad WorkerLoadCounterMap
	mutex      sync.RWMutex
	size       uint32
	minLoadID  uint32
}

func (lb *LoadBalancer) RegisterLoadCounter(id uint32) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	if nil != lb.workerLoad[id] {
		return errors.New("has registered")
	}
	lb.workerLoad[id] = new(LoadCounter)
	lb.size++
	return nil
}

func (lb *LoadBalancer) GetLoacCounter(id uint32) *LoadCounter {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	return lb.workerLoad[id]
}

func (lb *LoadBalancer) GetMinLoadID() uint32 {
	return lb.minLoadID
}

func (lb *LoadBalancer) refreshMinLoadID() {
	for {
		if lb.size > 0 {
			lb.mutex.Lock()
			var minLoad uint64 = 0
			var id uint32 = 0
			for i, l := range lb.workerLoad {
				ld := l.Load()
				if 0 == minLoad {
					minLoad = ld
					id = i
				}
				if 0 == ld {
					id = i
					break
				}
				if minLoad > ld {
					minLoad = ld
					id = i
				}
			}
			lb.minLoadID = id
			lb.mutex.Unlock()
		}
		time.Sleep(time.Millisecond)
	}
}

func (lb *LoadBalancer) ShowLoad() {
	fmt.Println("----------------------------------------")
	fmt.Println("id", "load")
	for id, lc := range lb.workerLoad {
		ll := lc.Load()
		if ll == 0 {
			continue
		}
		fmt.Println(id, ll)
	}
	fmt.Println("----------------------------------------")
}

func NewLoadBalancer() *LoadBalancer {
	lb := new(LoadBalancer)
	lb.workerLoad = make(WorkerLoadCounterMap)
	go lb.refreshMinLoadID()
	return lb
}
