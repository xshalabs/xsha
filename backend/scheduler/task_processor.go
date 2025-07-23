package scheduler

import (
	"log"
	"sleep0-backend/services"
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
	log.Println("开始处理待处理的任务对话...")

	if err := p.aiTaskExecutor.ProcessPendingConversations(); err != nil {
		log.Printf("处理任务失败: %v", err)
		return err
	}

	log.Println("任务处理完成")
	return nil
}
