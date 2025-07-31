package repository

import (
	"encoding/json"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type systemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

func (r *systemConfigRepository) Create(config *database.SystemConfig) error {
	return r.db.Create(config).Error
}

func (r *systemConfigRepository) GetByKey(key string) (*database.SystemConfig, error) {
	var config database.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *systemConfigRepository) ListAll() ([]database.SystemConfig, error) {
	var configs []database.SystemConfig
	err := r.db.Order("category, config_key").Find(&configs).Error
	return configs, err
}

func (r *systemConfigRepository) Update(config *database.SystemConfig) error {
	return r.db.Save(config).Error
}

func (r *systemConfigRepository) GetValue(key string) (string, error) {
	config, err := r.GetByKey(key)
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

func (r *systemConfigRepository) SetValue(key, value string) error {
	config, err := r.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			newConfig := &database.SystemConfig{
				ConfigKey:   key,
				ConfigValue: value,
				Category:    "general",
				IsEditable:  true,
			}
			return r.Create(newConfig)
		}
		return err
	}

	config.ConfigValue = value
	return r.Update(config)
}

func (r *systemConfigRepository) SetValueWithCategory(key, value, description, category string, isEditable bool) error {
	config, err := r.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			newConfig := &database.SystemConfig{
				ConfigKey:   key,
				ConfigValue: value,
				Description: description,
				Category:    category,
				IsEditable:  isEditable,
			}
			return r.Create(newConfig)
		}
		return err
	}

	config.ConfigValue = value
	config.Description = description
	config.Category = category
	config.IsEditable = isEditable
	return r.Update(config)
}

func (r *systemConfigRepository) InitializeDefaultConfigs() error {
	_, err := r.GetByKey("dev_environment_types")
	if err == nil {
		return nil
	}

	defaultDevEnvTypes := []map[string]interface{}{
		{
			"image": "claude-code:latest",
			"name":  "Claude Code",
			"key":   "claude-code",
		},
	}

	devEnvTypesJSON, err := json.Marshal(defaultDevEnvTypes)
	if err != nil {
		return err
	}

	err = r.SetValueWithCategory(
		"dev_environment_types",
		string(devEnvTypesJSON),
		"Development environment type configuration, defines available development environments and their corresponding Docker images",
		"dev_environment",
		true,
	)
	if err != nil {
		return err
	}

	return nil
}
