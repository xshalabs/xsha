package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

// TaskHandlers task handler struct
type TaskHandlers struct {
	taskService         services.TaskService
	conversationService services.TaskConversationService
	projectService      services.ProjectService
}

// NewTaskHandlers creates a task handler instance
func NewTaskHandlers(taskService services.TaskService, conversationService services.TaskConversationService, projectService services.ProjectService) *TaskHandlers {
	return &TaskHandlers{
		taskService:         taskService,
		conversationService: conversationService,
		projectService:      projectService,
	}
}

// CreateTaskRequest request structure for creating tasks
type CreateTaskRequest struct {
	Title            string `json:"title" binding:"required"`
	StartBranch      string `json:"start_branch" binding:"required"`
	ProjectID        uint   `json:"project_id" binding:"required"`
	DevEnvironmentID *uint  `json:"dev_environment_id" binding:"required"`
	RequirementDesc  string `json:"requirement_desc" binding:"required"` // Requirement description for creating conversation
	IncludeBranches  bool   `json:"include_branches"`                    // Whether to return project branch information
}

// CreateTaskResponse response structure for creating tasks
type CreateTaskResponse struct {
	Task            *database.Task `json:"task"`
	ProjectBranches []string       `json:"project_branches,omitempty"` // Project branch list
	BranchError     string         `json:"branch_error,omitempty"`     // Error information when getting branches
}

// UpdateTaskRequest request structure for updating tasks
type UpdateTaskRequest struct {
	Title string `json:"title" binding:"required"`
}

// CreateTask creates a task
func (h *TaskHandlers) CreateTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 创建任务
	task, err := h.taskService.CreateTask(req.Title, req.StartBranch, username.(string), req.ProjectID, req.DevEnvironmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	// 如果有需求描述，创建初始对话记录
	if strings.TrimSpace(req.RequirementDesc) != "" {
		_, err := h.conversationService.CreateConversation(
			task.ID,
			req.RequirementDesc,
			username.(string),
		)
		if err != nil {
			// 如果对话创建失败，记录错误但不影响任务创建
			// 可以考虑添加日志记录
		}
	}

	// 构建响应数据
	response := CreateTaskResponse{
		Task: task,
	}

	// 如果请求包含分支信息，获取项目分支
	if req.IncludeBranches {
		if task.Project != nil {
			branchResult, err := h.projectService.FetchRepositoryBranches(
				task.Project.RepoURL,
				task.Project.CredentialID,
				username.(string),
			)
			if err != nil {
				response.BranchError = err.Error()
			} else if branchResult.CanAccess {
				response.ProjectBranches = branchResult.Branches
			} else {
				response.BranchError = branchResult.ErrorMessage
			}
		} else {
			response.BranchError = i18n.T(lang, "task.project_info_incomplete")
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "task.create_success"),
		"data":    response,
	})
}

// GetTask 获取任务详情
func (h *TaskHandlers) GetTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取任务
	task, err := h.taskService.GetTask(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "task.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "task.get_success"),
		"data":    task,
	})
}

// ListTasks 获取任务列表
func (h *TaskHandlers) ListTasks(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var projectID *uint
	if pid := c.Query("project_id"); pid != "" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			pidUint := uint(id)
			projectID = &pidUint
		}
	}

	var status *database.TaskStatus
	if s := c.Query("status"); s != "" {
		taskStatus := database.TaskStatus(s)
		status = &taskStatus
	}

	var title *string
	if t := c.Query("title"); t != "" {
		title = &t
	}

	var branch *string
	if b := c.Query("branch"); b != "" {
		branch = &b
	}

	var devEnvID *uint
	if envID := c.Query("dev_environment_id"); envID != "" {
		if id, err := strconv.ParseUint(envID, 10, 32); err == nil {
			envIDUint := uint(id)
			devEnvID = &envIDUint
		}
	}

	// 获取任务列表
	tasks, total, err := h.taskService.ListTasks(projectID, username.(string), status, title, branch, devEnvID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "task.get_success"),
		"data": gin.H{
			"tasks":     tasks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateTask 更新任务
func (h *TaskHandlers) UpdateTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error()})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	updates["title"] = req.Title

	// 更新任务
	if err := h.taskService.UpdateTask(uint(id), username.(string), updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task.update_success")})
}

