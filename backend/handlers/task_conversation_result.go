package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type TaskConversationResultHandlers struct {
	resultService services.TaskConversationResultService
}

func NewTaskConversationResultHandlers(resultService services.TaskConversationResultService) *TaskConversationResultHandlers {
	return &TaskConversationResultHandlers{
		resultService: resultService,
	}
}

// @Description Update result request
type UpdateResultRequest struct {
	Updates map[string]interface{} `json:"updates"`
}

// GetResult retrieves a conversation result by ID
// @Summary Get conversation result
// @Description Get a conversation result by result ID
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Result ID"
// @Success 200 {object} object{message=string,data=object} "Result retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid result ID"
// @Failure 404 {object} object{error=string} "Result not found"
// @Router /conversation-results/{id} [get]
func (h *TaskConversationResultHandlers) GetResult(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	result, err := h.resultService.GetResult(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.result_not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_get_success"),
		"data":    result,
	})
}

// GetResultByConversationID retrieves a result by conversation ID
// @Summary Get result by conversation ID
// @Description Get a conversation result by conversation ID
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param conversation_id path int true "Conversation ID"
// @Success 200 {object} object{message=string,data=object} "Result retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 404 {object} object{error=string} "Result not found"
// @Router /conversations/{conversation_id}/result [get]
func (h *TaskConversationResultHandlers) GetResultByConversationID(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	result, err := h.resultService.GetResultByConversationID(uint(conversationID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.result_not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_get_success"),
		"data":    result,
	})
}

// ListResultsByTaskID lists results for a specific task
// @Summary List results by task ID
// @Description Get paginated list of conversation results for a specific task
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id query int true "Task ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size (1-100)" default(10)
// @Success 200 {object} object{message=string,data=object{results=[]object,total=int,page=int,page_size=int}} "Results retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /conversation-results [get]
func (h *TaskConversationResultHandlers) ListResultsByTaskID(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "task.id_required")})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	results, total, err := h.resultService.ListResultsByTaskID(uint(taskID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_list_success"),
		"data": gin.H{
			"results":   results,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ListResultsByProjectID lists results for a specific project
// @Summary List results by project ID
// @Description Get paginated list of conversation results for a specific project
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id query int true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size (1-100)" default(10)
// @Success 200 {object} object{message=string,data=object{items=[]object,total=int,page=int,page_size=int}} "Results retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /projects/{project_id}/conversation-results [get]
func (h *TaskConversationResultHandlers) ListResultsByProjectID(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "project.id_required")})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	results, total, err := h.resultService.ListResultsByProjectID(uint(projectID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_list_success"),
		"data": gin.H{
			"items":     results,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateResult updates a conversation result
// @Summary Update conversation result
// @Description Update specific fields of a conversation result
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Result ID"
// @Param result body UpdateResultRequest true "Result update information"
// @Success 200 {object} object{message=string} "Result updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /conversation-results/{id} [put]
func (h *TaskConversationResultHandlers) UpdateResult(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	var req UpdateResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error()})
		return
	}

	err = h.resultService.UpdateResult(uint(id), req.Updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_update_success"),
	})
}

// DeleteResult deletes a conversation result
// @Summary Delete conversation result
// @Description Delete a conversation result by ID
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Result ID"
// @Success 200 {object} object{message=string} "Result deleted successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /conversation-results/{id} [delete]
func (h *TaskConversationResultHandlers) DeleteResult(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	err = h.resultService.DeleteResult(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_delete_success"),
	})
}

// GetTaskStats retrieves statistics for a task
// @Summary Get task statistics
// @Description Get conversation result statistics for a specific task
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id path int true "Task ID"
// @Success 200 {object} object{message=string,data=object} "Task statistics retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /tasks/{task_id}/stats [get]
func (h *TaskConversationResultHandlers) GetTaskStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	stats, err := h.resultService.GetTaskStats(uint(taskID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.stats_get_success"),
		"data":    stats,
	})
}

// GetProjectStats retrieves statistics for a project
// @Summary Get project statistics
// @Description Get conversation result statistics for a specific project
// @Tags Task Conversation Results
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path int true "Project ID"
// @Success 200 {object} object{message=string,data=object} "Project statistics retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /projects/{project_id}/stats [get]
func (h *TaskConversationResultHandlers) GetProjectStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	stats, err := h.resultService.GetProjectStats(uint(projectID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.stats_get_success"),
		"data":    stats,
	})
}
