package executor

import (
	"encoding/json"
	"regexp"
	"strings"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/utils"
)

type resultParser struct {
	taskConvResultRepo    repository.TaskConversationResultRepository
	taskConvResultService services.TaskConversationResultService
	logLineJSONRegex      *regexp.Regexp
}

// NewResultParser 创建结果解析器
func NewResultParser(
	taskConvResultRepo repository.TaskConversationResultRepository,
	taskConvResultService services.TaskConversationResultService,
) ResultParser {
	// 预编译用于提取日志行中JSON的正则表达式
	logLineJSONRegex := regexp.MustCompile(`^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`)

	return &resultParser{
		taskConvResultRepo:    taskConvResultRepo,
		taskConvResultService: taskConvResultService,
		logLineJSONRegex:      logLineJSONRegex,
	}
}

// ParseAndCreate 解析执行日志中的结果并创建 TaskConversationResult 记录
func (r *resultParser) ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	// 从执行日志中解析结果 JSON
	resultData, err := r.ParseFromLogs(execLog.ExecutionLogs)
	if err != nil {
		utils.Warn("Failed to parse execution result from logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID,
			"error", err)
		return
	}

	if resultData == nil {
		// 没有找到结果数据，可能是正常情况（某些执行可能不产生结果JSON）
		utils.Info("No result data found in execution logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID)
		return
	}

	// 检查是否已存在结果记录
	exists, err := r.taskConvResultRepo.ExistsByConversationID(conv.ID)
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

	// 创建 TaskConversationResult 记录
	_, err = r.taskConvResultService.CreateResult(conv.ID, resultData)
	if err != nil {
		utils.Error("Failed to create task conversation result",
			"conversation_id", conv.ID,
			"error", err)
		return
	}

	utils.Info("Successfully created task conversation result",
		"conversation_id", conv.ID,
		"result_data", resultData)
}

// ParseFromLogs 从执行日志字符串中解析结果 JSON
func (r *resultParser) ParseFromLogs(executionLogs string) (map[string]interface{}, error) {
	if executionLogs == "" {
		return nil, nil
	}

	// 按行分割日志
	lines := strings.Split(executionLogs, "\n")

	// 从后往前查找，因为结果通常在日志末尾
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// 提取日志行中的 JSON 部分
		jsonStr := r.extractJSONFromLogLine(line)
		if jsonStr == "" {
			continue // 没有找到 JSON 部分
		}

		// 尝试解析为 JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			continue // 不是有效的 JSON，继续查找
		}

		// 检查是否是我们要找的结果类型
		if typeVal, ok := result["type"].(string); ok && typeVal == "result" {
			// 验证必需字段
			if _, hasSubtype := result["subtype"]; hasSubtype {
				if _, hasIsError := result["is_error"]; hasIsError {
					// 额外验证其他关键字段
					if r.validateResultData(result) {
						utils.Info("Found result JSON in execution logs",
							"line_index", i,
							"result_type", typeVal,
							"json_extract", jsonStr[:100]+"...") // 记录前100个字符用于调试
						return result, nil
					}
				}
			}
		}
	}

	return nil, nil // 没有找到符合条件的结果 JSON
}

// extractJSONFromLogLine 从日志行中提取 JSON 字符串
// 支持格式: [时间戳] 前缀: {JSON内容} 或纯 JSON
func (r *resultParser) extractJSONFromLogLine(line string) string {
	// 使用预编译的正则表达式匹配日志格式并提取 JSON
	// 模式说明:
	// ^                     - 行开始
	// (?:\[\d{2}:\d{2}:\d{2}\]\s*)?  - 可选的时间戳 [HH:MM:SS]
	// (?:\w+:\s*)?          - 可选的前缀如 STDOUT:, STDERR: 等
	// (\{.*\})              - 捕获组：JSON 对象（从 { 开始到 } 结束）
	// \s*$                  - 可选的空白字符直到行尾

	// 匹配并提取 JSON
	matches := r.logLineJSONRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) >= 2 {
		return matches[1] // 返回第一个捕获组（JSON部分）
	}

	// 如果正则匹配失败，检查是否是纯 JSON 行
	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, "{") && strings.HasSuffix(trimmedLine, "}") {
		return trimmedLine
	}

	return ""
}

// validateResultData 验证结果数据的完整性
func (r *resultParser) validateResultData(data map[string]interface{}) bool {
	// 检查必需字段是否存在
	requiredFields := []string{"type", "subtype", "is_error", "session_id"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			utils.Warn("Missing required field in result data", "field", field)
			return false
		}
	}

	// 检查数据类型
	if typeVal, ok := data["type"].(string); !ok || typeVal != "result" {
		utils.Warn("Invalid type field in result data", "type", data["type"])
		return false
	}

	if _, ok := data["is_error"].(bool); !ok {
		utils.Warn("Invalid is_error field in result data", "is_error", data["is_error"])
		return false
	}

	if sessionID, ok := data["session_id"].(string); !ok || sessionID == "" {
		utils.Warn("Invalid session_id field in result data", "session_id", data["session_id"])
		return false
	}

	return true
}