// UpdateTaskStatusRequest 更新任务状态请求结构
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateTaskStatus 更新任务状态
func (h *TaskHandlers) UpdateTaskStatus(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	var req UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error()})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 验证状态值
	var status database.TaskStatus
	switch req.Status {
	case "todo":
		status = database.TaskStatusTodo
	case "in_progress":
		status = database.TaskStatusInProgress
	case "done":
		status = database.TaskStatusDone
	case "cancelled":
		status = database.TaskStatusCancelled
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format")})
		return
	}

	// 更新任务状态
	if err := h.taskService.UpdateTaskStatus(uint(id), username.(string), status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task.update_success")})
}

// DeleteTask 删除任务
func (h *TaskHandlers) DeleteTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 删除任务
	if err := h.taskService.DeleteTask(uint(id), username.(string)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task.delete_success")})
}

// BatchUpdateTaskStatusRequest 批量更新任务状态请求结构
type BatchUpdateTaskStatusRequest struct {
	TaskIDs []uint `json:"task_ids" binding:"required,min=1,max=100"`
	Status  string `json:"status" binding:"required"`
}

// BatchUpdateTaskStatusResponse 批量更新任务状态响应结构
type BatchUpdateTaskStatusResponse struct {
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
	SuccessIDs   []uint `json:"success_ids"`
	FailedIDs    []uint `json:"failed_ids"`
}

// BatchUpdateTaskStatus 批量更新任务状态
func (h *TaskHandlers) BatchUpdateTaskStatus(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req BatchUpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error()})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 验证状态值
	var status database.TaskStatus
	switch req.Status {
	case "todo":
		status = database.TaskStatusTodo
	case "in_progress":
		status = database.TaskStatusInProgress
	case "done":
		status = database.TaskStatusDone
	case "cancelled":
		status = database.TaskStatusCancelled
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format")})
		return
	}

	// 批量更新任务状态
	successIDs, failedIDs, err := h.taskService.UpdateTaskStatusBatch(req.TaskIDs, username.(string), status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	response := BatchUpdateTaskStatusResponse{
		SuccessCount: len(successIDs),
		FailedCount:  len(failedIDs),
		SuccessIDs:   successIDs,
		FailedIDs:    failedIDs,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "task.batch_update_success"),
		"data":    response,
	})
}

// GetTaskGitDiff 获取任务Git变动
func (h *TaskHandlers) GetTaskGitDiff(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取任务ID
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	// 获取查询参数
	includeContent := c.DefaultQuery("include_content", "false") == "true"

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取任务详情
	task, err := h.taskService.GetTask(uint(taskID), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "tasks.errors.not_found"),
		})
		return
	}

	// 检查权限
	if task.CreatedBy != username.(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": i18n.T(lang, "common.no_permission"),
		})
		return
	}

	// 检查必要的分支信息
	if task.StartBranch == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "tasks.errors.no_start_branch"),
		})
		return
	}

	if task.WorkBranch == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "tasks.errors.no_work_branch"),
		})
		return
	}

	// 检查工作空间是否存在
	if task.WorkspacePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "tasks.errors.no_workspace"),
		})
		return
	}

	// 获取Git差异
	diff, err := h.taskService.GetTaskGitDiff(task, includeContent)
	if err != nil {
		utils.Error("Failed to get task Git diff", "taskID", taskID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "tasks.errors.git_diff_failed"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": diff,
	})
}

// GetTaskGitDiffFile 获取任务指定文件的Git变动详情
func (h *TaskHandlers) GetTaskGitDiffFile(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取任务ID
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	// 获取文件路径
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.file_path_required"),
		})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取任务详情
	task, err := h.taskService.GetTask(uint(taskID), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "tasks.errors.not_found"),
		})
		return
	}

	// 检查权限
	if task.CreatedBy != username.(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": i18n.T(lang, "common.no_permission"),
		})
		return
	}

	// 获取文件的Git差异内容
	diffContent, err := h.taskService.GetTaskGitDiffFile(task, filePath)
	if err != nil {
		utils.Error("Failed to get task file Git diff", "taskID", taskID, "filePath", filePath, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "tasks.errors.git_diff_file_failed"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"file_path":    filePath,
			"diff_content": diffContent,
		},
	})
}

// PushTaskBranch 推送任务分支到远程仓库
func (h *TaskHandlers) PushTaskBranch(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取任务ID
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 执行推送
	output, err := h.taskService.PushTaskBranch(uint(taskID), username.(string))
	if err != nil {
		utils.Error("Failed to push task branch", "taskID", taskID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   i18n.T(lang, "tasks.push_failed"),
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "tasks.push_success"),
		"data": gin.H{
			"output": output,
		},
	})
}
