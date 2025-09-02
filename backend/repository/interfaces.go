package repository

import (
	"time"
	"xsha-backend/database"
)

type TokenBlacklistRepository interface {
	Add(tokenID string, username string, expiresAt time.Time, reason string) error
	IsBlacklisted(tokenID string) (bool, error)
	CleanExpired() error
}

type LoginLogRepository interface {
	Add(username, ip, userAgent, reason string, success bool) error
	GetLogs(username, ip *string, success *bool, startTime, endTime *string, page, pageSize int) ([]database.LoginLog, int64, error)
	CleanOld(days int) error
}

type AdminRepository interface {
	Create(admin *database.Admin) error
	GetByID(id uint) (*database.Admin, error)
	GetByUsername(username string) (*database.Admin, error)
	List(search *string, isActive *bool, page, pageSize int) ([]database.Admin, int64, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	UpdateLastLogin(username, ip string) error
	CountAdmins() (int64, error)
	CountActiveAdminsByRole(role database.AdminRole) (int64, error)
	InitializeDefaultAdmin() error
}

type GitCredentialRepository interface {
	Create(credential *database.GitCredential) error
	GetByID(id uint) (*database.GitCredential, error)
	GetByName(name string) (*database.GitCredential, error)
	List(name *string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	Update(credential *database.GitCredential) error
	Delete(id uint) error
}

type ProjectRepository interface {
	Create(project *database.Project) error
	GetByID(id uint) (*database.Project, error)
	GetByName(name string) (*database.Project, error)
	List(name string, protocol *database.GitProtocolType, sortBy, sortDirection string, page, pageSize int) ([]database.Project, int64, error)
	Update(project *database.Project) error
	Delete(id uint) error

	UpdateLastUsed(id uint) error
	GetByCredentialID(credentialID uint) ([]database.Project, error)
	GetTaskCounts(projectIDs []uint) (map[uint]int64, error)
}

type AdminOperationLogRepository interface {
	Add(log *database.AdminOperationLog) error
	GetByID(id uint) (*database.AdminOperationLog, error)

	List(username string, operation *database.AdminOperationType, resource string,
		success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error)

	GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error)

	CleanOld(days int) error
}

type DevEnvironmentRepository interface {
	Create(env *database.DevEnvironment) error
	GetByID(id uint) (*database.DevEnvironment, error)
	GetByIDWithAdmins(id uint) (*database.DevEnvironment, error)
	GetByName(name string) (*database.DevEnvironment, error)
	List(name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	ListByAdminAccess(adminID uint, name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	Update(env *database.DevEnvironment) error
	Delete(id uint) error
	
	// Admin management methods
	AddAdmin(envID, adminID uint) error
	RemoveAdmin(envID, adminID uint) error
	GetAdmins(envID uint) ([]database.Admin, error)
	IsAdminForEnvironment(envID, adminID uint) (bool, error)
}

type TaskRepository interface {
	Create(task *database.Task) error
	GetByID(id uint) (*database.Task, error)
	List(projectID *uint, statuses []database.TaskStatus, title *string, branch *string, devEnvID *uint, sortBy, sortDirection string, page, pageSize int) ([]database.Task, int64, error)
	Update(task *database.Task) error
	Delete(id uint) error

	ListByProject(projectID uint) ([]database.Task, error)
	GetConversationCounts(taskIDs []uint) (map[uint]int64, error)
	GetLatestExecutionTimes(taskIDs []uint) (map[uint]*time.Time, error)
}

type TaskConversationRepository interface {
	Create(conversation *database.TaskConversation) error
	GetByID(id uint) (*database.TaskConversation, error)
	GetWithResult(id uint) (*database.TaskConversation, *database.TaskConversationResult, *database.TaskExecutionLog, error)
	List(taskID uint, page, pageSize int) ([]database.TaskConversation, int64, error)
	Update(conversation *database.TaskConversation) error
	Delete(id uint) error

	ListByTask(taskID uint) ([]database.TaskConversation, error)
	GetLatestByTask(taskID uint) (*database.TaskConversation, error)

	ListByStatus(status database.ConversationStatus) ([]database.TaskConversation, error)
	GetPendingConversationsWithDetails() ([]database.TaskConversation, error)
	HasPendingOrRunningConversations(taskID uint) (bool, error)
	UpdateCommitHash(id uint, commitHash string) error
}

type TaskExecutionLogRepository interface {
	Create(log *database.TaskExecutionLog) error
	GetByID(id uint) (*database.TaskExecutionLog, error)
	GetByConversationID(conversationID uint) (*database.TaskExecutionLog, error)
	Update(log *database.TaskExecutionLog) error
	AppendLog(id uint, logContent string) error
	UpdateMetadata(id uint, updates map[string]interface{}) error
	DeleteByConversationID(conversationID uint) error
}

type TaskConversationResultRepository interface {
	Create(result *database.TaskConversationResult) error
	ExistsByConversationID(conversationID uint) (bool, error)
	DeleteByConversationID(conversationID uint) error
	GetLatestByTaskID(taskID uint) (*database.TaskConversationResult, error)
}

type SystemConfigRepository interface {
	Create(config *database.SystemConfig) error
	GetByKey(key string) (*database.SystemConfig, error)
	ListAll() ([]database.SystemConfig, error)
	Update(config *database.SystemConfig) error

	GetValue(key string) (string, error)
	SetValue(key, value string) error
	SetValueWithCategoryAndSort(key, value, description, category, formType string, isEditable bool, sortOrder int) error
	InitializeDefaultConfigs() error
}

type DashboardRepository interface {
	GetDashboardStats() (map[string]interface{}, error)
	GetRecentTasks(limit int) ([]database.Task, error)
}

type TaskConversationAttachmentRepository interface {
	Create(attachment *database.TaskConversationAttachment) error
	GetByID(id uint) (*database.TaskConversationAttachment, error)
	GetByConversationID(conversationID uint) ([]database.TaskConversationAttachment, error)
	Update(attachment *database.TaskConversationAttachment) error
	Delete(id uint) error
	DeleteByConversationID(conversationID uint) error
}

type AdminAvatarRepository interface {
	Create(avatar *database.AdminAvatar) error
	GetByID(id uint) (*database.AdminAvatar, error)
	GetByUUID(uuid string) (*database.AdminAvatar, error)
	GetByAdminID(adminID uint) (*database.AdminAvatar, error)
	Update(avatar *database.AdminAvatar) error
	Delete(id uint) error
}
