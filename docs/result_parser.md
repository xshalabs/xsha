# Result Parser 任务执行结果解析器

## 概述

Result Parser 是 XSha 系统中负责从任务执行日志中解析和提取结构化结果数据的核心组件。它能够从多种格式的日志文件中智能识别并解析任务执行结果，支持 JSON、结构化文本等多种格式。

## 架构设计

### 核心组件

1. **Parser（解析器）**: 主要的解析接口和实现
2. **Strategy（解析策略）**: 不同格式的解析策略
3. **Factory（工厂模式）**: 创建和管理解析器和策略
4. **Validator（验证器）**: 验证解析结果的完整性和正确性
5. **Config（配置）**: 解析器的配置管理

### 文件结构

```
backend/services/executor/result_parser/
├── parser.go           # 主要解析器实现
├── config.go          # 配置管理
├── factory.go         # 工厂模式实现
├── parser_test.go     # 单元测试
├── strategies/        # 解析策略
│   ├── strategy.go    # 策略接口定义
│   ├── json_strategy.go      # JSON 解析策略
│   ├── text_strategy.go      # 文本解析策略
│   └── fallback_strategy.go  # 兜底解析策略
└── validator/         # 结果验证
    ├── validator.go   # 验证器实现
    └── rules.go      # 验证规则
```

## 核心接口

### Parser 接口

```go
type Parser interface {
    // 从日志中解析结果
    ParseFromLogs(executionLogs string) (map[string]interface{}, error)
    
    // 带上下文的日志解析（支持超时）
    ParseFromLogsWithContext(ctx context.Context, executionLogs string) (map[string]interface{}, error)
    
    // 解析并创建结果记录
    ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog)
    
    // 获取解析指标
    GetMetrics() *Metrics
}
```

### ParseStrategy 策略接口

```go
type ParseStrategy interface {
    // 策略名称
    Name() string
    
    // 检查是否能解析给定的日志内容
    CanParse(logs string) bool
    
    // 解析日志内容并返回结果数据
    Parse(ctx context.Context, logs string) (map[string]interface{}, error)
    
    // 策略优先级（数值越小优先级越高）
    Priority() int
    
    // 是否支持批量解析
    SupportsBatch() bool
    
    // 批量解析多个日志条目
    ParseBatch(ctx context.Context, logEntries []string) ([]map[string]interface{}, error)
}
```

## 解析策略

### 1. JSON 策略（优先级: 1）

**适用场景**: 日志中包含 JSON 格式的结果数据

**识别规则**:
- 包含 `"type":`、`"subtype":`、`"session_id":` 等 JSON 字段
- 匹配预定义的正则表达式模式

**正则模式**:
```go
// 带时间戳的日志行
`^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`

// 纯 JSON 行
`^(\{.*\})$`

// 带日志级别的行
`(?i)(?:info|debug|warn|error):\s*(\{.*\})`
```

**解析示例**:
```
输入: [12:34:56] INFO: {"type": "result", "subtype": "success", "is_error": false, "session_id": "test-123"}
输出: {
  "type": "result",
  "subtype": "success", 
  "is_error": false,
  "session_id": "test-123"
}
```

### 2. 优化的 JSON 策略（优先级: 1）

**特点**: 专为大日志文件优化，只扫描最后 N 行

**配置**: 通过 `MaxLogLines` 控制扫描行数（默认: 10000）

### 3. 结构化文本策略（优先级: 2）

**适用场景**: 包含 key=value 格式的结构化数据

**识别规则**:
- 包含 `type=`、`subtype=`、`session_id=` 等键值对
- 至少包含 2 个结构化指示符

**解析示例**:
```
输入: Task completed: type=result, subtype=success, is_error=false, session_id=test-456
输出: {
  "type": "result",
  "subtype": "success",
  "is_error": false,
  "session_id": "test-456"
}
```

### 4. 兜底策略（优先级: 99）

**适用场景**: 当所有其他策略都无法解析时使用

**行为**: 生成默认的结果结构，包含必需字段

**输出结构**:
```go
{
  "type": "result",
  "subtype": "unknown", 
  "is_error": true,
  "session_id": "unknown-" + timestamp
}
```

