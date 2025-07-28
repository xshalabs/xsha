package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskConversationResultService struct {
	repo             repository.TaskConversationResultRepository
	conversationRepo repository.TaskConversationRepository
	taskRepo         repository.TaskRepository
	projectRepo      repository.ProjectRepository
}

// NewTaskConversationResultService 创建任务对话结果服务实例
func NewTaskConversationResultService(
	repo repository.TaskConversationResultRepository,
	conversationRepo repository.TaskConversationRepository,
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
) TaskConversationResultService {
	return &taskConversationResultService{
		repo:             repo,
		conversationRepo: conversationRepo,
		taskRepo:         taskRepo,
		projectRepo:      projectRepo,
	}
}

// CreateResult 创建结果
func (s *taskConversationResultService) CreateResult(conversationID uint, resultData map[string]interface{}) (*database.TaskConversationResult, error) {
	// 验证输入数据
	if err := s.ValidateResultData(resultData); err != nil {
		return nil, err
	}

	// 检查是否已存在结果记录
	exists, err := s.repo.ExistsByConversationID(conversationID)
	if err != nil {
		return nil, errors.New("failed to check existing result")
	}
	if exists {
		return nil, errors.New("result already exists for this conversation")
	}

	// 构建结果对象
	result := &database.TaskConversationResult{
		ConversationID: conversationID,
	}

	// 设置基本字段
	if typeVal, ok := resultData["type"].(string); ok {
		result.Type = database.ResultType(typeVal)
	}
	if subtypeVal, ok := resultData["subtype"].(string); ok {
		result.Subtype = database.ResultSubtype(subtypeVal)
	}
	if isErrorVal, ok := resultData["is_error"].(bool); ok {
		result.IsError = isErrorVal
	}

	// 设置时间和性能字段
	if durationMs, ok := resultData["duration_ms"].(float64); ok {
		result.DurationMs = int64(durationMs)
	}
	if durationApiMs, ok := resultData["duration_api_ms"].(float64); ok {
		result.DurationApiMs = int64(durationApiMs)
	}
	if numTurns, ok := resultData["num_turns"].(float64); ok {
		result.NumTurns = int(numTurns)
	}

	// 设置结果内容
	if resultStr, ok := resultData["result"].(string); ok {
		result.Result = resultStr
	}

	// 设置会话ID
	if sessionID, ok := resultData["session_id"].(string); ok {
		result.SessionID = sessionID
	}

	// 设置成本
	if totalCost, ok := resultData["total_cost_usd"].(float64); ok {
		result.TotalCostUsd = totalCost
	}

	// 设置使用统计（JSON字符串）
	if usage, ok := resultData["usage"]; ok {
		usageBytes, err := json.Marshal(usage)
		if err != nil {
			utils.Warn("Failed to marshal usage data", "error", err)
		} else {
			result.Usage = string(usageBytes)
		}
	}

	// 创建结果记录
	if err := s.repo.Create(result); err != nil {
		return nil, fmt.Errorf("failed to create result: %w", err)
	}

	utils.Info("Task conversation result created successfully",
		"conversation_id", conversationID,
		"result_id", result.ID)

	return result, nil
}

// GetResult 获取结果
func (s *taskConversationResultService) GetResult(id uint) (*database.TaskConversationResult, error) {
	return s.repo.GetByID(id)
}

// GetResultByConversationID 根据对话ID获取结果
func (s *taskConversationResultService) GetResultByConversationID(conversationID uint) (*database.TaskConversationResult, error) {
	return s.repo.GetByConversationID(conversationID)
}

// UpdateResult 更新结果
func (s *taskConversationResultService) UpdateResult(id uint, updates map[string]interface{}) error {
	// 检查结果是否存在
	result, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("result not found")
	}

	// 更新字段
	if typeVal, ok := updates["type"].(string); ok {
		result.Type = database.ResultType(typeVal)
	}
	if subtypeVal, ok := updates["subtype"].(string); ok {
		result.Subtype = database.ResultSubtype(subtypeVal)
	}
	if isErrorVal, ok := updates["is_error"].(bool); ok {
		result.IsError = isErrorVal
	}
	if durationMs, ok := updates["duration_ms"].(int64); ok {
		result.DurationMs = durationMs
	}
	if durationApiMs, ok := updates["duration_api_ms"].(int64); ok {
		result.DurationApiMs = durationApiMs
	}
	if numTurns, ok := updates["num_turns"].(int); ok {
		result.NumTurns = numTurns
	}
	if resultStr, ok := updates["result"].(string); ok {
		result.Result = resultStr
	}
	if sessionID, ok := updates["session_id"].(string); ok {
		result.SessionID = sessionID
	}
	if totalCost, ok := updates["total_cost_usd"].(float64); ok {
		result.TotalCostUsd = totalCost
	}
	if usage, ok := updates["usage"].(string); ok {
		result.Usage = usage
	}

	return s.repo.Update(result)
}

