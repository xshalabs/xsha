package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type devEnvironmentRepository struct {
	db *gorm.DB
}

// NewDevEnvironmentRepository 创建开发环境仓库实例
func NewDevEnvironmentRepository(db *gorm.DB) DevEnvironmentRepository {
	return &devEnvironmentRepository{db: db}
}

// Create 创建开发环境
func (r *devEnvironmentRepository) Create(env *database.DevEnvironment) error {
	return r.db.Create(env).Error
}

// GetByID 根据ID获取开发环境
func (r *devEnvironmentRepository) GetByID(id uint, createdBy string) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Where("id = ? AND created_by = ?", id, createdBy).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// GetByName 根据名称获取开发环境
func (r *devEnvironmentRepository) GetByName(name, createdBy string) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Where("name = ? AND created_by = ?", name, createdBy).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// List 分页获取开发环境列表
func (r *devEnvironmentRepository) List(createdBy string, envType *database.DevEnvironmentType, name *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	var environments []database.DevEnvironment
	var total int64

	query := r.db.Model(&database.DevEnvironment{}).Where("created_by = ?", createdBy)

	if envType != nil {
		query = query.Where("type = ?", *envType)
	}

	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&environments).Error; err != nil {
		return nil, 0, err
	}

	return environments, total, nil
}

// Update 更新开发环境
func (r *devEnvironmentRepository) Update(env *database.DevEnvironment) error {
	return r.db.Save(env).Error
}

// Delete 删除开发环境
func (r *devEnvironmentRepository) Delete(id uint, createdBy string) error {
	return r.db.Where("id = ? AND created_by = ?", id, createdBy).Delete(&database.DevEnvironment{}).Error
}