## 解析流程

1. **策略选择**: 根据日志内容选择最合适的解析策略
2. **解析执行**: 使用选定策略解析日志内容
3. **结果验证**: 验证解析结果的完整性和正确性
4. **重试机制**: 解析失败时进行重试（默认 3 次）
5. **指标记录**: 记录解析性能和成功率指标

## 配置选项

### 环境变量配置

```bash
# 最大日志行数
export XSHA_PARSER_MAX_LOG_LINES=10000

# 解析超时时间
export XSHA_PARSER_TIMEOUT=30s

# 重试次数
export XSHA_PARSER_RETRY_ATTEMPTS=3

# 严格验证模式
export XSHA_PARSER_STRICT_VALIDATION=false

# 允许部分数据
export XSHA_PARSER_ALLOW_PARTIAL_DATA=true

# 缓冲区大小
export XSHA_PARSER_BUFFER_SIZE=4096

# 启用指标收集
export XSHA_PARSER_ENABLE_METRICS=true
```

### 默认配置

```go
config := &Config{
    MaxLogLines:   10000,
    ParseTimeout:  30 * time.Second,
    RetryAttempts: 3,
    JSONLogPatterns: []string{
        `^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`,
        `^(\{.*\})$`,
    },
    RequiredFields: []string{
        "type", "subtype", "is_error", "session_id",
    },
    OptionalFields: []string{
        "duration_ms", "duration_api_ms", "num_turns", 
        "result", "total_cost_usd", "usage",
    },
    StrictValidation: false,
    AllowPartialData: true,
    BufferSize:       4096,
    EnableMetrics:    true,
}
```

## 数据验证

### 必需字段

- `type`: 结果类型（必须为 "result"）
- `subtype`: 结果子类型（如 "success", "error"）
- `is_error`: 是否为错误结果（布尔值）
- `session_id`: 会话标识符

### 可选字段

- `duration_ms`: 执行时长（毫秒）
- `duration_api_ms`: API 调用时长（毫秒）
- `num_turns`: 对话轮数
- `result`: 任务结果内容
- `total_cost_usd`: 总成本（美元）
- `usage`: 资源使用情况

### 验证模式

- **严格模式**: 所有必需字段必须存在且类型正确
- **非严格模式**: 允许缺少某些字段，记录警告但继续处理

## 使用示例

### 基本使用

```go
// 创建解析器
parser := NewDefaultParser(
    taskConvResultRepo,
    taskConvResultService, 
    taskService,
)

// 解析日志
result, err := parser.ParseFromLogs(executionLogs)
if err != nil {
    log.Error("解析失败", "error", err)
    return
}

// 处理结果
fmt.Printf("解析结果: %+v\n", result)
```

### 使用自定义配置

```go
// 自定义配置
config := &Config{
    MaxLogLines:      5000,
    ParseTimeout:     15 * time.Second,
    RetryAttempts:    5,
    StrictValidation: true,
}

// 创建解析器
parser := NewParserWithConfig(config)
```

### 使用特定策略

```go
// 创建 JSON 策略
jsonStrategy := strategies.NewJSONStrategy()

// 创建解析器并使用特定策略
parser := NewParser(WithStrategy(jsonStrategy))
```

### 批量解析

```go
// 创建批量解析器
batchParser := NewBatchParser(parser, config)

// 并发解析多个日志
results, err := batchParser.ParseBatchConcurrent(
    ctx, logEntries, 4, // 4 个并发协程
)
```

### 流式解析

```go
// 创建流式解析器
streamingParser := NewStreamingParser(parser.(*DefaultParser), 1024)

// 从日志流解析
logStream := make(chan string)
go func() {
    // 向流中发送日志行
    for _, line := range logLines {
        logStream <- line
    }
    close(logStream)
}()

result, err := streamingParser.ParseFromStream(ctx, logStream)
```

### 缓存解析器

```go
// 创建缓存解析器
cachedParser := NewCachedParser(parser, 1000) // 缓存 1000 个结果

// 解析（自动缓存）
result, err := cachedParser.ParseFromLogs(logs)

// 查看缓存统计
stats := cachedParser.GetCacheStats()
fmt.Printf("缓存命中率: %.2f%%\n", stats["hit_rate"].(float64)*100)
```

