package result_parser

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/services/executor/result_parser/strategies"
	"xsha-backend/services/executor/result_parser/validator"
	"xsha-backend/utils"
)

// Parser 解析器接口
type Parser interface {
	// ParseFromLogs 从日志中解析结果
	ParseFromLogs(executionLogs string) (map[string]interface{}, error)
	
	// ParseFromLogsWithContext 带上下文的日志解析
	ParseFromLogsWithContext(ctx context.Context, executionLogs string) (map[string]interface{}, error)
	
	// ParseAndCreate 解析并创建结果记录
	ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog)
	
	// GetMetrics 获取解析指标
	GetMetrics() *Metrics
}

// DefaultParser 默认解析器实现
type DefaultParser struct {
	config          *Config
	strategyFactory StrategyFactory
	strategy        strategies.ParseStrategy // 固定策略（可选）
	validator       validator.Validator
	metrics         *Metrics
	
	// 依赖服务（用于向后兼容）
	taskConvResultRepo    repository.TaskConversationResultRepository
	taskConvResultService services.TaskConversationResultService
	taskService           services.TaskService
}

// NewDefaultParser 创建默认解析器
func NewDefaultParser(
	taskConvResultRepo repository.TaskConversationResultRepository,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
) Parser {
	config := DefaultConfig()
	config.LoadFromEnv()
	config.Validate()
	
	return &DefaultParser{
		config:                config,
		strategyFactory:       NewDefaultStrategyFactory(config),
		validator:             validator.NewResultValidator(config.StrictValidation),
		metrics:               NewMetrics(),
		taskConvResultRepo:    taskConvResultRepo,
		taskConvResultService: taskConvResultService,
		taskService:           taskService,
	}
}

// ParseFromLogs 从日志中解析结果
func (p *DefaultParser) ParseFromLogs(executionLogs string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ParseTimeout)
	defer cancel()
	
	return p.ParseFromLogsWithContext(ctx, executionLogs)
}

// ParseFromLogsWithContext 带上下文的日志解析
func (p *DefaultParser) ParseFromLogsWithContext(ctx context.Context, executionLogs string) (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		p.metrics.RecordParseAttempt(time.Since(startTime))
	}()
	
	if executionLogs == "" {
		p.metrics.RecordParseError("empty_logs")
		return nil, errors.New("execution logs are empty")
	}
	
	// 选择解析策略
	strategy := p.selectStrategy(executionLogs)
	p.metrics.RecordStrategyUsage(strategy.Name())
	
	// 尝试解析
	var lastErr error
	for attempt := 0; attempt <= p.config.RetryAttempts; attempt++ {
		select {
		case <-ctx.Done():
			p.metrics.RecordParseError("timeout")
			return nil, ctx.Err()
		default:
		}
		
		result, err := strategy.Parse(ctx, executionLogs)
		if err != nil {
			lastErr = err
			p.metrics.RecordRetry()
			
			// 如果不是最后一次尝试，稍等片刻再重试
			if attempt < p.config.RetryAttempts {
				time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
			}
			continue
		}
		
		// 验证解析结果
		if err := p.validateResult(result); err != nil {
			p.metrics.RecordValidationError()
			if p.config.StrictValidation {
				return nil, fmt.Errorf("validation failed: %v", err)
			}
			
			// 在非严格模式下，记录警告但继续
			utils.Warn("Result validation failed, continuing in non-strict mode",
				"error", err,
				"result", result)
		}
		
		p.metrics.RecordParseSuccess(strategy.Name())
		return result, nil
	}
	
	p.metrics.RecordParseError("max_retries_exceeded")
	return nil, fmt.Errorf("failed to parse after %d attempts, last error: %v", 
		p.config.RetryAttempts+1, lastErr)
}

