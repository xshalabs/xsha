package events

import (
	"time"
	"xsha-backend/database"
)

// 任务事件类型常量
const (
	EventTypeTaskCreated         = "task.created"
	EventTypeTaskUpdated         = "task.updated"
	EventTypeTaskStatusChanged   = "task.status.changed"
	EventTypeTaskDeleted         = "task.deleted"
	EventTypeTaskWorkspaceReady  = "task.workspace.ready"
	EventTypeTaskBranchCreated   = "task.branch.created"
	EventTypeTaskBranchPushed    = "task.branch.pushed"
	EventTypeTaskCompleted       = "task.completed"
	EventTypeTaskFailed          = "task.failed"
	EventTypeTaskCancelled       = "task.cancelled"
)

// TaskCreatedEvent 任务创建事件
type TaskCreatedEvent struct {
	BaseEvent
	TaskID           uint   `json:"task_id"`
	ProjectID        uint   `json:"project_id"`
	DevEnvironmentID *uint  `json:"dev_environment_id"`
	AdminID          *uint  `json:"admin_id"`
	Title            string `json:"title"`
	StartBranch      string `json:"start_branch"`
	WorkBranch       string `json:"work_branch"`
	CreatedBy        string `json:"created_by"`
}

// NewTaskCreatedEvent 创建任务创建事件
func NewTaskCreatedEvent(task *database.Task) *TaskCreatedEvent {
	event := &TaskCreatedEvent{
		BaseEvent:        NewBaseEvent(EventTypeTaskCreated),
		TaskID:           task.ID,
		ProjectID:        task.ProjectID,
		DevEnvironmentID: task.DevEnvironmentID,
		AdminID:          task.AdminID,
		Title:            task.Title,
		StartBranch:      task.StartBranch,
		WorkBranch:       task.WorkBranch,
		CreatedBy:        task.CreatedBy,
	}
	event.Payload = event
	return event
}

// TaskUpdatedEvent 任务更新事件
type TaskUpdatedEvent struct {
	BaseEvent
	TaskID      uint                   `json:"task_id"`
	ProjectID   uint                   `json:"project_id"`
	Changes     map[string]interface{} `json:"changes"`
	UpdatedBy   string                 `json:"updated_by"`
	UpdatedFields []string             `json:"updated_fields"`
}

// NewTaskUpdatedEvent 创建任务更新事件
func NewTaskUpdatedEvent(taskID, projectID uint, changes map[string]interface{}, updatedBy string) *TaskUpdatedEvent {
	fields := make([]string, 0, len(changes))
	for field := range changes {
		fields = append(fields, field)
	}
	
	event := &TaskUpdatedEvent{
		BaseEvent:     NewBaseEvent(EventTypeTaskUpdated),
		TaskID:        taskID,
		ProjectID:     projectID,
		Changes:       changes,
		UpdatedBy:     updatedBy,
		UpdatedFields: fields,
	}
	event.Payload = event
	return event
}

// TaskStatusChangedEvent 任务状态变更事件
type TaskStatusChangedEvent struct {
	BaseEvent
	TaskID      uint                  `json:"task_id"`
	ProjectID   uint                  `json:"project_id"`
	OldStatus   database.TaskStatus   `json:"old_status"`
	NewStatus   database.TaskStatus   `json:"new_status"`
	ChangedBy   string                `json:"changed_by"`
	Reason      string                `json:"reason"`
	Context     map[string]interface{} `json:"context"`
}

// NewTaskStatusChangedEvent 创建任务状态变更事件
func NewTaskStatusChangedEvent(taskID, projectID uint, oldStatus, newStatus database.TaskStatus, changedBy, reason string) *TaskStatusChangedEvent {
	event := &TaskStatusChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeTaskStatusChanged),
		TaskID:    taskID,
		ProjectID: projectID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ChangedBy: changedBy,
		Reason:    reason,
		Context:   make(map[string]interface{}),
	}
	event.Payload = event
	return event
}

