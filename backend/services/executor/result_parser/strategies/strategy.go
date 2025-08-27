package strategies

import (
	"context"
)

// ParseStrategy 定义解析策略接口
type ParseStrategy interface {
	// Name 返回策略名称
	Name() string
	
	// CanParse 检查是否能解析给定的日志内容
	CanParse(logs string) bool
	
	// Parse 解析日志内容并返回结果数据
	Parse(ctx context.Context, logs string) (map[string]interface{}, error)
	
	// Priority 返回策略优先级（数值越小优先级越高）
	Priority() int
}



// containsJSON 检查是否包含JSON格式
func containsJSON(logs string) bool {
	// 简化的JSON检测，查找典型的JSON模式
	jsonIndicators := []string{
		`"type":`,
		`"subtype":`,
		`"session_id":`,
		`{"`,
		`}`,
	}
	
	count := 0
	for _, indicator := range jsonIndicators {
		if containsString(logs, indicator) {
			count++
		}
	}
	
	// 如果包含多个JSON指示符，认为是JSON格式
	return count >= 3
}


// containsPlanMode 检查是否包含计划模式
func containsPlanMode(logs string) bool {
	// 检查计划模式特有的标识符
	planModeIndicators := []string{
		`"type":"assistant"`,
		`"name":"ExitPlanMode"`,
		`"tool_use"`,
		`ExitPlanMode`,
	}
	
	count := 0
	for _, indicator := range planModeIndicators {
		if containsString(logs, indicator) {
			count++
		}
	}
	
	// 至少需要包含2个计划模式指示符
	return count >= 2
}

// containsString 简单的字符串包含检查
func containsString(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	
	return false
}