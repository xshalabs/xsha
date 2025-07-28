package database

import (
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist token blacklist model
type TokenBlacklist struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"` // JWT token
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`        // Token expiration time
	Username  string         `gorm:"not null" json:"username"`          // Username
	Reason    string         `gorm:"default:'logout'" json:"reason"`    // Reason for adding to blacklist
}

// LoginLog 登录日志模型
type LoginLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"not null;index" json:"username"`   // 尝试登录的用户名
	Success   bool           `gorm:"not null;index" json:"success"`    // 登录是否成功
	IP        string         `gorm:"not null" json:"ip"`               // 客户端IP地址
	UserAgent string         `gorm:"type:text" json:"user_agent"`      // 用户代理字符串
	Reason    string         `gorm:"default:''" json:"reason"`         // 失败原因
	LoginTime time.Time      `gorm:"not null;index" json:"login_time"` // 登录时间
}

// GitCredentialType 定义Git凭据类型
type GitCredentialType string

const (
	GitCredentialTypePassword GitCredentialType = "password" // 用户名密码
	GitCredentialTypeToken    GitCredentialType = "token"    // Personal Access Token
	GitCredentialTypeSSHKey   GitCredentialType = "ssh_key"  // SSH Key
)

// GitCredential Git凭据模型
type GitCredential struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string            `gorm:"not null;uniqueIndex:idx_name_user" json:"name"` // 凭据名称
	Description string            `gorm:"type:text" json:"description"`                   // 描述
	Type        GitCredentialType `gorm:"not null;index" json:"type"`                     // 凭据类型

	// 认证信息
	Username string `gorm:"default:''" json:"username"` // 用户名（用于password和token类型）

	// 加密存储的敏感信息
	PasswordHash string `gorm:"type:text" json:"-"`          // 加密的密码/token
	PrivateKey   string `gorm:"type:text" json:"-"`          // 加密的SSH私钥
	PublicKey    string `gorm:"type:text" json:"public_key"` // SSH公钥（不敏感，可显示）

	// 元数据
	LastUsed *time.Time `json:"last_used"`                           // 最后使用时间
	IsActive bool       `gorm:"default:true;index" json:"is_active"` // 是否激活

	// 关联用户
	CreatedBy string `gorm:"not null;index;uniqueIndex:idx_name_user" json:"created_by"` // 创建者用户名
}

// GitProtocolType 定义Git协议类型
type GitProtocolType string

const (
	GitProtocolHTTPS GitProtocolType = "https" // HTTPS协议
	GitProtocolSSH   GitProtocolType = "ssh"   // SSH协议
)

// Project 项目模型
type Project struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string `gorm:"not null;uniqueIndex:idx_project_name_user" json:"name"` // 项目名称
	Description string `gorm:"type:text" json:"description"`                           // 项目描述

	// Git配置
	RepoURL      string          `gorm:"not null" json:"repo_url"`                  // Git仓库地址
	Protocol     GitProtocolType `gorm:"not null;index" json:"protocol"`            // Git协议类型
	CredentialID *uint           `gorm:"index" json:"credential_id"`                // 绑定的凭据ID
	Credential   *GitCredential  `gorm:"foreignKey:CredentialID" json:"credential"` // 关联的凭据

	// 元数据
	IsActive bool       `gorm:"default:true;index" json:"is_active"` // 是否激活
	LastUsed *time.Time `json:"last_used"`                           // 最后使用时间

	// 关联用户
	CreatedBy string `gorm:"not null;index;uniqueIndex:idx_project_name_user" json:"created_by"` // 创建者用户名
}

// AdminOperationType 管理员操作类型
type AdminOperationType string

const (
	AdminOperationCreate AdminOperationType = "create" // 创建
	AdminOperationRead   AdminOperationType = "read"   // 查询
	AdminOperationUpdate AdminOperationType = "update" // 更新
	AdminOperationDelete AdminOperationType = "delete" // 删除
	AdminOperationLogin  AdminOperationType = "login"  // 登录
	AdminOperationLogout AdminOperationType = "logout" // 登出
	AdminOperationExport AdminOperationType = "export" // 导出
	AdminOperationImport AdminOperationType = "import" // 导入
)

