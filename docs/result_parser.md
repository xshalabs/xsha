# Result Parser 任务执行结果解析器

## 概述

Result Parser 是 xsha 系统中负责从任务执行日志中解析和提取结构化结果数据的核心组件。经过精简重构，现在专注于支持两种核心结果类型：**计划模式（Plan Mode）结果**和**常规 JSON 结果**，提供简洁高效的解析能力。

## 架构设计

### 核心组件

1. **Parser（解析器）**: 主要的解析接口和实现
2. **Strategy（解析策略）**: 针对不同格式的解析策略
3. **Validator（验证器）**: 验证解析结果的完整性和正确性
4. **Config（配置）**: 精简的配置管理

### 文件结构

```
backend/services/executor/result_parser/
├── parser.go               # 主要解析器实现
├── config.go              # 精简的配置管理
├── plan_mode_test.go      # 计划模式测试
├── strategies/            # 解析策略
│   ├── strategy.go           # 策略接口定义
│   ├── plan_mode_strategy.go # 计划模式解析策略
│   └── json_strategy.go      # JSON 解析策略
└── validator/             # 结果验证
    ├── validator.go       # 验证器实现
    └── rules.go          # 验证规则
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
}
```

## 解析策略

### 1. 计划模式策略（优先级: 1）

**适用场景**: Claude Code 计划模式结果，包含 `ExitPlanMode` 工具调用

**识别规则**:
- 包含 `"type":"assistant"` 字段
- 包含 `"name":"ExitPlanMode"` 工具调用
- 包含 `"tool_use"` 或 `ExitPlanMode` 关键字
- 至少匹配 2 个以上的计划模式指示符

**输入示例**:
```json
{
  "type": "assistant",
  "message": {
    "content": [
      {
        "type": "tool_use",
        "name": "ExitPlanMode",
        "input": {
          "plan": "## 实施计划\n\n1. 分析需求\n2. 设计方案\n3. 实施步骤"
        }
      }
    ]
  },
  "session_id": "plan-session-123"
}
```

**输出结果**:
```json
{
  "type": "result",
  "subtype": "plan_mode",
  "is_error": false,
  "session_id": "plan-session-123",
  "result": "## 实施计划\n\n1. 分析需求\n2. 设计方案\n3. 实施步骤",
  "duration_ms": 0,
  "duration_api_ms": 0,
  "num_turns": 0,
  "total_cost_usd": 0.0,
  "usage": { /* 使用信息（如果存在） */ }
}
```

### 2. JSON 策略（优先级: 1）

**适用场景**: 日志中包含常规 JSON 格式的结果数据

**识别规则**:
- 包含 `"type":"result"`、`"subtype":`、`"session_id":` 等常规结果字段
- 匹配预定义的正则表达式模式
- **排除计划模式**: 自动检测并排除 `type=assistant` 的计划模式结果

**正则模式**:
```go
// 带时间戳的日志行
`^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`

// 纯 JSON 行
`^(\{.*\})$`

// 带日志级别的行
`(?i)(?:info|debug|warn|error):\s*(\{.*\})`
```

**性能优化**: 只扫描最后 1000 行日志以提高大文件解析性能

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

## 解析流程

1. **策略选择**: 根据日志内容选择合适的解析策略（计划模式策略优先）
2. **解析执行**: 使用选定策略解析日志内容
3. **结果验证**: 验证解析结果的完整性和正确性
4. **错误处理**: 解析失败时返回详细错误信息

### 策略选择逻辑

```go
// 按优先级顺序尝试每个策略
strategies := []ParseStrategy{
    NewPlanModeStrategy(),  // 优先级: 1
    NewJSONStrategy(),      // 优先级: 1
}

// 选择第一个能够解析的策略
for _, strategy := range strategies {
    if strategy.CanParse(logs) {
        return strategy
    }
}
```

## 精简配置

### 配置结构

```go
type Config struct {
    // 基本配置
    ParseTimeout time.Duration `json:"parse_timeout"`
    
    // 验证配置
    StrictValidation bool `json:"strict_validation"`
    RequiredFields  []string `json:"required_fields"`
}
```

