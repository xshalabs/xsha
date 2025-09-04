package services

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"
)

type AuthService interface {
	Login(username, password, clientIP, userAgent string) (bool, string, error)
	Logout(token, username, clientIP, userAgent string) error
	IsTokenBlacklisted(token string) (bool, error)
	CleanExpiredTokens() error
	CheckAdminStatus(username string) (bool, error)
}

type LoginLogService interface {
	GetLogs(username, ip *string, success *bool, startTime, endTime *string, page, pageSize int) ([]database.LoginLog, int64, error)
	CleanOldLogs(days int) error
}

type AdminService interface {
	CreateAdmin(username, password, name, email, createdBy string) (*database.Admin, error)
	CreateAdminWithRole(username, password, name, email string, role database.AdminRole, createdBy string) (*database.Admin, error)
	GetAdmin(id uint) (*database.Admin, error)
	GetAdminByUsername(username string) (*database.Admin, error)
	ListAdmins(search *string, isActive *bool, page, pageSize int) ([]database.Admin, int64, error)
	UpdateAdmin(id uint, updates map[string]interface{}) error
	UpdateAdminRole(id uint, role database.AdminRole) error
	DeleteAdmin(id uint) error
	ChangePassword(id uint, newPassword string) error
	ValidateCredentials(username, password string) (*database.Admin, error)
	InitializeDefaultAdmin() error
	SetAuthService(authService AuthService)
	SetDevEnvironmentService(devEnvService DevEnvironmentService)
	SetGitCredentialService(gitCredService GitCredentialService)
	HasPermission(admin *database.Admin, resource, action string, resourceOwnerID uint) bool
	CanAccessResource(admin *database.Admin, resource string, action string, resourceOwnerID uint) bool
	GetAvailableRoles() []database.AdminRole
}