// AdminOperationLog 管理员操作日志模型
type AdminOperationLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 操作者信息
	Username string `gorm:"not null;index" json:"username"` // 操作用户名

	// 操作信息
	Operation   AdminOperationType `gorm:"not null;index" json:"operation"` // 操作类型
	Resource    string             `gorm:"not null;index" json:"resource"`  // 操作的资源类型（如：project, credential, user等）
	ResourceID  string             `gorm:"default:''" json:"resource_id"`   // 操作的资源ID
	Description string             `gorm:"type:text" json:"description"`    // 操作描述
	Details     string             `gorm:"type:text" json:"details"`        // 操作详情（JSON格式存储）

	// 操作结果
	Success  bool   `gorm:"not null;index" json:"success"` // 操作是否成功
	ErrorMsg string `gorm:"type:text" json:"error_msg"`    // 错误信息（如果失败）

	// 客户端信息
	IP        string `gorm:"not null" json:"ip"`          // 客户端IP
	UserAgent string `gorm:"type:text" json:"user_agent"` // 用户代理

	// 请求信息
	Method string `gorm:"not null" json:"method"` // HTTP方法
	Path   string `gorm:"not null" json:"path"`   // 请求路径

	// 时间信息
	OperationTime time.Time `gorm:"not null;index" json:"operation_time"` // 操作时间
}

// DevEnvironmentType 开发环境类型
type DevEnvironmentType string

const (
	DevEnvTypeClaude   DevEnvironmentType = "claude_code" // Claude Code环境
	DevEnvTypeGemini   DevEnvironmentType = "gemini_cli"  // Gemini CLI环境
	DevEnvTypeOpenCode DevEnvironmentType = "opencode"    // OpenCode环境
)

// DevEnvironment 开发环境模型
type DevEnvironment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string             `gorm:"not null;uniqueIndex:idx_env_name_user" json:"name"` // 环境名称
	Description string             `gorm:"type:text" json:"description"`                       // 环境描述
	Type        DevEnvironmentType `gorm:"not null;index" json:"type"`                         // 环境类型

	// 资源限制
	CPULimit    float64 `gorm:"default:1.0" json:"cpu_limit"`     // CPU限制 (核心数)
	MemoryLimit int64   `gorm:"default:1024" json:"memory_limit"` // 内存限制 (MB)

	// 环境变量 (JSON格式存储)
	EnvVars string `gorm:"type:text" json:"env_vars"` // 环境变量JSON字符串

	// 关联用户
	CreatedBy string `gorm:"not null;index;uniqueIndex:idx_env_name_user" json:"created_by"` // 创建者用户名
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"        // 待处理
	TaskStatusInProgress TaskStatus = "in_progress" // 进行中
	TaskStatusDone       TaskStatus = "done"        // 已完成
	TaskStatusCancelled  TaskStatus = "cancelled"   // 已取消
)

// ConversationStatus 对话状态
type ConversationStatus string

const (
	ConversationStatusPending   ConversationStatus = "pending"   // 待处理
	ConversationStatusRunning   ConversationStatus = "running"   // 进行中
	ConversationStatusSuccess   ConversationStatus = "success"   // 执行成功
	ConversationStatusFailed    ConversationStatus = "failed"    // 执行失败
	ConversationStatusCancelled ConversationStatus = "cancelled" // 已撤销
)

