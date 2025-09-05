package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
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
	Title            string     `json:"title" binding:"required" example:"Fix user authentication bug"`
	StartBranch      string     `json:"start_branch" binding:"required" example:"main"`
	ProjectID        uint       `json:"project_id" example:"1"`
	DevEnvironmentID *uint      `json:"dev_environment_id" binding:"required" example:"1"`
	RequirementDesc  string     `json:"requirement_desc" binding:"required" example:"Fix the login validation issue"`
	IncludeBranches  bool       `json:"include_branches" example:"true"`
	ExecutionTime    *time.Time `json:"execution_time" example:"2024-01-01T10:00:00Z"`
	EnvParams        string     `json:"env_params" example:"{\"model\":\"sonnet\"}"`
	AttachmentIDs    []uint     `json:"attachment_ids" swaggertype:"array,integer" example:"1,2,3"`
}

// @Description Create task response
type CreateTaskResponse struct {
	Task            *database.Task `json:"task"`
	ProjectBranches []string       `json:"project_branches,omitempty" example:"main,develop,feature/user-auth"`
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

	// Extract project ID from URL path
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	username, _ := c.Get("username")
	adminIDInterface, _ := c.Get("admin_id")
	adminID, _ := adminIDInterface.(uint)

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProjectID = uint(projectID)

	task, err := h.taskService.CreateTask(req.Title, req.StartBranch, req.ProjectID, req.DevEnvironmentID, &adminID, username.(string))
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.RequirementDesc) != "" {
		var err error
		if len(req.AttachmentIDs) > 0 {
			_, err = h.conversationService.CreateConversationWithExecutionTimeAndAttachments(
				task.ID,
				req.RequirementDesc,
				username.(string),
				req.ExecutionTime,
				req.EnvParams,
				req.AttachmentIDs,
				&adminID,
			)
		} else {
			_, err = h.conversationService.CreateConversationWithExecutionTime(
				task.ID,
				req.RequirementDesc,
				username.(string),
				req.ExecutionTime,
				req.EnvParams,
				&adminID,
			)
		}
		if err != nil {
			utils.Error("Failed to create conversation", "taskID", task.ID, "error", err)
		}
	}

	response := CreateTaskResponse{
		Task: task,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "task.create_success"),
		"data":    response,
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

	// Extract project ID from URL path for validation
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
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

	// Verify task belongs to project before updating
	_, err = h.taskService.GetTaskByIDAndProject(uint(taskID), uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "task.not_found")})
		return
	}

	if err := h.taskService.UpdateTask(uint(taskID), updates); err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
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

	// Extract project ID from URL path for validation
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// Verify task belongs to project before deleting
	_, err = h.taskService.GetTaskByIDAndProject(uint(taskID), uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "task.not_found")})
		return
	}

	if err := h.taskService.DeleteTask(uint(taskID)); err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusNotFound, err)
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
// @Description Update the status of multiple tasks in a single request within a specific project
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param batch body BatchUpdateTaskStatusRequest true "Batch status update information"
// @Success 200 {object} object{message=string,data=BatchUpdateTaskStatusResponse} "Batch update completed"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Project not found or task not found"
// @Router /projects/{id}/tasks/batch/status [put]
func (h *TaskHandlers) BatchUpdateTaskStatus(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// Extract project ID from URL path
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

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

	successIDs, failedIDs, err := h.taskService.UpdateTaskStatusBatch(req.TaskIDs, status, uint(projectID))
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
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

	// Extract project ID from URL path for validation
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	includeContent := c.DefaultQuery("include_content", "false") == "true"

	task, err := h.taskService.GetTaskByIDAndProject(uint(taskID), uint(projectID))
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

	// Extract project ID from URL path for validation
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	taskIDStr := c.Param("taskId")
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

	task, err := h.taskService.GetTaskByIDAndProject(uint(taskID), uint(projectID))
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

	// Extract project ID from URL path for validation
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	taskIDStr := c.Param("taskId")
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

	// Verify task belongs to project before pushing
	_, err = h.taskService.GetTaskByIDAndProject(uint(taskID), uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "tasks.errors.not_found"),
		})
		return
	}

	output, err := h.taskService.PushTaskBranch(uint(taskID), req.ForcePush)
	if err != nil {
		utils.Error("Failed to push task branch", "taskID", taskID, "error", err)
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "tasks.push_success"),
		"data": gin.H{
			"output": output,
		},
	})
}

// @Description Get kanban tasks response
type GetKanbanTasksResponse struct {
	Todo       []database.Task `json:"todo"`
	InProgress []database.Task `json:"in_progress"`
	Done       []database.Task `json:"done"`
	Cancelled  []database.Task `json:"cancelled"`
}

// GetKanbanTasks retrieves tasks grouped by status for kanban view
// @Summary Get kanban tasks
// @Description Get tasks grouped by status for a specific project to display in kanban view
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} object{message=string,data=GetKanbanTasksResponse} "Kanban tasks retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid project ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Project not found"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /projects/{id}/kanban [get]
func (h *TaskHandlers) GetKanbanTasks(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	kanbanData, err := h.taskService.GetKanbanTasks(uint(projectID))
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusNotFound, err)
		return
	}

	response := GetKanbanTasksResponse{
		Todo:       kanbanData[database.TaskStatusTodo],
		InProgress: kanbanData[database.TaskStatusInProgress],
		Done:       kanbanData[database.TaskStatusDone],
		Cancelled:  kanbanData[database.TaskStatusCancelled],
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "task.kanban_get_success"),
		"data":    response,
	})
}
