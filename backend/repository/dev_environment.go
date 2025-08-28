package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type devEnvironmentRepository struct {
	db *gorm.DB
}

func NewDevEnvironmentRepository(db *gorm.DB) DevEnvironmentRepository {
	return &devEnvironmentRepository{db: db}
}

func (r *devEnvironmentRepository) Create(env *database.DevEnvironment) error {
	return r.db.Create(env).Error
}

func (r *devEnvironmentRepository) GetByID(id uint) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Where("id = ?", id).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *devEnvironmentRepository) GetByName(name string) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Where("name = ?", name).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *devEnvironmentRepository) List(name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	var environments []database.DevEnvironment
	var total int64

	query := r.db.Model(&database.DevEnvironment{})

	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}

	if dockerImage != nil && *dockerImage != "" {
		query = query.Where("docker_image LIKE ?", "%"+*dockerImage+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&environments).Error; err != nil {
		return nil, 0, err
	}

	return environments, total, nil
}

func (r *devEnvironmentRepository) Update(env *database.DevEnvironment) error {
	return r.db.Save(env).Error
}

func (r *devEnvironmentRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.DevEnvironment{}).Error
}