// Task 任务模型
type Task struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Title       string `gorm:"not null" json:"title"`              // 任务标题
	StartBranch string `gorm:"default:'main'" json:"start_branch"` // 开始开发的分支

	// 状态信息
	Status         TaskStatus `gorm:"not null;index" json:"status"`          // 任务状态
	HasPullRequest bool       `gorm:"default:false" json:"has_pull_request"` // 是否提交PR

	// 工作空间信息
	WorkspacePath string `gorm:"type:text" json:"workspace_path"` // 任务工作空间路径

	// 关联信息
	ProjectID        uint            `gorm:"not null;index" json:"project_id"`                   // 所属项目ID
	Project          *Project        `gorm:"foreignKey:ProjectID" json:"project"`                // 关联项目
	DevEnvironmentID *uint           `gorm:"index" json:"dev_environment_id"`                    // 绑定的开发环境ID
	DevEnvironment   *DevEnvironment `gorm:"foreignKey:DevEnvironmentID" json:"dev_environment"` // 关联开发环境

	// 元数据
	CreatedBy string `gorm:"not null;index" json:"created_by"` // 创建者

	// 关联对话
	Conversations []TaskConversation `gorm:"foreignKey:TaskID" json:"conversations"`
}

// TaskConversation 任务对话模型
type TaskConversation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联信息
	TaskID uint  `gorm:"not null;index" json:"task_id"` // 所属任务ID
	Task   *Task `gorm:"foreignKey:TaskID" json:"task"` // 关联任务

	// 对话信息
	Content string             `gorm:"type:text;not null" json:"content"` // 对话内容
	Status  ConversationStatus `gorm:"not null;index" json:"status"`      // 对话状态

	// 执行结果
	CommitHash string `gorm:"default:''" json:"commit_hash"` // 提交的hash（成功时）

	// 元数据
	CreatedBy string `gorm:"not null;index" json:"created_by"` // 创建者
}

// TaskExecutionLog 任务执行日志模型
type TaskExecutionLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联信息
	ConversationID uint              `gorm:"not null;index" json:"conversation_id"`         // 对话ID
	Conversation   *TaskConversation `gorm:"foreignKey:ConversationID" json:"conversation"` // 关联对话

	// 执行信息
	DockerCommand string `gorm:"type:text" json:"docker_command"`     // 执行的Docker命令
	ExecutionLogs string `gorm:"type:longtext" json:"execution_logs"` // 执行日志（实时追加）
	ErrorMessage  string `gorm:"type:text" json:"error_message"`      // 错误信息

	// 时间信息
	StartedAt   *time.Time `json:"started_at"`   // 开始执行时间
	CompletedAt *time.Time `json:"completed_at"` // 完成时间
}

// ResultType 结果类型
type ResultType string

const (
	ResultTypeResult ResultType = "result" // 结果
)

// ResultSubtype 结果子类型
type ResultSubtype string

const (
	ResultSubtypeSuccess ResultSubtype = "success" // 成功
	ResultSubtypeError   ResultSubtype = "error"   // 错误
)

// TaskConversationResult 任务对话结果模型
type TaskConversationResult struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联信息
	ConversationID uint              `gorm:"not null;index;unique" json:"conversation_id"`  // 对话ID (一对一关系)
	Conversation   *TaskConversation `gorm:"foreignKey:ConversationID" json:"conversation"` // 关联对话

	// 执行结果基本信息
	Type    ResultType    `gorm:"not null;index" json:"type"`     // 结果类型，固定为"result"
	Subtype ResultSubtype `gorm:"not null;index" json:"subtype"`  // 结果子类型，如"success"、"error"
	IsError bool          `gorm:"not null;index" json:"is_error"` // 是否错误

	// 时间和性能信息
	DurationMs    int64 `gorm:"not null" json:"duration_ms"`     // 执行持续时间（毫秒）
	DurationApiMs int64 `gorm:"not null" json:"duration_api_ms"` // API调用持续时间（毫秒）
	NumTurns      int   `gorm:"not null" json:"num_turns"`       // 对话轮数

	// 结果内容
	Result string `gorm:"type:text;not null" json:"result"` // 执行结果内容

	// 会话信息
	SessionID string `gorm:"not null;index" json:"session_id"` // 会话ID

	// 成本信息
	TotalCostUsd float64 `gorm:"type:decimal(10,6);not null;default:0" json:"total_cost_usd"` // 总成本（美元）

	// 使用统计（JSON格式）
	Usage string `gorm:"type:text" json:"usage"` // 使用统计信息JSON字符串
}
