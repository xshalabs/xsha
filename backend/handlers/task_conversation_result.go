package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

// TaskConversationResultHandlers 任务对话结果处理器结构体
type TaskConversationResultHandlers struct {
	resultService services.TaskConversationResultService
}

// NewTaskConversationResultHandlers 创建任务对话结果处理器实例
func NewTaskConversationResultHandlers(resultService services.TaskConversationResultService) *TaskConversationResultHandlers {
	return &TaskConversationResultHandlers{
		resultService: resultService,
	}
}

// UpdateResultRequest 更新结果请求结构
type UpdateResultRequest struct {
	Updates map[string]interface{} `json:"updates"`
}

// GetResult 获取结果详情
func (h *TaskConversationResultHandlers) GetResult(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取结果
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

// GetResultByConversationID 根据对话ID获取结果
func (h *TaskConversationResultHandlers) GetResultByConversationID(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取结果
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

// ListResultsByTaskID 根据任务ID获取结果列表
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

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取结果列表
	results, total, err := h.resultService.ListResultsByTaskID(uint(taskID), page, pageSize)
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

// ListResultsByProjectID 根据项目ID获取结果列表
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

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取结果列表
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

// UpdateResult 更新结果
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

	// 更新结果
	err = h.resultService.UpdateResult(uint(id), req.Updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_update_success"),
	})
}

// DeleteResult 删除结果
func (h *TaskConversationResultHandlers) DeleteResult(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 删除结果
	err = h.resultService.DeleteResult(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.result_delete_success"),
	})
}

// GetTaskStats 获取任务统计信息
func (h *TaskConversationResultHandlers) GetTaskStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取统计信息
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

// GetProjectStats 获取项目统计信息
func (h *TaskConversationResultHandlers) GetProjectStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取统计信息
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