// SetContext 设置上下文信息
func (e *TaskStatusChangedEvent) SetContext(key string, value interface{}) {
	e.Context[key] = value
}

// TaskDeletedEvent 任务删除事件
type TaskDeletedEvent struct {
	BaseEvent
	TaskID      uint   `json:"task_id"`
	ProjectID   uint   `json:"project_id"`
	Title       string `json:"title"`
	WorkBranch  string `json:"work_branch"`
	DeletedBy   string `json:"deleted_by"`
	Reason      string `json:"reason"`
}

// NewTaskDeletedEvent 创建任务删除事件
func NewTaskDeletedEvent(task *database.Task, deletedBy, reason string) *TaskDeletedEvent {
	event := &TaskDeletedEvent{
		BaseEvent:  NewBaseEvent(EventTypeTaskDeleted),
		TaskID:     task.ID,
		ProjectID:  task.ProjectID,
		Title:      task.Title,
		WorkBranch: task.WorkBranch,
		DeletedBy:  deletedBy,
		Reason:     reason,
	}
	event.Payload = event
	return event
}

// TaskWorkspaceReadyEvent 任务工作空间就绪事件
type TaskWorkspaceReadyEvent struct {
	BaseEvent
	TaskID        uint   `json:"task_id"`
	ProjectID     uint   `json:"project_id"`
	WorkspacePath string `json:"workspace_path"`
	GitRepoURL    string `json:"git_repo_url"`
	StartBranch   string `json:"start_branch"`
	WorkBranch    string `json:"work_branch"`
}

// NewTaskWorkspaceReadyEvent 创建任务工作空间就绪事件
func NewTaskWorkspaceReadyEvent(taskID, projectID uint, workspacePath, gitRepoURL, startBranch, workBranch string) *TaskWorkspaceReadyEvent {
	event := &TaskWorkspaceReadyEvent{
		BaseEvent:     NewBaseEvent(EventTypeTaskWorkspaceReady),
		TaskID:        taskID,
		ProjectID:     projectID,
		WorkspacePath: workspacePath,
		GitRepoURL:    gitRepoURL,
		StartBranch:   startBranch,
		WorkBranch:    workBranch,
	}
	event.Payload = event
	return event
}

// TaskBranchCreatedEvent 任务分支创建事件
type TaskBranchCreatedEvent struct {
	BaseEvent
	TaskID      uint   `json:"task_id"`
	ProjectID   uint   `json:"project_id"`
	BranchName  string `json:"branch_name"`
	BaseBranch  string `json:"base_branch"`
	CommitHash  string `json:"commit_hash"`
}

// NewTaskBranchCreatedEvent 创建任务分支创建事件
func NewTaskBranchCreatedEvent(taskID, projectID uint, branchName, baseBranch, commitHash string) *TaskBranchCreatedEvent {
	event := &TaskBranchCreatedEvent{
		BaseEvent:  NewBaseEvent(EventTypeTaskBranchCreated),
		TaskID:     taskID,
		ProjectID:  projectID,
		BranchName: branchName,
		BaseBranch: baseBranch,
		CommitHash: commitHash,
	}
	event.Payload = event
	return event
}

// TaskBranchPushedEvent 任务分支推送事件
type TaskBranchPushedEvent struct {
	BaseEvent
	TaskID      uint     `json:"task_id"`
	ProjectID   uint     `json:"project_id"`
	BranchName  string   `json:"branch_name"`
	CommitHash  string   `json:"commit_hash"`
	CommitCount int      `json:"commit_count"`
	ForcePush   bool     `json:"force_push"`
	FilesChanged []string `json:"files_changed"`
	PushedBy    string   `json:"pushed_by"`
}

// NewTaskBranchPushedEvent 创建任务分支推送事件
func NewTaskBranchPushedEvent(taskID, projectID uint, branchName, commitHash string, commitCount int, forcePush bool, filesChanged []string, pushedBy string) *TaskBranchPushedEvent {
	event := &TaskBranchPushedEvent{
		BaseEvent:    NewBaseEvent(EventTypeTaskBranchPushed),
		TaskID:       taskID,
		ProjectID:    projectID,
		BranchName:   branchName,
		CommitHash:   commitHash,
		CommitCount:  commitCount,
		ForcePush:    forcePush,
		FilesChanged: filesChanged,
		PushedBy:     pushedBy,
	}
	event.Payload = event
	return event
}

