package handlers

import (
	"fmt"
	"log"
	"time"
	"xsha-backend/events"
	"xsha-backend/services"
	"xsha-backend/utils"
)

// TaskEventHandlers 任务事件处理器
type TaskEventHandlers struct {
	auditService        services.AdminOperationLogService
	workspaceManager    *utils.WorkspaceManager
	taskService         services.TaskService
	projectService      services.ProjectService
	gitCredentialService services.GitCredentialService
	systemConfigService services.SystemConfigService
}

// NewTaskEventHandlers 创建任务事件处理器
func NewTaskEventHandlers(
	auditService services.AdminOperationLogService,
	workspaceManager *utils.WorkspaceManager,
	taskService services.TaskService,
	projectService services.ProjectService,
	gitCredService services.GitCredentialService,
	systemConfigService services.SystemConfigService,
) *TaskEventHandlers {
	return &TaskEventHandlers{
		auditService:         auditService,
		workspaceManager:     workspaceManager,
		taskService:          taskService,
		projectService:       projectService,
		gitCredentialService: gitCredService,
		systemConfigService:  systemConfigService,
	}
}

// HandleTaskCreated 处理任务创建事件
func (h *TaskEventHandlers) HandleTaskCreated(event events.Event) error {
	taskEvent, ok := event.(*events.TaskCreatedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task created handler")
	}

	log.Printf("Processing task created event: Task ID %d, Title: %s", taskEvent.TaskID, taskEvent.Title)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogCreate(
			taskEvent.CreatedBy,
			taskEvent.AdminID,
			"task",
			fmt.Sprintf("%d", taskEvent.TaskID),
			fmt.Sprintf("Created task: %s in project %d", taskEvent.Title, taskEvent.ProjectID),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log task creation audit: %v", err)
		}
	}()

	// 2. 准备工作空间（异步）
	go func() {
		if err := h.prepareTaskWorkspace(taskEvent); err != nil {
			log.Printf("Failed to prepare workspace for task %d: %v", taskEvent.TaskID, err)
		}
	}()

	// 3. 初始化任务统计
	go func() {
		h.initializeTaskStats(taskEvent)
	}()

	return nil
}

// HandleTaskStatusChanged 处理任务状态变更事件
func (h *TaskEventHandlers) HandleTaskStatusChanged(event events.Event) error {
	statusEvent, ok := event.(*events.TaskStatusChangedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task status changed handler")
	}

	log.Printf("Processing task status changed event: Task ID %d, %s -> %s", 
		statusEvent.TaskID, statusEvent.OldStatus, statusEvent.NewStatus)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogUpdate(
			statusEvent.ChangedBy,
			nil, // adminID 在context中获取
			"task",
			fmt.Sprintf("%d", statusEvent.TaskID),
			fmt.Sprintf("Changed task status from %s to %s", statusEvent.OldStatus, statusEvent.NewStatus),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log task status change audit: %v", err)
		}
	}()

	// 2. 根据状态变更执行相应操作
	switch statusEvent.NewStatus {
	case "completed":
		go h.handleTaskCompletion(statusEvent)
	case "failed":
		go h.handleTaskFailure(statusEvent)
	case "cancelled":
		go h.handleTaskCancellation(statusEvent)
	case "in_progress":
		go h.handleTaskStart(statusEvent)
	}

	return nil
}

// HandleTaskCompleted 处理任务完成事件
func (h *TaskEventHandlers) HandleTaskCompleted(event events.Event) error {
	completedEvent, ok := event.(*events.TaskCompletedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task completed handler")
	}

	log.Printf("Processing task completed event: Task ID %d, Duration: %v", 
		completedEvent.TaskID, completedEvent.Duration)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogUpdate(
			completedEvent.CompletedBy,
			nil,
			"task",
			fmt.Sprintf("%d", completedEvent.TaskID),
			fmt.Sprintf("Task completed successfully in %v", completedEvent.Duration),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log task completion audit: %v", err)
		}
	}()

	// 2. 清理工作空间（延迟清理，给用户时间查看结果）
	go func() {
		time.Sleep(time.Hour * 24) // 24小时后清理
		if err := h.cleanupTaskWorkspace(completedEvent.TaskID); err != nil {
			log.Printf("Failed to cleanup workspace for completed task %d: %v", completedEvent.TaskID, err)
		}
	}()

	// 3. 更新任务统计
	go func() {
		h.updateTaskCompletionStats(completedEvent)
	}()

	// 4. 生成任务报告
	go func() {
		h.generateTaskReport(completedEvent)
	}()

	return nil
}

