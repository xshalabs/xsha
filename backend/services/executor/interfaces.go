package executor

import (
	"context"
	"xsha-backend/database"
)

type DockerExecutor interface {
	CheckAvailability() error
	BuildCommandForLog(conv *database.TaskConversation, workspacePath string) string
	ExecuteWithContext(ctx context.Context, dockerCmd string, execLogID uint) error
	ExecuteWithContainerTracking(ctx context.Context, conv *database.TaskConversation, workspacePath string, execLogID uint) (string, error)
	StopAndRemoveContainer(containerID string) error
}

type ResultParser interface {
	ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog)
	ParseFromLogs(executionLogs string) (map[string]interface{}, error)
}

type WorkspaceCleaner interface {
	CleanupOnFailure(taskID uint, workspacePath string) error
	CleanupOnCancel(taskID uint, workspacePath string) error
	CleanupBeforeExecution(taskID uint, workspacePath string) error
}

type ConversationStateManager interface {
	SetFailed(conv *database.TaskConversation, errorMessage string)
	Rollback(conv *database.TaskConversation, errorMessage string)
	RollbackToState(conv *database.TaskConversation, execLog *database.TaskExecutionLog,
		status database.ConversationStatus, errorMessage string)
}

type LogAppender interface {
	AppendLog(execLogID uint, content string)
}
