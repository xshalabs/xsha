package events

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrEventNotFound = errors.New("event not found")
)

// EventStore 事件存储接口
type EventStore interface {
	// 保存事件
	Save(event Event) error
	
	// 查询事件
	GetByID(id string) (Event, error)
	GetByType(eventType string, limit int) ([]Event, error)
	GetByTypeAndTimeRange(eventType string, start, end time.Time, limit int) ([]Event, error)
	GetByTimeRange(start, end time.Time, limit int) ([]Event, error)
	
	// 事件重放
	Replay(from time.Time, handler EventHandler) error
	ReplayByType(eventType string, from time.Time, handler EventHandler) error
	
	// 清理过期事件
	CleanupOldEvents(before time.Time) error
	
	// 统计
	CountByType(eventType string) (int64, error)
	CountByTimeRange(start, end time.Time) (int64, error)
}

// StoredEvent 存储的事件模型
type StoredEvent struct {
	ID        string            `gorm:"primaryKey;size:36" json:"id"`
	Type      string            `gorm:"index;size:100;not null" json:"type"`
	Timestamp time.Time         `gorm:"index;not null" json:"timestamp"`
	PayloadJSON string          `gorm:"type:text" json:"payload_json"`
	Metadata  string            `gorm:"type:text" json:"metadata"`
	CreatedAt time.Time         `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (StoredEvent) TableName() string {
	return "events"
}

// ToEvent 转换为事件接口
func (se *StoredEvent) ToEvent() (Event, error) {
	var payload interface{}
	if se.PayloadJSON != "" {
		if err := json.Unmarshal([]byte(se.PayloadJSON), &payload); err != nil {
			return nil, err
		}
	}
	
	var metadata map[string]string
	if se.Metadata != "" {
		if err := json.Unmarshal([]byte(se.Metadata), &metadata); err != nil {
			return nil, err
		}
	}
	
	event := &BaseEvent{
		ID:        se.ID,
		Type:      se.Type,
		Timestamp: se.Timestamp,
		Payload:   payload,
		Metadata:  metadata,
	}
	
	return event, nil
}

// DBEventStore 基于数据库的事件存储实现
type DBEventStore struct {
	db *gorm.DB
}

// NewDBEventStore 创建数据库事件存储
func NewDBEventStore(db *gorm.DB) EventStore {
	// 自动迁移表结构
	db.AutoMigrate(&StoredEvent{})
	
	return &DBEventStore{db: db}
}

// Save 保存事件
func (s *DBEventStore) Save(event Event) error {
	payloadJSON, err := json.Marshal(event.GetPayload())
	if err != nil {
		return err
	}
	
	metadataJSON, err := json.Marshal(event.GetMetadata())
	if err != nil {
		return err
	}
	
	storedEvent := &StoredEvent{
		ID:          event.GetID(),
		Type:        event.GetType(),
		Timestamp:   event.GetTimestamp(),
		PayloadJSON: string(payloadJSON),
		Metadata:    string(metadataJSON),
	}
	
	return s.db.Create(storedEvent).Error
}

// GetByID 根据ID获取事件
func (s *DBEventStore) GetByID(id string) (Event, error) {
	var storedEvent StoredEvent
	
	if err := s.db.Where("id = ?", id).First(&storedEvent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	
	return storedEvent.ToEvent()
}

// GetByType 根据类型获取事件
func (s *DBEventStore) GetByType(eventType string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	
	var storedEvents []StoredEvent
	
	if err := s.db.Where("type = ?", eventType).
		Order("timestamp DESC").
		Limit(limit).
		Find(&storedEvents).Error; err != nil {
		return nil, err
	}
	
	events := make([]Event, len(storedEvents))
	for i, se := range storedEvents {
		event, err := se.ToEvent()
		if err != nil {
			return nil, err
		}
		events[i] = event
	}
	
	return events, nil
}

// GetByTypeAndTimeRange 根据类型和时间范围获取事件
func (s *DBEventStore) GetByTypeAndTimeRange(eventType string, start, end time.Time, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	
	var storedEvents []StoredEvent
	
	query := s.db.Where("type = ? AND timestamp >= ? AND timestamp <= ?", eventType, start, end)
	
	if err := query.Order("timestamp DESC").
		Limit(limit).
		Find(&storedEvents).Error; err != nil {
		return nil, err
	}
	
	events := make([]Event, len(storedEvents))
	for i, se := range storedEvents {
		event, err := se.ToEvent()
		if err != nil {
			return nil, err
		}
		events[i] = event
	}
	
	return events, nil
}

// GetByTimeRange 根据时间范围获取事件
func (s *DBEventStore) GetByTimeRange(start, end time.Time, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	
	var storedEvents []StoredEvent
	
	if err := s.db.Where("timestamp >= ? AND timestamp <= ?", start, end).
		Order("timestamp DESC").
		Limit(limit).
		Find(&storedEvents).Error; err != nil {
		return nil, err
	}
	
	events := make([]Event, len(storedEvents))
	for i, se := range storedEvents {
		event, err := se.ToEvent()
		if err != nil {
			return nil, err
		}
		events[i] = event
	}
	
	return events, nil
}

// Replay 重放事件
func (s *DBEventStore) Replay(from time.Time, handler EventHandler) error {
	const batchSize = 100
	var offset int
	
	for {
		var storedEvents []StoredEvent
		
		if err := s.db.Where("timestamp >= ?", from).
			Order("timestamp ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&storedEvents).Error; err != nil {
			return err
		}
		
		if len(storedEvents) == 0 {
			break
		}
		
		for _, se := range storedEvents {
			event, err := se.ToEvent()
			if err != nil {
				return err
			}
			
			if err := handler(event); err != nil {
				return err
			}
		}
		
		if len(storedEvents) < batchSize {
			break
		}
		
		offset += batchSize
	}
	
	return nil
}

// ReplayByType 按类型重放事件
func (s *DBEventStore) ReplayByType(eventType string, from time.Time, handler EventHandler) error {
	const batchSize = 100
	var offset int
	
	for {
		var storedEvents []StoredEvent
		
		if err := s.db.Where("type = ? AND timestamp >= ?", eventType, from).
			Order("timestamp ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&storedEvents).Error; err != nil {
			return err
		}
		
		if len(storedEvents) == 0 {
			break
		}
		
		for _, se := range storedEvents {
			event, err := se.ToEvent()
			if err != nil {
				return err
			}
			
			if err := handler(event); err != nil {
				return err
			}
		}
		
		
		if len(storedEvents) < batchSize {
			break
		}
		
		offset += batchSize
	}
	
	return nil
}

// CleanupOldEvents 清理过期事件
func (s *DBEventStore) CleanupOldEvents(before time.Time) error {
	return s.db.Where("timestamp < ?", before).Delete(&StoredEvent{}).Error
}

// CountByType 按类型统计事件数量
func (s *DBEventStore) CountByType(eventType string) (int64, error) {
	var count int64
	err := s.db.Model(&StoredEvent{}).Where("type = ?", eventType).Count(&count).Error
	return count, err
}

// CountByTimeRange 按时间范围统计事件数量
func (s *DBEventStore) CountByTimeRange(start, end time.Time) (int64, error) {
	var count int64
	err := s.db.Model(&StoredEvent{}).
		Where("timestamp >= ? AND timestamp <= ?", start, end).
		Count(&count).Error
	return count, err
}