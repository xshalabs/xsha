package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sleep0-backend/utils"
	"sync"
	"time"
)

// LogMessage 日志消息结构
type LogMessage struct {
	ConversationID uint      `json:"conversation_id"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
	LogType        string    `json:"log_type"` // "log", "status", "error"
}

// SSEClient SSE客户端连接
type SSEClient struct {
	ID       string
	Channel  chan LogMessage
	UserID   string // 用户ID，用于权限控制
	CloseCh  chan bool
	LastSeen time.Time
}

// LogBroadcaster 日志广播管理器
type LogBroadcaster struct {
	clients    map[string]*SSEClient
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan LogMessage
	mu         sync.RWMutex
	logger     *slog.Logger
}

// NewLogBroadcaster 创建日志广播管理器
func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients:    make(map[string]*SSEClient),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan LogMessage, 1000), // 缓冲区大小为1000
		logger:     utils.WithFields(map[string]interface{}{"component": "log_broadcaster"}),
	}
}

// Start 启动广播管理器
func (lb *LogBroadcaster) Start() {
	go lb.run()
	go lb.cleanupInactiveClients()
}

// run 运行广播循环
func (lb *LogBroadcaster) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-lb.register:
			lb.mu.Lock()
			lb.clients[client.ID] = client
			lb.mu.Unlock()

			lb.logger.Info("SSE client connected",
				"client_id", client.ID,
				"user_id", client.UserID,
			)

		case client := <-lb.unregister:
			lb.mu.Lock()
			if _, ok := lb.clients[client.ID]; ok {
				delete(lb.clients, client.ID)
				close(client.Channel)
			}
			lb.mu.Unlock()

			lb.logger.Info("SSE client disconnected",
				"client_id", client.ID,
				"user_id", client.UserID,
			)

		case message := <-lb.broadcast:
			lb.mu.RLock()
			for id, client := range lb.clients {
				select {
				case client.Channel <- message:
					client.LastSeen = time.Now()
				default:
					// 通道已满，客户端可能已断开
					lb.logger.Warn("Client channel full, removing client",
						"client_id", id,
						"user_id", client.UserID,
					)
					delete(lb.clients, id)
					close(client.Channel)
				}
			}
			lb.mu.RUnlock()

		case <-ticker.C:
			// 定期检查不活跃的客户端
			// 这个在cleanupInactiveClients中处理
		}
	}
}

// cleanupInactiveClients 清理不活跃的客户端
func (lb *LogBroadcaster) cleanupInactiveClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		lb.mu.Lock()
		for id, client := range lb.clients {
			if now.Sub(client.LastSeen) > 10*time.Minute {
				lb.logger.Info("Cleaning up inactive SSE client",
					"client_id", id,
					"user_id", client.UserID,
					"inactive_duration", now.Sub(client.LastSeen).String(),
				)
				delete(lb.clients, id)
				close(client.Channel)
			}
		}
		lb.mu.Unlock()
	}
}

// RegisterClient 注册客户端
func (lb *LogBroadcaster) RegisterClient(clientID, userID string) *SSEClient {
	client := &SSEClient{
		ID:       clientID,
		Channel:  make(chan LogMessage, 100),
		UserID:   userID,
		CloseCh:  make(chan bool),
		LastSeen: time.Now(),
	}

	lb.register <- client
	return client
}

// UnregisterClient 取消注册客户端
func (lb *LogBroadcaster) UnregisterClient(clientID string) {
	lb.mu.RLock()
	client, exists := lb.clients[clientID]
	lb.mu.RUnlock()

	if exists {
		lb.unregister <- client
	}
}

// BroadcastLog 广播日志消息
func (lb *LogBroadcaster) BroadcastLog(conversationID uint, content, logType string) {
	message := LogMessage{
		ConversationID: conversationID,
		Content:        content,
		Timestamp:      time.Now(),
		LogType:        logType,
	}

	select {
	case lb.broadcast <- message:
		// 消息已发送
	default:
		lb.logger.Warn("Broadcast channel full, dropping message",
			"conversation_id", conversationID,
			"log_type", logType,
		)
	}
}

// BroadcastStatus 广播状态消息
func (lb *LogBroadcaster) BroadcastStatus(conversationID uint, status string) {
	message := LogMessage{
		ConversationID: conversationID,
		Content:        status,
		Timestamp:      time.Now(),
		LogType:        "status",
	}

	select {
	case lb.broadcast <- message:
		// 消息已发送
	default:
		lb.logger.Warn("Broadcast channel full, dropping status message",
			"conversation_id", conversationID,
			"status", status,
		)
	}
}

// GetConnectedClientCount 获取连接的客户端数量
func (lb *LogBroadcaster) GetConnectedClientCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.clients)
}

// FormatSSEMessage 格式化SSE消息
func (client *SSEClient) FormatSSEMessage(message LogMessage) string {
	data, _ := json.Marshal(message)
	return fmt.Sprintf("data: %s\n\n", string(data))
}
