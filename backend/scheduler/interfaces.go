package scheduler

// Scheduler 定时器接口
type Scheduler interface {
	Start() error
	Stop() error
	IsRunning() bool
}

// TaskProcessor 任务处理器接口
type TaskProcessor interface {
	ProcessTasks() error
}
