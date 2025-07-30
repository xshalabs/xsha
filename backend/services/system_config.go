package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"xsha-backend/database"
	"xsha-backend/repository"
)

type systemConfigService struct {
	repo repository.SystemConfigRepository
}

// NewSystemConfigService creates a new system configuration service instance
func NewSystemConfigService(repo repository.SystemConfigRepository) SystemConfigService {
	return &systemConfigService{
		repo: repo,
	}
}

// GetConfig gets a system configuration by ID
func (s *systemConfigService) GetConfig(id uint) (*database.SystemConfig, error) {
	return s.repo.GetByID(id)
}

// GetConfigByKey gets a system configuration by key
func (s *systemConfigService) GetConfigByKey(key string) (*database.SystemConfig, error) {
	return s.repo.GetByKey(key)
}

// ListConfigs gets a list of system configurations
func (s *systemConfigService) ListConfigs(category string, page, pageSize int) ([]database.SystemConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.List(category, page, pageSize)
}

// UpdateConfig updates a system configuration
func (s *systemConfigService) UpdateConfig(id uint, updates map[string]interface{}) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Check if configuration is editable
	if !config.IsEditable {
		return errors.New("configuration is not editable")
	}

	// Validate updates
	if key, ok := updates["config_key"].(string); ok {
		if value, ok := updates["config_value"].(string); ok {
			if category, ok := updates["category"].(string); ok {
				if err := s.ValidateConfigData(key, value, category); err != nil {
					return err
				}
			}
		}
	}

	// Apply updates
	if key, ok := updates["config_key"]; ok {
		config.ConfigKey = key.(string)
	}
	if value, ok := updates["config_value"]; ok {
		config.ConfigValue = value.(string)
	}
	if description, ok := updates["description"]; ok {
		config.Description = description.(string)
	}
	if category, ok := updates["category"]; ok {
		config.Category = category.(string)
	}
	if isEditable, ok := updates["is_editable"]; ok {
		config.IsEditable = isEditable.(bool)
	}

	return s.repo.Update(config)
}

// GetValue gets a configuration value by key
func (s *systemConfigService) GetValue(key string) (string, error) {
	return s.repo.GetValue(key)
}

// SetValue sets a configuration value by key
func (s *systemConfigService) SetValue(key, value string) error {
	return s.repo.SetValue(key, value)
}

// GetConfigsByCategory gets all configurations in a category as key-value pairs
func (s *systemConfigService) GetConfigsByCategory(category string) (map[string]string, error) {
	return s.repo.GetConfigsByCategory(category)
}

// GetDevEnvironmentTypes gets available development environment types
func (s *systemConfigService) GetDevEnvironmentTypes() ([]map[string]interface{}, error) {
	value, err := s.repo.GetValue("dev_environment_types")
	if err != nil {
		return nil, err
	}

	var envTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(value), &envTypes); err != nil {
		return nil, fmt.Errorf("failed to parse development environment types: %v", err)
	}

	return envTypes, nil
}

// UpdateDevEnvironmentTypes updates available development environment types
func (s *systemConfigService) UpdateDevEnvironmentTypes(envTypes []map[string]interface{}) error {
	// Validate environment types
	for _, envType := range envTypes {
		name, nameOk := envType["name"].(string)
		image, imageOk := envType["image"].(string)
		if !nameOk || !imageOk || name == "" || image == "" {
			return errors.New("each environment type must have a valid name and image")
		}
	}

	value, err := json.Marshal(envTypes)
	if err != nil {
		return fmt.Errorf("failed to serialize environment types: %v", err)
	}

	return s.repo.SetValue("dev_environment_types", string(value))
}

// InitializeDefaultConfigs initializes default configurations
func (s *systemConfigService) InitializeDefaultConfigs() error {
	return s.repo.InitializeDefaultConfigs()
}

// ValidateConfigData validates configuration data
func (s *systemConfigService) ValidateConfigData(key, value, category string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("configuration key is required")
	}
	if strings.TrimSpace(value) == "" {
		return errors.New("configuration value is required")
	}
	if strings.TrimSpace(category) == "" {
		return errors.New("configuration category is required")
	}

	// Validate key format (only allow alphanumeric, underscore, and hyphen)
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-') {
			return errors.New("configuration key can only contain letters, numbers, underscores, and hyphens")
		}
	}

	return nil
}
