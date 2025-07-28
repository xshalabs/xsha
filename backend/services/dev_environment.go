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
	repo repository.DevEnvironmentRepository
}

// NewDevEnvironmentService 创建开发环境服务实例
func NewDevEnvironmentService(repo repository.DevEnvironmentRepository) DevEnvironmentService {
	return &devEnvironmentService{
		repo: repo,
	}
}

// CreateEnvironment 创建开发环境
func (s *devEnvironmentService) CreateEnvironment(name, description, envType, createdBy string, cpuLimit float64, memoryLimit int64, envVars map[string]string) (*database.DevEnvironment, error) {
	// 验证输入
	if err := s.validateEnvironmentData(name, envType, cpuLimit, memoryLimit); err != nil {
		return nil, err
	}

	// 验证环境变量
	if err := s.ValidateEnvVars(envVars); err != nil {
		return nil, err
	}

	// 检查环境名称是否已存在
	if existing, _ := s.repo.GetByName(name, createdBy); existing != nil {
		return nil, errors.New("environment name already exists")
	}

	// 序列化环境变量
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize environment variables: %v", err)
	}

	// 创建环境对象
	env := &database.DevEnvironment{
		Name:        name,
		Description: description,
		Type:        database.DevEnvironmentType(envType),
		CPULimit:    cpuLimit,
		MemoryLimit: memoryLimit,
		EnvVars:     string(envVarsJSON),
		CreatedBy:   createdBy,
	}

	// 保存到数据库
	if err := s.repo.Create(env); err != nil {
		return nil, err
	}

	return env, nil
}

// GetEnvironment 获取开发环境
func (s *devEnvironmentService) GetEnvironment(id uint, createdBy string) (*database.DevEnvironment, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListEnvironments 获取开发环境列表
func (s *devEnvironmentService) ListEnvironments(createdBy string, envType *database.DevEnvironmentType, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	return s.repo.List(createdBy, envType, page, pageSize)
}

// UpdateEnvironment 更新开发环境
func (s *devEnvironmentService) UpdateEnvironment(id uint, createdBy string, updates map[string]interface{}) error {
	env, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 更新基本信息
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

	// 验证资源限制
	if err := s.ValidateResourceLimits(env.CPULimit, env.MemoryLimit); err != nil {
		return err
	}

	return s.repo.Update(env)
}

// DeleteEnvironment 删除开发环境
func (s *devEnvironmentService) DeleteEnvironment(id uint, createdBy string) error {
	return s.repo.Delete(id, createdBy)
}

// ValidateEnvVars 验证环境变量
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

// GetEnvironmentVars 获取环境变量
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

// UpdateEnvironmentVars 更新环境变量
func (s *devEnvironmentService) UpdateEnvironmentVars(id uint, createdBy string, envVars map[string]string) error {
	env, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 验证环境变量
	if err := s.ValidateEnvVars(envVars); err != nil {
		return err
	}

	// 序列化环境变量
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return fmt.Errorf("failed to serialize environment variables: %v", err)
	}

	env.EnvVars = string(envVarsJSON)
	return s.repo.Update(env)
}

// ValidateResourceLimits 验证资源限制
func (s *devEnvironmentService) ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error {
	if cpuLimit <= 0 || cpuLimit > 16 {
		return errors.New("CPU limit must be between 0 and 16 cores")
	}
	if memoryLimit <= 0 || memoryLimit > 32768 {
		return errors.New("memory limit must be between 0 and 32GB (32768MB)")
	}
	return nil
}

// validateEnvironmentData 验证环境数据
func (s *devEnvironmentService) validateEnvironmentData(name, envType string, cpuLimit float64, memoryLimit int64) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("environment name is required")
	}
	if envType != string(database.DevEnvTypeClaude) &&
		envType != string(database.DevEnvTypeGemini) &&
		envType != string(database.DevEnvTypeOpenCode) {
		return errors.New("unsupported environment type")
	}
	return s.ValidateResourceLimits(cpuLimit, memoryLimit)
}
