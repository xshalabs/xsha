package services

import (
	"sleep0-backend/database"
	"sleep0-backend/utils"
	"time"
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
	GetCredentialByName(name, createdBy string) (*database.GitCredential, error)
	ListCredentials(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	UpdateCredential(id uint, createdBy string, updates map[string]interface{}, secretData map[string]string) error
	DeleteCredential(id uint, createdBy string) error

	// 凭据操作
	UseCredential(id uint, createdBy string) (*database.GitCredential, error)
	ToggleCredential(id uint, createdBy string, isActive bool) error
	ListActiveCredentials(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error)

	// 凭据验证和解密
	DecryptCredentialSecret(credential *database.GitCredential, secretType string) (string, error)
	ValidateCredentialData(credType string, data map[string]string) error
}

// ProjectService 定义项目服务接口
type ProjectService interface {
	// 项目管理
	CreateProject(name, description, repoURL, protocol, defaultBranch, createdBy string, credentialID *uint) (*database.Project, error)
	GetProject(id uint, createdBy string) (*database.Project, error)
	GetProjectByName(name, createdBy string) (*database.Project, error)
	ListProjects(createdBy string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	UpdateProject(id uint, createdBy string, updates map[string]interface{}) error
	DeleteProject(id uint, createdBy string) error

	// 项目操作

	// 凭据相关
	ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint, createdBy string) error
	GetCompatibleCredentials(protocol database.GitProtocolType, createdBy string) ([]database.GitCredential, error)

	// Git操作
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
	GetEnvironmentByName(name, createdBy string) (*database.DevEnvironment, error)
	ListEnvironments(createdBy string, envType *database.DevEnvironmentType, status *database.DevEnvironmentStatus, page, pageSize int) ([]database.DevEnvironment, int64, error)
	UpdateEnvironment(id uint, createdBy string, updates map[string]interface{}) error
	DeleteEnvironment(id uint, createdBy string) error

	// 环境操作
	StartEnvironment(id uint, createdBy string) error
	StopEnvironment(id uint, createdBy string) error
	RestartEnvironment(id uint, createdBy string) error
	UseEnvironment(id uint, createdBy string) (*database.DevEnvironment, error)

	// 环境变量操作
	ValidateEnvVars(envVars map[string]string) error
	GetEnvironmentVars(id uint, createdBy string) (map[string]string, error)
	UpdateEnvironmentVars(id uint, createdBy string, envVars map[string]string) error

	// 资源限制验证
	ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error
}

// TaskService 定义任务服务接口
type TaskService interface {
	// 任务管理
	CreateTask(title, description, startBranch, createdBy string, projectID uint) (*database.Task, error)
	GetTask(id uint, createdBy string) (*database.Task, error)
	ListTasks(projectID *uint, createdBy string, status *database.TaskStatus, page, pageSize int) ([]database.Task, int64, error)
	UpdateTask(id uint, createdBy string, updates map[string]interface{}) error
	DeleteTask(id uint, createdBy string) error

	// 任务状态管理
	UpdateTaskStatus(id uint, createdBy string, status database.TaskStatus) error
	UpdatePullRequestStatus(id uint, createdBy string, hasPullRequest bool) error

	// 任务统计
	GetTaskStats(projectID uint, createdBy string) (map[database.TaskStatus]int64, error)
	ListTasksByProject(projectID uint, createdBy string) ([]database.Task, error)

	// 验证操作
	ValidateTaskData(title, startBranch string, projectID uint, createdBy string) error
}

// TaskConversationService 定义任务对话服务接口
type TaskConversationService interface {
	// 对话管理
	CreateConversation(taskID uint, content, createdBy string, role database.ConversationRole) (*database.TaskConversation, error)
	GetConversation(id uint, createdBy string) (*database.TaskConversation, error)
	ListConversations(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error)
	UpdateConversation(id uint, createdBy string, updates map[string]interface{}) error
	DeleteConversation(id uint, createdBy string) error

	// 对话状态管理
	UpdateConversationStatus(id uint, createdBy string, status database.ConversationStatus) error

	// 对话业务操作
	ListConversationsByTask(taskID uint, createdBy string) ([]database.TaskConversation, error)
	GetLatestConversation(taskID uint, createdBy string) (*database.TaskConversation, error)

	// 验证操作
	ValidateConversationData(taskID uint, content string, createdBy string) error
}
