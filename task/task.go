package task

type TaskType uint32

type ITask interface {
	SetType(t TaskType)
	GetType() TaskType
	SetParam(p interface{})
	GetParam() interface{}
	SetDeadline(deadline int64)
	GetDeadline() int64
}

type Task struct {
	tType    TaskType
	param    interface{}
	deadline int64
}

func (t *Task) SetType(tt TaskType) {
	t.tType = tt
}

func (t *Task) GetType() TaskType {
	return t.tType
}

func (t *Task) SetParam(p interface{}) {
	t.param = p
}

func (t *Task) GetParam() interface{} {
	return t.param
}

func (t *Task) SetDeadline(deadline int64) {
	t.deadline = deadline
}

func (t *Task) GetDeadline() int64 {
	return t.deadline
}

func NewTask(tt TaskType, tp interface{}) ITask {
	t := new(Task)
	t.SetType(tt)
	t.SetParam(tp)
	return t
}
