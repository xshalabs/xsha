package result_parser

import (
	"context"
	"errors"
	"fmt"
	"time"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/services/executor/result_parser/strategies"
	"xsha-backend/services/executor/result_parser/validator"
	"xsha-backend/utils"
)

// Parser 解析器接口
type Parser interface {
	// ParseFromLogs 从日志中解析结果
	ParseFromLogs(executionLogs string) (map[string]interface{}, error)
	
	// ParseFromLogsWithContext 带上下文的日志解析
	ParseFromLogsWithContext(ctx context.Context, executionLogs string) (map[string]interface{}, error)
	
	// ParseAndCreate 解析并创建结果记录
	ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog)
}

// DefaultParser 默认解析器实现
type DefaultParser struct {
	config     *Config
	strategies []strategies.ParseStrategy
	validator  validator.Validator
	
	// 依赖服务
	taskConvResultRepo    repository.TaskConversationResultRepository
	taskConvResultService services.TaskConversationResultService
	taskService           services.TaskService
}

// NewDefaultParser 创建默认解析器
func NewDefaultParser(
	taskConvResultRepo repository.TaskConversationResultRepository,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
) Parser {
	config := DefaultConfig()
	config.LoadFromEnv()
	config.Validate()
	
	// 创建策略
	strategies := []strategies.ParseStrategy{
		strategies.NewPlanModeStrategy(),
		strategies.NewJSONStrategy(),
	}
	
	return &DefaultParser{
		config:                config,
		strategies:            strategies,
		validator:             validator.NewResultValidator(config.StrictValidation),
		taskConvResultRepo:    taskConvResultRepo,
		taskConvResultService: taskConvResultService,
		taskService:           taskService,
	}
}

// ParseFromLogs 从日志中解析结果
func (p *DefaultParser) ParseFromLogs(executionLogs string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ParseTimeout)
	defer cancel()
	
	return p.ParseFromLogsWithContext(ctx, executionLogs)
}

// ParseFromLogsWithContext 带上下文的日志解析
func (p *DefaultParser) ParseFromLogsWithContext(ctx context.Context, executionLogs string) (map[string]interface{}, error) {
	if executionLogs == "" {
		return nil, errors.New("execution logs are empty")
	}
	
	// 选择解析策略
	strategy := p.selectStrategy(executionLogs)
	
	// 解析日志
	result, err := strategy.Parse(ctx, executionLogs)
	if err != nil {
		return nil, err
	}
	
	// 验证解析结果
	if err := p.validateResult(result); err != nil {
		if p.config.StrictValidation {
			return nil, fmt.Errorf("validation failed: %v", err)
		}
		
		// 在非严格模式下，记录警告但继续
		utils.Warn("Result validation failed, continuing in non-strict mode",
			"error", err,
			"result", result)
	}
	
	return result, nil
}

// ParseAndCreate 解析并创建结果记录
func (p *DefaultParser) ParseAndCreate(conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	startTime := time.Now()
	
	resultData, err := p.ParseFromLogs(execLog.ExecutionLogs)
	if err != nil {
		utils.Warn("Failed to parse execution result from logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID,
			"error", err,
			"duration", time.Since(startTime))
		return
	}
	
	if resultData == nil {
		utils.Info("No result data found in execution logs",
			"conversation_id", conv.ID,
			"execution_log_id", execLog.ID,
			"duration", time.Since(startTime))
		return
	}
	
	// 检查是否已存在结果记录
	if p.taskConvResultRepo != nil {
		exists, err := p.taskConvResultRepo.ExistsByConversationID(conv.ID)
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
	}
	
	// 创建结果记录
	if p.taskConvResultService != nil {
		result, err := p.taskConvResultService.CreateResult(conv.ID, resultData)
		if err != nil {
			utils.Error("Failed to create task conversation result",
				"conversation_id", conv.ID,
				"error", err)
			return
		}
		
		utils.Info("Successfully created task conversation result",
			"conversation_id", conv.ID,
			"result_id", result.ID,
			"duration", time.Since(startTime))
		
		// 更新任务的会话ID
		if result.SessionID != "" && conv.Task != nil && p.taskService != nil {
			err = p.taskService.UpdateTaskSessionID(conv.Task.ID, result.SessionID)
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
}


// selectStrategy 选择解析策略
func (p *DefaultParser) selectStrategy(logs string) strategies.ParseStrategy {
	// 找到第一个可以解析的策略
	for _, strategy := range p.strategies {
		if strategy.CanParse(logs) {
			return strategy
		}
	}
	
	// 如果没有策略可以解析，返回JSON策略作为默认
	return strategies.NewJSONStrategy()
}

// validateResult 验证解析结果
func (p *DefaultParser) validateResult(result map[string]interface{}) error {
	if p.validator == nil {
		return nil
	}
	
	// 对于计划模式结果，使用更宽松的验证
	if p.isPlanModeResult(result) {
		return p.validatePlanModeResult(result)
	}
	
	return p.validator.Validate(result)
}

// isPlanModeResult 检查是否是计划模式结果
func (p *DefaultParser) isPlanModeResult(result map[string]interface{}) bool {
	if typeVal, ok := result["type"].(string); ok && typeVal == "result" {
		if subtypeVal, ok := result["subtype"].(string); ok && subtypeVal == "plan_mode" {
			return true
		}
	}
	return false
}

// validatePlanModeResult 验证计划模式结果
func (p *DefaultParser) validatePlanModeResult(result map[string]interface{}) error {
	// 检查必需字段
	requiredFields := []string{"type", "subtype", "is_error", "session_id", "result"}
	
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	
	// 验证type字段
	if typeVal, ok := result["type"].(string); !ok || typeVal != "result" {
		return fmt.Errorf("invalid type field, expected 'result', got: %v", result["type"])
	}
	
	// 验证subtype字段
	if subtypeVal, ok := result["subtype"].(string); !ok || subtypeVal != "plan_mode" {
		return fmt.Errorf("invalid subtype field, expected 'plan_mode', got: %v", result["subtype"])
	}
	
	// 验证session_id字段
	if sessionID, ok := result["session_id"].(string); !ok || sessionID == "" {
		return fmt.Errorf("invalid or empty session_id field: %v", result["session_id"])
	}
	
	// 验证result字段（应包含计划内容）
	if planResult, ok := result["result"].(string); !ok || planResult == "" {
		return fmt.Errorf("invalid or empty result field: %v", result["result"])
	}
	
	return nil
}



// NewResultParser 创建结果解析器（向后兼容）
func NewResultParser(
	taskConvResultRepo repository.TaskConversationResultRepository,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
) Parser {
	return NewDefaultParser(taskConvResultRepo, taskConvResultService, taskService)
}