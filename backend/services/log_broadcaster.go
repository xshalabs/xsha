package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"xsha-backend/utils"
)

// LogMessage log message structure
type LogMessage struct {
	ConversationID uint      `json:"conversation_id"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
	LogType        string    `json:"log_type"` // "log", "status", "error"
}

// SSEClient SSE client connection
type SSEClient struct {
	ID       string
	Channel  chan LogMessage
	UserID   string // User ID for access control
	CloseCh  chan bool
	LastSeen time.Time
}

// LogBroadcaster log broadcast manager
type LogBroadcaster struct {
	clients    map[string]*SSEClient
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan LogMessage
	mu         sync.RWMutex
}

// NewLogBroadcaster creates a log broadcast manager
func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients:    make(map[string]*SSEClient),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan LogMessage, 1000), // Buffer size is 1000
	}
}

// Start starts the broadcast manager
func (lb *LogBroadcaster) Start() {
	go lb.run()
	go lb.cleanupInactiveClients()
}

// run runs the broadcast loop
func (lb *LogBroadcaster) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-lb.register:
			lb.mu.Lock()
			lb.clients[client.ID] = client
			lb.mu.Unlock()

			utils.Info("SSE client connected",
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

			utils.Info("SSE client disconnected",
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
					// Channel is full, client may have disconnected
					utils.Warn("Client channel full, removing client",
						"client_id", id,
						"user_id", client.UserID,
					)
					delete(lb.clients, id)
					close(client.Channel)
				}
			}
			lb.mu.RUnlock()

		case <-ticker.C:
			// Periodically check inactive clients
			// This is handled in cleanupInactiveClients
		}
	}
}

// cleanupInactiveClients cleans up inactive clients
func (lb *LogBroadcaster) cleanupInactiveClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		lb.mu.Lock()
		for id, client := range lb.clients {
			if now.Sub(client.LastSeen) > 10*time.Minute {
				utils.Info("Cleaning up inactive SSE client",
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

// RegisterClient registers a client
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

// UnregisterClient unregisters a client
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
		// Message sent
	default:
		utils.Warn("Broadcast channel full, dropping message",
			"conversation_id", conversationID,
			"log_type", logType,
		)
	}
}

// BroadcastStatus broadcasts status messages
func (lb *LogBroadcaster) BroadcastStatus(conversationID uint, status string) {
	message := LogMessage{
		ConversationID: conversationID,
		Content:        status,
		Timestamp:      time.Now(),
		LogType:        "status",
	}

	select {
	case lb.broadcast <- message:
		// Message sent
	default:
		utils.Warn("Broadcast channel full, dropping status message",
			"conversation_id", conversationID,
			"status", status,
		)
	}
}

// GetConnectedClientCount gets the number of connected clients
func (lb *LogBroadcaster) GetConnectedClientCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.clients)
}

// FormatSSEMessage formats SSE messages
func (client *SSEClient) FormatSSEMessage(message LogMessage) string {
	data, _ := json.Marshal(message)
	return fmt.Sprintf("data: %s\n\n", string(data))
}
