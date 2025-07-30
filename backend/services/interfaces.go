package services

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"
)

// AuthService 定义认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(username, password, clientIP, userAgent string) (bool, string, error)

	// Logout 用户登出
	Logout(token, username string) error

	// IsTokenBlacklisted 检查Token是否在黑名单
	IsTokenBlacklisted(token string) (bool, error)

	// CleanExpiredTokens 清理过期Token
	CleanExpiredTokens() error
}

// LoginLogService 定义登录日志服务接口
type LoginLogService interface {
	// GetLogs 获取登录日志
	GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error)

	// CleanOldLogs 清理旧日志
	CleanOldLogs(days int) error
}

// GitCredentialService 定义Git凭据服务接口
type GitCredentialService interface {
	// 凭据管理
	CreateCredential(name, description, credType, username, createdBy string, secretData map[string]string) (*database.GitCredential, error)
	GetCredential(id uint, createdBy string) (*database.GitCredential, error)

	ListCredentials(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	UpdateCredential(id uint, createdBy string, updates map[string]interface{}, secretData map[string]string) error
	DeleteCredential(id uint, createdBy string) error

	// 凭据操作
	ListActiveCredentials(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error)

	// 凭据验证和解密
	DecryptCredentialSecret(credential *database.GitCredential, secretType string) (string, error)
	ValidateCredentialData(credType string, data map[string]string) error
}

// ProjectService 定义项目服务接口
type ProjectService interface {
	// 项目管理
	CreateProject(name, description, repoURL, protocol, createdBy string, credentialID *uint) (*database.Project, error)
	GetProject(id uint, createdBy string) (*database.Project, error)

	ListProjects(createdBy string, name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	UpdateProject(id uint, createdBy string, updates map[string]interface{}) error
	DeleteProject(id uint, createdBy string) error

	ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint, createdBy string) error
	GetCompatibleCredentials(protocol database.GitProtocolType, createdBy string) ([]database.GitCredential, error)
	FetchRepositoryBranches(repoURL string, credentialID *uint, createdBy string) (*utils.GitAccessResult, error)
	ValidateRepositoryAccess(repoURL string, credentialID *uint, createdBy string) error
}

// AdminOperationLogService 定义管理员操作日志服务接口
type AdminOperationLogService interface {
	// 日志记录
	LogOperation(username, operation, resource, resourceID, description, details string,
		success bool, errorMsg, ip, userAgent, method, path string) error

	// 便捷记录方法
	LogCreate(username, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogUpdate(username, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogDelete(username, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogRead(username, resource, resourceID, description, ip, userAgent, path string) error
	LogLogin(username, ip, userAgent string, success bool, errorMsg string) error
	LogLogout(username, ip, userAgent string, success bool, errorMsg string) error

	// 查询操作
	GetLogs(username string, operation *database.AdminOperationType, resource string,
		success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error)
	GetLog(id uint) (*database.AdminOperationLog, error)

	// 统计操作
	GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error)

	// 清理操作
	CleanOldLogs(days int) error
}

// DevEnvironmentService 定义开发环境服务接口
type DevEnvironmentService interface {
	// 环境管理
	CreateEnvironment(name, description, envType, createdBy string, cpuLimit float64, memoryLimit int64, envVars map[string]string) (*database.DevEnvironment, error)
	GetEnvironment(id uint, createdBy string) (*database.DevEnvironment, error)

	ListEnvironments(createdBy string, envType *database.DevEnvironmentType, name *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	UpdateEnvironment(id uint, createdBy string, updates map[string]interface{}) error
	DeleteEnvironment(id uint, createdBy string) error

	// 环境变量操作
	ValidateEnvVars(envVars map[string]string) error
	GetEnvironmentVars(id uint, createdBy string) (map[string]string, error)
	UpdateEnvironmentVars(id uint, createdBy string, envVars map[string]string) error

	// 资源限制验证
	ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error

	// 获取可用的环境类型选项
	GetAvailableEnvironmentTypes() ([]map[string]interface{}, error)
}

// TaskService 定义任务服务接口
type TaskService interface {
	// 任务管理
	CreateTask(title, startBranch, createdBy string, projectID uint, devEnvironmentID *uint) (*database.Task, error)
	GetTask(id uint, createdBy string) (*database.Task, error)
	ListTasks(projectID *uint, createdBy string, status *database.TaskStatus, title *string, branch *string, devEnvID *uint, page, pageSize int) ([]database.Task, int64, error)
	UpdateTask(id uint, createdBy string, updates map[string]interface{}) error
	UpdateTaskStatus(id uint, createdBy string, status database.TaskStatus) error
	UpdateTaskStatusBatch(taskIDs []uint, createdBy string, status database.TaskStatus) ([]uint, []uint, error) // 批量更新状态，返回成功和失败的任务ID
	DeleteTask(id uint, createdBy string) error

	// 验证操作
	ValidateTaskData(title, startBranch string, projectID uint, createdBy string) error

	// Git diff 操作
	GetTaskGitDiff(task *database.Task, includeContent bool) (*utils.GitDiffSummary, error)
	GetTaskGitDiffFile(task *database.Task, filePath string) (string, error)

	// Git push 操作
	PushTaskBranch(id uint, createdBy string) (string, error)
}

// TaskConversationService 定义任务对话服务接口
type TaskConversationService interface {
	// 对话管理
	CreateConversation(taskID uint, content, createdBy string) (*database.TaskConversation, error)
	GetConversation(id uint, createdBy string) (*database.TaskConversation, error)
	ListConversations(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error)
	UpdateConversation(id uint, createdBy string, updates map[string]interface{}) error
	DeleteConversation(id uint, createdBy string) error

	// 对话业务操作

	GetLatestConversation(taskID uint, createdBy string) (*database.TaskConversation, error)

	// Git 差异操作
	GetConversationGitDiff(conversationID uint, createdBy string, includeContent bool) (*utils.GitDiffSummary, error)
	GetConversationGitDiffFile(conversationID uint, createdBy string, filePath string) (string, error)

	// 验证操作
	ValidateConversationData(taskID uint, content string, createdBy string) error
}

// TaskConversationResultService 定义任务对话结果服务接口
type TaskConversationResultService interface {
	// 结果管理
	CreateResult(conversationID uint, resultData map[string]interface{}) (*database.TaskConversationResult, error)
	GetResult(id uint) (*database.TaskConversationResult, error)
	GetResultByConversationID(conversationID uint) (*database.TaskConversationResult, error)
	UpdateResult(id uint, updates map[string]interface{}) error
	DeleteResult(id uint) error

	// 查询操作
	ListResultsByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error)
	ListResultsByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error)

	// 统计操作
	GetTaskStats(taskID uint) (map[string]interface{}, error)
	GetProjectStats(projectID uint) (map[string]interface{}, error)

	// 业务操作
	ExistsForConversation(conversationID uint) (bool, error)

	// 验证操作
	ValidateResultData(resultData map[string]interface{}) error
}

// AITaskExecutorService 定义AI任务执行服务接口
type AITaskExecutorService interface {
	// 处理待处理的对话
	ProcessPendingConversations() error

	// 获取执行日志
	GetExecutionLog(conversationID uint) (*database.TaskExecutionLog, error)

	// 取消执行
	CancelExecution(conversationID uint, createdBy string) error

	// 重试执行
	RetryExecution(conversationID uint, createdBy string) error

	// 获取执行状态信息
	GetExecutionStatus() map[string]interface{}

	// 工作空间清理
	CleanupWorkspaceOnFailure(taskID uint, workspacePath string) error
	CleanupWorkspaceOnCancel(taskID uint, workspacePath string) error
}

// SystemConfigService 定义系统配置服务接口
type SystemConfigService interface {
	// 配置管理
	GetConfig(id uint) (*database.SystemConfig, error)
	GetConfigByKey(key string) (*database.SystemConfig, error)
	ListConfigs(category string, page, pageSize int) ([]database.SystemConfig, int64, error)
	UpdateConfig(id uint, updates map[string]interface{}) error

	// 配置值操作
	GetValue(key string) (string, error)
	SetValue(key, value string) error
	GetConfigsByCategory(category string) (map[string]string, error)

	// 开发环境类型配置
	GetDevEnvironmentTypes() ([]map[string]interface{}, error)
	UpdateDevEnvironmentTypes(envTypes []map[string]interface{}) error

	// 系统初始化
	InitializeDefaultConfigs() error

	// 验证操作
	ValidateConfigData(key, value, category string) error
}
