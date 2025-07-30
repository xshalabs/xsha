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

// GetByID 根据ID获取配置
func (r *systemConfigRepository) GetByID(id uint) (*database.SystemConfig, error) {
	var config database.SystemConfig
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// List 分页获取配置列表
func (r *systemConfigRepository) List(category string, page, pageSize int) ([]database.SystemConfig, int64, error) {
	var configs []database.SystemConfig
	var total int64

	query := r.db.Model(&database.SystemConfig{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("category, config_key").Find(&configs).Error
	if err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

// ListByCategory 根据分类获取所有配置
func (r *systemConfigRepository) ListByCategory(category string) ([]database.SystemConfig, error) {
	var configs []database.SystemConfig
	err := r.db.Where("category = ?", category).Order("config_key").Find(&configs).Error
	return configs, err
}

// Update 更新配置
func (r *systemConfigRepository) Update(config *database.SystemConfig) error {
	return r.db.Save(config).Error
}

// Delete 删除配置
func (r *systemConfigRepository) Delete(id uint) error {
	return r.db.Delete(&database.SystemConfig{}, id).Error
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

// GetConfigsByCategory 根据分类获取配置键值对
func (r *systemConfigRepository) GetConfigsByCategory(category string) (map[string]string, error) {
	configs, err := r.ListByCategory(category)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, config := range configs {
		result[config.ConfigKey] = config.ConfigValue
	}
	return result, nil
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
			"name":  "Claude Code",
			"image": "claude-code:latest",
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