// HandleTaskFailed 处理任务失败事件
func (h *TaskEventHandlers) HandleTaskFailed(event events.Event) error {
	failedEvent, ok := event.(*events.TaskFailedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task failed handler")
	}

	log.Printf("Processing task failed event: Task ID %d, Error: %s", 
		failedEvent.TaskID, failedEvent.ErrorMessage)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogUpdate(
			failedEvent.FailedBy,
			nil,
			"task",
			fmt.Sprintf("%d", failedEvent.TaskID),
			fmt.Sprintf("Task failed: %s", failedEvent.ErrorMessage),
			"", "", "", false, failedEvent.ErrorMessage,
		); err != nil {
			log.Printf("Failed to log task failure audit: %v", err)
		}
	}()

	// 2. 保留工作空间用于调试
	go func() {
		h.preserveFailedTaskWorkspace(failedEvent)
	}()

	// 3. 更新错误统计
	go func() {
		h.updateTaskErrorStats(failedEvent)
	}()

	// 4. 发送失败通知
	go func() {
		h.sendTaskFailureNotification(failedEvent)
	}()

	return nil
}

// HandleTaskDeleted 处理任务删除事件
func (h *TaskEventHandlers) HandleTaskDeleted(event events.Event) error {
	deletedEvent, ok := event.(*events.TaskDeletedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task deleted handler")
	}

	log.Printf("Processing task deleted event: Task ID %d, Title: %s", 
		deletedEvent.TaskID, deletedEvent.Title)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogDelete(
			deletedEvent.DeletedBy,
			nil,
			"task",
			fmt.Sprintf("%d", deletedEvent.TaskID),
			fmt.Sprintf("Deleted task: %s (Reason: %s)", deletedEvent.Title, deletedEvent.Reason),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log task deletion audit: %v", err)
		}
	}()

	// 2. 立即清理工作空间和分支
	go func() {
		if err := h.cleanupTaskWorkspace(deletedEvent.TaskID); err != nil {
			log.Printf("Failed to cleanup workspace for deleted task %d: %v", deletedEvent.TaskID, err)
		}
		
		if err := h.cleanupTaskBranch(deletedEvent); err != nil {
			log.Printf("Failed to cleanup branch for deleted task %d: %v", deletedEvent.TaskID, err)
		}
	}()

	// 3. 清理相关数据
	go func() {
		h.cleanupTaskRelatedData(deletedEvent.TaskID)
	}()

	return nil
}

// HandleTaskWorkspaceReady 处理任务工作空间就绪事件
func (h *TaskEventHandlers) HandleTaskWorkspaceReady(event events.Event) error {
	workspaceEvent, ok := event.(*events.TaskWorkspaceReadyEvent)
	if !ok {
		return fmt.Errorf("invalid event type for task workspace ready handler")
	}

	log.Printf("Processing task workspace ready event: Task ID %d, Workspace: %s", 
		workspaceEvent.TaskID, workspaceEvent.WorkspacePath)

	// 1. 创建工作分支
	go func() {
		if err := h.createTaskBranch(workspaceEvent); err != nil {
			log.Printf("Failed to create task branch for task %d: %v", workspaceEvent.TaskID, err)
		}
	}()

	// 2. 更新任务状态
	go func() {
		if err := h.taskService.UpdateTaskStatus(workspaceEvent.TaskID, "ready"); err != nil {
			log.Printf("Failed to update task status to ready for task %d: %v", workspaceEvent.TaskID, err)
		}
	}()

	return nil
}