// TaskCompletedEvent 任务完成事件
type TaskCompletedEvent struct {
	BaseEvent
	TaskID          uint          `json:"task_id"`
	ProjectID       uint          `json:"project_id"`
	Title           string        `json:"title"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	ConversationCount int         `json:"conversation_count"`
	FilesChanged    []string      `json:"files_changed"`
	CommitsCount    int           `json:"commits_count"`
	CompletedBy     string        `json:"completed_by"`
	Result          interface{}   `json:"result"`
}

// NewTaskCompletedEvent 创建任务完成事件
func NewTaskCompletedEvent(task *database.Task, duration time.Duration, conversationCount, commitsCount int, filesChanged []string, completedBy string, result interface{}) *TaskCompletedEvent {
	now := time.Now()
	event := &TaskCompletedEvent{
		BaseEvent:         NewBaseEvent(EventTypeTaskCompleted),
		TaskID:            task.ID,
		ProjectID:         task.ProjectID,
		Title:             task.Title,
		StartTime:         task.CreatedAt,
		EndTime:           now,
		Duration:          duration,
		ConversationCount: conversationCount,
		FilesChanged:      filesChanged,
		CommitsCount:      commitsCount,
		CompletedBy:       completedBy,
		Result:            result,
	}
	event.Payload = event
	return event
}

// TaskFailedEvent 任务失败事件
type TaskFailedEvent struct {
	BaseEvent
	TaskID      uint                   `json:"task_id"`
	ProjectID   uint                   `json:"project_id"`
	Title       string                 `json:"title"`
	ErrorType   string                 `json:"error_type"`
	ErrorMessage string                `json:"error_message"`
	FailedAt    time.Time              `json:"failed_at"`
	Context     map[string]interface{} `json:"context"`
	FailedBy    string                 `json:"failed_by"`
}

// NewTaskFailedEvent 创建任务失败事件
func NewTaskFailedEvent(task *database.Task, errorType, errorMessage, failedBy string) *TaskFailedEvent {
	event := &TaskFailedEvent{
		BaseEvent:    NewBaseEvent(EventTypeTaskFailed),
		TaskID:       task.ID,
		ProjectID:    task.ProjectID,
		Title:        task.Title,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		FailedAt:     time.Now(),
		Context:      make(map[string]interface{}),
		FailedBy:     failedBy,
	}
	event.Payload = event
	return event
}

// SetContext 设置上下文信息
func (e *TaskFailedEvent) SetContext(key string, value interface{}) {
	e.Context[key] = value
}

// TaskCancelledEvent 任务取消事件
type TaskCancelledEvent struct {
	BaseEvent
	TaskID       uint      `json:"task_id"`
	ProjectID    uint      `json:"project_id"`
	Title        string    `json:"title"`
	CancelledBy  string    `json:"cancelled_by"`
	CancelledAt  time.Time `json:"cancelled_at"`
	Reason       string    `json:"reason"`
	WasRunning   bool      `json:"was_running"`
	Progress     float64   `json:"progress"`
}

// NewTaskCancelledEvent 创建任务取消事件
func NewTaskCancelledEvent(task *database.Task, cancelledBy, reason string, wasRunning bool, progress float64) *TaskCancelledEvent {
	event := &TaskCancelledEvent{
		BaseEvent:   NewBaseEvent(EventTypeTaskCancelled),
		TaskID:      task.ID,
		ProjectID:   task.ProjectID,
		Title:       task.Title,
		CancelledBy: cancelledBy,
		CancelledAt: time.Now(),
		Reason:      reason,
		WasRunning:  wasRunning,
		Progress:    progress,
	}
	event.Payload = event
	return event
}