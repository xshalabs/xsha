package strategies

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// JSONStrategy JSON解析策略
type JSONStrategy struct {
	name     string
	priority int
}

// NewJSONStrategy 创建JSON解析策略
func NewJSONStrategy() *JSONStrategy {
	return &JSONStrategy{
		name:     "json",
		priority: 1, // 高优先级
	}
}

// Name 返回策略名称
func (s *JSONStrategy) Name() string {
	return s.name
}

// Priority 返回策略优先级
func (s *JSONStrategy) Priority() int {
	return s.priority
}

// CanParse 检查是否能解析给定的日志内容
func (s *JSONStrategy) CanParse(logs string) bool {
	if logs == "" {
		return false
	}

	// 快速检查是否包含JSON指示符
	if !containsJSON(logs) {
		return false
	}

	// 尝试提取JSON并验证
	lines := strings.Split(logs, "\n")
	for i := len(lines) - 1; i >= 0 && i >= len(lines)-10; i-- { // 只检查最后10行
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if jsonStr := s.extractJSONFromLine(line); jsonStr != "" {
			var testData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &testData); err == nil {
				if typeVal, ok := testData["type"].(string); ok && typeVal == "result" {
					return true
				}
			}
		}
	}

	return false
}

// Parse 解析日志内容
func (s *JSONStrategy) Parse(ctx context.Context, logs string) (map[string]interface{}, error) {
	if logs == "" {
		return nil, errors.New("empty logs")
	}

	lines := strings.Split(logs, "\n")

	// 限制处理的行数（优化性能），只查看最后1000行
	maxLines := 1000
	startIndex := 0
	if len(lines) > maxLines {
		startIndex = len(lines) - maxLines
	}

	// 从末尾开始查找有效的JSON结果
	for i := len(lines) - 1; i >= startIndex; i-- {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		jsonStr := s.extractJSONFromLine(line)
		if jsonStr == "" {
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			continue
		}

		// 检查是否是有效的结果JSON
		if s.isValidResultJSON(result) {
			return result, nil
		}
	}

	return nil, errors.New("no valid result JSON found")
}

// extractJSONFromLine 从单行日志中提取JSON
func (s *JSONStrategy) extractJSONFromLine(line string) string {
	// 查找 "STDOUT: " 的位置
	idx := strings.Index(line, "STDOUT: ")
	if idx != -1 {
		// 提取 STDOUT: 后面的内容
		jsonStr := strings.TrimSpace(line[idx+8:])
		// 验证是否以 { 开头并以 } 结尾
		if strings.HasPrefix(jsonStr, "{") && strings.HasSuffix(jsonStr, "}") {
			return jsonStr
		}
	}

	// 如果没有 STDOUT: 前缀，检查整行是否为 JSON
	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, "{") && strings.HasSuffix(trimmedLine, "}") {
		return trimmedLine
	}

	return ""
}

// isValidResultJSON 检查是否是有效的结果JSON
func (s *JSONStrategy) isValidResultJSON(data map[string]interface{}) bool {
	// 首先检查是否是计划模式结果
	if s.isPlanModeResult(data) {
		return false // 计划模式结果应该由PlanModeStrategy处理
	}

	// 检查必需的字段
	typeVal, hasType := data["type"].(string)
	if !hasType || typeVal != "result" {
		return false
	}

	// 检查subtype字段
	if _, hasSubtype := data["subtype"]; !hasSubtype {
		return false
	}

	// 检查is_error字段
	if _, hasIsError := data["is_error"]; !hasIsError {
		return false
	}

	// 检查session_id字段
	if sessionID, hasSessionID := data["session_id"].(string); !hasSessionID || sessionID == "" {
		return false
	}

	return true
}

// isPlanModeResult 检查是否是计划模式结果
func (s *JSONStrategy) isPlanModeResult(data map[string]interface{}) bool {
	// 检查是否是assistant类型
	typeVal, hasType := data["type"].(string)
	if !hasType || typeVal != "assistant" {
		return false
	}

	// 检查message字段
	message, hasMessage := data["message"]
	if !hasMessage {
		return false
	}

	messageMap, ok := message.(map[string]interface{})
	if !ok {
		return false
	}

	// 检查content字段
	content, hasContent := messageMap["content"]
	if !hasContent {
		return false
	}

	// content应该是一个数组
	contentArray, ok := content.([]interface{})
	if !ok {
		return false
	}

	// 检查是否包含ExitPlanMode工具使用
	for _, item := range contentArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if toolType, hasType := itemMap["type"].(string); hasType && toolType == "tool_use" {
				if name, hasName := itemMap["name"].(string); hasName && name == "ExitPlanMode" {
					return true
				}
			}
		}
	}

	return false
}
