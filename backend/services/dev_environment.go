package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"xsha-backend/config"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"
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

func (s *devEnvironmentService) CreateEnvironment(name, description, systemPrompt, envType, dockerImage string, cpuLimit float64, memoryLimit int64, envVars map[string]string, adminID uint, createdBy string) (*database.DevEnvironment, error) {
	if err := s.validateEnvironmentData(name, envType, cpuLimit, memoryLimit); err != nil {
		return nil, err
	}

	if err := s.ValidateEnvVars(envVars); err != nil {
		return nil, err
	}

	if existing, _ := s.repo.GetByName(name); existing != nil {
		return nil, appErrors.ErrEnvironmentNameExists
	}

	if strings.TrimSpace(dockerImage) == "" {
		return nil, appErrors.ErrEnvironmentDockerImageRequired
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
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		Type:         envType,
		DockerImage:  dockerImage,
		CPULimit:     cpuLimit,
		MemoryLimit:  memoryLimit,
		EnvVars:      string(envVarsJSON),
		SessionDir:   sessionDir,
		AdminID:      &adminID,
		CreatedBy:    createdBy,
	}

	if err := s.repo.Create(env); err != nil {
		return nil, err
	}

	return env, nil
}

func (s *devEnvironmentService) GetEnvironment(id uint) (*database.DevEnvironment, error) {
	return s.repo.GetByID(id)
}

func (s *devEnvironmentService) ListEnvironments(name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	return s.repo.List(name, dockerImage, page, pageSize)
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
	if systemPrompt, ok := updates["system_prompt"]; ok {
		env.SystemPrompt = systemPrompt.(string)
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

	tasks, _, err := s.taskRepo.List(nil, nil, nil, nil, &env.ID, "created_at", "desc", 1, 1)
	if err != nil {
		return fmt.Errorf("failed to check environment usage: %v", err)
	}
	if len(tasks) > 0 {
		return appErrors.ErrEnvironmentUsedByTasks
	}

	// Delete session directory if it exists
	if env.SessionDir != "" {
		// Convert relative path to absolute path for deletion
		absoluteSessionDir := s.getAbsoluteSessionPath(env.SessionDir)
		if err := os.RemoveAll(absoluteSessionDir); err != nil {
			// Log the error but don't fail the deletion
			// as the database record should still be removed
			utils.Warn("Failed to delete session directory", "sessionDir", absoluteSessionDir, "error", err)
		}
	}

	return s.repo.Delete(id)
}

func (s *devEnvironmentService) ValidateEnvVars(envVars map[string]string) error {
	for key, value := range envVars {
		if strings.TrimSpace(key) == "" {
			return appErrors.ErrEnvironmentVarKeyEmpty
		}
		if strings.Contains(key, "=") {
			return appErrors.ErrEnvironmentVarKeyInvalidChar
		}
		_ = value
	}
	return nil
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
		return appErrors.ErrEnvironmentCPULimitInvalid
	}
	if memoryLimit <= 0 || memoryLimit > 32768 {
		return appErrors.ErrEnvironmentMemoryLimitInvalid
	}
	return nil
}

func (s *devEnvironmentService) validateEnvironmentData(name, envType string, cpuLimit float64, memoryLimit int64) error {
	if strings.TrimSpace(name) == "" {
		return appErrors.ErrEnvironmentNameRequired
	}

	envImagesJSON, err := s.configService.GetValue("dev_environment_images")
	if err != nil {
		return appErrors.ErrEnvironmentImagesConfigFailed
	}

	var envImages []map[string]interface{}
	if err := json.Unmarshal([]byte(envImagesJSON), &envImages); err != nil {
		return appErrors.ErrEnvironmentImagesConfigParseError
	}

	found := false
	for _, envImage := range envImages {
		if imageType, ok := envImage["type"].(string); ok && imageType == envType {
			found = true
			break
		}
	}
	if !found {
		return appErrors.ErrEnvironmentUnsupportedType
	}

	return s.ValidateResourceLimits(cpuLimit, memoryLimit)
}

func (s *devEnvironmentService) GetAvailableEnvironmentImages() ([]map[string]interface{}, error) {
	envImagesJSON, err := s.configService.GetValue("dev_environment_images")
	if err != nil {
		return nil, fmt.Errorf("failed to get dev environment images config: %v", err)
	}

	var envImages []map[string]interface{}
	if err := json.Unmarshal([]byte(envImagesJSON), &envImages); err != nil {
		return nil, fmt.Errorf("failed to parse dev environment images: %v", err)
	}

	return envImages, nil
}

// generateSessionDir creates a unique session directory for the dev environment
func (s *devEnvironmentService) generateSessionDir() (string, error) {
	// Create base sessions directory if it doesn't exist
	if err := os.MkdirAll(s.config.DevSessionsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create dev sessions base directory: %v", err)
	}

	// Generate unique directory name using safe characters only
	// Use timestamp and random suffix to ensure uniqueness
	timestamp := utils.Now().Unix()
	// Generate a short random suffix for better uniqueness
	randomSuffix := utils.Now().Nanosecond() % 10000
	dirName := fmt.Sprintf("env-%d-%04d", timestamp, randomSuffix)

	// Create the absolute path for directory creation
	absoluteSessionDir := filepath.Join(s.config.DevSessionsDir, dirName)
	if err := os.MkdirAll(absoluteSessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %v", err)
	}

	// Return relative path for database storage
	return dirName, nil
}

// getAbsoluteSessionPath converts a relative session path to absolute path
func (s *devEnvironmentService) getAbsoluteSessionPath(relativePath string) string {
	if relativePath == "" {
		return ""
	}

	// If already absolute, return as is
	if filepath.IsAbs(relativePath) {
		return relativePath
	}

	return filepath.Join(s.config.DevSessionsDir, relativePath)
}
