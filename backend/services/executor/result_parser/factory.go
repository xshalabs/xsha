package result_parser

import (
	"context"
	"fmt"
	"sort"
	"time"
	"xsha-backend/services/executor/result_parser/strategies"
	"xsha-backend/services/executor/result_parser/validator"
)

// StrategyFactory 策略工厂接口
type StrategyFactory interface {
	// CreateStrategy 根据日志内容创建合适的解析策略
	CreateStrategy(logs string) strategies.ParseStrategy
	
	// CreateStrategies 创建所有可用的策略
	CreateStrategies() []strategies.ParseStrategy
	
	// RegisterStrategy 注册自定义策略
	RegisterStrategy(strategy strategies.ParseStrategy)
	
	// GetBestStrategy 获取最适合的策略
	GetBestStrategy(logs string) strategies.ParseStrategy
}

// DefaultStrategyFactory 默认策略工厂
type DefaultStrategyFactory struct {
	strategies      []strategies.ParseStrategy
	customStrategies []strategies.ParseStrategy
	config          *Config
}

// NewDefaultStrategyFactory 创建默认策略工厂
func NewDefaultStrategyFactory(config *Config) *DefaultStrategyFactory {
	factory := &DefaultStrategyFactory{
		config: config,
	}
	
	// 创建默认策略
	factory.initializeDefaultStrategies()
	
	return factory
}

// initializeDefaultStrategies 初始化默认策略
func (f *DefaultStrategyFactory) initializeDefaultStrategies() {
	// JSON策略（高优先级）
	jsonStrategy := strategies.NewJSONStrategy(f.config.JSONLogPatterns...)
	
	// 优化的JSON策略（用于大日志）
	optimizedJSONStrategy := strategies.NewOptimizedJSONStrategy(
		f.config.MaxLogLines,
		f.config.JSONLogPatterns...,
	)
	
	// 文本策略（中等优先级）
	textStrategy := strategies.NewTextStrategy()
	
	// 兜底策略（最低优先级）
	fallbackStrategy := strategies.NewFallbackStrategy()
	
	f.strategies = []strategies.ParseStrategy{
		optimizedJSONStrategy,
		jsonStrategy,
		textStrategy,
		fallbackStrategy,
	}
}

// CreateStrategy 根据日志内容创建合适的解析策略
func (f *DefaultStrategyFactory) CreateStrategy(logs string) strategies.ParseStrategy {
	return f.GetBestStrategy(logs)
}

// CreateStrategies 创建所有可用的策略
func (f *DefaultStrategyFactory) CreateStrategies() []strategies.ParseStrategy {
	allStrategies := make([]strategies.ParseStrategy, 0, len(f.strategies)+len(f.customStrategies))
	allStrategies = append(allStrategies, f.strategies...)
	allStrategies = append(allStrategies, f.customStrategies...)
	
	// 按优先级排序
	sort.Slice(allStrategies, func(i, j int) bool {
		return allStrategies[i].Priority() < allStrategies[j].Priority()
	})
	
	return allStrategies
}

// RegisterStrategy 注册自定义策略
func (f *DefaultStrategyFactory) RegisterStrategy(strategy strategies.ParseStrategy) {
	f.customStrategies = append(f.customStrategies, strategy)
}

// GetBestStrategy 获取最适合的策略
func (f *DefaultStrategyFactory) GetBestStrategy(logs string) strategies.ParseStrategy {
	allStrategies := f.CreateStrategies()
	
	// 找到第一个可以解析的策略
	for _, strategy := range allStrategies {
		if strategy.CanParse(logs) {
			return strategy
		}
	}
	
	// 如果没有策略可以解析，返回兜底策略
	return strategies.NewFallbackStrategy()
}

// ParserFactory 解析器工厂接口
type ParserFactory interface {
	// CreateParser 创建解析器
	CreateParser(options ...ParserOption) Parser
	
	// CreateParserWithStrategy 使用指定策略创建解析器
	CreateParserWithStrategy(strategy strategies.ParseStrategy, options ...ParserOption) Parser
	
	// CreateParserWithValidator 使用指定验证器创建解析器
	CreateParserWithValidator(validator validator.Validator, options ...ParserOption) Parser
}

// DefaultParserFactory 默认解析器工厂
type DefaultParserFactory struct {
	config          *Config
	strategyFactory StrategyFactory
	validator       validator.Validator
}

