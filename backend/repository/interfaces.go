package repository

import (
	"time"
	"xsha-backend/database"
)

// TokenBlacklistRepository defines Token blacklist repository interface
type TokenBlacklistRepository interface {
	Add(token string, username string, expiresAt time.Time, reason string) error
	IsBlacklisted(token string) (bool, error)
	CleanExpired() error
}

// LoginLogRepository defines login log repository interface
type LoginLogRepository interface {
	Add(username, ip, userAgent, reason string, success bool) error
	GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error)
	CleanOld(days int) error
}

// GitCredentialRepository defines Git credential repository interface
type GitCredentialRepository interface {
	// Basic CRUD operations
	Create(credential *database.GitCredential) error
	GetByID(id uint, createdBy string) (*database.GitCredential, error)
	GetByName(name, createdBy string) (*database.GitCredential, error)
	List(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	Update(credential *database.GitCredential) error
	Delete(id uint, createdBy string) error

	// Business operations
}

// ProjectRepository defines project repository interface
type ProjectRepository interface {
	// Basic CRUD operations
	Create(project *database.Project) error
	GetByID(id uint, createdBy string) (*database.Project, error)
	GetByName(name, createdBy string) (*database.Project, error)
	List(createdBy string, name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	Update(project *database.Project) error
	Delete(id uint, createdBy string) error

	// Business operations
	UpdateLastUsed(id uint, createdBy string) error
	GetByCredentialID(credentialID uint, createdBy string) ([]database.Project, error)
}

// AdminOperationLogRepository defines admin operation log repository interface
type AdminOperationLogRepository interface {
	// Basic operations
	Add(log *database.AdminOperationLog) error
	GetByID(id uint) (*database.AdminOperationLog, error)

	// Query operations
	List(username string, operation *database.AdminOperationType, resource string,
		success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error)

	// Statistics operations
	GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error)

	// Cleanup operations
	CleanOld(days int) error
}

// DevEnvironmentRepository defines development environment repository interface
type DevEnvironmentRepository interface {
	// Basic CRUD operations
	Create(env *database.DevEnvironment) error
	GetByID(id uint, createdBy string) (*database.DevEnvironment, error)
	GetByName(name, createdBy string) (*database.DevEnvironment, error)
	List(createdBy string, envType *database.DevEnvironmentType, name *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	Update(env *database.DevEnvironment) error
	Delete(id uint, createdBy string) error
}

// TaskRepository defines task repository interface
type TaskRepository interface {
	// Basic CRUD operations
	Create(task *database.Task) error
	GetByID(id uint, createdBy string) (*database.Task, error)
	List(projectID *uint, createdBy string, status *database.TaskStatus, title *string, branch *string, devEnvID *uint, page, pageSize int) ([]database.Task, int64, error)
	Update(task *database.Task) error
	Delete(id uint, createdBy string) error

	// Business operations
	ListByProject(projectID uint, createdBy string) ([]database.Task, error)
}

// TaskConversationRepository defines task conversation repository interface
type TaskConversationRepository interface {
	// Basic CRUD operations
	Create(conversation *database.TaskConversation) error
	GetByID(id uint, createdBy string) (*database.TaskConversation, error)
	List(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error)
	Update(conversation *database.TaskConversation) error
	Delete(id uint, createdBy string) error

	// Business operations
	ListByTask(taskID uint, createdBy string) ([]database.TaskConversation, error)
	GetLatestByTask(taskID uint, createdBy string) (*database.TaskConversation, error)

	// New methods: query by status
	ListByStatus(status database.ConversationStatus) ([]database.TaskConversation, error)
	// Get pending conversations with complete association information
	GetPendingConversationsWithDetails() ([]database.TaskConversation, error)
	// Check if task has pending or running conversations
	HasPendingOrRunningConversations(taskID uint, createdBy string) (bool, error)
	// Update conversation commit hash
	UpdateCommitHash(id uint, commitHash string) error
}

// TaskExecutionLogRepository defines task execution log repository interface
type TaskExecutionLogRepository interface {
	Create(log *database.TaskExecutionLog) error
	GetByID(id uint) (*database.TaskExecutionLog, error)
	GetByConversationID(conversationID uint) (*database.TaskExecutionLog, error)
	Update(log *database.TaskExecutionLog) error
	AppendLog(id uint, logContent string) error
	UpdateMetadata(id uint, updates map[string]interface{}) error
	DeleteByConversationID(conversationID uint) error
}

// TaskConversationResultRepository defines task conversation result repository interface
type TaskConversationResultRepository interface {
	// Basic CRUD operations
	Create(result *database.TaskConversationResult) error
	GetByID(id uint) (*database.TaskConversationResult, error)
	GetByConversationID(conversationID uint) (*database.TaskConversationResult, error)
	Update(result *database.TaskConversationResult) error
	Delete(id uint) error

	// Query operations
	ListByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error)
	ListByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error)

	// Statistics operations
	GetSuccessRate(taskID uint) (float64, error)
	GetTotalCost(taskID uint) (float64, error)
	GetAverageDuration(taskID uint) (float64, error)

	// Business operations
	ExistsByConversationID(conversationID uint) (bool, error)
	DeleteByConversationID(conversationID uint) error
}

// SystemConfigRepository defines system configuration repository interface
type SystemConfigRepository interface {
	// Basic CRUD operations
	Create(config *database.SystemConfig) error
	GetByKey(key string) (*database.SystemConfig, error)
	ListAll() ([]database.SystemConfig, error)
	Update(config *database.SystemConfig) error

	// Business operations
	GetValue(key string) (string, error)
	SetValue(key, value string) error
	SetValueWithCategory(key, value, description, category string, isEditable bool) error
	InitializeDefaultConfigs() error
}