type GitCredentialService interface {
	CreateCredential(name, description, credType, username string, secretData map[string]string, createdBy string, adminID *uint) (*database.GitCredential, error)
	GetCredential(id uint) (*database.GitCredential, error)
	GetCredentialWithAdmins(id uint) (*database.GitCredential, error)
	ListCredentials(name *string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	ListCredentialsByAdminAccess(adminID uint, name *string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	UpdateCredential(id uint, updates map[string]interface{}, secretData map[string]string) error
	DeleteCredential(id uint) error
	ListActiveCredentials(credType *database.GitCredentialType) ([]database.GitCredential, error)
	ListActiveCredentialsByAdminAccess(adminID uint, credType *database.GitCredentialType) ([]database.GitCredential, error)
	DecryptCredentialSecret(credential *database.GitCredential, secretType string) (string, error)
	ValidateCredentialData(credType string, data map[string]string) error

	// Admin management methods
	AddAdminToCredential(credentialID, adminID uint) error
	RemoveAdminFromCredential(credentialID, adminID uint) error
	GetCredentialAdmins(credentialID uint) ([]database.Admin, error)
	CanAdminAccessCredential(credentialID, adminID uint) (bool, error)
}

type ProjectService interface {
	CreateProject(name, description, systemPrompt, repoURL, protocol string, credentialID *uint, adminID *uint, createdBy string) (*database.Project, error)
	GetProject(id uint) (*database.Project, error)
	ListProjects(name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	ListProjectsWithTaskCount(name string, protocol *database.GitProtocolType, sortBy, sortDirection string, page, pageSize int) (interface{}, int64, error)
	UpdateProject(id uint, updates map[string]interface{}) error
	DeleteProject(id uint) error
	ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint) error
	GetCompatibleCredentials(protocol database.GitProtocolType, admin *database.Admin) ([]database.GitCredential, error)
	FetchRepositoryBranches(repoURL string, credentialID *uint) (*utils.GitAccessResult, error)
}

type AdminOperationLogService interface {
	LogOperation(username string, adminID *uint, operation, resource, resourceID, description, details string,
		success bool, errorMsg, ip, userAgent, method, path string) error
	LogCreate(username string, adminID *uint, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogUpdate(username string, adminID *uint, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogDelete(username string, adminID *uint, resource, resourceID, description, ip, userAgent, path string, success bool, errorMsg string) error
	LogRead(username string, adminID *uint, resource, resourceID, description, ip, userAgent, path string) error
	LogLogin(username string, adminID *uint, ip, userAgent string, success bool, errorMsg string) error
	LogLogout(username string, adminID *uint, ip, userAgent string, success bool, errorMsg string) error
	GetLogs(username string, operation *database.AdminOperationType, resource string,
		success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error)
	GetLog(id uint) (*database.AdminOperationLog, error)
	GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error)
	CleanOldLogs(days int) error
}

type DevEnvironmentService interface {
	CreateEnvironment(name, description, systemPrompt, envType, dockerImage string, cpuLimit float64, memoryLimit int64, envVars map[string]string, adminID uint, createdBy string) (*database.DevEnvironment, error)
	GetEnvironment(id uint) (*database.DevEnvironment, error)
	GetEnvironmentWithAdmins(id uint) (*database.DevEnvironment, error)
	ListEnvironments(name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	ListEnvironmentsByAdminAccess(adminID uint, name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error)
	UpdateEnvironment(id uint, updates map[string]interface{}) error
	DeleteEnvironment(id uint) error
	ValidateEnvVars(envVars map[string]string) error
	UpdateEnvironmentVars(id uint, envVars map[string]string) error
	ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error
	GetAvailableEnvironmentImages() ([]map[string]interface{}, error)

	// Admin management methods
	AddAdminToEnvironment(envID, adminID uint) error
	RemoveAdminFromEnvironment(envID, adminID uint) error
	GetEnvironmentAdmins(envID uint) ([]database.Admin, error)
	CanAdminAccessEnvironment(envID, adminID uint) (bool, error)
}

type TaskService interface {
	CreateTask(title, startBranch string, projectID uint, devEnvironmentID *uint, adminID *uint, createdBy string) (*database.Task, error)
	GetTask(id uint) (*database.Task, error)
	GetTaskByIDAndProject(taskID, projectID uint) (*database.Task, error)
	GetKanbanTasks(projectID uint) (map[database.TaskStatus][]database.Task, error)
	UpdateTask(id uint, updates map[string]interface{}) error
	UpdateTaskStatus(id uint, status database.TaskStatus) error
	UpdateTaskSessionID(id uint, sessionID string) error
	UpdateTaskStatusBatch(taskIDs []uint, status database.TaskStatus, projectID uint) ([]uint, []uint, error)
	DeleteTask(id uint) error
	ValidateTaskData(title, startBranch string, projectID uint) error
	GetTaskGitDiff(task *database.Task, includeContent bool) (*utils.GitDiffSummary, error)
	GetTaskGitDiffFile(task *database.Task, filePath string) (string, error)
	PushTaskBranch(id uint, forcePush bool) (string, error)
}

type TaskConversationService interface {
	CreateConversationWithExecutionTime(taskID uint, content, createdBy string, executionTime *time.Time, envParams string, adminID *uint) (*database.TaskConversation, error)
	CreateConversationWithExecutionTimeAndAttachments(taskID uint, content, createdBy string, executionTime *time.Time, envParams string, attachmentIDs []uint, adminID *uint) (*database.TaskConversation, error)
	GetConversation(id uint) (*database.TaskConversation, error)
	GetConversationWithResult(id uint) (map[string]interface{}, error)
	ListConversations(taskID uint, page, pageSize int) ([]database.TaskConversation, int64, error)
	UpdateConversation(id uint, updates map[string]interface{}) error
	DeleteConversation(id uint) error
	GetLatestConversation(taskID uint) (*database.TaskConversation, error)
	GetConversationGitDiff(conversationID uint, includeContent bool) (*utils.GitDiffSummary, error)
	GetConversationGitDiffFile(conversationID uint, filePath string) (string, error)
	ValidateConversationData(taskID uint, content string) error
}

type TaskConversationResultService interface {
	CreateResult(conversationID uint, resultData map[string]interface{}) (*database.TaskConversationResult, error)
	ValidateResultData(resultData map[string]interface{}) error
}

type AITaskExecutorService interface {
	ProcessPendingConversations() error
	CancelExecution(conversationID uint, createdBy string) error
	RetryExecution(conversationID uint, createdBy string) error
	GetExecutionStatus() map[string]interface{}
	CleanupWorkspaceOnFailure(taskID uint, workspacePath string) error
	CleanupWorkspaceOnCancel(taskID uint, workspacePath string) error
}

type ConfigUpdateItem struct {
	ConfigKey   string
	ConfigValue string
}

type SystemConfigService interface {
	ListAllConfigs() ([]database.SystemConfig, error)
	BatchUpdateConfigs(configs []ConfigUpdateItem) error
	GetValue(key string) (string, error)
	SetValue(key, value string) error
	InitializeDefaultConfigs() error
	ValidateConfigData(key, value, category string) error
	GetGitProxyConfig() (*utils.GitProxyConfig, error)
	GetGitCloneTimeout() (time.Duration, error)
	GetGitSSLVerify() (bool, error)
	GetDockerTimeout() (time.Duration, error)
}

type DashboardService interface {
	GetDashboardStats() (map[string]interface{}, error)
	GetRecentTasks(limit int) ([]database.Task, error)
}

type TaskConversationAttachmentService interface {
	UploadAttachment(fileName, originalName, contentType string, fileSize int64, filePath string, attachmentType database.AttachmentType, adminID uint, createdBy string) (*database.TaskConversationAttachment, error)
	AssociateWithConversation(attachmentID, conversationID uint) error
	GetAttachment(id uint) (*database.TaskConversationAttachment, error)
	GetAttachmentsByConversation(conversationID uint) ([]database.TaskConversationAttachment, error)
	UpdateAttachment(id uint, attachment *database.TaskConversationAttachment) error
	DeleteAttachment(id uint) error
	DeleteAttachmentsByConversation(conversationID uint) error
	ProcessContentWithAttachments(content string, attachments []database.TaskConversationAttachment, conversationID uint) string
	ParseAttachmentTags(content string) []string
	GetAttachmentStorageDir() string
	// Workspace attachment handling methods
	CopyAttachmentsToWorkspace(conversationID uint, workspacePath string) ([]database.TaskConversationAttachment, error)
	ReplaceAttachmentTagsWithPaths(content string, attachments []database.TaskConversationAttachment, workspacePath string) string
	CleanupWorkspaceAttachments(workspacePath string) error
}

type AdminAvatarService interface {
	UploadAvatar(fileName, originalName, contentType string, fileSize int64, filePath string, adminID uint, createdBy string) (*database.AdminAvatar, error)
	GetAvatarByUUID(uuid string) (*database.AdminAvatar, error)
	UpdateAdminAvatarByUUID(avatarUUID string, adminID uint) error
	GetAvatarStorageDir() string
	GenerateAvatarFileName(originalName string) string
	GetFullAvatarPath(relativePath string) string
}
