package events

import (
	"context"
	"log"
	"sync"
	"time"
)

// WorkerTask 工作任务
type WorkerTask struct {
	Event        Event
	Subscription *Subscription
	Context      context.Context
}

// WorkerPool 工作池
type WorkerPool struct {
	workerCount int
	taskQueue   chan *WorkerTask
	workers     []*Worker
	
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Worker 工作者
type Worker struct {
	id        int
	taskQueue chan *WorkerTask
	quit      chan bool
	wg        *sync.WaitGroup
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 5
	}
	
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan *WorkerTask, workerCount*10), // 缓冲队列
		workers:     make([]*Worker, workerCount),
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	
	// 启动工作者
	for i := 0; i < wp.workerCount; i++ {
		worker := &Worker{
			id:        i + 1,
			taskQueue: wp.taskQueue,
			quit:      make(chan bool),
			wg:        &wp.wg,
		}
		
		wp.workers[i] = worker
		wp.wg.Add(1)
		go worker.start()
	}
	
	log.Printf("Worker pool started with %d workers", wp.workerCount)
}

// Stop 停止工作池
func (wp *WorkerPool) Stop(ctx context.Context) error {
	if wp.cancel != nil {
		wp.cancel()
	}
	
	// 通知所有工作者停止
	for _, worker := range wp.workers {
		close(worker.quit)
	}
	
	// 等待所有工作者停止
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Println("Worker pool stopped gracefully")
		return nil
	case <-ctx.Done():
		log.Println("Worker pool stopped with timeout")
		return ctx.Err()
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task *WorkerTask) {
	select {
	case wp.taskQueue <- task:
		// 任务提交成功
	default:
		// 队列已满，丢弃任务
		log.Printf("Worker pool queue full, dropping task for event %s", task.Event.GetID())
	}
}

// GetQueueSize 获取队列大小
func (wp *WorkerPool) GetQueueSize() int {
	return len(wp.taskQueue)
}

// start 启动工作者
func (w *Worker) start() {
	defer w.wg.Done()
	
	log.Printf("Worker %d started", w.id)
	
	for {
		select {
		case task := <-w.taskQueue:
			w.processTask(task)
			
		case <-w.quit:
			log.Printf("Worker %d stopping", w.id)
			return
		}
	}
}

// processTask 处理任务
func (w *Worker) processTask(task *WorkerTask) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d panic while processing event %s: %v", 
				w.id, task.Event.GetID(), r)
		}
	}()
	
	startTime := time.Now()
	
	// 设置超时上下文
	ctx := task.Context
	if timeout := 5 * time.Minute; timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	
	// 执行事件处理器
	err := task.Subscription.Handler(task.Event)
	
	duration := time.Since(startTime)
	
	if err != nil {
		log.Printf("Worker %d failed to process event %s: %v (took %v)", 
			w.id, task.Event.GetID(), err, duration)
	} else {
		log.Printf("Worker %d successfully processed event %s (took %v)", 
			w.id, task.Event.GetID(), duration)
	}
}