// ParseAndCreate 解析并创建结果记录
func (p *DefaultParser) ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	startTime := time.Now()
	
	resultData, err := p.ParseFromLogs(execLog.ExecutionLogs)
	if err != nil {
		utils.Warn("Failed to parse execution result from logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID,
			"error", err,
			"duration", time.Since(startTime))
		return
	}
	
	if resultData == nil {
		utils.Info("No result data found in execution logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID,
			"duration", time.Since(startTime))
		return
	}
	
	// 检查是否已存在结果记录
	if p.taskConvResultRepo != nil {
		exists, err := p.taskConvResultRepo.ExistsByConversationID(conv.ID)
		if err != nil {
			utils.Error("Failed to check existing task conversation result",
				"conversation_id", conv.ID,
				"error", err)
			return
		}
		
		if exists {
			utils.Info("Task conversation result already exists, skipping creation",
				"conversation_id", conv.ID)
			return
		}
	}
	
	// 创建结果记录
	if p.taskConvResultService != nil {
		result, err := p.taskConvResultService.CreateResult(conv.ID, resultData)
		if err != nil {
			utils.Error("Failed to create task conversation result",
				"conversation_id", conv.ID,
				"error", err)
			return
		}
		
		utils.Info("Successfully created task conversation result",
			"conversation_id", conv.ID,
			"result_id", result.ID,
			"duration", time.Since(startTime))
		
		// 更新任务的会话ID
		if result.SessionID != "" && conv.Task != nil && p.taskService != nil {
			err = p.taskService.UpdateTaskSessionID(conv.Task.ID, result.SessionID)
			if err != nil {
				utils.Error("Failed to update task session ID",
					"task_id", conv.Task.ID,
					"session_id", result.SessionID,
					"error", err)
			} else {
				utils.Info("Successfully updated task session ID",
					"task_id", conv.Task.ID,
					"session_id", result.SessionID)
			}
		}
	}
}

// GetMetrics 获取解析指标
func (p *DefaultParser) GetMetrics() *Metrics {
	return p.metrics
}

// selectStrategy 选择解析策略
func (p *DefaultParser) selectStrategy(logs string) strategies.ParseStrategy {
	// 如果设置了固定策略，直接使用
	if p.strategy != nil {
		return p.strategy
	}
	
	// 否则使用工厂选择最佳策略
	return p.strategyFactory.GetBestStrategy(logs)
}

// validateResult 验证解析结果
func (p *DefaultParser) validateResult(result map[string]interface{}) error {
	if p.validator == nil {
		return nil
	}
	
	if p.config.AllowPartialData {
		return p.validator.ValidatePartial(result)
	}
	
	return p.validator.Validate(result)
}

// Metrics 解析指标
type Metrics struct {
	parseAttempts     int64
	parseSuccesses    int64
	parseErrors       int64
	retryCount        int64
	validationErrors  int64
	totalParseTime    int64 // 纳秒
	strategyUsage     map[string]int64
	errorTypes        map[string]int64
}

// NewMetrics 创建新的指标实例
func NewMetrics() *Metrics {
	return &Metrics{
		strategyUsage: make(map[string]int64),
		errorTypes:    make(map[string]int64),
	}
}

// RecordParseAttempt 记录解析尝试
func (m *Metrics) RecordParseAttempt(duration time.Duration) {
	atomic.AddInt64(&m.parseAttempts, 1)
	atomic.AddInt64(&m.totalParseTime, int64(duration))
}

// RecordParseSuccess 记录解析成功
func (m *Metrics) RecordParseSuccess(strategy string) {
	atomic.AddInt64(&m.parseSuccesses, 1)
	m.strategyUsage[strategy]++
}

// RecordParseError 记录解析错误
func (m *Metrics) RecordParseError(errorType string) {
	atomic.AddInt64(&m.parseErrors, 1)
	m.errorTypes[errorType]++
}

// RecordRetry 记录重试
func (m *Metrics) RecordRetry() {
	atomic.AddInt64(&m.retryCount, 1)
}

