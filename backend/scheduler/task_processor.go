package scheduler

import (
	"xsha-backend/services"
	"xsha-backend/utils"
)

type taskProcessor struct {
	aiTaskExecutor services.AITaskExecutorService
}

func NewTaskProcessor(aiTaskExecutor services.AITaskExecutorService) TaskProcessor {
	return &taskProcessor{
		aiTaskExecutor: aiTaskExecutor,
	}
}

func (p *taskProcessor) ProcessTasks() error {
	if err := p.aiTaskExecutor.ProcessPendingConversations(); err != nil {
		utils.Error("Task processing failed", "error", err)
		return err
	}
	return nil
}
