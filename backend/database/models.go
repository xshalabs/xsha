package database

import (
	"time"

	"gorm.io/gorm"
)

type TokenBlacklist struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	Username  string         `gorm:"not null" json:"username"`
	Reason    string         `gorm:"default:'logout'" json:"reason"`
}

type LoginLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"not null;index" json:"username"`
	Success   bool           `gorm:"not null;index" json:"success"`
	IP        string         `gorm:"not null" json:"ip"`
	UserAgent string         `gorm:"type:text" json:"user_agent"`
	Reason    string         `gorm:"default:''" json:"reason"`
	LoginTime time.Time      `gorm:"not null;index" json:"login_time"`
}

type GitCredentialType string

const (
	GitCredentialTypePassword GitCredentialType = "password"
	GitCredentialTypeToken    GitCredentialType = "token"
	GitCredentialTypeSSHKey   GitCredentialType = "ssh_key"
)

type GitCredential struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string            `gorm:"not null" json:"name"`
	Description string            `gorm:"type:text" json:"description"`
	Type        GitCredentialType `gorm:"not null;index" json:"type"`

	Username string `gorm:"default:''" json:"username"`

	PasswordHash string `gorm:"type:text" json:"-"`
	PrivateKey   string `gorm:"type:text" json:"-"`
	PublicKey    string `gorm:"type:text" json:"public_key"`

	CreatedBy string `gorm:"not null;index" json:"created_by"`
}

type GitProtocolType string

const (
	GitProtocolHTTPS GitProtocolType = "https"
	GitProtocolSSH   GitProtocolType = "ssh"
)

type Project struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`

	RepoURL      string          `gorm:"not null" json:"repo_url"`
	Protocol     GitProtocolType `gorm:"not null;index" json:"protocol"`
	CredentialID *uint           `gorm:"index" json:"credential_id"`
	Credential   *GitCredential  `gorm:"foreignKey:CredentialID" json:"credential"`

	CreatedBy string `gorm:"not null;index" json:"created_by"`
}

type AdminOperationType string

const (
	AdminOperationCreate AdminOperationType = "create"
	AdminOperationRead   AdminOperationType = "read"
	AdminOperationUpdate AdminOperationType = "update"
	AdminOperationDelete AdminOperationType = "delete"
	AdminOperationLogin  AdminOperationType = "login"
	AdminOperationLogout AdminOperationType = "logout"
	AdminOperationExport AdminOperationType = "export"
	AdminOperationImport AdminOperationType = "import"
)

type AdminOperationLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username string `gorm:"not null;index" json:"username"`

	Operation   AdminOperationType `gorm:"not null;index" json:"operation"`
	Resource    string             `gorm:"not null;index" json:"resource"`
	ResourceID  string             `gorm:"default:''" json:"resource_id"`
	Description string             `gorm:"type:text" json:"description"`
	Details     string             `gorm:"type:text" json:"details"`

	Success  bool   `gorm:"not null;index" json:"success"`
	ErrorMsg string `gorm:"type:text" json:"error_msg"`

	IP        string `gorm:"not null" json:"ip"`
	UserAgent string `gorm:"type:text" json:"user_agent"`

	Method string `gorm:"not null" json:"method"`
	Path   string `gorm:"not null" json:"path"`

	OperationTime time.Time `gorm:"not null;index" json:"operation_time"`
}

