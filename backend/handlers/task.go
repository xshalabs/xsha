package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

// TaskHandlers 任务处理器结构体
type TaskHandlers struct {
	taskService         services.TaskService
	conversationService services.TaskConversationService
	projectService      services.ProjectService
}

// NewTaskHandlers 创建任务处理器实例
func NewTaskHandlers(taskService services.TaskService, conversationService services.TaskConversationService, projectService services.ProjectService) *TaskHandlers {
	return &TaskHandlers{
		taskService:         taskService,
		conversationService: conversationService,
		projectService:      projectService,
	}
}

// CreateTaskRequest 创建任务请求结构
type CreateTaskRequest struct {
	Title            string `json:"title" binding:"required"`
	StartBranch      string `json:"start_branch" binding:"required"`
	ProjectID        uint   `json:"project_id" binding:"required"`
	DevEnvironmentID *uint  `json:"dev_environment_id"`
	RequirementDesc  string `json:"requirement_desc"` // 需求描述，用于创建conversation
	IncludeBranches  bool   `json:"include_branches"` // 是否返回项目分支信息
}

// CreateTaskResponse 创建任务响应结构
type CreateTaskResponse struct {
	Task            *database.Task `json:"task"`
	ProjectBranches []string       `json:"project_branches,omitempty"` // 项目分支列表
	BranchError     string         `json:"branch_error,omitempty"`     // 获取分支时的错误信息
}

// UpdateTaskRequest 更新任务请求结构
type UpdateTaskRequest struct {
	Title string `json:"title" binding:"required"`
}

// CreateTask 创建任务
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

	// 获取任务列表
	tasks, total, err := h.taskService.ListTasks(projectID, username.(string), status, page, pageSize)
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

// GetTaskStats 获取任务统计
func (h *TaskHandlers) GetTaskStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "task.project_id_required")})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
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

	// 获取任务统计
	stats, err := h.taskService.GetTaskStats(uint(projectID), username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "task.get_success"),
		"data":    stats,
	})
}
