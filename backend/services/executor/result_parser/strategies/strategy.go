package strategies

import (
	"context"
	"time"
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
	
	// SupportsBatch 是否支持批量解析
	SupportsBatch() bool
	
	// ParseBatch 批量解析多个日志条目
	ParseBatch(ctx context.Context, logEntries []string) ([]map[string]interface{}, error)
}

// ParseResult 解析结果
type ParseResult struct {
	Data      map[string]interface{} `json:"data"`
	Strategy  string                 `json:"strategy"`
	ParseTime time.Time              `json:"parse_time"`
	Duration  time.Duration          `json:"duration"`
	Error     error                  `json:"error,omitempty"`
}

// LogFormat 日志格式类型
type LogFormat int

const (
	LogFormatUnknown LogFormat = iota
	LogFormatJSON
	LogFormatStructuredText
	LogFormatPlainText
)

func (f LogFormat) String() string {
	switch f {
	case LogFormatJSON:
		return "json"
	case LogFormatStructuredText:
		return "structured_text"
	case LogFormatPlainText:
		return "plain_text"
	default:
		return "unknown"
	}
}

// DetectLogFormat 检测日志格式
func DetectLogFormat(logs string) LogFormat {
	if logs == "" {
		return LogFormatUnknown
	}
	
	// 简单的格式检测逻辑
	if containsPlanMode(logs) {
		return LogFormatJSON // 计划模式也是JSON格式
	}
	
	if containsJSON(logs) {
		return LogFormatJSON
	}
	
	if containsStructuredText(logs) {
		return LogFormatStructuredText
	}
	
	return LogFormatPlainText
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

// containsStructuredText 检查是否包含结构化文本
func containsStructuredText(logs string) bool {
	// 检查常见的结构化文本模式
	structuredIndicators := []string{
		"type=",
		"subtype=",
		"session_id=",
		"result=",
	}
	
	count := 0
	for _, indicator := range structuredIndicators {
		if containsString(logs, indicator) {
			count++
		}
	}
	
	return count >= 2
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
	return len(haystack) >= len(needle) && 
		   findSubstring(haystack, needle) != -1
}

// findSubstring 查找子字符串位置
func findSubstring(haystack, needle string) int {
	if len(needle) == 0 {
		return 0
	}
	
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	
	return -1
}