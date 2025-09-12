package events

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Subscription 订阅信息
type Subscription struct {
	ID         SubscriptionID `json:"id"`
	EventType  string         `json:"event_type"`
	Handler    EventHandler   `json:"-"`
	Filter     EventFilter    `json:"-"`
	Priority   int            `json:"priority"`
	Async      bool           `json:"async"`
	MaxRetries int            `json:"max_retries"`
	RetryDelay time.Duration  `json:"retry_delay"`
	CreatedAt  time.Time      `json:"created_at"`
}

// SubscriptionManager 订阅管理器
type SubscriptionManager struct {
	subscriptions map[string][]*Subscription // eventType -> subscriptions
	idMap         map[SubscriptionID]*Subscription
	mu            sync.RWMutex
}

// NewSubscriptionManager 创建新的订阅管理器
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string][]*Subscription),
		idMap:         make(map[SubscriptionID]*Subscription),
	}
}

// Subscribe 订阅事件
func (sm *SubscriptionManager) Subscribe(eventType string, handler EventHandler, filter EventFilter, priority int) SubscriptionID {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := SubscriptionID(uuid.New().String())
	subscription := &Subscription{
		ID:         id,
		EventType:  eventType,
		Handler:    handler,
		Filter:     filter,
		Priority:   priority,
		Async:      false,
		MaxRetries: 3,
		RetryDelay: time.Second * 5,
		CreatedAt:  time.Now(),
	}

	// 添加到事件类型映射
	sm.subscriptions[eventType] = append(sm.subscriptions[eventType], subscription)
	
	// 按优先级排序（高优先级在前）
	sort.Slice(sm.subscriptions[eventType], func(i, j int) bool {
		return sm.subscriptions[eventType][i].Priority > sm.subscriptions[eventType][j].Priority
	})

	// 添加到ID映射
	sm.idMap[id] = subscription

	return id
}

// SubscribeAsync 异步订阅事件
func (sm *SubscriptionManager) SubscribeAsync(eventType string, handler EventHandler, filter EventFilter, priority int) SubscriptionID {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := SubscriptionID(uuid.New().String())
	subscription := &Subscription{
		ID:         id,
		EventType:  eventType,
		Handler:    handler,
		Filter:     filter,
		Priority:   priority,
		Async:      true,
		MaxRetries: 3,
		RetryDelay: time.Second * 5,
		CreatedAt:  time.Now(),
	}

	sm.subscriptions[eventType] = append(sm.subscriptions[eventType], subscription)
	
	sort.Slice(sm.subscriptions[eventType], func(i, j int) bool {
		return sm.subscriptions[eventType][i].Priority > sm.subscriptions[eventType][j].Priority
	})

	sm.idMap[id] = subscription

	return id
}

// Unsubscribe 取消订阅
func (sm *SubscriptionManager) Unsubscribe(id SubscriptionID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscription, exists := sm.idMap[id]
	if !exists {
		return ErrSubscriptionNotFound
	}

	// 从事件类型映射中移除
	eventType := subscription.EventType
	subscriptions := sm.subscriptions[eventType]
	
	for i, sub := range subscriptions {
		if sub.ID == id {
			// 从切片中移除
			sm.subscriptions[eventType] = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}

	// 如果没有订阅者了，删除这个事件类型
	if len(sm.subscriptions[eventType]) == 0 {
		delete(sm.subscriptions, eventType)
	}

	// 从ID映射中移除
	delete(sm.idMap, id)

	return nil
}

// GetSubscriptions 获取特定事件类型的所有订阅
func (sm *SubscriptionManager) GetSubscriptions(eventType string) []*Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriptions := sm.subscriptions[eventType]
	result := make([]*Subscription, len(subscriptions))
	copy(result, subscriptions)
	return result
}

// GetAllSubscriptions 获取所有订阅
func (sm *SubscriptionManager) GetAllSubscriptions() map[string][]*Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string][]*Subscription)
	for eventType, subscriptions := range sm.subscriptions {
		result[eventType] = make([]*Subscription, len(subscriptions))
		copy(result[eventType], subscriptions)
	}
	return result
}

// GetSubscriptionByID 根据ID获取订阅
func (sm *SubscriptionManager) GetSubscriptionByID(id SubscriptionID) (*Subscription, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscription, exists := sm.idMap[id]
	return subscription, exists
}

// HasSubscriptions 检查是否有特定事件类型的订阅
func (sm *SubscriptionManager) HasSubscriptions(eventType string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriptions, exists := sm.subscriptions[eventType]
	return exists && len(subscriptions) > 0
}

// Count 获取订阅总数
func (sm *SubscriptionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.idMap)
}

// CountByEventType 获取特定事件类型的订阅数
func (sm *SubscriptionManager) CountByEventType(eventType string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.subscriptions[eventType])
}

// Clear 清空所有订阅
func (sm *SubscriptionManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscriptions = make(map[string][]*Subscription)
	sm.idMap = make(map[SubscriptionID]*Subscription)
}

// GetEventTypes 获取所有有订阅的事件类型
func (sm *SubscriptionManager) GetEventTypes() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	types := make([]string, 0, len(sm.subscriptions))
	for eventType := range sm.subscriptions {
		types = append(types, eventType)
	}
	return types
}

// FilterSubscriptions 根据过滤器过滤订阅
func (sm *SubscriptionManager) FilterSubscriptions(eventType string, event Event) []*Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriptions := sm.subscriptions[eventType]
	var filtered []*Subscription

	for _, sub := range subscriptions {
		if sub.Filter == nil || sub.Filter(event) {
			filtered = append(filtered, sub)
		}
	}

	return filtered
}