type DevEnvironment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"not null;index" json:"type"`

	CPULimit    float64 `gorm:"default:1.0" json:"cpu_limit"`
	MemoryLimit int64   `gorm:"default:1024" json:"memory_limit"`

	EnvVars string `gorm:"type:text" json:"env_vars"`

	CreatedBy string `gorm:"not null;index" json:"created_by"`
}

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type ConversationStatus string

const (
	ConversationStatusPending   ConversationStatus = "pending"
	ConversationStatusRunning   ConversationStatus = "running"
	ConversationStatusSuccess   ConversationStatus = "success"
	ConversationStatusFailed    ConversationStatus = "failed"
	ConversationStatusCancelled ConversationStatus = "cancelled"
)

type Task struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"not null" json:"title"`
	StartBranch string `gorm:"default:'main'" json:"start_branch"`
	WorkBranch  string `gorm:"not null;default:''" json:"work_branch"`

	Status         TaskStatus `gorm:"not null;index" json:"status"`
	HasPullRequest bool       `gorm:"default:false" json:"has_pull_request"`

	WorkspacePath string `gorm:"type:text" json:"workspace_path"`

	ProjectID        uint            `gorm:"not null;index" json:"project_id"`
	Project          *Project        `gorm:"foreignKey:ProjectID" json:"project"`
	DevEnvironmentID *uint           `gorm:"index" json:"dev_environment_id"`
	DevEnvironment   *DevEnvironment `gorm:"foreignKey:DevEnvironmentID" json:"dev_environment"`

	CreatedBy string `gorm:"not null;index" json:"created_by"`

	Conversations     []TaskConversation `gorm:"foreignKey:TaskID" json:"conversations"`
	ConversationCount int64              `gorm:"-" json:"conversation_count"`
}

type TaskConversation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TaskID uint  `gorm:"not null;index" json:"task_id"`
	Task   *Task `gorm:"foreignKey:TaskID" json:"task"`

	Content string             `gorm:"type:text;not null" json:"content"`
	Status  ConversationStatus `gorm:"not null;index" json:"status"`

	CommitHash string `gorm:"default:''" json:"commit_hash"`

	CreatedBy string `gorm:"not null;index" json:"created_by"`
}

type TaskExecutionLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ConversationID uint              `gorm:"not null;index" json:"conversation_id"`
	Conversation   *TaskConversation `gorm:"foreignKey:ConversationID" json:"conversation"`

	DockerCommand string `gorm:"type:text" json:"docker_command"`
	ExecutionLogs string `gorm:"type:longtext" json:"execution_logs"`
	ErrorMessage  string `gorm:"type:text" json:"error_message"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type ResultType string

const (
	ResultTypeResult ResultType = "result"
)

type ResultSubtype string

const (
	ResultSubtypeSuccess ResultSubtype = "success"
	ResultSubtypeError   ResultSubtype = "error"
)

type TaskConversationResult struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ConversationID uint              `gorm:"not null;index;unique" json:"conversation_id"`
	Conversation   *TaskConversation `gorm:"foreignKey:ConversationID" json:"conversation"`

	Type    ResultType    `gorm:"not null;index" json:"type"`
	Subtype ResultSubtype `gorm:"not null;index" json:"subtype"`
	IsError bool          `gorm:"not null;index" json:"is_error"`

	DurationMs    int64 `gorm:"not null" json:"duration_ms"`
	DurationApiMs int64 `gorm:"not null" json:"duration_api_ms"`
	NumTurns      int   `gorm:"not null" json:"num_turns"`

	Result string `gorm:"type:text;not null" json:"result"`

	SessionID string `gorm:"not null;index" json:"session_id"`

	TotalCostUsd float64 `gorm:"type:decimal(10,6);not null;default:0" json:"total_cost_usd"`

	Usage string `gorm:"type:text" json:"usage"`
}

type ConfigFormType string

const (
	ConfigFormTypeInput    ConfigFormType = "input"
	ConfigFormTypeTextarea ConfigFormType = "textarea"
	ConfigFormTypeSwitch   ConfigFormType = "switch"
	ConfigFormTypeSelect   ConfigFormType = "select"
	ConfigFormTypeNumber   ConfigFormType = "number"
	ConfigFormTypePassword ConfigFormType = "password"
)

type SystemConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ConfigKey   string         `gorm:"not null;uniqueIndex" json:"config_key"`
	ConfigValue string         `gorm:"type:text;not null" json:"config_value"`
	Description string         `gorm:"type:text" json:"description"`
	Category    string         `gorm:"not null;index;default:'general'" json:"category"`
	FormType    ConfigFormType `gorm:"not null;default:'input'" json:"form_type"`
	IsEditable  bool           `gorm:"not null;default:true" json:"is_editable"`
}
