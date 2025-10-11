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

func (r *systemConfigRepository) GetByKeys(keys []string) (map[string]*database.SystemConfig, error) {
	var configs []database.SystemConfig
	err := r.db.Where("config_key IN ?", keys).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]*database.SystemConfig)
	for i := range configs {
		result[configs[i].ConfigKey] = &configs[i]
	}

	return result, nil
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

func (r *systemConfigRepository) CreateOrUpdate(key, value, name, description, category, formType string, isEditable bool, sortOrder int) error {
	config, err := r.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if formType == "" {
				formType = string(database.ConfigFormTypeInput)
			}
			newConfig := &database.SystemConfig{
				ConfigKey:   key,
				ConfigValue: value,
				Name:        name,
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

	config.Name = name
	config.Description = description
	config.Category = category
	if formType != "" {
		config.FormType = database.ConfigFormType(formType)
	}
	config.IsEditable = isEditable
	config.SortOrder = sortOrder
	return r.Update(config)
}

// MergeDevEnvImages merges default dev environment images with existing user-configured images
func (r *systemConfigRepository) MergeDevEnvImages(defaultImages []map[string]interface{}, name, description, category, formType string, sortOrder int) error {
	config, err := r.GetByKey("dev_environment_images")

	// If config doesn't exist, create it with default images
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			defaultImagesJSON, err := json.Marshal(defaultImages)
			if err != nil {
				return err
			}

			newConfig := &database.SystemConfig{
				ConfigKey:   "dev_environment_images",
				ConfigValue: string(defaultImagesJSON),
				Name:        name,
				Description: description,
				Category:    category,
				FormType:    database.ConfigFormType(formType),
				IsEditable:  true,
				SortOrder:   sortOrder,
			}
			return r.Create(newConfig)
		}
		return err
	}

	// Parse existing images
	var existingImages []map[string]interface{}
	if err := json.Unmarshal([]byte(config.ConfigValue), &existingImages); err != nil {
		return err
	}

	// Merge images: preserve order while deduplicating based on image URL
	// Use map only for O(1) lookup, not for storing final results
	imageSet := make(map[string]bool)

	// 1. First, keep all existing images in their original order (preserving user customizations)
	mergedImages := make([]map[string]interface{}, 0, len(existingImages)+len(defaultImages))
	for _, img := range existingImages {
		if imageURL, ok := img["image"].(string); ok {
			mergedImages = append(mergedImages, img)
			imageSet[imageURL] = true
		}
	}

	// 2. Then, append default images that don't already exist (in their defined order)
	for _, img := range defaultImages {
		if imageURL, ok := img["image"].(string); ok {
			if !imageSet[imageURL] {
				mergedImages = append(mergedImages, img)
				imageSet[imageURL] = true
			}
		}
	}

	// Update config value
	mergedJSON, err := json.Marshal(mergedImages)
	if err != nil {
		return err
	}

	config.ConfigValue = string(mergedJSON)
	config.Name = name
	config.Description = description
	config.Category = category
	if formType != "" {
		config.FormType = database.ConfigFormType(formType)
	}
	config.SortOrder = sortOrder

	return r.Update(config)
}

