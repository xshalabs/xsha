package handlers

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

// SSELogHandlers SSE log handler
type SSELogHandlers struct {
	logBroadcaster *services.LogBroadcaster
}

// NewSSELogHandlers creates an SSE log handler
func NewSSELogHandlers(logBroadcaster *services.LogBroadcaster) *SSELogHandlers {
	return &SSELogHandlers{
		logBroadcaster: logBroadcaster,
	}
}

// StreamLogs SSE log streaming interface
// @Summary Real-time log stream
// @Description Push real-time logs of task execution via SSE
// @Tags SSE Logs
// @Produce text/plain
// @Param conversationId query int false "Conversation ID for filtering logs of specific conversations"
// @Success 200 {string} string "event-stream"
// @Security BearerAuth
// @Router /api/v1/logs/stream [get]
func (h *SSELogHandlers) StreamLogs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// Get user information
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// Get optional conversation ID filter parameter
	var filterConversationID uint
	if conversationIDStr := c.Query("conversationId"); conversationIDStr != "" {
		if id, err := strconv.ParseUint(conversationIDStr, 10, 32); err == nil {
			filterConversationID = uint(id)
		}
	}

	// Generate client ID
	clientID := generateClientID()

	// Register client
	client := h.logBroadcaster.RegisterClient(clientID, username.(string))
	defer h.logBroadcaster.UnregisterClient(clientID)

	// Send connection confirmation message
	c.Stream(func(w io.Writer) bool {
		fmt.Fprintf(w, "data: %s\n\n", `{"message":"Connected successfully","type":"connection","timestamp":"`+time.Now().Format(time.RFC3339)+`"}`)
		return false
	})

	// Listen for messages and send them
	for {
		select {
		case message := <-client.Channel:
			// If conversation ID filter is set, only send matching messages
			if filterConversationID != 0 && message.ConversationID != filterConversationID {
				continue
			}

			// Format and send SSE message
			formattedMessage := client.FormatSSEMessage(message)
			c.Stream(func(w io.Writer) bool {
				fmt.Fprint(w, formattedMessage)
				return false
			})

		case <-client.CloseCh:
			// Client connection closed
			return

		case <-c.Request.Context().Done():
			// Request context cancelled
			return

		case <-time.After(30 * time.Second):
			// Send heartbeat message
			heartbeat := `{"message":"heartbeat","type":"heartbeat","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
			c.Stream(func(w io.Writer) bool {
				fmt.Fprintf(w, "data: %s\n\n", heartbeat)
				return false
			})
		}
	}
}

// GetLogStats gets log statistics
// @Summary Get SSE connection statistics
// @Description Get current SSE connection count and other statistical information
// @Tags SSE Logs
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/logs/stats [get]
func (h *SSELogHandlers) GetLogStats(c *gin.Context) {
	stats := map[string]interface{}{
		"connected_clients": h.logBroadcaster.GetConnectedClientCount(),
		"timestamp":         time.Now(),
	}

	c.JSON(http.StatusOK, stats)
}

// SendTestMessage sends test message (for debugging only)
// @Summary Send test log message
// @Description Send a test log message for debugging SSE functionality
// @Tags SSE Logs
// @Accept json
// @Produce json
// @Param conversationId path int true "Conversation ID"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/logs/test/{conversationId} [post]
func (h *SSELogHandlers) SendTestMessage(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// Send test message
	testMessage := fmt.Sprintf(i18n.T(lang, "sse_log.test_message_content"), time.Now().Format("15:04:05"))
	h.logBroadcaster.BroadcastLog(uint(conversationID), testMessage, "test")

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "sse_log.test_message_sent"),
		"content": testMessage,
	})
}

// generateClientID generates client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("client_%x_%d", bytes, time.Now().Unix())
}
