package scheduler

import (
	"sync"
	"time"
	"xsha-backend/utils"
)

type schedulerManager struct {
	processor TaskProcessor
	ticker    *time.Ticker
	quit      chan struct{}
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
	interval  time.Duration
}

func NewSchedulerManager(processor TaskProcessor, interval time.Duration) Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	return &schedulerManager{
		processor: processor,
		interval:  interval,
		quit:      make(chan struct{}),
	}
}

func (s *schedulerManager) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.ticker = time.NewTicker(s.interval)
	s.running = true

	s.wg.Add(1)
	go s.run()

	return nil
}

func (s *schedulerManager) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	close(s.quit)
	s.ticker.Stop()
	s.wg.Wait()
	s.running = false

	utils.Info("Scheduler stopped")
	return nil
}

func (s *schedulerManager) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *schedulerManager) run() {
	defer s.wg.Done()

	if err := s.processor.ProcessTasks(); err != nil {
		utils.Error("Initial task processing failed", "error", err)
	}

	for {
		select {
		case <-s.ticker.C:
			if err := s.processor.ProcessTasks(); err != nil {
				utils.Error("Scheduled task processing failed", "error", err)
			}
		case <-s.quit:
			return
		}
	}
}
