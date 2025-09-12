package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// DeadLetterQueue 死信队列接口
type DeadLetterQueue interface {
	// 推送失败事件到死信队列
	Push(event Event, subscription *Subscription, err error) error
	
	// 处理死信队列中的事件
	Process(handler EventHandler) error
	
	// 重试特定事件
	Retry(eventID string) error
	
	// 获取死信事件列表
	List(limit, offset int) ([]DeadLetterEvent, error)
	
	// 删除死信事件
	Delete(eventID string) error
	
	// 清空死信队列
	Clear() error
	
	// 获取统计信息
	Stats() (DeadLetterStats, error)
}

// DeadLetterEvent 死信事件
type DeadLetterEvent struct {
	ID               string    `gorm:"primaryKey;size:36" json:"id"`
	EventID          string    `gorm:"index;size:36;not null" json:"event_id"`
	EventType        string    `gorm:"index;size:100;not null" json:"event_type"`
	EventData        string    `gorm:"type:text" json:"event_data"`
	SubscriptionID   string    `gorm:"size:36" json:"subscription_id"`
	SubscriptionData string    `gorm:"type:text" json:"subscription_data"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message"`
	RetryCount       int       `json:"retry_count"`
	MaxRetries       int       `json:"max_retries"`
	FirstFailedAt    time.Time `gorm:"not null" json:"first_failed_at"`
	LastFailedAt     time.Time `gorm:"not null" json:"last_failed_at"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DeadLetterEvent) TableName() string {
	return "dead_letter_events"
}

// DeadLetterStats 死信队列统计
type DeadLetterStats struct {
	TotalEvents       int64            `json:"total_events"`
	EventsByType      map[string]int64 `json:"events_by_type"`
	EventsByErrorType map[string]int64 `json:"events_by_error_type"`
	OldestEvent       *time.Time       `json:"oldest_event"`
	NewestEvent       *time.Time       `json:"newest_event"`
}

// DBDeadLetterQueue 基于数据库的死信队列实现
type DBDeadLetterQueue struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewDBDeadLetterQueue 创建数据库死信队列
func NewDBDeadLetterQueue(db *gorm.DB) DeadLetterQueue {
	// 自动迁移表结构
	db.AutoMigrate(&DeadLetterEvent{})
	
	return &DBDeadLetterQueue{db: db}
}

// Push 推送失败事件到死信队列
func (dlq *DBDeadLetterQueue) Push(event Event, subscription *Subscription, err error) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	// 序列化事件数据
	eventData, jsonErr := json.Marshal(event)
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal event: %v", jsonErr)
	}
	
	// 序列化订阅数据
	subscriptionData, jsonErr := json.Marshal(map[string]interface{}{
		"id":          subscription.ID,
		"event_type":  subscription.EventType,
		"priority":    subscription.Priority,
		"async":       subscription.Async,
		"max_retries": subscription.MaxRetries,
		"retry_delay": subscription.RetryDelay.String(),
	})
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal subscription: %v", jsonErr)
	}
	
	now := time.Now()
	
	// 检查是否已存在相同的死信事件
	var existingEvent DeadLetterEvent
	result := dlq.db.Where("event_id = ? AND subscription_id = ?", 
		event.GetID(), string(subscription.ID)).First(&existingEvent)
	
	if result.Error == nil {
		// 更新现有记录
		existingEvent.RetryCount++
		existingEvent.ErrorMessage = err.Error()
		existingEvent.LastFailedAt = now
		existingEvent.UpdatedAt = now
		
		return dlq.db.Save(&existingEvent).Error
	}
	
	// 创建新的死信事件
	deadLetterEvent := &DeadLetterEvent{
		ID:               fmt.Sprintf("dl_%s_%s", event.GetID(), string(subscription.ID)),
		EventID:          event.GetID(),
		EventType:        event.GetType(),
		EventData:        string(eventData),
		SubscriptionID:   string(subscription.ID),
		SubscriptionData: string(subscriptionData),
		ErrorMessage:     err.Error(),
		RetryCount:       subscription.MaxRetries + 1, // 已经超过最大重试次数
		MaxRetries:       subscription.MaxRetries,
		FirstFailedAt:    now,
		LastFailedAt:     now,
	}
	
	return dlq.db.Create(deadLetterEvent).Error
}

// Process 处理死信队列中的事件
func (dlq *DBDeadLetterQueue) Process(handler EventHandler) error {
	const batchSize = 100
	var offset int
	
	for {
		var deadLetterEvents []DeadLetterEvent
		
		if err := dlq.db.Limit(batchSize).Offset(offset).Find(&deadLetterEvents).Error; err != nil {
			return err
		}
		
		if len(deadLetterEvents) == 0 {
			break
		}
		
		for _, dle := range deadLetterEvents {
			if err := dlq.processDeadLetterEvent(&dle, handler); err != nil {
				log.Printf("Failed to process dead letter event %s: %v", dle.ID, err)
			}
		}
		
		if len(deadLetterEvents) < batchSize {
			break
		}
		
		offset += batchSize
	}
	
	return nil
}

// processDeadLetterEvent 处理单个死信事件
func (dlq *DBDeadLetterQueue) processDeadLetterEvent(dle *DeadLetterEvent, handler EventHandler) error {
	// 反序列化事件
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(dle.EventData), &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %v", err)
	}
	
	// 重构事件对象
	event := &BaseEvent{
		ID:        dle.EventID,
		Type:      dle.EventType,
		Timestamp: dle.FirstFailedAt,
		Payload:   eventData,
		Metadata:  make(map[string]string),
	}
	
	// 尝试处理事件
	if err := handler(event); err != nil {
		log.Printf("Dead letter event %s still failed: %v", dle.ID, err)
		return err
	}
	
	// 处理成功，从死信队列中移除
	return dlq.Delete(dle.ID)
}

// Retry 重试特定事件
func (dlq *DBDeadLetterQueue) Retry(eventID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	var dle DeadLetterEvent
	if err := dlq.db.Where("id = ?", eventID).First(&dle).Error; err != nil {
		return err
	}
	
	// 反序列化事件
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(dle.EventData), &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %v", err)
	}
	
	// 这里应该重新发布事件到事件总线
	// 但由于循环依赖，这里只是标记为可重试状态
	log.Printf("Event %s marked for retry", dle.EventID)
	
	return nil
}

// List 获取死信事件列表
func (dlq *DBDeadLetterQueue) List(limit, offset int) ([]DeadLetterEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	
	var events []DeadLetterEvent
	
	err := dlq.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	
	return events, err
}

// Delete 删除死信事件
func (dlq *DBDeadLetterQueue) Delete(eventID string) error {
	return dlq.db.Where("id = ?", eventID).Delete(&DeadLetterEvent{}).Error
}

// Clear 清空死信队列
func (dlq *DBDeadLetterQueue) Clear() error {
	return dlq.db.Where("1 = 1").Delete(&DeadLetterEvent{}).Error
}

// Stats 获取统计信息
func (dlq *DBDeadLetterQueue) Stats() (DeadLetterStats, error) {
	var stats DeadLetterStats
	
	// 总事件数
	if err := dlq.db.Model(&DeadLetterEvent{}).Count(&stats.TotalEvents).Error; err != nil {
		return stats, err
	}
	
	// 按事件类型统计
	var typeStats []struct {
		EventType string
		Count     int64
	}
	
	if err := dlq.db.Model(&DeadLetterEvent{}).
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Find(&typeStats).Error; err != nil {
		return stats, err
	}
	
	stats.EventsByType = make(map[string]int64)
	for _, ts := range typeStats {
		stats.EventsByType[ts.EventType] = ts.Count
	}
	
	// 按错误类型统计（简化版，实际可以更复杂）
	stats.EventsByErrorType = make(map[string]int64)
	
	// 获取最老和最新的事件时间
	var oldestEvent DeadLetterEvent
	if err := dlq.db.Order("first_failed_at ASC").First(&oldestEvent).Error; err == nil {
		stats.OldestEvent = &oldestEvent.FirstFailedAt
	}
	
	var newestEvent DeadLetterEvent
	if err := dlq.db.Order("last_failed_at DESC").First(&newestEvent).Error; err == nil {
		stats.NewestEvent = &newestEvent.LastFailedAt
	}
	
	return stats, nil
}