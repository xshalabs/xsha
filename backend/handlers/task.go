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

type TaskHandlers struct {
	taskService         services.TaskService
	conversationService services.TaskConversationService
	projectService      services.ProjectService
}

func NewTaskHandlers(taskService services.TaskService, conversationService services.TaskConversationService, projectService services.ProjectService) *TaskHandlers {
	return &TaskHandlers{
		taskService:         taskService,
		conversationService: conversationService,
		projectService:      projectService,
	}
}

// @Description Create task request
type CreateTaskRequest struct {
	Title            string `json:"title" binding:"required" example:"Fix user authentication bug"`
	StartBranch      string `json:"start_branch" binding:"required" example:"main"`
	ProjectID        uint   `json:"project_id" binding:"required" example:"1"`
	DevEnvironmentID *uint  `json:"dev_environment_id" binding:"required" example:"1"`
	RequirementDesc  string `json:"requirement_desc" binding:"required" example:"Fix the login validation issue"`
	IncludeBranches  bool   `json:"include_branches" example:"true"`
}

// @Description Create task response
type CreateTaskResponse struct {
	Task            *database.Task `json:"task"`
	ProjectBranches []string       `json:"project_branches,omitempty" example:"main,develop,feature/user-auth"`
	BranchError     string         `json:"branch_error,omitempty" example:"Failed to fetch branches"`
}

// @Description Update task request
type UpdateTaskRequest struct {
	Title string `json:"title" binding:"required" example:"Updated task title"`
}

// CreateTask creates a new task
// @Summary Create task
// @Description Create a new task with optional requirement description and branch fetching
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body CreateTaskRequest true "Task information"
// @Success 201 {object} object{message=string,data=CreateTaskResponse} "Task created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /tasks [post]
func (h *TaskHandlers) CreateTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "auth.unauthorized"),
		})
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.CreateTask(req.Title, req.StartBranch, req.ProjectID, req.DevEnvironmentID, username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	if strings.TrimSpace(req.RequirementDesc) != "" {
		_, err := h.conversationService.CreateConversation(
			task.ID,
			req.RequirementDesc,
			username.(string),
		)
		if err != nil {
			utils.Error("Failed to create conversation", "taskID", task.ID, "error", err)
		}
	}

	response := CreateTaskResponse{
		Task: task,
	}

	if req.IncludeBranches {
		if task.Project != nil {
			branchResult, err := h.projectService.FetchRepositoryBranches(
				task.Project.RepoURL,
				task.Project.CredentialID,
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

// GetTask retrieves a specific task
// @Summary Get task
// @Description Get a task by ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} object{message=string,data=database.Task} "Task retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid task ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Task not found"
// @Router /tasks/{id} [get]
func (h *TaskHandlers) GetTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	task, err := h.taskService.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "task.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "task.get_success"),
		"data":    task,
	})
}

// ListTasks retrieves tasks with pagination and filtering
// @Summary List tasks
// @Description Get a paginated list of tasks with optional filtering by project, status, title, branch, and dev environment
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param page_size query int false "Number of items per page (default: 20)" default(20)
// @Param project_id query int false "Filter by project ID"
// @Param status query string false "Filter by task status" Enums(todo,in_progress,done,cancelled)
// @Param title query string false "Filter by task title (partial match)"
// @Param branch query string false "Filter by branch name"
// @Param dev_environment_id query int false "Filter by development environment ID"
// @Success 200 {object} object{message=string,data=object{tasks=[]database.Task,total=int,page=int,page_size=int}} "Tasks retrieved successfully"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /tasks [get]
func (h *TaskHandlers) ListTasks(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

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

	tasks, total, err := h.taskService.ListTasks(projectID, status, title, branch, devEnvID, page, pageSize)
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

// UpdateTask updates an existing task
// @Summary Update task
// @Description Update task information
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param task body UpdateTaskRequest true "Task update information"
// @Success 200 {object} object{message=string} "Task updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /tasks/{id} [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error())})
		return
	}

	updates := make(map[string]interface{})
	updates["title"] = req.Title

	if err := h.taskService.UpdateTask(uint(id), updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task.update_success")})
}

// @Description Update task status request
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required" example:"in_progress" enums:"todo,in_progress,done,cancelled"`
}

