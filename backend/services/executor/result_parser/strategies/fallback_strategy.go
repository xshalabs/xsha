package strategies

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"time"
)

// FallbackStrategy 兜底解析策略
type FallbackStrategy struct {
	name     string
	priority int
}

// NewFallbackStrategy 创建兜底解析策略
func NewFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		name:     "fallback",
		priority: 10, // 最低优先级
	}
}

// Name 返回策略名称
func (s *FallbackStrategy) Name() string {
	return s.name
}

// Priority 返回策略优先级
func (s *FallbackStrategy) Priority() int {
	return s.priority
}

// CanParse 总是返回true，作为兜底策略
func (s *FallbackStrategy) CanParse(logs string) bool {
	return logs != ""
}

// Parse 尽最大努力从日志中提取信息
func (s *FallbackStrategy) Parse(ctx context.Context, logs string) (map[string]interface{}, error) {
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
	
	// 尝试各种启发式方法提取信息
	result := s.extractHeuristically(ctx, lines)
	
	if result == nil || len(result) == 0 {
		// 如果无法提取任何有用信息，创建一个基本的错误结果
		return s.createFallbackResult(logs), nil
	}
	
	// 填充缺失的必需字段
	s.fillMissingFields(result)
	
	return result, nil
}

// SupportsBatch 支持批量解析
func (s *FallbackStrategy) SupportsBatch() bool {
	return true
}

// ParseBatch 批量解析
func (s *FallbackStrategy) ParseBatch(ctx context.Context, logEntries []string) ([]map[string]interface{}, error) {
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

// extractHeuristically 使用启发式方法提取信息
func (s *FallbackStrategy) extractHeuristically(ctx context.Context, lines []string) map[string]interface{} {
	result := make(map[string]interface{})
	
	// 从不同的行中提取信息
	for _, line := range lines {
		select {
		case <-ctx.Done():
			return result
		default:
		}
		
		s.extractFromLine(line, result)
	}
	
	return result
}

// extractFromLine 从单行中提取信息
func (s *FallbackStrategy) extractFromLine(line string, result map[string]interface{}) {
	lowerLine := strings.ToLower(line)
	
	// 尝试识别session ID
	if sessionID := s.extractSessionID(line); sessionID != "" {
		result["session_id"] = sessionID
	}
	
	// 尝试识别错误状态
	if s.containsErrorIndicator(lowerLine) {
		result["is_error"] = true
		result["subtype"] = "error"
	} else if s.containsSuccessIndicator(lowerLine) {
		result["is_error"] = false
		result["subtype"] = "success"
	}
	
	// 尝试识别持续时间
	if duration := s.extractDuration(line); duration > 0 {
		result["duration_ms"] = duration
	}
	
	// 尝试识别成本信息
	if cost := s.extractCost(line); cost > 0 {
		result["total_cost_usd"] = cost
	}
	
	// 尝试识别轮数
	if turns := s.extractTurns(line); turns > 0 {
		result["num_turns"] = turns
	}
}

// extractSessionID 提取session ID
func (s *FallbackStrategy) extractSessionID(line string) string {
	patterns := []string{
		"session_id",
		"sessionid",
		"session-id",
		"sid",
	}
	
	lowerLine := strings.ToLower(line)
	
	for _, pattern := range patterns {
		if idx := strings.Index(lowerLine, pattern); idx >= 0 {
			// 查找模式后的值
			after := line[idx+len(pattern):]
			if value := s.extractValueAfterPattern(after); value != "" {
				return value
			}
		}
	}
	
	return ""
}

// extractValueAfterPattern 从模式后提取值
func (s *FallbackStrategy) extractValueAfterPattern(text string) string {
	// 移除前导的分隔符
	text = strings.TrimLeft(text, ":= \"'")
	
	// 找到值的结尾
	for i, char := range text {
		if char == ' ' || char == ',' || char == ';' || char == '"' || char == '\'' {
			if i > 0 {
				return text[:i]
			}
		}
	}
	
	// 如果没有找到分隔符，返回整个剩余文本（限制长度）
	if len(text) > 50 {
		return text[:50]
	}
	return text
}

// containsErrorIndicator 检查是否包含错误指示符
func (s *FallbackStrategy) containsErrorIndicator(line string) bool {
	errorIndicators := []string{
		"error", "failed", "failure", "exception", "panic",
		"timeout", "abort", "cancelled", "invalid",
	}
	
	for _, indicator := range errorIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}
	
	return false
}

// containsSuccessIndicator 检查是否包含成功指示符
func (s *FallbackStrategy) containsSuccessIndicator(line string) bool {
	successIndicators := []string{
		"success", "completed", "finished", "done",
		"ok", "passed", "successful",
	}
	
	for _, indicator := range successIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}
	
	return false
}

