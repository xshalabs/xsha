package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"
	"xsha-backend/services/executor"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

type TaskConversationHandlers struct {
	conversationService services.TaskConversationService
	logStreamingService executor.LogStreamingService
}

func NewTaskConversationHandlers(conversationService services.TaskConversationService, logStreamingService executor.LogStreamingService) *TaskConversationHandlers {
	return &TaskConversationHandlers{
		conversationService: conversationService,
		logStreamingService: logStreamingService,
	}
}

// @Description Create conversation request
type CreateConversationRequest struct {
	TaskID        uint       `json:"task_id" binding:"required" example:"1"`
	Content       string     `json:"content" binding:"required" example:"Please implement the user authentication feature"`
	ExecutionTime *time.Time `json:"execution_time" example:"2024-01-01T10:00:00Z"`
	EnvParams     string     `json:"env_params" example:"{\"model\":\"sonnet\"}"`
	AttachmentIDs []uint     `json:"attachment_ids,omitempty" example:"[1,2]"`
}

// @Description Update conversation request
type UpdateConversationRequest struct {
	Content string `json:"content" example:"Updated conversation content"`
}

// CreateConversation creates a new task conversation
// @Summary Create task conversation
// @Description Create a new conversation for a specific task
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param conversation body CreateConversationRequest true "Conversation information"
// @Success 201 {object} object{message=string,data=object} "Conversation created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /conversations [post]
func (h *TaskConversationHandlers) CreateConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error())})
		return
	}

	var conversation *database.TaskConversation
	var err error

	if len(req.AttachmentIDs) > 0 {
		conversation, err = h.conversationService.CreateConversationWithExecutionTimeAndAttachments(req.TaskID, req.Content, username.(string), req.ExecutionTime, req.EnvParams, req.AttachmentIDs)
	} else {
		conversation, err = h.conversationService.CreateConversationWithExecutionTime(req.TaskID, req.Content, username.(string), req.ExecutionTime, req.EnvParams)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "taskConversation.create_success"),
		"data":    conversation,
	})
}

// GetConversation retrieves a specific conversation
// @Summary Get task conversation
// @Description Get a conversation by ID
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Success 200 {object} object{message=string,data=object} "Conversation retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Conversation not found"
// @Router /conversations/{id} [get]
func (h *TaskConversationHandlers) GetConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	conversation, err := h.conversationService.GetConversation(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
		"data":    conversation,
	})
}

// GetConversationDetails retrieves a conversation with its result details
// @Summary Get conversation details
// @Description Get a conversation with its associated result information
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Success 200 {object} object{message=string,data=object} "Conversation details retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Conversation not found"
// @Router /conversations/{id}/details [get]
func (h *TaskConversationHandlers) GetConversationDetails(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	details, err := h.conversationService.GetConversationWithResult(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
		"data":    details,
	})
}

// ListConversations lists conversations for a task
// @Summary List task conversations
// @Description Get paginated list of conversations for a specific task
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id query int true "Task ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} object{message=string,data=object{conversations=[]object,total=int,page=int,page_size=int}} "Conversations retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /conversations [get]
func (h *TaskConversationHandlers) ListConversations(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "task.project_id_required")})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	conversations, total, err := h.conversationService.ListConversations(uint(taskID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
		"data": gin.H{
			"conversations": conversations,
			"total":         total,
			"page":          page,
			"page_size":     pageSize,
		},
	})
}

// UpdateConversation updates a conversation
// @Summary Update task conversation
// @Description Update a conversation's content
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Param conversation body UpdateConversationRequest true "Conversation update information"
// @Success 200 {object} object{message=string} "Conversation updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /conversations/{id} [put]
func (h *TaskConversationHandlers) UpdateConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error())})
		return
	}

	updates := make(map[string]interface{})
	if req.Content != "" {
		updates["content"] = req.Content
	}

	if err := h.conversationService.UpdateConversation(uint(id), updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "taskConversation.update_success")})
}

// DeleteConversation deletes a conversation
// @Summary Delete task conversation
// @Description Delete a conversation by ID
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Success 200 {object} object{message=string} "Conversation deleted successfully"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Conversation not found"
// @Router /conversations/{id} [delete]
func (h *TaskConversationHandlers) DeleteConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	if err := h.conversationService.DeleteConversation(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "taskConversation.update_success")})
}

