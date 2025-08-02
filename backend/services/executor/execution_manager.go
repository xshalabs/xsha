package executor

import (
	"context"
	"sync"
)

type ExecutionInfo struct {
	CancelFunc  context.CancelFunc
	ContainerID string
}

type ExecutionManager struct {
	runningConversations map[uint]*ExecutionInfo
	maxConcurrency       int
	currentCount         int
	mu                   sync.RWMutex
}

func NewExecutionManager(maxConcurrency int) *ExecutionManager {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}
	return &ExecutionManager{
		runningConversations: make(map[uint]*ExecutionInfo),
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

	em.runningConversations[conversationID] = &ExecutionInfo{
		CancelFunc:  cancelFunc,
		ContainerID: "", // Will be set later
	}
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

func (em *ExecutionManager) SetContainerID(conversationID uint, containerID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if execInfo, exists := em.runningConversations[conversationID]; exists {
		execInfo.ContainerID = containerID
	}
}

func (em *ExecutionManager) GetContainerID(conversationID uint) string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if execInfo, exists := em.runningConversations[conversationID]; exists {
		return execInfo.ContainerID
	}
	return ""
}

func (em *ExecutionManager) CancelExecution(conversationID uint) (context.CancelFunc, string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if execInfo, exists := em.runningConversations[conversationID]; exists {
		cancelFunc := execInfo.CancelFunc
		containerID := execInfo.ContainerID
		delete(em.runningConversations, conversationID)
		em.currentCount--
		return cancelFunc, containerID
	}
	return nil, ""
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
