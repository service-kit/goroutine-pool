package worker

import (
	"errors"
	"fmt"
	"sync"

	"github.com/service-kit/goroutine-pool/task"
	"time"
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

type ILoadCounter interface {
	AddLoad()
	ReduceLoad()
	Load() uint64
}

type WorkerType uint32
type WorkerFunc func(wp interface{}) error
type WorkerTaskMap map[task.TaskType]WorkerFunc

type IWorker interface {
	ID() uint32
	RegisterTask(t task.TaskType, f WorkerFunc)
	Start() error
	Stop() error
	Suspend() error
	Resume() error
	Interrupt() error
	AddTask(t task.ITask) error
	Status() WorkerStatus
	WorkStatus() WorkerWorkStatus
	SetLoadCounter(lc ILoadCounter)
	SetTimeout(timeout time.Duration)
	GetTimeout() time.Duration
}

type Worker struct {
	id       uint32
	wtm      WorkerTaskMap
	capacity uint32
	size     uint32
	taskCh   chan task.ITask
	mutex    sync.Mutex
	signalCh chan WorkerSignal
	status   WorkerStatus
	lc       ILoadCounter
	timeout  time.Duration
	errCh    chan error
}

func (w *Worker) ID() uint32 {
	return w.id
}

func (w *Worker) SetLoadCounter(lc ILoadCounter) {
	w.lc = lc
}

func (w *Worker) RegisterTask(t task.TaskType, f WorkerFunc) {
	w.wtm[t] = f
}

func (w *Worker) AddTask(t task.ITask) (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.size == w.capacity {
		return errors.New("worker busy")
	}
	t.SetDeadline(time.Now().Add(w.timeout).UnixNano() / int64(time.Millisecond))
	w.taskCh <- t
	w.lc.AddLoad()
	return nil
}

func (w *Worker) Start() (err error) {
	w.status = WS_RUNNING
	go w.safeWorkLoop()
	return
}

func (w *Worker) SetTimeout(timeout time.Duration) {
	w.timeout = timeout
}

func (w *Worker) GetTimeout() time.Duration {
	return w.timeout
}

func (w *Worker) safeWorkLoop() {
	for {
		w.workLoop()
	}
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
			err := w.proessTask(t)
			if nil != err {
				fmt.Println("process task err ", err.Error())
			}
			w.lc.ReduceLoad()
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

func (w *Worker) proessTask(t task.ITask) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("process task panic ", err)
		}
	}()
	f := w.wtm[t.GetType()]
	if nil == f {
		errStr := fmt.Sprintf("invalid task type:%v", t.GetType())
		return errors.New(errStr)
	}
	nowMilSec := time.Now().UnixNano() / int64(time.Millisecond)
	checkTime := t.GetDeadline() - nowMilSec
	if checkTime <= 0 {
		return errors.New("task timeout")
	}
	wf := func(param interface{}) chan error {
		w.errCh <- f(t.GetParam())
		return w.errCh
	}
	timeTicker := time.NewTicker(time.Duration(checkTime) * time.Millisecond)
	select {
	case <-timeTicker.C:
		return errors.New("task timeout")
	case err := <-wf(t.GetParam()):
		return err
	}
}

func NewWorker(id, capacity uint32) IWorker {
	w := new(Worker)
	w.id = id
	w.capacity = capacity
	w.taskCh = make(chan task.ITask, capacity)
	w.signalCh = make(chan WorkerSignal, 1)
	w.errCh = make(chan error, 1)
	w.status = WS_NULL
	w.wtm = make(WorkerTaskMap)
	return w
}
