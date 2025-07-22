package repository

import (
	"sleep0-backend/database"
	"time"
)

// TokenBlacklistRepository 定义Token黑名单仓库接口
type TokenBlacklistRepository interface {
	Add(token string, username string, expiresAt time.Time, reason string) error
	IsBlacklisted(token string) (bool, error)
	CleanExpired() error
}

// LoginLogRepository 定义登录日志仓库接口
type LoginLogRepository interface {
	Add(username, ip, userAgent, reason string, success bool) error
	GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error)
	CleanOld(days int) error
}

// GitCredentialRepository 定义Git凭据仓库接口
type GitCredentialRepository interface {
	// 基本CRUD操作
	Create(credential *database.GitCredential) error
	GetByID(id uint, createdBy string) (*database.GitCredential, error)
	GetByName(name, createdBy string) (*database.GitCredential, error)
	List(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	Update(credential *database.GitCredential) error
	Delete(id uint, createdBy string) error

	// 业务操作
	UpdateLastUsed(id uint, createdBy string) error
	SetActive(id uint, createdBy string, isActive bool) error
	ListActive(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error)
}

// ProjectRepository 定义项目仓库接口
type ProjectRepository interface {
	// 基本CRUD操作
	Create(project *database.Project) error
	GetByID(id uint, createdBy string) (*database.Project, error)
	GetByName(name, createdBy string) (*database.Project, error)
	List(createdBy string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	Update(project *database.Project) error
	Delete(id uint, createdBy string) error

	// 业务操作
	UpdateLastUsed(id uint, createdBy string) error
	GetByCredentialID(credentialID uint, createdBy string) ([]database.Project, error)
}

// AdminOperationLogRepository 定义管理员操作日志仓库接口
type AdminOperationLogRepository interface {
	// 基本操作
	Add(log *database.AdminOperationLog) error
	GetByID(id uint) (*database.AdminOperationLog, error)

	// 查询操作
	List(username string, operation *database.AdminOperationType, resource string,
		success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error)

	// 统计操作
	GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error)

	// 清理操作
	CleanOld(days int) error
}

// DevEnvironmentRepository 定义开发环境仓库接口
type DevEnvironmentRepository interface {
	// 基本CRUD操作
	Create(env *database.DevEnvironment) error
	GetByID(id uint, createdBy string) (*database.DevEnvironment, error)
	GetByName(name, createdBy string) (*database.DevEnvironment, error)
	List(createdBy string, envType *database.DevEnvironmentType, status *database.DevEnvironmentStatus, page, pageSize int) ([]database.DevEnvironment, int64, error)
	Update(env *database.DevEnvironment) error
	Delete(id uint, createdBy string) error

	// 业务操作
	UpdateLastUsed(id uint, createdBy string) error
	UpdateStatus(id uint, createdBy string, status database.DevEnvironmentStatus) error
	ListByStatus(createdBy string, status database.DevEnvironmentStatus) ([]database.DevEnvironment, error)
}

// TaskRepository 定义任务仓库接口
type TaskRepository interface {
	// 基本CRUD操作
	Create(task *database.Task) error
	GetByID(id uint, createdBy string) (*database.Task, error)
	List(projectID *uint, createdBy string, status *database.TaskStatus, page, pageSize int) ([]database.Task, int64, error)
	Update(task *database.Task) error
	Delete(id uint, createdBy string) error

	// 业务操作
	UpdateStatus(id uint, createdBy string, status database.TaskStatus) error
	UpdatePullRequestStatus(id uint, createdBy string, hasPullRequest bool) error
	ListByProject(projectID uint, createdBy string) ([]database.Task, error)
	CountByStatus(projectID uint, createdBy string) (map[database.TaskStatus]int64, error)
}

// TaskConversationRepository 定义任务对话仓库接口
type TaskConversationRepository interface {
	// 基本CRUD操作
	Create(conversation *database.TaskConversation) error
	GetByID(id uint, createdBy string) (*database.TaskConversation, error)
	List(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error)
	Update(conversation *database.TaskConversation) error
	Delete(id uint, createdBy string) error

	// 业务操作
	UpdateStatus(id uint, createdBy string, status database.ConversationStatus) error
	ListByTask(taskID uint, createdBy string) ([]database.TaskConversation, error)
	GetLatestByTask(taskID uint, createdBy string) (*database.TaskConversation, error)
}