// NewDefaultParserFactory 创建默认解析器工厂
func NewDefaultParserFactory(config *Config) *DefaultParserFactory {
	if config == nil {
		config = DefaultConfig()
		config.LoadFromEnv()
		config.Validate()
	}
	
	return &DefaultParserFactory{
		config:          config,
		strategyFactory: NewDefaultStrategyFactory(config),
		validator:       validator.NewResultValidator(config.StrictValidation),
	}
}

// CreateParser 创建解析器
func (f *DefaultParserFactory) CreateParser(options ...ParserOption) Parser {
	parser := &DefaultParser{
		config:          f.config,
		strategyFactory: f.strategyFactory,
		validator:       f.validator,
		metrics:         NewMetrics(),
	}
	
	// 应用选项
	for _, option := range options {
		option(parser)
	}
	
	return parser
}

// CreateParserWithStrategy 使用指定策略创建解析器
func (f *DefaultParserFactory) CreateParserWithStrategy(strategy strategies.ParseStrategy, options ...ParserOption) Parser {
	return f.CreateParser(append(options, WithStrategy(strategy))...)
}

// CreateParserWithValidator 使用指定验证器创建解析器
func (f *DefaultParserFactory) CreateParserWithValidator(val validator.Validator, options ...ParserOption) Parser {
	return f.CreateParser(append(options, WithValidator(val))...)
}

// ParserOption 解析器选项
type ParserOption func(*DefaultParser)

// WithConfig 设置配置选项
func WithConfig(config *Config) ParserOption {
	return func(p *DefaultParser) {
		p.config = config
	}
}

// WithStrategy 设置策略选项
func WithStrategy(strategy strategies.ParseStrategy) ParserOption {
	return func(p *DefaultParser) {
		p.strategy = strategy
	}
}

// WithValidator 设置验证器选项
func WithValidator(validator validator.Validator) ParserOption {
	return func(p *DefaultParser) {
		p.validator = validator
	}
}

// WithMetrics 设置指标选项
func WithMetrics(metrics *Metrics) ParserOption {
	return func(p *DefaultParser) {
		p.metrics = metrics
	}
}

// WithTimeout 设置超时选项
func WithTimeout(timeout time.Duration) ParserOption {
	return func(p *DefaultParser) {
		p.config.ParseTimeout = timeout
	}
}

// WithRetryAttempts 设置重试次数选项
func WithRetryAttempts(attempts int) ParserOption {
	return func(p *DefaultParser) {
		p.config.RetryAttempts = attempts
	}
}

// WithStrictValidation 设置严格验证选项
func WithStrictValidation(strict bool) ParserOption {
	return func(p *DefaultParser) {
		p.config.StrictValidation = strict
		p.validator = validator.NewResultValidator(strict)
	}
}

// AdaptiveStrategyFactory 自适应策略工厂
type AdaptiveStrategyFactory struct {
	*DefaultStrategyFactory
	cache    map[string]strategies.ParseStrategy
	maxCache int
}

// NewAdaptiveStrategyFactory 创建自适应策略工厂
func NewAdaptiveStrategyFactory(config *Config, maxCache int) *AdaptiveStrategyFactory {
	return &AdaptiveStrategyFactory{
		DefaultStrategyFactory: NewDefaultStrategyFactory(config),
		cache:                  make(map[string]strategies.ParseStrategy),
		maxCache:               maxCache,
	}
}

// GetBestStrategy 获取最适合的策略（带缓存）
func (f *AdaptiveStrategyFactory) GetBestStrategy(logs string) strategies.ParseStrategy {
	// 生成缓存键
	cacheKey := f.generateCacheKey(logs)
	
	// 检查缓存
	if strategy, exists := f.cache[cacheKey]; exists {
		return strategy
	}
	
	// 获取最佳策略
	strategy := f.DefaultStrategyFactory.GetBestStrategy(logs)
	
	// 缓存策略
	if len(f.cache) < f.maxCache {
		f.cache[cacheKey] = strategy
	}
	
	return strategy
}

// generateCacheKey 生成缓存键
func (f *AdaptiveStrategyFactory) generateCacheKey(logs string) string {
	// 使用日志格式作为缓存键
	format := strategies.DetectLogFormat(logs)
	return format.String()
}

// BatchParser 批量解析器
type BatchParser struct {
	parser Parser
	config *Config
}

// NewBatchParser 创建批量解析器
func NewBatchParser(parser Parser, config *Config) *BatchParser {
	return &BatchParser{
		parser: parser,
		config: config,
	}
}