func (r *systemConfigRepository) InitializeDefaultConfigs() error {
	defaultDevEnvImages := []map[string]interface{}{
		// 2.0.13
		{
			"image": "ghcr.io/xshalabs/dev-image-registry/claude-code:node18-2.0.13",
			"name":  "Claude Code node18_2.0.13",
			"type":  "claude-code",
		},
		{
			"image": "ghcr.io/xshalabs/dev-image-registry/claude-code:node20-2.0.13",
			"name":  "Claude Code node20_2.0.13",
			"type":  "claude-code",
		},
		{
			"image": "registry.cn-hangzhou.aliyuncs.com/hzbs/claude-code:node18-2.0.13",
			"name":  "[CN]Claude Code node18_2.0.13",
			"type":  "claude-code",
		},
		{
			"image": "registry.cn-hangzhou.aliyuncs.com/hzbs/claude-code:node20-2.0.13",
			"name":  "[CN]Claude Code node20_2.0.13",
			"type":  "claude-code",
		},
	}

	// Merge dev environment images with existing configuration
	if err := r.MergeDevEnvImages(
		defaultDevEnvImages,
		"config.name.dev_environment.images",
		"config.description.dev_environment.images",
		"dev_environment",
		string(database.ConfigFormTypeTextarea),
		10,
	); err != nil {
		return err
	}

	defaultConfigs := []struct {
		key         string
		value       string
		name        string
		description string
		category    string
		formType    string
		sortOrder   int
	}{
		{
			key:         "git_proxy_enabled",
			value:       "false",
			name:        "config.name.git.proxy_enabled",
			description: "config.description.git.proxy_enabled",
			category:    "git",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   20,
		},
		{
			key:         "git_proxy_http",
			value:       "",
			name:        "config.name.git.proxy_http",
			description: "config.description.git.proxy_http",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   30,
		},
		{
			key:         "git_proxy_https",
			value:       "",
			name:        "config.name.git.proxy_https",
			description: "config.description.git.proxy_https",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   40,
		},
		{
			key:         "git_proxy_no_proxy",
			value:       "",
			name:        "config.name.git.proxy_no_proxy",
			description: "config.description.git.proxy_no_proxy",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   50,
		},
		{
			key:         "git_clone_timeout",
			value:       "5m",
			name:        "config.name.git.clone_timeout",
			description: "config.description.git.clone_timeout",
			category:    "git",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   60,
		},
		{
			key:         "git_ssl_verify",
			value:       "false",
			name:        "config.name.git.ssl_verify",
			description: "config.description.git.ssl_verify",
			category:    "git",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   70,
		},
		{
			key:         "docker_timeout",
			value:       "120m",
			name:        "config.name.docker.timeout",
			description: "config.description.docker.timeout",
			category:    "docker",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   80,
		},
		{
			key:         "smtp_enabled",
			value:       "false",
			name:        "config.name.email.smtp_enabled",
			description: "config.description.email.smtp_enabled",
			category:    "email",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   90,
		},
		{
			key:         "smtp_host",
			value:       "",
			name:        "config.name.email.smtp_host",
			description: "config.description.email.smtp_host",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   100,
		},
		{
			key:         "smtp_port",
			value:       "587",
			name:        "config.name.email.smtp_port",
			description: "config.description.email.smtp_port",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   110,
		},
		{
			key:         "smtp_username",
			value:       "",
			name:        "config.name.email.smtp_username",
			description: "config.description.email.smtp_username",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   120,
		},
		{
			key:         "smtp_password",
			value:       "",
			name:        "config.name.email.smtp_password",
			description: "config.description.email.smtp_password",
			category:    "email",
			formType:    string(database.ConfigFormTypePassword),
			sortOrder:   130,
		},
		{
			key:         "smtp_from",
			value:       "",
			name:        "config.name.email.smtp_from",
			description: "config.description.email.smtp_from",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   140,
		},
		{
			key:         "smtp_from_name",
			value:       "xsha Platform",
			name:        "config.name.email.smtp_from_name",
			description: "config.description.email.smtp_from_name",
			category:    "email",
			formType:    string(database.ConfigFormTypeInput),
			sortOrder:   150,
		},
		{
			key:         "smtp_use_tls",
			value:       "true",
			name:        "config.name.email.smtp_use_tls",
			description: "config.description.email.smtp_use_tls",
			category:    "email",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   160,
		},
		{
			key:         "smtp_skip_verify",
			value:       "false",
			name:        "config.name.email.smtp_skip_verify",
			description: "config.description.email.smtp_skip_verify",
			category:    "email",
			formType:    string(database.ConfigFormTypeSwitch),
			sortOrder:   170,
		},
	}

	for _, config := range defaultConfigs {
		if err := r.CreateOrUpdate(
			config.key,
			config.value,
			config.name,
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
