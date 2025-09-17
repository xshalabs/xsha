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
	defaultDevEnvImages := []map[string]interface{}{
		{
			"image": "ghcr.io/xshalabs/dev-image-registry/claude-code:node18-1.0.67",
			"name":  "Claude Code node18_1.0.67",
			"type":  "claude-code",
		},
		{
			"image": "ghcr.io/xshalabs/dev-image-registry/claude-code:node20-1.0.67",
			"name":  "Claude Code node20_1.0.67",
			"type":  "claude-code",
		},
		{
			"image": "registry.cn-hangzhou.aliyuncs.com/hzbs/claude-code:node18-1.0.67",
			"name":  "[CN]Claude Code node18_1.0.67",
			"type":  "claude-code",
		},
		{
			"image": "registry.cn-hangzhou.aliyuncs.com/hzbs/claude-code:node20-1.0.67",
			"name":  "[CN]Claude Code node20_1.0.67",
			"type":  "claude-code",
		},
	}

	devEnvImagesJSON, err := json.Marshal(defaultDevEnvImages)
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
			key:         "dev_environment_images",
			value:       string(devEnvImagesJSON),
			description: "Development environment image configuration, defines available Docker images and their corresponding environment images",
			category:    "dev_environment",
			formType:    string(database.ConfigFormTypeTextarea),
			sortOrder:   10,
		},
		{
			key:         "git_proxy_enabled",
			value:       "false",
			description: "Enable or disable HTTP proxy for Git operations",
			category:    "git",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   20,
		},
		{
			key:         "git_proxy_http",
			value:       "",
			description: "HTTP proxy URL for Git operations (e.g., http://proxy.example.com:8080)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   30,
		},
		{
			key:         "git_proxy_https",
			value:       "",
			description: "HTTPS proxy URL for Git operations (e.g., https://proxy.example.com:8080)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   40,
		},
		{
			key:         "git_proxy_no_proxy",
			value:       "",
			description: "Comma-separated list of domains to bypass proxy (e.g., localhost,127.0.0.1,.local)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   50,
		},
		{
			key:         "git_clone_timeout",
			value:       "5m",
			description: "Timeout for Git clone operations (e.g., 5m, 300s)",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   60,
		},
		{
			key:         "git_ssl_verify",
			value:       "false",
			description: "Enable or disable SSL verification for Git operations",
			category:    "git",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   70,
		},
		{
			key:         "docker_timeout",
			value:       "120m",
			description: "Timeout for Docker execution operations (e.g., 120m, 7200s)",
			category:    "docker",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   80,
		},
		{
			key:         "smtp_enabled",
			value:       "false",
			description: "Enable or disable email service for sending welcome emails and notifications",
			category:    "email",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   90,
		},
		{
			key:         "smtp_host",
			value:       "",
			description: "SMTP server hostname (e.g., smtp.gmail.com, smtp.163.com)",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   100,
		},
		{
			key:         "smtp_port",
			value:       "587",
			description: "SMTP server port (usually 587 for TLS, 465 for SSL, 25 for plain)",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   110,
		},
		{
			key:         "smtp_username",
			value:       "",
			description: "SMTP authentication username (usually your email address)",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   120,
		},
		{
			key:         "smtp_password",
			value:       "",
			description: "SMTP authentication password or app-specific password",
			category:    "email",
			formType:    string(database.ConfigFormTypePassword),
			sortOrder:   130,
		},
		{
			key:         "smtp_from",
			value:       "",
			description: "Sender email address (must be authorized by SMTP server)",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   140,
		},
		{
			key:         "smtp_from_name",
			value:       "xsha Platform",
			description: "Sender display name that appears in email",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   150,
		},
		{
			key:         "smtp_use_tls",
			value:       "true",
			description: "Use TLS encryption for SMTP connection",
			category:    "email",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   160,
		},
		{
			key:         "smtp_skip_verify",
			value:       "false",
			description: "Skip TLS certificate verification (not recommended for production)",
			category:    "email",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   170,
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
