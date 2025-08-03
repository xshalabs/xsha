package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/repository"
)

type devEnvironmentService struct {
	repo          repository.DevEnvironmentRepository
	taskRepo      repository.TaskRepository
	configService SystemConfigService
	config        *config.Config
}

func NewDevEnvironmentService(repo repository.DevEnvironmentRepository, taskRepo repository.TaskRepository, configService SystemConfigService, cfg *config.Config) DevEnvironmentService {
	return &devEnvironmentService{
		repo:          repo,
		taskRepo:      taskRepo,
		configService: configService,
		config:        cfg,
	}
}

func (s *devEnvironmentService) CreateEnvironment(name, description, envType string, cpuLimit float64, memoryLimit int64, envVars map[string]string, createdBy string) (*database.DevEnvironment, error) {
	if err := s.validateEnvironmentData(name, envType, cpuLimit, memoryLimit); err != nil {
		return nil, err
	}

	if err := s.ValidateEnvVars(envVars); err != nil {
		return nil, err
	}

	if existing, _ := s.repo.GetByName(name); existing != nil {
		return nil, errors.New("environment name already exists")
	}

	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize environment variables: %v", err)
	}

	// Generate session directory
	sessionDir, err := s.generateSessionDir()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session directory: %v", err)
	}

	env := &database.DevEnvironment{
		Name:        name,
		Description: description,
		Type:        envType,
		CPULimit:    cpuLimit,
		MemoryLimit: memoryLimit,
		EnvVars:     string(envVarsJSON),
		SessionDir:  sessionDir,
		CreatedBy:   createdBy,
	}

	if err := s.repo.Create(env); err != nil {
		return nil, err
	}

	return env, nil
}

func (s *devEnvironmentService) GetEnvironment(id uint) (*database.DevEnvironment, error) {
	return s.repo.GetByID(id)
}

func (s *devEnvironmentService) ListEnvironments(envType *string, name *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	return s.repo.List(envType, name, page, pageSize)
}

func (s *devEnvironmentService) UpdateEnvironment(id uint, updates map[string]interface{}) error {
	env, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

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

	if err := s.ValidateResourceLimits(env.CPULimit, env.MemoryLimit); err != nil {
		return err
	}

	return s.repo.Update(env)
}

func (s *devEnvironmentService) DeleteEnvironment(id uint) error {
	env, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	tasks, _, err := s.taskRepo.List(nil, nil, nil, nil, &env.ID, 1, 1)
	if err != nil {
		return fmt.Errorf("failed to check environment usage: %v", err)
	}
	if len(tasks) > 0 {
		return errors.New("dev_environment.delete_used_by_tasks")
	}

	return s.repo.Delete(id)
}

func (s *devEnvironmentService) ValidateEnvVars(envVars map[string]string) error {
	for key, value := range envVars {
		if strings.TrimSpace(key) == "" {
			return errors.New("environment variable key cannot be empty")
		}
		if strings.Contains(key, "=") {
			return errors.New("environment variable key cannot contain '=' character")
		}
		_ = value
	}
	return nil
}

func (s *devEnvironmentService) GetEnvironmentVars(id uint) (map[string]string, error) {
	env, err := s.repo.GetByID(id)
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

func (s *devEnvironmentService) UpdateEnvironmentVars(id uint, envVars map[string]string) error {
	env, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.ValidateEnvVars(envVars); err != nil {
		return err
	}

	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return fmt.Errorf("failed to serialize environment variables: %v", err)
	}

	env.EnvVars = string(envVarsJSON)
	return s.repo.Update(env)
}

func (s *devEnvironmentService) ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error {
	if cpuLimit <= 0 || cpuLimit > 16 {
		return errors.New("CPU limit must be between 0 and 16 cores")
	}
	if memoryLimit <= 0 || memoryLimit > 32768 {
		return errors.New("memory limit must be between 0 and 32GB (32768MB)")
	}
	return nil
}

func (s *devEnvironmentService) validateEnvironmentData(name, envType string, cpuLimit float64, memoryLimit int64) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("environment name is required")
	}

	envTypesJSON, err := s.configService.GetValue("dev_environment_types")
	if err != nil {
		return errors.New("failed to get environment types configuration")
	}

	var envTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(envTypesJSON), &envTypes); err != nil {
		return errors.New("failed to parse environment types configuration")
	}

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

// generateSessionDir creates a unique session directory for the dev environment
func (s *devEnvironmentService) generateSessionDir() (string, error) {
	// Create base sessions directory if it doesn't exist
	if err := os.MkdirAll(s.config.DevSessionsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create dev sessions base directory: %v", err)
	}

	// Generate unique directory name using safe characters only
	// Use timestamp and random suffix to ensure uniqueness
	timestamp := time.Now().Unix()
	// Generate a short random suffix for better uniqueness
	randomSuffix := time.Now().Nanosecond() % 10000
	dirName := fmt.Sprintf("env-%d-%04d", timestamp, randomSuffix)
	sessionDir := filepath.Join(s.config.DevSessionsDir, dirName)

	// Create the session directory
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %v", err)
	}

	return sessionDir, nil
}