### 环境变量配置

```bash
# 解析超时时间（默认: 30s）
export XSHA_PARSER_TIMEOUT=30s

# 严格验证模式（默认: false）
export XSHA_PARSER_STRICT_VALIDATION=false
```

### 默认配置

```go
config := &Config{
    ParseTimeout: 30 * time.Second,
    StrictValidation: false,
    RequiredFields: []string{
        "type", "subtype", "is_error", "session_id",
    },
}
```

## 数据验证

### 通用必需字段

- `type`: 结果类型（必须为 "result"）
- `subtype`: 结果子类型（如 "success", "error", "plan_mode"）
- `is_error`: 是否为错误结果（布尔值）
- `session_id`: 会话标识符

### 计划模式专用验证

**计划模式结果**（`subtype: "plan_mode"`）的特殊验证规则：

- `type`: 必须为 "result"
- `subtype`: 必须为 "plan_mode"
- `result`: 必须包含计划内容（非空字符串）
- `session_id`: 必须存在且非空
- **自动设置默认值**:
  - `duration_ms`: 0
  - `duration_api_ms`: 0
  - `num_turns`: 0
  - `total_cost_usd`: 0.0

### 常规结果字段

- `duration_ms`: 执行时长（毫秒）
- `duration_api_ms`: API 调用时长（毫秒）
- `num_turns`: 对话轮数
- `result`: 任务结果内容
- `total_cost_usd`: 总成本（美元）
- `usage`: 资源使用情况

### 验证模式

- **严格模式**: 所有必需字段必须存在且类型正确
- **非严格模式**: 允许缺少某些字段，记录警告但继续处理
- **计划模式验证**: 使用专用验证逻辑，更宽松的字段要求

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

### 计划模式解析示例

```go
// 计划模式日志内容
planModeLog := `{
  "type": "assistant",
  "message": {
    "content": [{
      "type": "tool_use",
      "name": "ExitPlanMode",
      "input": {
        "plan": "## 项目实施计划\n\n1. 需求分析\n2. 架构设计\n3. 编码实现"
      }
    }]
  },
  "session_id": "plan-session-123"
}`

// 解析计划模式结果
result, err := parser.ParseFromLogs(planModeLog)
if err != nil {
    log.Error("计划模式解析失败", "error", err)
    return
}

// 验证计划模式结果
if result["subtype"] == "plan_mode" {
    planContent := result["result"].(string)
    sessionID := result["session_id"].(string)
    
    fmt.Printf("解析到计划模式结果:\n")
    fmt.Printf("会话ID: %s\n", sessionID)
    fmt.Printf("计划内容: %s\n", planContent)
}
```

### 使用自定义配置

```go
// 自定义配置
config := &Config{
    ParseTimeout:     15 * time.Second,
    StrictValidation: true,
}

// 应用配置
config.LoadFromEnv()
config.Validate()

// 创建解析器
parser := NewDefaultParser(taskConvResultRepo, taskConvResultService, taskService)
```

### 带上下文的解析

```go
// 设置解析超时
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 带超时的解析
result, err := parser.ParseFromLogsWithContext(ctx, executionLogs)
if err != nil {
    if err == context.DeadlineExceeded {
        log.Error("解析超时")
    } else {
        log.Error("解析失败", "error", err)
    }
    return
}
```

## 错误处理

### 常见错误类型

- `empty_logs`: 日志为空
- `timeout`: 解析超时
- `validation_failed`: 验证失败
- `parse_error`: 解析错误
- `no_valid_result_found`: 未找到有效结果
- `missing_plan_content`: 缺少计划内容（计划模式专用）
- `invalid_exit_plan_mode`: 无效的 ExitPlanMode 结构（计划模式专用）

### 错误处理策略

1. **简洁错误处理**: 移除复杂的重试机制，直接返回明确的错误信息
2. **策略降级**: 计划模式解析失败时尝试常规 JSON 解析
3. **容错模式**: 非严格模式下允许部分数据缺失
4. **详细日志**: 记录详细错误信息用于调试

