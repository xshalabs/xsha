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
