package repository

import (
	"fmt"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type notifierRepository struct {
	db *gorm.DB
}

func NewNotifierRepository(db *gorm.DB) NotifierRepository {
	return &notifierRepository{db: db}
}

func (r *notifierRepository) Create(notifier *database.Notifier) error {
	return r.db.Create(notifier).Error
}

func (r *notifierRepository) GetByID(id uint) (*database.Notifier, error) {
	var notifier database.Notifier
	err := r.db.First(&notifier, id).Error
	if err != nil {
		return nil, err
	}
	return &notifier, nil
}

func (r *notifierRepository) GetByIDWithAdmin(id uint) (*database.Notifier, error) {
	var notifier database.Notifier
	err := r.db.Preload("Admin").Preload("Admin.Avatar").First(&notifier, id).Error
	if err != nil {
		return nil, err
	}
	return &notifier, nil
}

func (r *notifierRepository) GetByName(name string) (*database.Notifier, error) {
	var notifier database.Notifier
	err := r.db.Where("name = ?", name).First(&notifier).Error
	if err != nil {
		return nil, err
	}
	return &notifier, nil
}

func (r *notifierRepository) List(name *string, notifierTypes []database.NotifierType, isEnabled *bool, page, pageSize int) ([]database.Notifier, int64, error) {
	var notifiers []database.Notifier
	var total int64

	query := r.db.Model(&database.Notifier{}).Preload("Admin").Preload("Admin.Avatar")

	// Apply filters
	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}
	if len(notifierTypes) > 0 {
		query = query.Where("type IN ?", notifierTypes)
	}
	if isEnabled != nil {
		query = query.Where("is_enabled = ?", *isEnabled)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifiers).Error; err != nil {
		return nil, 0, err
	}

	return notifiers, total, nil
}

func (r *notifierRepository) ListByAdminAccess(adminID uint, role database.AdminRole, name *string, notifierTypes []database.NotifierType, isEnabled *bool, page, pageSize int) ([]database.Notifier, int64, error) {
	var notifiers []database.Notifier
	var total int64

	query := r.db.Model(&database.Notifier{}).Preload("Admin").Preload("Admin.Avatar")

	// Apply role-based filtering
	if role != database.AdminRoleSuperAdmin {
		query = query.Where("admin_id = ?", adminID)
	}

	// Apply other filters
	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}
	if len(notifierTypes) > 0 {
		query = query.Where("type IN ?", notifierTypes)
	}
	if isEnabled != nil {
		query = query.Where("is_enabled = ?", *isEnabled)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifiers).Error; err != nil {
		return nil, 0, err
	}

	return notifiers, total, nil
}

func (r *notifierRepository) Update(notifier *database.Notifier) error {
	return r.db.Save(notifier).Error
}

func (r *notifierRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First remove all project associations
		if err := tx.Exec("DELETE FROM project_notifiers WHERE notifier_id = ?", id).Error; err != nil {
			return err
		}

		// Then delete the notifier
		return tx.Delete(&database.Notifier{}, id).Error
	})
}

// Project association methods

func (r *notifierRepository) AddProject(notifierID, projectID uint) error {
	// Check if association already exists
	var count int64
	err := r.db.Table("project_notifiers").
		Where("notifier_id = ? AND project_id = ?", notifierID, projectID).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("notifier already associated with project")
	}

	// Create association
	return r.db.Exec("INSERT INTO project_notifiers (notifier_id, project_id) VALUES (?, ?)",
		notifierID, projectID).Error
}

func (r *notifierRepository) RemoveProject(notifierID, projectID uint) error {
	return r.db.Exec("DELETE FROM project_notifiers WHERE notifier_id = ? AND project_id = ?",
		notifierID, projectID).Error
}

func (r *notifierRepository) GetProjects(notifierID uint) ([]database.Project, error) {
	var projects []database.Project
	err := r.db.
		Joins("JOIN project_notifiers ON projects.id = project_notifiers.project_id").
		Where("project_notifiers.notifier_id = ?", notifierID).
		Find(&projects).Error
	return projects, err
}

func (r *notifierRepository) GetProjectNotifiers(projectID uint) ([]database.Notifier, error) {
	var notifiers []database.Notifier
	err := r.db.
		Preload("Admin").Preload("Admin.Avatar").
		Joins("JOIN project_notifiers ON notifiers.id = project_notifiers.notifier_id").
		Where("project_notifiers.project_id = ?", projectID).
		Find(&notifiers).Error
	return notifiers, err
}

func (r *notifierRepository) GetEnabledProjectNotifiers(projectID uint) ([]database.Notifier, error) {
	var notifiers []database.Notifier
	err := r.db.
		Preload("Admin").Preload("Admin.Avatar").
		Joins("JOIN project_notifiers ON notifiers.id = project_notifiers.notifier_id").
		Where("project_notifiers.project_id = ? AND notifiers.is_enabled = ?", projectID, true).
		Find(&notifiers).Error
	return notifiers, err
}

func (r *notifierRepository) IsAssociatedWithProject(notifierID, projectID uint) (bool, error) {
	var count int64
	err := r.db.Table("project_notifiers").
		Where("notifier_id = ? AND project_id = ?", notifierID, projectID).
		Count(&count).Error
	return count > 0, err
}

// Permission helper methods

func (r *notifierRepository) IsOwner(notifierID, adminID uint) (bool, error) {
	var notifier database.Notifier
	err := r.db.Select("admin_id").First(&notifier, notifierID).Error
	if err != nil {
		return false, err
	}
	return notifier.AdminID != nil && *notifier.AdminID == adminID, nil
}

func (r *notifierRepository) CountByAdminID(adminID uint) (int64, error) {
	var count int64
	err := r.db.Model(&database.Notifier{}).Where("admin_id = ?", adminID).Count(&count).Error
	return count, err
}
