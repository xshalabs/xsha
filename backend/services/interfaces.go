package services

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"
)

type AuthService interface {
	Login(username, password, clientIP, userAgent string) (bool, string, error)
	Logout(token, username string) error
	IsTokenBlacklisted(token string) (bool, error)
	CleanExpiredTokens() error
}

type LoginLogService interface {
	GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error)
	CleanOldLogs(days int) error
}

type GitCredentialService interface {
	CreateCredential(name, description, credType, username, createdBy string, secretData map[string]string) (*database.GitCredential, error)
	GetCredential(id uint, createdBy string) (*database.GitCredential, error)
	ListCredentials(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	UpdateCredential(id uint, createdBy string, updates map[string]interface{}, secretData map[string]string) error
	DeleteCredential(id uint, createdBy string) error
	ListActiveCredentials(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error)
	DecryptCredentialSecret(credential *database.GitCredential, secretType string) (string, error)
	ValidateCredentialData(credType string, data map[string]string) error
}

type ProjectService interface {
	CreateProject(name, description, repoURL, protocol, createdBy string, credentialID *uint) (*database.Project, error)
	GetProject(id uint, createdBy string) (*database.Project, error)
	ListProjects(createdBy string, name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	ListProjectsWithTaskCount(createdBy string, name string, protocol *database.GitProtocolType, page, pageSize int) (interface{}, int64, error)
	UpdateProject(id uint, createdBy string, updates map[string]interface{}) error
	DeleteProject(id uint, createdBy string) error
	ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint, createdBy string) error
	GetCompatibleCredentials(protocol database.GitProtocolType, createdBy string) ([]database.GitCredential, error)
	FetchRepositoryBranches(repoURL string, credentialID *uint, createdBy string) (*utils.GitAccessResult, error)
	ValidateRepositoryAccess(repoURL string, credentialID *uint, createdBy string) error
}

type AdminOperationLogService interface {
	LogOperation(username, operation, resource, resourceID, description, details string,
		success bool, errorMsg, ip, userAgent, method, path string) error
	LogCreate(username, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogUpdate(username, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogDelete(username, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogRead(username, resource, resourceID, description, ip, userAgent, path string) error
	LogLogin(username, ip, userAgent string, success bool, errorMsg string) error
	LogLogout(username, ip, userAgent string, success bool, errorMsg string) error
	GetLogs(username string, operation *database.AdminOperationType, resource string,
		success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error)
	GetLog(id uint) (*database.AdminOperationLog, error)
	GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	CleanOldLogs(days int) error
}

type DevEnvironmentService interface {
	CreateEnvironment(name, description, envType, createdBy string, cpuLimit float64, memoryLimit int64, envVars map[string]string) (*database.DevEnvironment, error)
	GetEnvironment(id uint, createdBy string) (*database.DevEnvironment, error)
	ListEnvironments(createdBy string, envType *string, name *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	UpdateEnvironment(id uint, createdBy string, updates map[string]interface{}) error
	DeleteEnvironment(id uint, createdBy string) error
	ValidateEnvVars(envVars map[string]string) error
	GetEnvironmentVars(id uint, createdBy string) (map[string]string, error)
	UpdateEnvironmentVars(id uint, createdBy string, envVars map[string]string) error
	ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error
	GetAvailableEnvironmentTypes() ([]map[string]interface{}, error)
}

type TaskService interface {
	CreateTask(title, startBranch, createdBy string, projectID uint, devEnvironmentID *uint) (*database.Task, error)
	GetTask(id uint, createdBy string) (*database.Task, error)
	ListTasks(projectID *uint, createdBy string, status *database.TaskStatus, title *string, branch *string, devEnvID *uint, page, pageSize int) ([]database.Task, int64, error)
	UpdateTask(id uint, createdBy string, updates map[string]interface{}) error
	UpdateTaskStatus(id uint, createdBy string, status database.TaskStatus) error
	UpdateTaskStatusBatch(taskIDs []uint, createdBy string, status database.TaskStatus) ([]uint, []uint, error)
	DeleteTask(id uint, createdBy string) error
	ValidateTaskData(title, startBranch string, projectID uint, createdBy string) error
	GetTaskGitDiff(task *database.Task, includeContent bool) (*utils.GitDiffSummary, error)
	GetTaskGitDiffFile(task *database.Task, filePath string) (string, error)
	PushTaskBranch(id uint, createdBy string, forcePush bool) (string, error)
}

type TaskConversationService interface {
	CreateConversation(taskID uint, content, createdBy string) (*database.TaskConversation, error)
	GetConversation(id uint, createdBy string) (*database.TaskConversation, error)
	ListConversations(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error)
	UpdateConversation(id uint, createdBy string, updates map[string]interface{}) error
	DeleteConversation(id uint, createdBy string) error
	GetLatestConversation(taskID uint, createdBy string) (*database.TaskConversation, error)
	GetConversationGitDiff(conversationID uint, createdBy string, includeContent bool) (*utils.GitDiffSummary, error)
	GetConversationGitDiffFile(conversationID uint, createdBy string, filePath string) (string, error)
	ValidateConversationData(taskID uint, content string, createdBy string) error
}

type TaskConversationResultService interface {
	CreateResult(conversationID uint, resultData map[string]interface{}) (*database.TaskConversationResult, error)
	GetResult(id uint) (*database.TaskConversationResult, error)
	GetResultByConversationID(conversationID uint) (*database.TaskConversationResult, error)
	UpdateResult(id uint, updates map[string]interface{}) error
	DeleteResult(id uint) error
	ListResultsByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error)
	ListResultsByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error)
	GetTaskStats(taskID uint) (map[string]interface{}, error)
	GetProjectStats(projectID uint) (map[string]interface{}, error)
	ExistsForConversation(conversationID uint) (bool, error)
	ValidateResultData(resultData map[string]interface{}) error
}

type AITaskExecutorService interface {
	ProcessPendingConversations() error
	GetExecutionLog(conversationID uint) (*database.TaskExecutionLog, error)
	CancelExecution(conversationID uint, createdBy string) error
	RetryExecution(conversationID uint, createdBy string) error
	GetExecutionStatus() map[string]interface{}
	CleanupWorkspaceOnFailure(taskID uint, workspacePath string) error
	CleanupWorkspaceOnCancel(taskID uint, workspacePath string) error
}

type ConfigUpdateItem struct {
	ConfigKey   string
	ConfigValue string
	Description string
	Category    string
	IsEditable  *bool
}

type SystemConfigService interface {
	ListAllConfigs() ([]database.SystemConfig, error)
	BatchUpdateConfigs(configs []ConfigUpdateItem) error
	GetValue(key string) (string, error)
	SetValue(key, value string) error
	InitializeDefaultConfigs() error
	ValidateConfigData(key, value, category string) error
	GetGitProxyConfig() (*utils.GitProxyConfig, error)
}
