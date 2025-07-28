package scheduler

import (
	"xsha-backend/services"
	"xsha-backend/utils"
)

type taskProcessor struct {
	aiTaskExecutor services.AITaskExecutorService
}

// NewTaskProcessor 创建任务处理器
func NewTaskProcessor(aiTaskExecutor services.AITaskExecutorService) TaskProcessor {
	return &taskProcessor{
		aiTaskExecutor: aiTaskExecutor,
	}
}

// ProcessTasks 处理任务
func (p *taskProcessor) ProcessTasks() error {
	utils.Info("开始处理待处理的任务对话...")

	if err := p.aiTaskExecutor.ProcessPendingConversations(); err != nil {
		utils.Error("处理任务失败", "error", err)
		return err
	}

	utils.Info("任务处理完成")
	return nil
}