## 性能监控

### 指标类型

- **解析尝试次数**: 总的解析请求数
- **解析成功次数**: 成功解析的数量
- **解析错误次数**: 失败的解析数量
- **重试次数**: 重试的总次数
- **验证错误次数**: 验证失败的次数
- **成功率**: 解析成功率
- **平均解析时间**: 平均解析耗时
- **策略使用统计**: 各策略的使用频率
- **错误类型统计**: 各类错误的发生频率

### 获取指标

```go
metrics := parser.GetMetrics()
stats := metrics.GetStats()

fmt.Printf("解析成功率: %.2f%%\n", stats["success_rate"].(float64)*100)
fmt.Printf("平均解析时间: %.2fms\n", stats["avg_parse_time_ms"].(float64))
fmt.Printf("策略使用: %+v\n", stats["strategy_usage"])
```

## 错误处理

### 常见错误类型

- `empty_logs`: 日志为空
- `timeout`: 解析超时
- `max_retries_exceeded`: 超过最大重试次数
- `validation_failed`: 验证失败
- `parse_error`: 解析错误

### 错误处理策略

1. **重试机制**: 解析失败时自动重试
2. **降级策略**: 高级策略失败时使用兜底策略
3. **容错模式**: 非严格模式下允许部分数据缺失
4. **日志记录**: 详细记录错误信息用于调试

## 扩展指南

### 添加新的解析策略

1. **实现 ParseStrategy 接口**:

```go
type CustomStrategy struct {
    name     string
    priority int
}

func (s *CustomStrategy) Name() string {
    return s.name
}

func (s *CustomStrategy) CanParse(logs string) bool {
    // 实现检测逻辑
    return true
}

func (s *CustomStrategy) Parse(ctx context.Context, logs string) (map[string]interface{}, error) {
    // 实现解析逻辑
    return result, nil
}

func (s *CustomStrategy) Priority() int {
    return s.priority
}
```

2. **注册策略**:

```go
factory := NewDefaultStrategyFactory(config)
factory.RegisterStrategy(&CustomStrategy{
    name:     "custom",
    priority: 5,
})
```

### 自定义验证规则

```go
type CustomValidator struct {
    // 自定义字段
}

func (v *CustomValidator) Validate(data map[string]interface{}) error {
    // 实现验证逻辑
    return nil
}

// 使用自定义验证器
parser := NewParser(WithValidator(customValidator))
```

## 最佳实践

1. **合理设置超时时间**: 根据日志大小调整解析超时
2. **选择合适的缓冲区大小**: 平衡内存使用和性能
3. **监控解析指标**: 定期检查解析成功率和性能
4. **使用缓存**: 对于重复解析相同内容，启用缓存
5. **错误日志**: 保留详细的错误日志便于问题排查
6. **批量处理**: 对于大量日志，使用批量解析提高效率

## 故障排查

### 常见问题

1. **解析失败率高**: 
   - 检查日志格式是否符合预期
   - 调整正则表达式模式
   - 启用非严格验证模式

2. **解析超时**:
   - 减少 MaxLogLines 限制
   - 增加 ParseTimeout 时间
   - 使用优化的解析策略

3. **内存使用过高**:
   - 调整 BufferSize 大小
   - 使用流式解析
   - 限制并发解析数量

### 调试方法

1. **启用详细日志**: 设置日志级别为 DEBUG
2. **检查指标**: 定期查看解析指标和错误统计
3. **单元测试**: 使用测试用例验证解析逻辑
4. **性能分析**: 使用基准测试找出性能瓶颈

## 总结

Result Parser 是 XSha 系统中的关键组件，通过灵活的策略模式和完善的错误处理机制，能够可靠地从各种格式的执行日志中提取结构化结果数据。其模块化的设计使得系统易于扩展和维护，而丰富的配置选项和监控指标则确保了在生产环境中的稳定运行。

新程序员在使用时，建议从基本的使用示例开始，逐步了解各个组件的功能，并根据实际需求进行配置调整和功能扩展。