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

// SSELogHandlers SSE日志处理器
type SSELogHandlers struct {
	logBroadcaster *services.LogBroadcaster
}

// NewSSELogHandlers 创建SSE日志处理器
func NewSSELogHandlers(logBroadcaster *services.LogBroadcaster) *SSELogHandlers {
	return &SSELogHandlers{
		logBroadcaster: logBroadcaster,
	}
}

// StreamLogs SSE日志流接口
// @Summary 实时日志流
// @Description 通过SSE推送任务执行的实时日志
// @Tags SSE日志
// @Produce text/plain
// @Param conversationId query int false "对话ID，用于过滤特定对话的日志"
// @Success 200 {string} string "event-stream"
// @Security BearerAuth
// @Router /api/v1/logs/stream [get]
func (h *SSELogHandlers) StreamLogs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// 获取用户信息
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取可选的对话ID过滤参数
	var filterConversationID uint
	if conversationIDStr := c.Query("conversationId"); conversationIDStr != "" {
		if id, err := strconv.ParseUint(conversationIDStr, 10, 32); err == nil {
			filterConversationID = uint(id)
		}
	}

	// 生成客户端ID
	clientID := generateClientID()

	// 注册客户端
	client := h.logBroadcaster.RegisterClient(clientID, username.(string))
	defer h.logBroadcaster.UnregisterClient(clientID)

	// 发送连接确认消息
	c.Stream(func(w io.Writer) bool {
		fmt.Fprintf(w, "data: %s\n\n", `{"message":"连接成功","type":"connection","timestamp":"`+time.Now().Format(time.RFC3339)+`"}`)
		return false
	})

	// 监听消息并发送
	for {
		select {
		case message := <-client.Channel:
			// 如果设置了对话ID过滤，只发送匹配的消息
			if filterConversationID != 0 && message.ConversationID != filterConversationID {
				continue
			}

			// 格式化并发送SSE消息
			formattedMessage := client.FormatSSEMessage(message)
			c.Stream(func(w io.Writer) bool {
				fmt.Fprint(w, formattedMessage)
				return false
			})

		case <-client.CloseCh:
			// 客户端连接关闭
			return

		case <-c.Request.Context().Done():
			// 请求上下文取消
			return

		case <-time.After(30 * time.Second):
			// 发送心跳消息
			heartbeat := `{"message":"heartbeat","type":"heartbeat","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
			c.Stream(func(w io.Writer) bool {
				fmt.Fprintf(w, "data: %s\n\n", heartbeat)
				return false
			})
		}
	}
}

// GetLogStats 获取日志统计信息
// @Summary 获取SSE连接统计
// @Description 获取当前SSE连接数等统计信息
// @Tags SSE日志
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

// SendTestMessage 发送测试消息（仅用于调试）
// @Summary 发送测试日志消息
// @Description 发送一条测试日志消息用于调试SSE功能
// @Tags SSE日志
// @Accept json
// @Produce json
// @Param conversationId path int true "对话ID"
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

	// 发送测试消息
	testMessage := fmt.Sprintf(i18n.T(lang, "sse_log.test_message_content"), time.Now().Format("15:04:05"))
	h.logBroadcaster.BroadcastLog(uint(conversationID), testMessage, "test")

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "sse_log.test_message_sent"),
		"content": testMessage,
	})
}

// generateClientID 生成客户端ID
func generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("client_%x_%d", bytes, time.Now().Unix())
}