// GetLatestConversation retrieves the latest conversation for a task
// @Summary Get latest task conversation
// @Description Get the most recent conversation for a specific task
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task_id query int true "Task ID"
// @Success 200 {object} object{message=string,data=object} "Latest conversation retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Conversation not found"
// @Router /conversations/latest [get]
func (h *TaskConversationHandlers) GetLatestConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "task.project_id_required")})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	conversation, err := h.conversationService.GetLatestConversation(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
		"data":    conversation,
	})
}

// GetConversationGitDiff retrieves Git diff for a conversation
// @Summary Get conversation Git diff
// @Description Get Git diff information for a specific conversation
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Param include_content query bool false "Include file content in diff" default(false)
// @Success 200 {object} object{data=object} "Git diff retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 500 {object} object{error=string} "Failed to get Git diff"
// @Router /conversations/{id}/git-diff [get]
func (h *TaskConversationHandlers) GetConversationGitDiff(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	includeContent := c.DefaultQuery("include_content", "false") == "true"

	diff, err := h.conversationService.GetConversationGitDiff(uint(conversationID), includeContent)
	if err != nil {
		utils.Error("Failed to get conversation Git diff", "conversationID", conversationID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "taskConversation.git_diff_failed"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": diff,
	})
}

// GetConversationGitDiffFile retrieves Git diff for a specific file
// @Summary Get conversation file Git diff
// @Description Get Git diff for a specific file in a conversation
// @Tags Task Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Param file_path query string true "File path"
// @Success 200 {object} object{data=object{file_path=string,diff_content=string}} "File Git diff retrieved successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 500 {object} object{error=string} "Failed to get file Git diff"
// @Router /conversations/{id}/git-diff/file [get]
func (h *TaskConversationHandlers) GetConversationGitDiffFile(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
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

	diffContent, err := h.conversationService.GetConversationGitDiffFile(uint(conversationID), filePath)
	if err != nil {
		utils.Error("Failed to get conversation file Git diff", "conversationID", conversationID, "filePath", filePath, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "taskConversation.git_diff_file_failed"),
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

// StreamConversationLogs streams real-time execution logs for a conversation
// @Summary Stream conversation execution logs
// @Description Get real-time or historical execution logs for a specific conversation via Server-Sent Events (SSE)
// @Tags Task Conversations
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Success 200 {string} string "Real-time log stream"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Conversation not found"
// @Failure 500 {object} object{error=string} "Failed to stream logs"
// @Router /conversations/{id}/logs/stream [get]
func (h *TaskConversationHandlers) StreamConversationLogs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Create context that will be cancelled when client disconnects
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Start streaming logs
	logChan, errChan, err := h.logStreamingService.StreamConversationLogs(ctx, uint(conversationID))
	if err != nil {
		utils.Error("Failed to start log streaming", "conversationID", conversationID, "error", err)
		c.SSEvent("error", gin.H{"message": i18n.T(lang, "taskConversation.log_stream_failed")})
		return
	}

	utils.Info("Started log streaming", "conversationID", conversationID)

	// Send initial connection message
	c.SSEvent("connected", gin.H{
		"conversation_id": conversationID,
		"timestamp":       time.Now().Unix(),
	})
	c.Writer.Flush()

	// Stream logs
	for {
		select {
		case <-ctx.Done():
			utils.Info("Log streaming cancelled by client", "conversationID", conversationID)
			return
		case logLine, ok := <-logChan:
			if !ok {
				// Log channel closed, conversation finished
				utils.Info("Log streaming completed", "conversationID", conversationID)
				c.SSEvent("finished", gin.H{
					"conversation_id": conversationID,
					"timestamp":       time.Now().Unix(),
				})
				c.Writer.Flush()
				return
			}

			// Send log line to client
			c.SSEvent("log", gin.H{
				"line":      logLine,
				"timestamp": time.Now().Unix(),
			})
			c.Writer.Flush()

		case streamErr, ok := <-errChan:
			if !ok {
				continue
			}

			utils.Error("Error during log streaming", "conversationID", conversationID, "error", streamErr)
			c.SSEvent("error", gin.H{
				"message":   fmt.Sprintf("Log streaming error: %v", streamErr),
				"timestamp": time.Now().Unix(),
			})
			c.Writer.Flush()
			return
		}
	}
}
