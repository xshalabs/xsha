package executor

import (
	"context"
	"sync"
)

// ExecutionManager 执行管理器 - 负责并发控制
type ExecutionManager struct {
	runningConversations map[uint]context.CancelFunc // 正在运行的对话及其取消函数
	maxConcurrency       int                         // 最大并发数
	currentCount         int                         // 当前执行数量
	mu                   sync.RWMutex                // 读写锁
}

// NewExecutionManager 创建执行管理器
func NewExecutionManager(maxConcurrency int) *ExecutionManager {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // 默认最大并发数为5
	}
	return &ExecutionManager{
		runningConversations: make(map[uint]context.CancelFunc),
		maxConcurrency:       maxConcurrency,
	}
}

// CanExecute 检查是否可以执行新任务
func (em *ExecutionManager) CanExecute() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.currentCount < em.maxConcurrency
}

// AddExecution 添加执行任务
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

// RemoveExecution 移除执行任务
func (em *ExecutionManager) RemoveExecution(conversationID uint) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.runningConversations[conversationID]; exists {
		delete(em.runningConversations, conversationID)
		em.currentCount--
	}
}

// CancelExecution 取消特定执行
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

// GetRunningCount 获取当前运行数量
func (em *ExecutionManager) GetRunningCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.currentCount
}

// IsRunning 检查特定对话是否在运行
func (em *ExecutionManager) IsRunning(conversationID uint) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	_, exists := em.runningConversations[conversationID]
	return exists
}
