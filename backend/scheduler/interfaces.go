package scheduler

type Scheduler interface {
	Start() error
	Stop() error
	IsRunning() bool
}

type TaskProcessor interface {
	ProcessTasks() error
}