// RecordValidationError 记录验证错误
func (m *Metrics) RecordValidationError() {
	atomic.AddInt64(&m.validationErrors, 1)
}

// RecordStrategyUsage 记录策略使用
func (m *Metrics) RecordStrategyUsage(strategy string) {
	m.strategyUsage[strategy]++
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() map[string]interface{} {
	attempts := atomic.LoadInt64(&m.parseAttempts)
	successes := atomic.LoadInt64(&m.parseSuccesses)
	errors := atomic.LoadInt64(&m.parseErrors)
	retries := atomic.LoadInt64(&m.retryCount)
	validationErrors := atomic.LoadInt64(&m.validationErrors)
	totalTime := atomic.LoadInt64(&m.totalParseTime)
	
	successRate := 0.0
	if attempts > 0 {
		successRate = float64(successes) / float64(attempts)
	}
	
	avgTime := int64(0)
	if attempts > 0 {
		avgTime = totalTime / attempts
	}
	
	return map[string]interface{}{
		"parse_attempts":     attempts,
		"parse_successes":    successes,
		"parse_errors":       errors,
		"retry_count":        retries,
		"validation_errors":  validationErrors,
		"success_rate":       successRate,
		"avg_parse_time_ns":  avgTime,
		"avg_parse_time_ms":  float64(avgTime) / 1000000,
		"strategy_usage":     m.strategyUsage,
		"error_types":        m.errorTypes,
	}
}

// Reset 重置指标
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.parseAttempts, 0)
	atomic.StoreInt64(&m.parseSuccesses, 0)
	atomic.StoreInt64(&m.parseErrors, 0)
	atomic.StoreInt64(&m.retryCount, 0)
	atomic.StoreInt64(&m.validationErrors, 0)
	atomic.StoreInt64(&m.totalParseTime, 0)
	
	m.strategyUsage = make(map[string]int64)
	m.errorTypes = make(map[string]int64)
}

// StreamingParser 流式解析器
type StreamingParser struct {
	*DefaultParser
	bufferSize int
}

// NewStreamingParser 创建流式解析器
func NewStreamingParser(baseParser *DefaultParser, bufferSize int) *StreamingParser {
	return &StreamingParser{
		DefaultParser: baseParser,
		bufferSize:    bufferSize,
	}
}

// ParseFromStream 从流中解析
func (s *StreamingParser) ParseFromStream(ctx context.Context, logStream <-chan string) (map[string]interface{}, error) {
	buffer := make([]string, 0, s.bufferSize)
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case logLine, ok := <-logStream:
			if !ok {
				// 流结束，解析缓冲区中的内容
				if len(buffer) == 0 {
					return nil, errors.New("no log data received")
				}
				
				combinedLogs := ""
				for _, line := range buffer {
					combinedLogs += line + "\n"
				}
				
				return s.ParseFromLogsWithContext(ctx, combinedLogs)
			}
			
			buffer = append(buffer, logLine)
			
			// 如果缓冲区满了，尝试解析
			if len(buffer) >= s.bufferSize {
				combinedLogs := ""
				for _, line := range buffer {
					combinedLogs += line + "\n"
				}
				
				// 尝试找到结果
				if result, err := s.ParseFromLogsWithContext(ctx, combinedLogs); err == nil && result != nil {
					return result, nil
				}
				
				// 移除最旧的一半数据，保留最新的
				copy(buffer, buffer[s.bufferSize/2:])
				buffer = buffer[:s.bufferSize-s.bufferSize/2]
			}
		}
	}
}

// ResultParser 结果解析器接口（向后兼容）
type ResultParser interface {
	ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog)
	ParseFromLogs(executionLogs string) (map[string]interface{}, error)
}

// NewResultParser 创建结果解析器（向后兼容）
func NewResultParser(
	taskConvResultRepo repository.TaskConversationResultRepository,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
) ResultParser {
	return NewDefaultParser(taskConvResultRepo, taskConvResultService, taskService)
}