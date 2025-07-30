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

func NewConversationStateManager(
	taskConvRepo repository.TaskConversationRepository,
	execLogRepo repository.TaskExecutionLogRepository,
) ConversationStateManager {
	return &conversationStateManager{
		taskConvRepo: taskConvRepo,
		execLogRepo:  execLogRepo,
	}
}

func (c *conversationStateManager) SetFailed(conv *database.TaskConversation, errorMessage string) {
	conv.Status = database.ConversationStatusFailed
	if updateErr := c.taskConvRepo.Update(conv); updateErr != nil {
		utils.Error("failed to update conversation status to failed", "error", updateErr)
	}

	execLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		ErrorMessage:   errorMessage,
		ExecutionLogs:  "",
	}
	if logErr := c.execLogRepo.Create(execLog); logErr != nil {
		utils.Error("failed to create execution log", "error", logErr)
	}
}

func (c *conversationStateManager) Rollback(conv *database.TaskConversation, errorMessage string) {
	conv.Status = database.ConversationStatusFailed
	if updateErr := c.taskConvRepo.Update(conv); updateErr != nil {
		utils.Error("failed to rollback conversation status to failed", "error", updateErr)
	}

	failedExecLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		ErrorMessage:   errorMessage,
		ExecutionLogs:  "",
	}
	if logErr := c.execLogRepo.Create(failedExecLog); logErr != nil {
		utils.Error("failed to create failed execution log", "error", logErr)
	}
}

func (c *conversationStateManager) RollbackToState(
	conv *database.TaskConversation,
	execLog *database.TaskExecutionLog,
	convStatus database.ConversationStatus,
	errorMessage string,
) {
	conv.Status = convStatus
	if updateErr := c.taskConvRepo.Update(conv); updateErr != nil {
		utils.Error("failed to rollback conversation status", "status", convStatus, "error", updateErr)
	}

	errorUpdates := map[string]interface{}{
		"error_message": errorMessage,
	}
	if updateErr := c.execLogRepo.UpdateMetadata(execLog.ID, errorUpdates); updateErr != nil {
		utils.Error("failed to update execution log", "error", updateErr)
	}
}
