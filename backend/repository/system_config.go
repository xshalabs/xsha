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
	err := r.db.Order("sort_order ASC").Find(&configs).Error
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

func (r *systemConfigRepository) SetValueWithCategoryAndSort(key, value, description, category, formType string, isEditable bool, sortOrder int) error {
	config, err := r.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if formType == "" {
				formType = string(database.ConfigFormTypeInput)
			}
			newConfig := &database.SystemConfig{
				ConfigKey:   key,
				ConfigValue: value,
				Description: description,
				Category:    category,
				FormType:    database.ConfigFormType(formType),
				IsEditable:  isEditable,
				SortOrder:   sortOrder,
			}
			return r.Create(newConfig)
		}
		return err
	}

	config.ConfigValue = value
	config.Description = description
	config.Category = category
	if formType != "" {
		config.FormType = database.ConfigFormType(formType)
	}
	config.IsEditable = isEditable
	config.SortOrder = sortOrder
	return r.Update(config)
}

func (r *systemConfigRepository) InitializeDefaultConfigs() error {
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

	defaultConfigs := []struct {
		key         string
		value       string
		description string
		category    string
		formType    string
		sortOrder   int
	}{
		{
			key:         "admin_user",
			value:       "xshauser",
			description: "Administrator username for system login",
			category:    "auth",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   10,
		},
		{
			key:         "admin_password",
			value:       "xshapass",
			description: "Administrator password for system login",
			category:    "auth",
			formType:    string(database.ConfigFormTypePassword),
			sortOrder:   20,
		},
		{
			key:         "dev_environment_types",
			value:       string(devEnvTypesJSON),
			description: "Development environment type configuration, defines available development environments and their corresponding Docker images",
			category:    "dev_environment",
			formType:    string(database.ConfigFormTypeTextarea),
			sortOrder:   30,
		},
		{
			key:         "git_proxy_enabled",
			value:       "false",
			description: "Enable or disable HTTP proxy for Git operations",
			category:    "git",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   40,
		},
		{
			key:         "git_proxy_http",
			value:       "",
			description: "HTTP proxy URL for Git operations (e.g., http://proxy.example.com:8080)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   50,
		},
		{
			key:         "git_proxy_https",
			value:       "",
			description: "HTTPS proxy URL for Git operations (e.g., https://proxy.example.com:8080)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   60,
		},
		{
			key:         "git_proxy_no_proxy",
			value:       "",
			description: "Comma-separated list of domains to bypass proxy (e.g., localhost,127.0.0.1,.local)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   70,
		},
	}

	for _, config := range defaultConfigs {
		existingConfig, err := r.GetByKey(config.key)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if existingConfig != nil {
			continue
		}

		if err := r.SetValueWithCategoryAndSort(
			config.key,
			config.value,
			config.description,
			config.category,
			config.formType,
			true,
			config.sortOrder,
		); err != nil {
			return err
		}
	}

	return nil
}
