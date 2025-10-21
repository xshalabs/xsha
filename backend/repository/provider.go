package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type providerRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) ProviderRepository {
	return &providerRepository{db: db}
}

func (r *providerRepository) Create(provider *database.Provider) error {
	return r.db.Create(provider).Error
}

func (r *providerRepository) GetByID(id uint) (*database.Provider, error) {
	var provider database.Provider
	err := r.db.First(&provider, id).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *providerRepository) GetByName(name string) (*database.Provider, error) {
	var provider database.Provider
	err := r.db.Where("name = ?", name).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *providerRepository) List(name *string, providerType *string, page, pageSize int) ([]database.Provider, int64, error) {
	var providers []database.Provider
	var total int64

	query := r.db.Model(&database.Provider{}).Preload("Admin").Preload("Admin.Avatar")

	// Apply filters
	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}
	if providerType != nil && *providerType != "" {
		query = query.Where("type = ?", *providerType)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *providerRepository) ListByAdminAccess(adminID uint, role database.AdminRole, name *string, providerType *string, page, pageSize int) ([]database.Provider, int64, error) {
	var providers []database.Provider
	var total int64

	query := r.db.Model(&database.Provider{}).Preload("Admin").Preload("Admin.Avatar")

	// Apply role-based filtering
	if role != database.AdminRoleSuperAdmin {
		query = query.Where("admin_id = ?", adminID)
	}

	// Apply other filters
	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}
	if providerType != nil && *providerType != "" {
		query = query.Where("type = ?", *providerType)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *providerRepository) Update(provider *database.Provider) error {
	return r.db.Save(provider).Error
}

func (r *providerRepository) Delete(id uint) error {
	return r.db.Delete(&database.Provider{}, id).Error
}

// Permission helper methods

func (r *providerRepository) IsOwner(providerID, adminID uint) (bool, error) {
	var provider database.Provider
	err := r.db.Select("admin_id").First(&provider, providerID).Error
	if err != nil {
		return false, err
	}
	return provider.AdminID != nil && *provider.AdminID == adminID, nil
}

func (r *providerRepository) CountByAdminID(adminID uint) (int64, error) {
	var count int64
	err := r.db.Model(&database.Provider{}).Where("admin_id = ?", adminID).Count(&count).Error
	return count, err
}

// DevEnvironment association helper

func (r *providerRepository) CountByProviderID(providerID uint) (int64, error) {
	var count int64
	err := r.db.Model(&database.DevEnvironment{}).Where("provider_id = ?", providerID).Count(&count).Error
	return count, err
}