// ParseBatch 批量解析
func (b *BatchParser) ParseBatch(ctx context.Context, logEntries []string) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, len(logEntries))
	
	for i, logs := range logEntries {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}
		
		result, err := b.parser.ParseFromLogs(logs)
		if err != nil {
			// 记录错误但继续处理
			fmt.Printf("Failed to parse log entry %d: %v\n", i, err)
			continue
		}
		
		if result != nil {
			results = append(results, result)
		}
	}
	
	return results, nil
}

// ParseBatchConcurrent 并发批量解析
func (b *BatchParser) ParseBatchConcurrent(ctx context.Context, logEntries []string, concurrency int) ([]map[string]interface{}, error) {
	if concurrency <= 0 {
		concurrency = 1
	}
	
	if len(logEntries) == 0 {
		return []map[string]interface{}{}, nil
	}
	
	// 创建工作通道
	jobs := make(chan int, len(logEntries))
	results := make(chan strategies.ParseResult, len(logEntries))
	
	// 启动工作协程
	for w := 0; w < concurrency; w++ {
		go func() {
			for jobIndex := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				
				logs := logEntries[jobIndex]
				startTime := time.Now()
				
				data, err := b.parser.ParseFromLogs(logs)
				
				results <- strategies.ParseResult{
					Data:     data,
					Strategy: fmt.Sprintf("batch_worker_%d", w),
					ParseTime: startTime,
					Duration: time.Since(startTime),
					Error:    err,
				}
			}
		}()
	}
	
	// 发送任务
	for i := 0; i < len(logEntries); i++ {
		select {
		case <-ctx.Done():
			close(jobs)
			return nil, ctx.Err()
		case jobs <- i:
		}
	}
	close(jobs)
	
	// 收集结果
	var parseResults []map[string]interface{}
	for i := 0; i < len(logEntries); i++ {
		select {
		case <-ctx.Done():
			return parseResults, ctx.Err()
		case result := <-results:
			if result.Error == nil && result.Data != nil {
				parseResults = append(parseResults, result.Data)
			}
		}
	}
	
	return parseResults, nil
}

// CachedParser 缓存解析器
type CachedParser struct {
	parser    Parser
	cache     map[string]map[string]interface{}
	maxCache  int
	hitCount  int
	missCount int
}

// NewCachedParser 创建缓存解析器
func NewCachedParser(parser Parser, maxCache int) *CachedParser {
	return &CachedParser{
		parser:   parser,
		cache:    make(map[string]map[string]interface{}),
		maxCache: maxCache,
	}
}

// ParseFromLogs 解析日志（带缓存）
func (c *CachedParser) ParseFromLogs(logs string) (map[string]interface{}, error) {
	// 生成缓存键
	cacheKey := c.generateHash(logs)
	
	// 检查缓存
	if result, exists := c.cache[cacheKey]; exists {
		c.hitCount++
		return result, nil
	}
	
	// 解析日志
	result, err := c.parser.ParseFromLogs(logs)
	if err != nil {
		c.missCount++
		return nil, err
	}
	
	// 缓存结果
	if len(c.cache) < c.maxCache && result != nil {
		c.cache[cacheKey] = result
	}
	
	c.missCount++
	return result, nil
}

// GetCacheStats 获取缓存统计
func (c *CachedParser) GetCacheStats() map[string]interface{} {
	total := c.hitCount + c.missCount
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hitCount) / float64(total)
	}
	
	return map[string]interface{}{
		"hit_count":    c.hitCount,
		"miss_count":   c.missCount,
		"hit_rate":     hitRate,
		"cache_size":   len(c.cache),
		"max_cache":    c.maxCache,
	}
}

// generateHash 生成简单哈希
func (c *CachedParser) generateHash(input string) string {
	hash := int64(0)
	for _, char := range input {
		hash = hash*31 + int64(char)
	}
	
	if hash < 0 {
		hash = -hash
	}
	
	return fmt.Sprintf("%d", hash)
}

// NewParser 创建新的解析器（便捷函数）
func NewParser(options ...ParserOption) Parser {
	factory := NewDefaultParserFactory(nil)
	return factory.CreateParser(options...)
}

// NewParserWithConfig 使用配置创建解析器
func NewParserWithConfig(config *Config, options ...ParserOption) Parser {
	factory := NewDefaultParserFactory(config)
	return factory.CreateParser(options...)
}