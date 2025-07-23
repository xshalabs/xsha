package services

import (
	"encoding/json"
	"fmt"
	"log"
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
}

// NewLogBroadcaster 创建日志广播管理器
func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients:    make(map[string]*SSEClient),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan LogMessage, 1000), // 缓冲区大小为1000
	}
}

// Start 启动广播管理器
func (lb *LogBroadcaster) Start() {
	go lb.run()
	go lb.cleanupInactiveClients()
}

// run 运行广播循环
func (lb *LogBroadcaster) run() {
	for {
		select {
		case client := <-lb.register:
			lb.mu.Lock()
			lb.clients[client.ID] = client
			lb.mu.Unlock()
			log.Printf("SSE客户端已连接: %s (用户: %s)", client.ID, client.UserID)

		case client := <-lb.unregister:
			lb.mu.Lock()
			if _, exists := lb.clients[client.ID]; exists {
				delete(lb.clients, client.ID)
				close(client.Channel)
				close(client.CloseCh)
				log.Printf("SSE客户端已断开: %s (用户: %s)", client.ID, client.UserID)
			}
			lb.mu.Unlock()

		case message := <-lb.broadcast:
			lb.mu.RLock()
			for _, client := range lb.clients {
				// 检查用户权限（简单实现：所有用户都能看到所有日志）
				// 在实际应用中，这里应该根据用户权限过滤消息
				select {
				case client.Channel <- message:
					client.LastSeen = time.Now()
				default:
					// 客户端通道已满，断开连接
					go lb.UnregisterClient(client.ID)
				}
			}
			lb.mu.RUnlock()
		}
	}
}

// cleanupInactiveClients 清理非活跃客户端
func (lb *LogBroadcaster) cleanupInactiveClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		lb.mu.Lock()
		now := time.Now()
		for id, client := range lb.clients {
			if now.Sub(client.LastSeen) > 10*time.Minute {
				delete(lb.clients, id)
				close(client.Channel)
				close(client.CloseCh)
				log.Printf("清理非活跃SSE客户端: %s", id)
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
func (lb *LogBroadcaster) BroadcastLog(conversationID uint, content string, logType string) {
	message := LogMessage{
		ConversationID: conversationID,
		Content:        content,
		Timestamp:      time.Now(),
		LogType:        logType,
	}

	select {
	case lb.broadcast <- message:
	default:
		log.Printf("警告：广播通道已满，丢弃消息")
	}
}

// BroadcastStatusChange 广播状态变化
func (lb *LogBroadcaster) BroadcastStatusChange(conversationID uint, status string, message string) {
	statusMessage := LogMessage{
		ConversationID: conversationID,
		Content:        fmt.Sprintf("状态变更: %s - %s", status, message),
		Timestamp:      time.Now(),
		LogType:        "status",
	}

	select {
	case lb.broadcast <- statusMessage:
	default:
		log.Printf("警告：广播通道已满，丢弃状态消息")
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
