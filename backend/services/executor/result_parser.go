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
	taskService           services.TaskService
	logLineJSONRegex      *regexp.Regexp
}

func NewResultParser(
	taskConvResultRepo repository.TaskConversationResultRepository,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
) ResultParser {
	logLineJSONRegex := regexp.MustCompile(`^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`)

	return &resultParser{
		taskConvResultRepo:    taskConvResultRepo,
		taskConvResultService: taskConvResultService,
		taskService:           taskService,
		logLineJSONRegex:      logLineJSONRegex,
	}
}

func (r *resultParser) ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	resultData, err := r.ParseFromLogs(execLog.ExecutionLogs)
	if err != nil {
		utils.Warn("Failed to parse execution result from logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID,
			"error", err)
		return
	}

	if resultData == nil {
		utils.Info("No result data found in execution logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID)
		return
	}

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

	result, err := r.taskConvResultService.CreateResult(conv.ID, resultData)
	if err != nil {
		utils.Error("Failed to create task conversation result",
			"conversation_id", conv.ID,
			"error", err)
		return
	}

	utils.Info("Successfully created task conversation result",
		"conversation_id", conv.ID,
		"result_data", resultData)

	// Update task session_id if result has a session_id
	if result.SessionID != "" && conv.Task != nil {
		err = r.taskService.UpdateTaskSessionID(conv.Task.ID, result.SessionID)
		if err != nil {
			utils.Error("Failed to update task session ID",
				"task_id", conv.Task.ID,
				"session_id", result.SessionID,
				"error", err)
		} else {
			utils.Info("Successfully updated task session ID",
				"task_id", conv.Task.ID,
				"session_id", result.SessionID)
		}
	}
}

func (r *resultParser) ParseFromLogs(executionLogs string) (map[string]interface{}, error) {
	if executionLogs == "" {
		return nil, nil
	}

	lines := strings.Split(executionLogs, "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		jsonStr := r.extractJSONFromLogLine(line)
		if jsonStr == "" {
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			continue
		}

		if typeVal, ok := result["type"].(string); ok && typeVal == "result" {
			if _, hasSubtype := result["subtype"]; hasSubtype {
				if _, hasIsError := result["is_error"]; hasIsError {
					if r.validateResultData(result) {
						utils.Info("Found result JSON in execution logs",
							"line_index", i,
							"result_type", typeVal,
							"json_extract", jsonStr[:100]+"...")
						return result, nil
					}
				}
			}
		}
	}

	return nil, nil
}

func (r *resultParser) extractJSONFromLogLine(line string) string {
	matches := r.logLineJSONRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) >= 2 {
		return matches[1]
	}

	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, "{") && strings.HasSuffix(trimmedLine, "}") {
		return trimmedLine
	}

	return ""
}

func (r *resultParser) validateResultData(data map[string]interface{}) bool {
	requiredFields := []string{"type", "subtype", "is_error", "session_id"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			utils.Warn("Missing required field in result data", "field", field)
			return false
		}
	}

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
