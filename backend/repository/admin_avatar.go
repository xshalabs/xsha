package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type adminAvatarRepository struct {
	db *gorm.DB
}

func NewAdminAvatarRepository(db *gorm.DB) AdminAvatarRepository {
	return &adminAvatarRepository{db: db}
}

func (r *adminAvatarRepository) Create(avatar *database.AdminAvatar) error {
	return r.db.Create(avatar).Error
}

func (r *adminAvatarRepository) GetByID(id uint) (*database.AdminAvatar, error) {
	var avatar database.AdminAvatar
	err := r.db.Preload("Admin").First(&avatar, id).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *adminAvatarRepository) GetByUUID(uuid string) (*database.AdminAvatar, error) {
	var avatar database.AdminAvatar
	err := r.db.Preload("Admin").Where("uuid = ?", uuid).First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *adminAvatarRepository) GetByAdminID(adminID uint) (*database.AdminAvatar, error) {
	var avatar database.AdminAvatar
	err := r.db.Where("admin_id = ?", adminID).First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *adminAvatarRepository) Update(avatar *database.AdminAvatar) error {
	return r.db.Save(avatar).Error
}

func (r *adminAvatarRepository) Delete(id uint) error {
	return r.db.Delete(&database.AdminAvatar{}, id).Error
}