## 性能特性

### 优化措施

1. **限制扫描范围**: JSON 策略只扫描最后 1000 行日志
2. **简化架构**: 移除复杂的工厂模式和缓存机制
3. **直接策略选择**: 简化策略选择逻辑
4. **减少内存分配**: 优化字符串处理和数据结构

### 性能指标

- **解析时间**: 大幅减少解析耗时
- **内存使用**: 显著降低内存占用
- **代码复杂度**: 代码量减少约 60%

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
    return containsCustomFormat(logs)
}

func (s *CustomStrategy) Parse(ctx context.Context, logs string) (map[string]interface{}, error) {
    // 实现解析逻辑
    return parseCustomFormat(logs), nil
}

func (s *CustomStrategy) Priority() int {
    return s.priority
}
```

2. **修改解析器创建逻辑**:

```go
// 在 NewDefaultParser 中添加自定义策略
strategies := []strategies.ParseStrategy{
    strategies.NewPlanModeStrategy(),
    strategies.NewJSONStrategy(),
    &CustomStrategy{name: "custom", priority: 2}, // 添加自定义策略
}
```

### 自定义验证规则

```go
type CustomValidator struct {
    strictMode bool
}

func (v *CustomValidator) Validate(data map[string]interface{}) error {
    // 实现自定义验证逻辑
    if v.strictMode {
        return validateStrict(data)
    }
    return validateLenient(data)
}
```

## 最佳实践

1. **合理设置超时时间**: 根据日志大小调整解析超时
2. **选择合适的验证模式**: 生产环境建议使用非严格模式
3. **错误日志**: 保留详细的错误日志便于问题排查
4. **结果类型检查**: 在业务逻辑中区分处理计划模式和常规结果
5. **上下文使用**: 对于可能耗时的操作使用带超时的上下文

## 故障排查

### 常见问题

1. **计划模式检测失败**:
   - 验证日志中是否包含 `"type":"assistant"` 字段
   - 确认存在 `"name":"ExitPlanMode"` 工具调用
   - 检查计划内容是否在 `input.plan` 字段中

2. **常规 JSON 解析失败**:
   - 检查 JSON 格式是否正确
   - 验证必需字段是否存在
   - 确认不是计划模式结果被误识别

3. **解析超时**:
   - 检查日志文件大小
   - 增加 ParseTimeout 时间
   - 考虑预处理日志以减少大小

### 调试方法

1. **启用详细日志**: 设置日志级别为 DEBUG
2. **单步调试**: 使用测试用例验证解析逻辑
3. **策略测试**: 单独测试各个策略的 CanParse 方法
4. **结果验证**: 检查解析结果是否符合预期格式

## 总结

经过精简重构的 Result Parser 现在更加简洁高效，专注于核心功能：

### 主要改进

- ✅ **代码精简**: 移除约 60% 的冗余代码
- ✅ **架构简化**: 去除复杂的工厂模式和缓存机制
- ✅ **性能优化**: 减少内存使用和解析时间
- ✅ **维护性**: 更清晰的代码结构，易于理解和维护

### 保留的核心功能

- ✅ **计划模式支持**: 完整支持 ExitPlanMode 结果解析
- ✅ **常规结果解析**: 支持标准 JSON 格式结果
- ✅ **智能检测**: 自动识别结果类型并选择合适策略
- ✅ **结果验证**: 确保解析结果的完整性和正确性
- ✅ **错误处理**: 提供清晰的错误信息
- ✅ **超时支持**: 防止长时间阻塞

### 支持的结果类型

1. **常规任务结果**: `{"type":"result","subtype":"success","is_error":false,...}`
2. **计划模式结果**: `{"type":"assistant","message":{"content":[{"name":"ExitPlanMode",...}]}}`
3. **错误结果**: `{"type":"result","subtype":"error","is_error":true,...}`

新的 Result Parser 在保持所有核心业务功能的同时，提供了更好的性能和可维护性，是一个精简而强大的解析引擎。