package executor

import (
	"context"
	"xsha-backend/database"
)

// DockerExecutor Docker执行器接口
type DockerExecutor interface {
	// 检查Docker可用性
	CheckAvailability() error
	// 构建Docker命令
	BuildCommand(conv *database.TaskConversation, workspacePath string) string
	// 构建用于日志记录的Docker命令(敏感信息已掩码)
	BuildCommandForLog(conv *database.TaskConversation, workspacePath string) string
	// 执行Docker命令
	ExecuteWithContext(ctx context.Context, dockerCmd string, execLogID uint) error
}

// ResultParser 结果解析器接口
type ResultParser interface {
	// 解析并创建任务结果
	ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog)
	// 从执行日志中解析结果
	ParseFromLogs(executionLogs string) (map[string]interface{}, error)
}

// WorkspaceCleaner 工作空间清理器接口
type WorkspaceCleaner interface {
	// 清理失败任务的工作空间
	CleanupOnFailure(taskID uint, workspacePath string) error
	// 清理取消任务的工作空间
	CleanupOnCancel(taskID uint, workspacePath string) error
}

// ConversationStateManager 对话状态管理器接口
type ConversationStateManager interface {
	// 设置对话为失败状态
	SetFailed(conv *database.TaskConversation, errorMessage string)
	// 回滚对话状态
	Rollback(conv *database.TaskConversation, errorMessage string)
	// 回滚到指定状态
	RollbackToState(conv *database.TaskConversation, execLog *database.TaskExecutionLog,
		status database.ConversationStatus, errorMessage string)
}

// LogAppender 日志追加器接口
type LogAppender interface {
	// 追加日志并广播
	AppendLog(execLogID uint, content string)
}