// UpdateTaskStatus updates the status of a task
// @Summary Update task status
// @Description Update the status of a specific task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param status body UpdateTaskStatusRequest true "Task status update information"
// @Success 200 {object} object{message=string} "Task status updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /tasks/{id}/status [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error())})
		return
	}

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

	if err := h.taskService.UpdateTaskStatus(uint(id), status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task.update_success")})
}

// DeleteTask deletes a task
// @Summary Delete task
// @Description Delete a specific task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} object{message=string} "Task deleted successfully"
// @Failure 400 {object} object{error=string} "Invalid task ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Task not found"
// @Router /tasks/{id} [delete]
func (h *TaskHandlers) DeleteTask(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	if err := h.taskService.DeleteTask(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task.delete_success")})
}

// @Description Batch update task status request
type BatchUpdateTaskStatusRequest struct {
	TaskIDs []uint `json:"task_ids" binding:"required,min=1,max=100" example:"1,2,3"`
	Status  string `json:"status" binding:"required" example:"done" enums:"todo,in_progress,done,cancelled"`
}

// @Description Batch update task status response
type BatchUpdateTaskStatusResponse struct {
	SuccessCount int    `json:"success_count" example:"2"`
	FailedCount  int    `json:"failed_count" example:"1"`
	SuccessIDs   []uint `json:"success_ids" example:"1,2"`
	FailedIDs    []uint `json:"failed_ids" example:"3"`
}

// BatchUpdateTaskStatus updates the status of multiple tasks
// @Summary Batch update task status
// @Description Update the status of multiple tasks in a single request
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param batch body BatchUpdateTaskStatusRequest true "Batch status update information"
// @Success 200 {object} object{message=string,data=BatchUpdateTaskStatusResponse} "Batch update completed"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /tasks/batch/status [put]
func (h *TaskHandlers) BatchUpdateTaskStatus(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req BatchUpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error())})
		return
	}

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

	successIDs, failedIDs, err := h.taskService.UpdateTaskStatusBatch(req.TaskIDs, status)
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

// GetTaskGitDiff retrieves the git diff for a task
// @Summary Get task git diff
// @Description Get the git diff between start branch and work branch for a task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param include_content query bool false "Include file content in diff" default(false)
// @Success 200 {object} object{data=object} "Git diff retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid task ID or missing workspace"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 403 {object} object{error=string} "No permission to access task"
// @Failure 404 {object} object{error=string} "Task not found"
// @Failure 500 {object} object{error=string} "Failed to get git diff"
// @Router /tasks/{id}/git-diff [get]
func (h *TaskHandlers) GetTaskGitDiff(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	includeContent := c.DefaultQuery("include_content", "false") == "true"

	task, err := h.taskService.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "tasks.errors.not_found"),
		})
		return
	}

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

	if task.WorkspacePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "tasks.errors.no_workspace"),
		})
		return
	}

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

// GetTaskGitDiffFile retrieves the git diff for a specific file in a task
// @Summary Get task git diff file
// @Description Get the git diff for a specific file between start branch and work branch
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param file_path query string true "File path to get diff for"
// @Success 200 {object} object{data=object{file_path=string,diff_content=string}} "File diff retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid task ID or missing file path"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 403 {object} object{error=string} "No permission to access task"
// @Failure 404 {object} object{error=string} "Task not found"
// @Failure 500 {object} object{error=string} "Failed to get file diff"
// @Router /tasks/{id}/git-diff/file [get]
func (h *TaskHandlers) GetTaskGitDiffFile(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.file_path_required"),
		})
		return
	}

	task, err := h.taskService.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "tasks.errors.not_found"),
		})
		return
	}

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

// PushTaskBranch pushes the task's work branch to the remote repository
// @Summary Push task branch
// @Description Push the task's work branch to the remote repository
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param request body object{force_push=bool} false "Push options"
// @Success 200 {object} object{message=string,data=object{output=string}} "Branch pushed successfully"
// @Failure 400 {object} object{error=string} "Invalid task ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 500 {object} object{error=string,details=string} "Failed to push branch"
// @Router /tasks/{id}/push [post]
func (h *TaskHandlers) PushTaskBranch(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var req struct {
		ForcePush bool `json:"force_push"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.ForcePush = false
	}

	output, err := h.taskService.PushTaskBranch(uint(taskID), req.ForcePush)
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
