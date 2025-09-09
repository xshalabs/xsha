package repository

import (
	"xsha-backend/database"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *database.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) GetByID(id uint) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Credential").Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetByName(name string) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Credential").Where("name = ?", name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) List(name string, protocol *database.GitProtocolType, sortBy, sortDirection string, page, pageSize int) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	query := r.db.Model(&database.Project{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if protocol != nil {
		query = query.Where("protocol = ?", *protocol)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Handle sorting
	var orderClause string
	switch sortBy {
	case "name":
		orderClause = "name " + sortDirection
	case "created_at":
		orderClause = "created_at " + sortDirection
	case "task_count":
		// For task_count sorting, we need to join with task counts
		// We'll use a subquery to get task counts and order by it
		subQuery := r.db.Table("tasks").
			Select("project_id, COUNT(*) as task_count").
			Where("deleted_at IS NULL").
			Group("project_id")

		query = query.
			Select("projects.*, COALESCE(task_counts.task_count, 0) as task_count").
			Joins("LEFT JOIN (?) as task_counts ON projects.id = task_counts.project_id", subQuery).
			Order("task_count " + sortDirection + ", projects.created_at DESC")
	default:
		orderClause = "created_at " + sortDirection
	}

	if sortBy != "task_count" {
		query = query.Order(orderClause)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Credential").Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *projectRepository) Update(project *database.Project) error {
	err := r.db.Save(project).Error
	if err != nil {
		return err
	}
	
	// Reload the credential relationship after saving to ensure consistency
	return r.db.Preload("Credential").Where("id = ?", project.ID).First(project).Error
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.Project{}).Error
}

func (r *projectRepository) UpdateLastUsed(id uint) error {
	now := utils.Now()
	return r.db.Model(&database.Project{}).
		Where("id = ?", id).
		Update("last_used", now).Error
}

func (r *projectRepository) GetByCredentialID(credentialID uint) ([]database.Project, error) {
	var projects []database.Project
	err := r.db.Where("credential_id = ?", credentialID).Find(&projects).Error
	return projects, err
}

func (r *projectRepository) GetTaskCounts(projectIDs []uint) (map[uint]int64, error) {
	if len(projectIDs) == 0 {
		return make(map[uint]int64), nil
	}

	type TaskCountResult struct {
		ProjectID uint  `gorm:"column:project_id"`
		Count     int64 `gorm:"column:count"`
	}

	var results []TaskCountResult
	err := r.db.Table("tasks").
		Select("project_id, COUNT(*) as count").
		Where("project_id IN ? AND deleted_at IS NULL", projectIDs).
		Group("project_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	taskCounts := make(map[uint]int64)
	for _, projectID := range projectIDs {
		taskCounts[projectID] = 0
	}

	for _, result := range results {
		taskCounts[result.ProjectID] = result.Count
	}

	return taskCounts, nil
}

func (r *projectRepository) GetAdminCounts(projectIDs []uint) (map[uint]int64, error) {
	if len(projectIDs) == 0 {
		return make(map[uint]int64), nil
	}

	type AdminCountResult struct {
		ProjectID uint  `gorm:"column:project_id"`
		Count     int64 `gorm:"column:count"`
	}

	var results []AdminCountResult

	// Count admins from many-to-many relationship table
	err := r.db.Table("project_admins").
		Select("project_id, COUNT(DISTINCT admin_id) as count").
		Where("project_id IN ?", projectIDs).
		Group("project_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// Initialize all projects with count 0
	adminCounts := make(map[uint]int64)
	for _, projectID := range projectIDs {
		adminCounts[projectID] = 0
	}

	// Set counts from many-to-many relationships
	for _, result := range results {
		adminCounts[result.ProjectID] = result.Count
	}

	// Now we need to add primary admins (AdminID field) to the count
	// Get projects with their AdminID to count primary admins
	var projects []database.Project
	err = r.db.Select("id, admin_id").Where("id IN ?", projectIDs).Find(&projects).Error
	if err != nil {
		return nil, err
	}

	// Add primary admin count (1 if AdminID is not nil)
	for _, project := range projects {
		if project.AdminID != nil {
			adminCounts[project.ID]++
		}
	}

	return adminCounts, nil
}

// GetByIDWithAdmins retrieves a project with its admin relationships preloaded
func (r *projectRepository) GetByIDWithAdmins(id uint) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Admins").Preload("Credential").Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListByAdminAccess lists projects that an admin has access to
func (r *projectRepository) ListByAdminAccess(adminID uint, name string, protocol *database.GitProtocolType, sortBy, sortDirection string, page, pageSize int) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	// Base query joining with the many-to-many relationship table
	query := r.db.Model(&database.Project{}).
		Joins("LEFT JOIN project_admins ON projects.id = project_admins.project_id").
		Where("projects.admin_id = ? OR project_admins.admin_id = ?", adminID, adminID)

	if name != "" {
		query = query.Where("projects.name LIKE ?", "%"+name+"%")
	}

	if protocol != nil {
		query = query.Where("projects.protocol = ?", *protocol)
	}

	// Count total records
	if err := query.Distinct("projects.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Handle sorting - similar to List method
	var orderClause string
	switch sortBy {
	case "name":
		orderClause = "projects.name " + sortDirection
	case "created_at":
		orderClause = "projects.created_at " + sortDirection
	case "task_count":
		// For task_count sorting with admin access
		subQuery := r.db.Table("tasks").
			Select("project_id, COUNT(*) as task_count").
			Where("deleted_at IS NULL").
			Group("project_id")

		query = query.
			Select("projects.*, COALESCE(task_counts.task_count, 0) as task_count").
			Joins("LEFT JOIN (?) as task_counts ON projects.id = task_counts.project_id", subQuery).
			Order("task_count " + sortDirection + ", projects.created_at DESC")
	default:
		orderClause = "projects.created_at " + sortDirection
	}

	if sortBy != "task_count" {
		query = query.Order(orderClause)
	}

	offset := (page - 1) * pageSize
	if err := query.
		Preload("Credential").
		Preload("Admins", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Avatar", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, uuid, original_name")
			})
		}).
		Distinct("projects.*").
		Offset(offset).Limit(pageSize).
		Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// AddAdmin adds an admin to the project's admin list
func (r *projectRepository) AddAdmin(projectID, adminID uint) error {
	// Check if the relationship already exists
	var count int64
	err := r.db.Table("project_admins").
		Where("project_id = ? AND admin_id = ?", projectID, adminID).
		Count(&count).Error
	if err != nil {
		return err
	}

	// If relationship doesn't exist, create it
	if count == 0 {
		return r.db.Exec("INSERT INTO project_admins (project_id, admin_id) VALUES (?, ?)", projectID, adminID).Error
	}

	return nil // Already exists, no error
}

// RemoveAdmin removes an admin from the project's admin list
func (r *projectRepository) RemoveAdmin(projectID, adminID uint) error {
	// Use direct SQL to delete from the many-to-many relationship table
	return r.db.Exec("DELETE FROM project_admins WHERE project_id = ? AND admin_id = ?", projectID, adminID).Error
}

// GetAdmins retrieves all admins for a specific project
func (r *projectRepository) GetAdmins(projectID uint) ([]database.Admin, error) {
	var admins []database.Admin
	err := r.db.Table("admins").
		Preload("Avatar", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, uuid, original_name")
		}).
		Joins("JOIN project_admins ON admins.id = project_admins.admin_id").
		Where("project_admins.project_id = ?", projectID).
		Find(&admins).Error
	return admins, err
}

// IsAdminForProject checks if an admin has access to a project (either through direct ownership or many-to-many relationship)
func (r *projectRepository) IsAdminForProject(projectID, adminID uint) (bool, error) {
	var count int64

	// Check if admin is the direct owner of the project (AdminID field)
	err := r.db.Model(&database.Project{}).
		Where("id = ? AND admin_id = ?", projectID, adminID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check if admin is in the many-to-many relationship table
	err = r.db.Table("project_admins").
		Where("project_id = ? AND admin_id = ?", projectID, adminID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
