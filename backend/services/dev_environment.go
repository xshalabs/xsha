package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"xsha-backend/database"
	"xsha-backend/repository"
)

type devEnvironmentService struct {
	repo          repository.DevEnvironmentRepository
	configService SystemConfigService
}

// NewDevEnvironmentService creates a new development environment service instance
func NewDevEnvironmentService(repo repository.DevEnvironmentRepository, configService SystemConfigService) DevEnvironmentService {
	return &devEnvironmentService{
		repo:          repo,
		configService: configService,
	}
}

// CreateEnvironment creates a development environment
func (s *devEnvironmentService) CreateEnvironment(name, description, envType, createdBy string, cpuLimit float64, memoryLimit int64, envVars map[string]string) (*database.DevEnvironment, error) {
	// Validate input
	if err := s.validateEnvironmentData(name, envType, cpuLimit, memoryLimit); err != nil {
		return nil, err
	}

	// Validate environment variables
	if err := s.ValidateEnvVars(envVars); err != nil {
		return nil, err
	}

	// Check if environment name already exists
	if existing, _ := s.repo.GetByName(name, createdBy); existing != nil {
		return nil, errors.New("environment name already exists")
	}

	// Serialize environment variables
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize environment variables: %v", err)
	}

	// Create environment object
	env := &database.DevEnvironment{
		Name:        name,
		Description: description,
		Type:        envType, // 直接存储 key 值
		CPULimit:    cpuLimit,
		MemoryLimit: memoryLimit,
		EnvVars:     string(envVarsJSON),
		CreatedBy:   createdBy,
	}

	// Save to database
	if err := s.repo.Create(env); err != nil {
		return nil, err
	}

	return env, nil
}

// GetEnvironment gets a development environment
func (s *devEnvironmentService) GetEnvironment(id uint, createdBy string) (*database.DevEnvironment, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListEnvironments gets a list of development environments
func (s *devEnvironmentService) ListEnvironments(createdBy string, envType *string, name *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	return s.repo.List(createdBy, envType, name, page, pageSize)
}

// UpdateEnvironment updates a development environment
func (s *devEnvironmentService) UpdateEnvironment(id uint, createdBy string, updates map[string]interface{}) error {
	env, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// Update basic information
	if name, ok := updates["name"]; ok {
		env.Name = name.(string)
	}
	if description, ok := updates["description"]; ok {
		env.Description = description.(string)
	}
	if cpuLimit, ok := updates["cpu_limit"]; ok {
		env.CPULimit = cpuLimit.(float64)
	}
	if memoryLimit, ok := updates["memory_limit"]; ok {
		env.MemoryLimit = memoryLimit.(int64)
	}

	// Validate resource limits
	if err := s.ValidateResourceLimits(env.CPULimit, env.MemoryLimit); err != nil {
		return err
	}

	return s.repo.Update(env)
}

// DeleteEnvironment deletes a development environment
func (s *devEnvironmentService) DeleteEnvironment(id uint, createdBy string) error {
	return s.repo.Delete(id, createdBy)
}

// ValidateEnvVars validates environment variables
func (s *devEnvironmentService) ValidateEnvVars(envVars map[string]string) error {
	for key, value := range envVars {
		if strings.TrimSpace(key) == "" {
			return errors.New("environment variable key cannot be empty")
		}
		if strings.Contains(key, "=") {
			return errors.New("environment variable key cannot contain '=' character")
		}
		// 可以添加更多验证规则
		_ = value // 暂时不验证值
	}
	return nil
}

// GetEnvironmentVars gets environment variables
func (s *devEnvironmentService) GetEnvironmentVars(id uint, createdBy string) (map[string]string, error) {
	env, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return nil, err
	}

	var envVars map[string]string
	if env.EnvVars != "" {
		if err := json.Unmarshal([]byte(env.EnvVars), &envVars); err != nil {
			return nil, fmt.Errorf("failed to parse environment variables: %v", err)
		}
	}

	if envVars == nil {
		envVars = make(map[string]string)
	}

	return envVars, nil
}

// UpdateEnvironmentVars updates environment variables
func (s *devEnvironmentService) UpdateEnvironmentVars(id uint, createdBy string, envVars map[string]string) error {
	env, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// Validate environment variables
	if err := s.ValidateEnvVars(envVars); err != nil {
		return err
	}

	// Serialize environment variables
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return fmt.Errorf("failed to serialize environment variables: %v", err)
	}

	env.EnvVars = string(envVarsJSON)
	return s.repo.Update(env)
}

// ValidateResourceLimits validates resource limits
func (s *devEnvironmentService) ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error {
	if cpuLimit <= 0 || cpuLimit > 16 {
		return errors.New("CPU limit must be between 0 and 16 cores")
	}
	if memoryLimit <= 0 || memoryLimit > 32768 {
		return errors.New("memory limit must be between 0 and 32GB (32768MB)")
	}
	return nil
}

// validateEnvironmentData validates environment data
func (s *devEnvironmentService) validateEnvironmentData(name, envType string, cpuLimit float64, memoryLimit int64) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("environment name is required")
	}

	// 从系统配置获取支持的环境类型
	envTypesJSON, err := s.configService.GetValue("dev_environment_types")
	if err != nil {
		return errors.New("failed to get environment types configuration")
	}

	// 解析环境类型配置
	var envTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(envTypesJSON), &envTypes); err != nil {
		return errors.New("failed to parse environment types configuration")
	}

	// 验证是否为配置支持的环境类型 key
	found := false
	for _, supportedType := range envTypes {
		if key, ok := supportedType["key"].(string); ok && key == envType {
			found = true
			break
		}
	}
	if !found {
		return errors.New("unsupported environment type")
	}

	return s.ValidateResourceLimits(cpuLimit, memoryLimit)
}

// GetAvailableEnvironmentTypes gets available environment types from system configuration
func (s *devEnvironmentService) GetAvailableEnvironmentTypes() ([]map[string]interface{}, error) {
	envTypesJSON, err := s.configService.GetValue("dev_environment_types")
	if err != nil {
		return nil, fmt.Errorf("failed to get dev environment types config: %v", err)
	}

	var envTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(envTypesJSON), &envTypes); err != nil {
		return nil, fmt.Errorf("failed to parse dev environment types: %v", err)
	}

	return envTypes, nil
}
