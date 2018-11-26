package task

type TaskType uint32

type ITask interface {
	SetType(t TaskType)
	GetType() TaskType
	SetParam(p interface{})
	GetParam() interface{}
}

type Task struct {
	tType TaskType
	param interface{}
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

func NewTask(tt TaskType, tp interface{}) ITask {
	t := new(Task)
	t.SetType(tt)
	t.SetParam(tp)
	return t
}
