package strategies

import (
	"bufio"
	"context"
	"errors"
	"strconv"
	"strings"
)

// TextStrategy 结构化文本解析策略
type TextStrategy struct {
	name     string
	priority int
}

// NewTextStrategy 创建文本解析策略
func NewTextStrategy() *TextStrategy {
	return &TextStrategy{
		name:     "structured_text",
		priority: 2, // 中等优先级
	}
}

// Name 返回策略名称
func (s *TextStrategy) Name() string {
	return s.name
}

// Priority 返回策略优先级
func (s *TextStrategy) Priority() int {
	return s.priority
}

// CanParse 检查是否能解析给定的日志内容
func (s *TextStrategy) CanParse(logs string) bool {
	if logs == "" {
		return false
	}
	
	// 如果是计划模式，应该由PlanModeStrategy处理
	if containsPlanMode(logs) {
		return false
	}
	
	// 检查是否包含结构化文本指示符
	return containsStructuredText(logs)
}

// Parse 解析结构化文本日志
func (s *TextStrategy) Parse(ctx context.Context, logs string) (map[string]interface{}, error) {
	if logs == "" {
		return nil, errors.New("empty logs")
	}
	
	scanner := bufio.NewScanner(strings.NewReader(logs))
	lines := make([]string, 0)
	
	// 收集所有非空行
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
		
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	// 从末尾开始查找结果行
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		
		if result := s.parseResultLine(line); result != nil {
			if s.isValidResult(result) {
				return result, nil
			}
		}
		
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	
	// 如果没有找到单行结果，尝试多行解析
	return s.parseMultiLine(ctx, lines)
}

// SupportsBatch 支持批量解析
func (s *TextStrategy) SupportsBatch() bool {
	return true
}

// ParseBatch 批量解析
func (s *TextStrategy) ParseBatch(ctx context.Context, logEntries []string) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, len(logEntries))
	
	for _, logs := range logEntries {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}
		
		if result, err := s.Parse(ctx, logs); err == nil {
			results = append(results, result)
		}
	}
	
	return results, nil
}

// parseResultLine 解析单行结果
func (s *TextStrategy) parseResultLine(line string) map[string]interface{} {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	
	// 检查是否包含result类型指示符
	if !strings.Contains(line, "type=result") && 
	   !strings.Contains(line, "TYPE=RESULT") &&
	   !strings.Contains(strings.ToLower(line), "result:") {
		return nil
	}
	
	result := make(map[string]interface{})
	
	// 解析key=value格式
	if s.parseKeyValueFormat(line, result) {
		return result
	}
	
	// 解析key: value格式
	if s.parseColonFormat(line, result) {
		return result
	}
	
	return nil
}

// parseKeyValueFormat 解析key=value格式
func (s *TextStrategy) parseKeyValueFormat(line string, result map[string]interface{}) bool {
	pairs := strings.Split(line, " ")
	foundPairs := 0
	
	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				
				// 移除引号
				value = strings.Trim(value, "\"'")
				
				result[key] = s.convertValue(key, value)
				foundPairs++
			}
		}
	}
	
	return foundPairs >= 3 // 至少需要3个键值对
}

// parseColonFormat 解析key: value格式
func (s *TextStrategy) parseColonFormat(line string, result map[string]interface{}) bool {
	// 处理类似 "Result: type=result, subtype=success, session_id=abc123" 的格式
	if strings.Contains(line, ":") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			content := strings.TrimSpace(parts[1])
			return s.parseKeyValueFormat(content, result)
		}
	}
	
	return false
}

// parseMultiLine 解析多行结果
func (s *TextStrategy) parseMultiLine(ctx context.Context, lines []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	inResultBlock := false
	
	// 从末尾开始查找结果块
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		if line == "" {
			continue
		}
		
		// 检查是否是结果块的开始
		if strings.Contains(strings.ToLower(line), "result") {
			inResultBlock = true
		}
		
		if inResultBlock {
			// 尝试解析当前行的键值对
			s.parseLineForKeyValue(line, result)
		}
		
		// 如果已经收集了足够的字段，停止解析
		if len(result) >= 4 && s.isValidResult(result) {
			return result, nil
		}
	}
	
	if len(result) > 0 && s.isValidResult(result) {
		return result, nil
	}
	
	return nil, errors.New("no valid structured text result found")
}

// parseLineForKeyValue 从行中解析键值对
func (s *TextStrategy) parseLineForKeyValue(line string, result map[string]interface{}) {
	// 尝试不同的分隔符
	separators := []string{"=", ":", " "}
	
	for _, sep := range separators {
		if strings.Contains(line, sep) {
			parts := strings.SplitN(line, sep, 2)
			if len(parts) == 2 {
				key := s.extractKey(parts[0])
				value := strings.TrimSpace(parts[1])
				
				if key != "" && value != "" {
					// 移除常见的前后缀
					value = strings.Trim(value, "\"',;")
					result[key] = s.convertValue(key, value)
				}
			}
		}
	}
}

// extractKey 提取键名
func (s *TextStrategy) extractKey(rawKey string) string {
	key := strings.TrimSpace(rawKey)
	key = strings.ToLower(key)
	
	// 移除常见的前缀
	prefixes := []string{"result ", "info ", "debug ", "warn ", "error "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			key = strings.TrimPrefix(key, prefix)
			break
		}
	}
	
	// 只保留已知的键
	knownKeys := []string{
		"type", "subtype", "is_error", "session_id",
		"duration_ms", "duration_api_ms", "num_turns",
		"result", "total_cost_usd", "usage",
	}
	
	for _, knownKey := range knownKeys {
		if key == knownKey || strings.Contains(key, knownKey) {
			return knownKey
		}
	}
	
	return ""
}

// convertValue 转换值的类型
func (s *TextStrategy) convertValue(key, value string) interface{} {
	switch key {
	case "is_error":
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
		return strings.ToLower(value) == "true" || value == "1"
		
	case "duration_ms", "duration_api_ms":
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
		
	case "num_turns":
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		
	case "total_cost_usd":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	
	return value
}

// isValidResult 检查结果是否有效
func (s *TextStrategy) isValidResult(result map[string]interface{}) bool {
	// 检查必需字段
	requiredFields := []string{"type", "subtype", "is_error", "session_id"}
	
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			return false
		}
	}
	
	// 检查type字段值
	if typeVal, ok := result["type"].(string); !ok || typeVal != "result" {
		return false
	}
	
	// 检查session_id不为空
	if sessionID, ok := result["session_id"].(string); !ok || sessionID == "" {
		return false
	}
	
	return true
}