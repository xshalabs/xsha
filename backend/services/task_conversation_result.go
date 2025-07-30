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

func (s *taskConversationResultService) CreateResult(conversationID uint, resultData map[string]interface{}) (*database.TaskConversationResult, error) {
	if err := s.ValidateResultData(resultData); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsByConversationID(conversationID)
	if err != nil {
		return nil, errors.New("failed to check existing result")
	}
	if exists {
		return nil, errors.New("result already exists for this conversation")
	}

	result := &database.TaskConversationResult{
		ConversationID: conversationID,
	}

	if typeVal, ok := resultData["type"].(string); ok {
		result.Type = database.ResultType(typeVal)
	}
	if subtypeVal, ok := resultData["subtype"].(string); ok {
		result.Subtype = database.ResultSubtype(subtypeVal)
	}
	if isErrorVal, ok := resultData["is_error"].(bool); ok {
		result.IsError = isErrorVal
	}

	if durationMs, ok := resultData["duration_ms"].(float64); ok {
		result.DurationMs = int64(durationMs)
	}
	if durationApiMs, ok := resultData["duration_api_ms"].(float64); ok {
		result.DurationApiMs = int64(durationApiMs)
	}
	if numTurns, ok := resultData["num_turns"].(float64); ok {
		result.NumTurns = int(numTurns)
	}

	if resultStr, ok := resultData["result"].(string); ok {
		result.Result = resultStr
	}

	if sessionID, ok := resultData["session_id"].(string); ok {
		result.SessionID = sessionID
	}

	if totalCost, ok := resultData["total_cost_usd"].(float64); ok {
		result.TotalCostUsd = totalCost
	}

	if usage, ok := resultData["usage"]; ok {
		usageBytes, err := json.Marshal(usage)
		if err != nil {
			utils.Warn("Failed to marshal usage data", "error", err)
		} else {
			result.Usage = string(usageBytes)
		}
	}

	if err := s.repo.Create(result); err != nil {
		return nil, fmt.Errorf("failed to create result: %v", err)
	}

	utils.Info("Task conversation result created successfully",
		"conversation_id", conversationID,
		"result_id", result.ID)

	return result, nil
}

func (s *taskConversationResultService) GetResult(id uint) (*database.TaskConversationResult, error) {
	return s.repo.GetByID(id)
}

func (s *taskConversationResultService) GetResultByConversationID(conversationID uint) (*database.TaskConversationResult, error) {
	return s.repo.GetByConversationID(conversationID)
}

func (s *taskConversationResultService) UpdateResult(id uint, updates map[string]interface{}) error {
	result, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("result not found")
	}

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

func (s *taskConversationResultService) DeleteResult(id uint) error {
	return s.repo.Delete(id)
}

func (s *taskConversationResultService) ListResultsByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	return s.repo.ListByTaskID(taskID, page, pageSize)
}

func (s *taskConversationResultService) ListResultsByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	return s.repo.ListByProjectID(projectID, page, pageSize)
}

func (s *taskConversationResultService) GetTaskStats(taskID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	successRate, err := s.repo.GetSuccessRate(taskID)
	if err != nil {
		utils.Warn("Failed to get success rate", "task_id", taskID, "error", err)
		successRate = 0
	}
	stats["success_rate"] = successRate

	totalCost, err := s.repo.GetTotalCost(taskID)
	if err != nil {
		utils.Warn("Failed to get total cost", "task_id", taskID, "error", err)
		totalCost = 0
	}
	stats["total_cost_usd"] = totalCost

	avgDuration, err := s.repo.GetAverageDuration(taskID)
	if err != nil {
		utils.Warn("Failed to get average duration", "task_id", taskID, "error", err)
		avgDuration = 0
	}
	stats["average_duration_ms"] = avgDuration

	return stats, nil
}

func (s *taskConversationResultService) GetProjectStats(projectID uint) (map[string]interface{}, error) {
	results, _, err := s.repo.ListByProjectID(projectID, 1, 1000000)
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

func (s *taskConversationResultService) ExistsForConversation(conversationID uint) (bool, error) {
	return s.repo.ExistsByConversationID(conversationID)
}

func (s *taskConversationResultService) ValidateResultData(resultData map[string]interface{}) error {
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
