package executor

import (
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type conversationStateManager struct {
	taskConvRepo repository.TaskConversationRepository
	execLogRepo  repository.TaskExecutionLogRepository
}

// NewConversationStateManager 创建状态管理器
func NewConversationStateManager(
	taskConvRepo repository.TaskConversationRepository,
	execLogRepo repository.TaskExecutionLogRepository,
) ConversationStateManager {
	return &conversationStateManager{
		taskConvRepo: taskConvRepo,
		execLogRepo:  execLogRepo,
	}
}

// SetFailed 设置对话状态为失败并创建执行日志
func (c *conversationStateManager) SetFailed(conv *database.TaskConversation, errorMessage string) {
	// 更新对话状态为失败
	conv.Status = database.ConversationStatusFailed
	if updateErr := c.taskConvRepo.Update(conv); updateErr != nil {
		utils.Error("failed to update conversation status to failed", "error", updateErr)
	}

	// 创建执行日志记录失败原因
	execLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		ErrorMessage:   errorMessage,
		ExecutionLogs:  "", // 初始化为空字符串，避免NULL值问题
	}
	if logErr := c.execLogRepo.Create(execLog); logErr != nil {
		utils.Error("failed to create execution log", "error", logErr)
	}
}

// Rollback 回滚对话状态为失败
func (c *conversationStateManager) Rollback(conv *database.TaskConversation, errorMessage string) {
	conv.Status = database.ConversationStatusFailed
	if updateErr := c.taskConvRepo.Update(conv); updateErr != nil {
		utils.Error("failed to rollback conversation status to failed", "error", updateErr)
	}

	// 尝试创建或更新执行日志记录失败原因
	failedExecLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		ErrorMessage:   errorMessage,
		ExecutionLogs:  "", // 初始化为空字符串，避免NULL值问题
	}
	if logErr := c.execLogRepo.Create(failedExecLog); logErr != nil {
		utils.Error("failed to create failed execution log", "error", logErr)
	}
}

// RollbackToState 回滚对话和执行日志到指定状态
func (c *conversationStateManager) RollbackToState(
	conv *database.TaskConversation,
	execLog *database.TaskExecutionLog,
	convStatus database.ConversationStatus,
	errorMessage string,
) {
	// 回滚对话状态
	conv.Status = convStatus
	if updateErr := c.taskConvRepo.Update(conv); updateErr != nil {
		utils.Error("failed to rollback conversation status", "status", convStatus, "error", updateErr)
	}

	// 更新执行日志错误信息
	errorUpdates := map[string]interface{}{
		"error_message": errorMessage,
	}
	if updateErr := c.execLogRepo.UpdateMetadata(execLog.ID, errorUpdates); updateErr != nil {
		utils.Error("failed to update execution log", "error", updateErr)
	}
}