// DeleteResult 删除结果
func (s *taskConversationResultService) DeleteResult(id uint) error {
	return s.repo.Delete(id)
}

// ListResultsByTaskID 根据任务ID获取结果列表
func (s *taskConversationResultService) ListResultsByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	return s.repo.ListByTaskID(taskID, page, pageSize)
}

// ListResultsByProjectID 根据项目ID获取结果列表
func (s *taskConversationResultService) ListResultsByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	return s.repo.ListByProjectID(projectID, page, pageSize)
}

// GetTaskStats 获取任务统计信息
func (s *taskConversationResultService) GetTaskStats(taskID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取成功率
	successRate, err := s.repo.GetSuccessRate(taskID)
	if err != nil {
		utils.Warn("Failed to get success rate", "task_id", taskID, "error", err)
		successRate = 0
	}
	stats["success_rate"] = successRate

	// 获取总成本
	totalCost, err := s.repo.GetTotalCost(taskID)
	if err != nil {
		utils.Warn("Failed to get total cost", "task_id", taskID, "error", err)
		totalCost = 0
	}
	stats["total_cost_usd"] = totalCost

	// 获取平均执行时间
	avgDuration, err := s.repo.GetAverageDuration(taskID)
	if err != nil {
		utils.Warn("Failed to get average duration", "task_id", taskID, "error", err)
		avgDuration = 0
	}
	stats["average_duration_ms"] = avgDuration

	return stats, nil
}

// GetProjectStats 获取项目统计信息
func (s *taskConversationResultService) GetProjectStats(projectID uint) (map[string]interface{}, error) {
	// 获取项目下的结果列表（不分页，用于统计）
	results, _, err := s.repo.ListByProjectID(projectID, 1, 1000000) // 大页面获取所有数据
	if err != nil {
		return nil, fmt.Errorf("failed to get project results: %w", err)
	}

	stats := make(map[string]interface{})
	totalCount := len(results)
	successCount := 0
	totalCost := 0.0
	totalDuration := int64(0)

	for _, result := range results {
		if !result.IsError {
			successCount++
		}
		totalCost += result.TotalCostUsd
		totalDuration += result.DurationMs
	}

	stats["total_conversations"] = totalCount
	stats["success_count"] = successCount
	stats["error_count"] = totalCount - successCount
	stats["success_rate"] = 0.0
	if totalCount > 0 {
		stats["success_rate"] = float64(successCount) / float64(totalCount)
	}
	stats["total_cost_usd"] = totalCost
	stats["average_duration_ms"] = 0.0
	if totalCount > 0 {
		stats["average_duration_ms"] = float64(totalDuration) / float64(totalCount)
	}

	return stats, nil
}

// ProcessResultFromJSON 从JSON字符串处理结果
func (s *taskConversationResultService) ProcessResultFromJSON(jsonStr string, conversationID uint) (*database.TaskConversationResult, error) {
	// 解析JSON
	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &resultData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 检查type字段是否为"result"
	if typeVal, ok := resultData["type"].(string); !ok || typeVal != "result" {
		return nil, errors.New("invalid result type: must be 'result'")
	}

	// 创建结果
	return s.CreateResult(conversationID, resultData)
}

// ExistsForConversation 检查对话是否已有结果
func (s *taskConversationResultService) ExistsForConversation(conversationID uint) (bool, error) {
	return s.repo.ExistsByConversationID(conversationID)
}

// ValidateResultData 验证结果数据
func (s *taskConversationResultService) ValidateResultData(resultData map[string]interface{}) error {
	// 检查必需字段
	if typeVal, ok := resultData["type"].(string); !ok || typeVal == "" {
		return errors.New("type is required")
	}

	if subtypeVal, ok := resultData["subtype"].(string); !ok || subtypeVal == "" {
		return errors.New("subtype is required")
	}

	if _, ok := resultData["is_error"].(bool); !ok {
		return errors.New("is_error is required and must be boolean")
	}

	if resultStr, ok := resultData["result"].(string); !ok || resultStr == "" {
		return errors.New("result content is required")
	}

	if sessionID, ok := resultData["session_id"].(string); !ok || sessionID == "" {
		return errors.New("session_id is required")
	}

	// 验证数值字段
	if durationMs, ok := resultData["duration_ms"]; ok {
		if _, isFloat := durationMs.(float64); !isFloat {
			return errors.New("duration_ms must be a number")
		}
	}

	if durationApiMs, ok := resultData["duration_api_ms"]; ok {
		if _, isFloat := durationApiMs.(float64); !isFloat {
			return errors.New("duration_api_ms must be a number")
		}
	}

	if numTurns, ok := resultData["num_turns"]; ok {
		if _, isFloat := numTurns.(float64); !isFloat {
			return errors.New("num_turns must be a number")
		}
	}

	if totalCost, ok := resultData["total_cost_usd"]; ok {
		if _, isFloat := totalCost.(float64); !isFloat {
			return errors.New("total_cost_usd must be a number")
		}
	}

	return nil
}
