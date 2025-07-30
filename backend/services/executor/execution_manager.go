package executor

import (
	"context"
	"sync"
)

type ExecutionManager struct {
	runningConversations map[uint]context.CancelFunc
	maxConcurrency       int
	currentCount         int
	mu                   sync.RWMutex
}

func NewExecutionManager(maxConcurrency int) *ExecutionManager {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}
	return &ExecutionManager{
		runningConversations: make(map[uint]context.CancelFunc),
		maxConcurrency:       maxConcurrency,
	}
}

func (em *ExecutionManager) CanExecute() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.currentCount < em.maxConcurrency
}

func (em *ExecutionManager) AddExecution(conversationID uint, cancelFunc context.CancelFunc) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.currentCount >= em.maxConcurrency {
		return false
	}

	em.runningConversations[conversationID] = cancelFunc
	em.currentCount++
	return true
}

func (em *ExecutionManager) RemoveExecution(conversationID uint) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.runningConversations[conversationID]; exists {
		delete(em.runningConversations, conversationID)
		em.currentCount--
	}
}

func (em *ExecutionManager) CancelExecution(conversationID uint) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	if cancelFunc, exists := em.runningConversations[conversationID]; exists {
		cancelFunc()
		delete(em.runningConversations, conversationID)
		em.currentCount--
		return true
	}
	return false
}

func (em *ExecutionManager) GetRunningCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.currentCount
}

func (em *ExecutionManager) IsRunning(conversationID uint) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	_, exists := em.runningConversations[conversationID]
	return exists
}
