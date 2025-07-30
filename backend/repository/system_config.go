package repository

import (
	"encoding/json"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type systemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓库实例
func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

// Create 创建系统配置
func (r *systemConfigRepository) Create(config *database.SystemConfig) error {
	return r.db.Create(config).Error
}

// GetByKey 根据配置键获取配置
func (r *systemConfigRepository) GetByKey(key string) (*database.SystemConfig, error) {
	var config database.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ListAll 获取所有配置列表
func (r *systemConfigRepository) ListAll() ([]database.SystemConfig, error) {
	var configs []database.SystemConfig
	err := r.db.Order("category, config_key").Find(&configs).Error
	return configs, err
}

// Update 更新配置
func (r *systemConfigRepository) Update(config *database.SystemConfig) error {
	return r.db.Save(config).Error
}

// GetValue 根据配置键获取配置值
func (r *systemConfigRepository) GetValue(key string) (string, error) {
	config, err := r.GetByKey(key)
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

// SetValue 设置配置值
func (r *systemConfigRepository) SetValue(key, value string) error {
	config, err := r.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 配置不存在，创建新配置
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

// SetValueWithCategory 设置配置值及其他属性
func (r *systemConfigRepository) SetValueWithCategory(key, value, description, category string, isEditable bool) error {
	config, err := r.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 配置不存在，创建新配置
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

// InitializeDefaultConfigs 初始化默认配置
func (r *systemConfigRepository) InitializeDefaultConfigs() error {
	// 检查是否已初始化
	_, err := r.GetByKey("dev_environment_types")
	if err == nil {
		// 配置已存在，跳过初始化
		return nil
	}

	// 初始化开发环境类型配置
	defaultDevEnvTypes := []map[string]interface{}{
		{
			"image": "claude-code:latest",
			"name":  "Claude Code",
			"key":   "claude-code",
		},
		{
			"image": "claude-code:latest",
			"name":  "Claude Code - CN",
			"key":   "claude-code-cn",
		},
	}

	devEnvTypesJSON, err := json.Marshal(defaultDevEnvTypes)
	if err != nil {
		return err
	}

	err = r.SetValueWithCategory(
		"dev_environment_types",
		string(devEnvTypesJSON),
		"开发环境类型配置，定义可用的开发环境及其对应的Docker镜像",
		"dev_environment",
		true,
	)
	if err != nil {
		return err
	}

	return nil
}
