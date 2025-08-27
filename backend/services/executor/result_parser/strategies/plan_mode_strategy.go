package strategies

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// PlanModeStrategy 计划模式解析策略
type PlanModeStrategy struct {
	name     string
	priority int
}

// NewPlanModeStrategy 创建计划模式解析策略
func NewPlanModeStrategy() *PlanModeStrategy {
	return &PlanModeStrategy{
		name:     "plan_mode",
		priority: 1, // 高优先级，与JSON策略相同
	}
}

// Name 返回策略名称
func (s *PlanModeStrategy) Name() string {
	return s.name
}

// Priority 返回策略优先级
func (s *PlanModeStrategy) Priority() int {
	return s.priority
}

// CanParse 检查是否能解析给定的日志内容
func (s *PlanModeStrategy) CanParse(logs string) bool {
	if logs == "" {
		return false
	}

	// 快速检查是否包含计划模式指示符
	if !containsPlanMode(logs) {
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
				if s.isPlanModeJSON(testData) {
					return true
				}
			}
		}
	}

	return false
}

// Parse 解析日志内容
func (s *PlanModeStrategy) Parse(ctx context.Context, logs string) (map[string]interface{}, error) {
	if logs == "" {
		return nil, errors.New("empty logs")
	}

	// 使用上下文监控超时
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	lines := strings.Split(logs, "\n")

	// 限制处理的行数，只查看最后1000行
	maxLines := 1000
	startIndex := 0
	if len(lines) > maxLines {
		startIndex = len(lines) - maxLines
	}

	// 从末尾开始查找有效的计划模式结果
	for i := len(lines) - 1; i >= startIndex; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		jsonStr := s.extractJSONFromLine(line)
		if jsonStr == "" {
			continue
		}

		var rawData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
			continue
		}

		// 检查是否是有效的计划模式JSON
		if !s.isPlanModeJSON(rawData) {
			continue
		}

		// 转换为标准结果格式
		result, err := s.convertToResultFormat(rawData)
		if err != nil {
			continue
		}

		return result, nil
	}

	return nil, errors.New("no valid plan mode result JSON found")
}

// extractJSONFromLine 从单行日志中提取JSON
func (s *PlanModeStrategy) extractJSONFromLine(line string) string {
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

// isPlanModeJSON 检查是否是有效的计划模式JSON
func (s *PlanModeStrategy) isPlanModeJSON(data map[string]interface{}) bool {
	// 检查必需的字段
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

// convertToResultFormat 将计划模式JSON转换为标准结果格式
func (s *PlanModeStrategy) convertToResultFormat(rawData map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 设置基本字段
	result["type"] = "result"
	result["subtype"] = "plan_mode"
	result["is_error"] = false

	// 提取session_id
	if sessionID, hasSession := rawData["session_id"].(string); hasSession && sessionID != "" {
		result["session_id"] = sessionID
	} else {
		// 生成一个默认session_id
		result["session_id"] = fmt.Sprintf("plan-mode-%d", time.Now().Unix())
	}

	// 设置默认值
	result["duration_ms"] = int64(0)
	result["duration_api_ms"] = int64(0)
	result["num_turns"] = 0
	result["total_cost_usd"] = 0.0

	// 提取计划内容
	message, hasMessage := rawData["message"]
	if !hasMessage {
		return nil, errors.New("missing message field")
	}

	messageMap, ok := message.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid message format")
	}

	content, hasContent := messageMap["content"]
	if !hasContent {
		return nil, errors.New("missing content field")
	}

	contentArray, ok := content.([]interface{})
	if !ok {
		return nil, errors.New("invalid content format")
	}

	// 查找ExitPlanMode工具使用并提取计划内容
	var planContent string
	for _, item := range contentArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if toolType, hasType := itemMap["type"].(string); hasType && toolType == "tool_use" {
				if name, hasName := itemMap["name"].(string); hasName && name == "ExitPlanMode" {
					if input, hasInput := itemMap["input"]; hasInput {
						if inputMap, ok := input.(map[string]interface{}); ok {
							if plan, hasPlan := inputMap["plan"].(string); hasPlan {
								planContent = plan
								break
							}
						}
					}
				}
			}
		}
	}

	if planContent == "" {
		return nil, errors.New("no plan content found")
	}

	result["result"] = planContent

	// 提取usage信息（如果存在）
	if messageMap["usage"] != nil {
		result["usage"] = messageMap["usage"]
	}

	return result, nil
}