// prepareTaskWorkspace 准备任务工作空间
func (h *TaskEventHandlers) prepareTaskWorkspace(taskEvent *events.TaskCreatedEvent) error {
	// 获取项目信息
	project, err := h.projectService.GetProject(taskEvent.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %v", err)
	}

	// 创建工作空间
	workspacePath, err := h.workspaceManager.GetOrCreateTaskWorkspace(taskEvent.TaskID, "")
	if err != nil {
		return fmt.Errorf("failed to create workspace: %v", err)
	}

	// 克隆仓库
	var credential *utils.GitCredentialInfo
	if project.CredentialID != nil {
		_, err := h.gitCredentialService.GetCredential(*project.CredentialID)
		if err != nil {
			return fmt.Errorf("failed to get git credential: %v", err)
		}
		// 这里需要转换credential格式，暂时使用nil
		credential = nil
	}

	if err := h.workspaceManager.CloneRepositoryWithConfig(workspacePath, project.RepoURL, taskEvent.StartBranch, credential, true, nil); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	log.Printf("Workspace prepared for task %d at %s", taskEvent.TaskID, workspacePath)
	return nil
}

// 其他辅助方法的实现...
func (h *TaskEventHandlers) handleTaskCompletion(statusEvent *events.TaskStatusChangedEvent) {
	log.Printf("Handling task completion for task %d", statusEvent.TaskID)
	// 实现任务完成后续处理逻辑
}

func (h *TaskEventHandlers) handleTaskFailure(statusEvent *events.TaskStatusChangedEvent) {
	log.Printf("Handling task failure for task %d", statusEvent.TaskID)
	// 实现任务失败后续处理逻辑
}

func (h *TaskEventHandlers) handleTaskCancellation(statusEvent *events.TaskStatusChangedEvent) {
	log.Printf("Handling task cancellation for task %d", statusEvent.TaskID)
	// 实现任务取消后续处理逻辑
}

func (h *TaskEventHandlers) handleTaskStart(statusEvent *events.TaskStatusChangedEvent) {
	log.Printf("Handling task start for task %d", statusEvent.TaskID)
	// 实现任务开始后续处理逻辑
}

func (h *TaskEventHandlers) cleanupTaskWorkspace(taskID uint) error {
	// 需要通过taskID获取workspace路径，这里先简化处理
	return nil // h.workspaceManager.CleanupTaskWorkspace(workspacePath)
}

func (h *TaskEventHandlers) cleanupTaskBranch(deletedEvent *events.TaskDeletedEvent) error {
	log.Printf("Cleaning up branch %s for deleted task %d", deletedEvent.WorkBranch, deletedEvent.TaskID)
	// 实现分支清理逻辑
	return nil
}

func (h *TaskEventHandlers) cleanupTaskRelatedData(taskID uint) {
	log.Printf("Cleaning up related data for task %d", taskID)
	// 实现相关数据清理逻辑
}

func (h *TaskEventHandlers) createTaskBranch(workspaceEvent *events.TaskWorkspaceReadyEvent) error {
	log.Printf("Creating branch %s for task %d", workspaceEvent.WorkBranch, workspaceEvent.TaskID)
	// 实现分支创建逻辑
	return nil
}

func (h *TaskEventHandlers) initializeTaskStats(taskEvent *events.TaskCreatedEvent) {
	log.Printf("Initializing stats for task %d", taskEvent.TaskID)
	// 实现任务统计初始化
}

func (h *TaskEventHandlers) updateTaskCompletionStats(completedEvent *events.TaskCompletedEvent) {
	log.Printf("Updating completion stats for task %d", completedEvent.TaskID)
	// 实现任务完成统计更新
}

func (h *TaskEventHandlers) updateTaskErrorStats(failedEvent *events.TaskFailedEvent) {
	log.Printf("Updating error stats for task %d", failedEvent.TaskID)
	// 实现任务错误统计更新
}

func (h *TaskEventHandlers) generateTaskReport(completedEvent *events.TaskCompletedEvent) {
	log.Printf("Generating report for completed task %d", completedEvent.TaskID)
	// 实现任务报告生成
}

func (h *TaskEventHandlers) preserveFailedTaskWorkspace(failedEvent *events.TaskFailedEvent) {
	log.Printf("Preserving workspace for failed task %d", failedEvent.TaskID)
	// 实现失败任务工作空间保留逻辑
}

func (h *TaskEventHandlers) sendTaskFailureNotification(failedEvent *events.TaskFailedEvent) {
	log.Printf("Sending failure notification for task %d", failedEvent.TaskID)
	// 实现失败通知发送逻辑
}