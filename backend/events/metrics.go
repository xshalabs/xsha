package events

import (
	"sync"
	"time"
)

// EventMetrics 事件指标收集器
type EventMetrics struct {
	// 事件发布统计
	publishedCount map[string]int64
	
	// 事件处理统计
	processedCount map[string]map[string]int64 // [eventType][status]count
	
	// 错误统计
	errorCount map[string]map[string]int64 // [eventType][errorType]count
	
	// 延迟统计
	latencyStats map[string]*LatencyStats
	
	// 重试统计
	retrySuccessCount map[string]int64
	retryFailedCount  map[string]int64
	
	// 死信队列统计
	deadLetterCount map[string]int64
	
	// 丢弃事件统计
	droppedCount map[string]int64
	
	mu sync.RWMutex
}

// LatencyStats 延迟统计
type LatencyStats struct {
	Count    int64
	Total    time.Duration
	Min      time.Duration
	Max      time.Duration
	Average  time.Duration
}

// NewEventMetrics 创建事件指标收集器
func NewEventMetrics() *EventMetrics {
	return &EventMetrics{
		publishedCount:    make(map[string]int64),
		processedCount:    make(map[string]map[string]int64),
		errorCount:        make(map[string]map[string]int64),
		latencyStats:      make(map[string]*LatencyStats),
		retrySuccessCount: make(map[string]int64),
		retryFailedCount:  make(map[string]int64),
		deadLetterCount:   make(map[string]int64),
		droppedCount:      make(map[string]int64),
	}
}

// EventPublished 记录事件发布
func (m *EventMetrics) EventPublished(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.publishedCount[eventType]++
}

// EventProcessed 记录事件处理
func (m *EventMetrics) EventProcessed(eventType, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.processedCount[eventType] == nil {
		m.processedCount[eventType] = make(map[string]int64)
	}
	
	m.processedCount[eventType][status]++
}

// EventError 记录事件错误
func (m *EventMetrics) EventError(eventType, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.errorCount[eventType] == nil {
		m.errorCount[eventType] = make(map[string]int64)
	}
	
	m.errorCount[eventType][errorType]++
}

// EventLatency 记录事件处理延迟
func (m *EventMetrics) EventLatency(eventType string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	stats, exists := m.latencyStats[eventType]
	if !exists {
		stats = &LatencyStats{
			Min: latency,
			Max: latency,
		}
		m.latencyStats[eventType] = stats
	}
	
	stats.Count++
	stats.Total += latency
	stats.Average = time.Duration(int64(stats.Total) / stats.Count)
	
	if latency < stats.Min {
		stats.Min = latency
	}
	if latency > stats.Max {
		stats.Max = latency
	}
}

// EventRetrySuccess 记录重试成功
func (m *EventMetrics) EventRetrySuccess(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.retrySuccessCount[eventType]++
}

// EventRetryFailed 记录重试失败
func (m *EventMetrics) EventRetryFailed(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.retryFailedCount[eventType]++
}

// EventDeadLetter 记录死信事件
func (m *EventMetrics) EventDeadLetter(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.deadLetterCount[eventType]++
}

// EventDropped 记录丢弃事件
func (m *EventMetrics) EventDropped(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.droppedCount[eventType]++
}

// GetPublishedCount 获取发布统计
func (m *EventMetrics) GetPublishedCount() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]int64)
	for k, v := range m.publishedCount {
		result[k] = v
	}
	return result
}

// GetProcessedCount 获取处理统计
func (m *EventMetrics) GetProcessedCount() map[string]map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]map[string]int64)
	for eventType, statusMap := range m.processedCount {
		result[eventType] = make(map[string]int64)
		for status, count := range statusMap {
			result[eventType][status] = count
		}
	}
	return result
}

// GetErrorCount 获取错误统计
func (m *EventMetrics) GetErrorCount() map[string]map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]map[string]int64)
	for eventType, errorMap := range m.errorCount {
		result[eventType] = make(map[string]int64)
		for errorType, count := range errorMap {
			result[eventType][errorType] = count
		}
	}
	return result
}

// GetLatencyStats 获取延迟统计
func (m *EventMetrics) GetLatencyStats() map[string]*LatencyStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]*LatencyStats)
	for eventType, stats := range m.latencyStats {
		result[eventType] = &LatencyStats{
			Count:   stats.Count,
			Total:   stats.Total,
			Min:     stats.Min,
			Max:     stats.Max,
			Average: stats.Average,
		}
	}
	return result
}

// GetRetryStats 获取重试统计
func (m *EventMetrics) GetRetryStats() (map[string]int64, map[string]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	successCount := make(map[string]int64)
	failedCount := make(map[string]int64)
	
	for k, v := range m.retrySuccessCount {
		successCount[k] = v
	}
	
	for k, v := range m.retryFailedCount {
		failedCount[k] = v
	}
	
	return successCount, failedCount
}

// GetDeadLetterCount 获取死信统计
func (m *EventMetrics) GetDeadLetterCount() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]int64)
	for k, v := range m.deadLetterCount {
		result[k] = v
	}
	return result
}

// GetDroppedCount 获取丢弃统计
func (m *EventMetrics) GetDroppedCount() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]int64)
	for k, v := range m.droppedCount {
		result[k] = v
	}
	return result
}

// Reset 重置所有统计
func (m *EventMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.publishedCount = make(map[string]int64)
	m.processedCount = make(map[string]map[string]int64)
	m.errorCount = make(map[string]map[string]int64)
	m.latencyStats = make(map[string]*LatencyStats)
	m.retrySuccessCount = make(map[string]int64)
	m.retryFailedCount = make(map[string]int64)
	m.deadLetterCount = make(map[string]int64)
	m.droppedCount = make(map[string]int64)
}

// GetSummary 获取统计摘要
func (m *EventMetrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	summary := make(map[string]interface{})
	
	// 总体统计
	totalPublished := int64(0)
	totalProcessed := int64(0)
	totalErrors := int64(0)
	
	for _, count := range m.publishedCount {
		totalPublished += count
	}
	
	for _, statusMap := range m.processedCount {
		for _, count := range statusMap {
			totalProcessed += count
		}
	}
	
	for _, errorMap := range m.errorCount {
		for _, count := range errorMap {
			totalErrors += count
		}
	}
	
	summary["total_published"] = totalPublished
	summary["total_processed"] = totalProcessed
	summary["total_errors"] = totalErrors
	summary["success_rate"] = float64(totalProcessed-totalErrors) / float64(totalProcessed)
	
	// 详细统计
	summary["published"] = m.GetPublishedCount()
	summary["processed"] = m.GetProcessedCount()
	summary["errors"] = m.GetErrorCount()
	summary["latency"] = m.GetLatencyStats()
	
	retrySuccess, retryFailed := m.GetRetryStats()
	summary["retry_success"] = retrySuccess
	summary["retry_failed"] = retryFailed
	
	summary["dead_letter"] = m.GetDeadLetterCount()
	summary["dropped"] = m.GetDroppedCount()
	
	return summary
}