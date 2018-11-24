package worker

import (
	"github.com/kataras/iris/core/errors"
	"github.com/service-kit/goroutine-pool/task"
	"sync"
	"fmt"
)

type WorkerStatus uint32
type WorkerWorkStatus uint32

const (
	WS_NULL WorkerStatus = iota
	WS_RUNNING
	WS_SUSPENDED
	WS_STOPED
	WS_INTERRUPTED
)

const (
	WWS_NULL WorkerWorkStatus = iota
	WWS_NORMAL
	WWS_BUSY
	WWS_FULL
)

type WorkerSignal byte

const (
	WSI_NULL WorkerSignal = iota
	WSI_SUSPEND
	WSI_RESUME
	WSI_INTERRUPT
	WSI_STOP
)

type WorkerType uint32
type WorkerFunc func(wp interface{}) error
type WorkerTaskMap map[task.TaskType]WorkerFunc

type IWorker interface {
	RegisterTask(t task.TaskType,f WorkerFunc)
	Start() error
	Stop() error
	Suspend() error
	Resume() error
	Interrupt() error
	AddTask(t task.Task) error
	Status() WorkerStatus
	WorkStatus() WorkerWorkStatus
}

type Worker struct {
	wtm       WorkerTaskMap
	capacity uint32
	size     uint32
	taskCh   chan task.Task
	mutex    sync.Mutex
	signalCh chan WorkerSignal
	status   WorkerStatus
}

func (w *Worker) RegisterTask(t task.TaskType,f WorkerFunc) {
	w.wtm[t] = f
}

func (w *Worker) AddTask(t task.Task) (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.size == w.capacity {
		return errors.New("worker busy")
	}
	w.taskCh <- t
	return nil
}

func (w *Worker) Start() (err error) {
	w.status = WS_RUNNING
	go w.workLoop()
	return
}

func (w *Worker) Stop() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.status == WS_STOPED {
		return errors.New("worker has stoped")
	}
	w.status = WS_STOPED
	w.signalCh <- WSI_STOP
	return
}

func (w *Worker) Suspend() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.status == WS_SUSPENDED {
		return errors.New("worker has suspended")
	}
	w.status = WS_SUSPENDED
	w.signalCh <- WSI_SUSPEND
	return
}

func (w *Worker) Resume() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.status != WS_SUSPENDED {
		return errors.New("worker has not suspended")
	}
	w.status = WS_RUNNING
	w.signalCh <- WSI_RESUME
	return
}

func (w *Worker) Interrupt() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.status == WS_INTERRUPTED {
		return errors.New("worker has interrupted")
	}
	w.status = WS_INTERRUPTED
	w.signalCh <- WSI_INTERRUPT
	return
}

func (w *Worker) Status() WorkerStatus {
	return w.status
}

func (w *Worker) WorkStatus() WorkerWorkStatus {
	if w.size < w.capacity/2 {
		return WWS_NORMAL
	} else if w.size <= w.capacity {
		return WWS_BUSY
	} else {
		return WWS_FULL
	}
}

func (w *Worker) workLoop() {
	for {
		select {
		case t := <-w.taskCh:
			f := w.wtm[t.GetType()]
			if nil == f {
				fmt.Errorf("invalid type:%v\n",t.GetType())
				continue
			}
			e := f(t.GetParam())
			if nil != e {
				fmt.Errorf("do task err:%v\n",e)
			}
		case sig := <-w.signalCh:
			if sig == WSI_SUSPEND {
				for {
					s := <-w.signalCh
					if s == WSI_STOP || s == WSI_INTERRUPT {
						return
					}
					if s == WSI_RESUME {
						break
					}
				}
			} else if sig == WSI_STOP || sig == WSI_INTERRUPT {
				return
			}
		}
	}
}

func NewWorker(capacity uint32) IWorker {
	w := new(Worker)
	w.capacity = capacity
	w.taskCh = make(chan task.Task, capacity)
	w.signalCh = make(chan WorkerSignal)
	w.status = WS_NULL
	return w
}