// extractDuration 提取持续时间（毫秒）
func (s *FallbackStrategy) extractDuration(line string) int64 {
	patterns := []string{
		"duration", "elapsed", "took", "time",
	}
	
	lowerLine := strings.ToLower(line)
	
	for _, pattern := range patterns {
		if strings.Contains(lowerLine, pattern) {
			// 简化的数字提取
			words := strings.Fields(line)
			for i, word := range words {
				if strings.Contains(strings.ToLower(word), pattern) && i+1 < len(words) {
					if num := s.extractNumber(words[i+1]); num > 0 {
						// 假设是毫秒，如果太小可能是秒
						if num < 1000 {
							return num * 1000
						}
						return num
					}
				}
			}
		}
	}
	
	return 0
}

// extractCost 提取成本
func (s *FallbackStrategy) extractCost(line string) float64 {
	if strings.Contains(strings.ToLower(line), "cost") ||
	   strings.Contains(line, "$") ||
	   strings.Contains(strings.ToLower(line), "usd") {
		words := strings.Fields(line)
		for _, word := range words {
			if num := s.extractFloat(word); num > 0 {
				return num
			}
		}
	}
	return 0
}

// extractTurns 提取轮数
func (s *FallbackStrategy) extractTurns(line string) int {
	if strings.Contains(strings.ToLower(line), "turn") {
		words := strings.Fields(line)
		for _, word := range words {
			if num := s.extractNumber(word); num > 0 && num < 1000 {
				return int(num)
			}
		}
	}
	return 0
}

// extractNumber 从字符串中提取数字
func (s *FallbackStrategy) extractNumber(text string) int64 {
	// 移除非数字字符
	numStr := ""
	for _, char := range text {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		}
	}
	
	if numStr == "" {
		return 0
	}
	
	if len(numStr) > 10 { // 防止过大的数字
		numStr = numStr[:10]
	}
	
	// 简单的字符串到数字转换
	result := int64(0)
	for _, char := range numStr {
		result = result*10 + int64(char-'0')
	}
	
	return result
}

// extractFloat 从字符串中提取浮点数
func (s *FallbackStrategy) extractFloat(text string) float64 {
	// 简化的浮点数提取
	numStr := ""
	hasDecimal := false
	
	for _, char := range text {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else if char == '.' && !hasDecimal {
			numStr += "."
			hasDecimal = true
		}
	}
	
	if numStr == "" || numStr == "." {
		return 0
	}
	
	// 简化的字符串到浮点数转换
	parts := strings.Split(numStr, ".")
	if len(parts) > 2 {
		return 0
	}
	
	intPart := int64(0)
	for _, char := range parts[0] {
		intPart = intPart*10 + int64(char-'0')
	}
	
	result := float64(intPart)
	
	if len(parts) == 2 && parts[1] != "" {
		fracPart := int64(0)
		divisor := int64(1)
		
		for _, char := range parts[1] {
			fracPart = fracPart*10 + int64(char-'0')
			divisor *= 10
		}
		
		result += float64(fracPart) / float64(divisor)
	}
	
	return result
}

// createFallbackResult 创建兜底结果
func (s *FallbackStrategy) createFallbackResult(logs string) map[string]interface{} {
	// 生成一个基本的会话ID
	sessionID := s.generateFallbackSessionID(logs)
	
	return map[string]interface{}{
		"type":        "result",
		"subtype":     "fallback",
		"is_error":    true,
		"session_id":  sessionID,
		"result":      "Failed to parse execution result",
		"duration_ms": int64(0),
		"num_turns":   1,
	}
}

// generateFallbackSessionID 生成兜底会话ID
func (s *FallbackStrategy) generateFallbackSessionID(logs string) string {
	// 使用日志内容的简单哈希作为会话ID
	hash := int64(0)
	for _, char := range logs {
		hash = hash*31 + int64(char)
	}
	
	if hash < 0 {
		hash = -hash
	}
	
	return "fallback_" + s.intToString(hash)
}

// intToString 整数转字符串
func (s *FallbackStrategy) intToString(num int64) string {
	if num == 0 {
		return "0"
	}
	
	result := ""
	for num > 0 {
		result = string(rune('0'+num%10)) + result
		num /= 10
	}
	
	return result
}

// fillMissingFields 填充缺失的必需字段
func (s *FallbackStrategy) fillMissingFields(result map[string]interface{}) {
	// 确保type字段
	if _, exists := result["type"]; !exists {
		result["type"] = "result"
	}
	
	// 确保subtype字段
	if _, exists := result["subtype"]; !exists {
		if isError, ok := result["is_error"].(bool); ok && isError {
			result["subtype"] = "error"
		} else {
			result["subtype"] = "success"
		}
	}
	
	// 确保is_error字段
	if _, exists := result["is_error"]; !exists {
		result["is_error"] = false
	}
	
	// 确保session_id字段
	if _, exists := result["session_id"]; !exists {
		result["session_id"] = "unknown_" + s.intToString(time.Now().Unix())
	}
}