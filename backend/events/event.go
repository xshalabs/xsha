package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event 事件基础接口
type Event interface {
	GetID() string
	GetType() string
	GetTimestamp() time.Time
	GetPayload() interface{}
	GetMetadata() map[string]string
	SetMetadata(key, value string)
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Payload   interface{}       `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// GetID 获取事件ID
func (e *BaseEvent) GetID() string {
	return e.ID
}

// GetType 获取事件类型
func (e *BaseEvent) GetType() string {
	return e.Type
}

// GetTimestamp 获取事件时间戳
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetPayload 获取事件负载
func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata 获取事件元数据
func (e *BaseEvent) GetMetadata() map[string]string {
	return e.Metadata
}

// SetMetadata 设置元数据
func (e *BaseEvent) SetMetadata(key, value string) {
	e.Metadata[key] = value
}

// ToJSON 将事件转换为JSON
func (e *BaseEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON 从JSON创建事件
func FromJSON(data []byte) (*BaseEvent, error) {
	var event BaseEvent
	err := json.Unmarshal(data, &event)
	return &event, err
}

// EventHandler 事件处理器函数类型
type EventHandler func(event Event) error

// EventFilter 事件过滤器函数类型
type EventFilter func(event Event) bool

// SubscriptionID 订阅ID类型
type SubscriptionID string

// Priority 优先级常量
const (
	PriorityLow    = 1
	PriorityNormal = 5
	PriorityHigh   = 10
)

// EventError 事件处理错误
type EventError struct {
	EventID string
	Handler string
	Err     error
}

func (e *EventError) Error() string {
	return e.Err.Error()
}