package services

import (
	"errors"
	"fmt"
	"strings"
	"xsha-backend/database"
	"xsha-backend/repository"

	"gorm.io/gorm"
)

type systemConfigService struct {
	repo repository.SystemConfigRepository
}

func NewSystemConfigService(repo repository.SystemConfigRepository) SystemConfigService {
	return &systemConfigService{
		repo: repo,
	}
}

func (s *systemConfigService) ListAllConfigs() ([]database.SystemConfig, error) {
	return s.repo.ListAll()
}

func (s *systemConfigService) BatchUpdateConfigs(configItems []ConfigUpdateItem) error {
	for _, item := range configItems {
		if err := s.ValidateConfigData(item.ConfigKey, item.ConfigValue, item.Category); err != nil {
			return fmt.Errorf("validation failed for key %s: %v", item.ConfigKey, err)
		}

		existingConfig, err := s.repo.GetByKey(item.ConfigKey)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing config for key %s: %v", item.ConfigKey, err)
		}

		if existingConfig != nil {
			if !existingConfig.IsEditable {
				return fmt.Errorf("configuration %s is not editable", item.ConfigKey)
			}

			existingConfig.ConfigValue = item.ConfigValue
			if item.Description != "" {
				existingConfig.Description = item.Description
			}
			if item.Category != "" {
				existingConfig.Category = item.Category
			}
			if item.IsEditable != nil {
				existingConfig.IsEditable = *item.IsEditable
			}

			if err := s.repo.Update(existingConfig); err != nil {
				return fmt.Errorf("failed to update config %s: %v", item.ConfigKey, err)
			}
		} else {
			isEditable := true
			if item.IsEditable != nil {
				isEditable = *item.IsEditable
			}

			category := item.Category
			if category == "" {
				category = "general"
			}

			if err := s.repo.SetValueWithCategory(item.ConfigKey, item.ConfigValue, item.Description, category, isEditable); err != nil {
				return fmt.Errorf("failed to create config %s: %v", item.ConfigKey, err)
			}
		}
	}

	return nil
}

func (s *systemConfigService) GetValue(key string) (string, error) {
	return s.repo.GetValue(key)
}

func (s *systemConfigService) SetValue(key, value string) error {
	return s.repo.SetValue(key, value)
}

func (s *systemConfigService) InitializeDefaultConfigs() error {
	return s.repo.InitializeDefaultConfigs()
}

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

	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-') {
			return errors.New("configuration key can only contain letters, numbers, underscores, and hyphens")
		}
	}

	return nil
}
