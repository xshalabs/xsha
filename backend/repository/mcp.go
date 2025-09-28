package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type mcpRepository struct {
	db *gorm.DB
}

func NewMCPRepository(db *gorm.DB) MCPRepository {
	return &mcpRepository{db: db}
}

func (r *mcpRepository) Create(mcp *database.MCP) error {
	return r.db.Create(mcp).Error
}

func (r *mcpRepository) GetByID(id uint) (*database.MCP, error) {
	var mcp database.MCP
	err := r.db.Where("id = ?", id).First(&mcp).Error
	if err != nil {
		return nil, err
	}
	return &mcp, nil
}

func (r *mcpRepository) GetByIDWithAdmin(id uint) (*database.MCP, error) {
	var mcp database.MCP
	err := r.db.Preload("Admin").Where("id = ?", id).First(&mcp).Error
	if err != nil {
		return nil, err
	}
	return &mcp, nil
}

func (r *mcpRepository) GetByName(name string) (*database.MCP, error) {
	var mcp database.MCP
	err := r.db.Where("name = ?", name).First(&mcp).Error
	if err != nil {
		return nil, err
	}
	return &mcp, nil
}

func (r *mcpRepository) List(name *string, enabled *bool, page, pageSize int) ([]database.MCP, int64, error) {
	var mcps []database.MCP
	var total int64

	query := r.db.Model(&database.MCP{})

	if name != nil {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}

	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Admin").Offset(offset).Limit(pageSize).Find(&mcps).Error; err != nil {
		return nil, 0, err
	}

	return mcps, total, nil
}

func (r *mcpRepository) ListByAdminAccess(adminID uint, role database.AdminRole, name *string, enabled *bool, page, pageSize int) ([]database.MCP, int64, error) {
	var mcps []database.MCP
	var total int64

	query := r.db.Model(&database.MCP{})

	// Super admin can see all, admin can only see their own
	if role != database.AdminRoleSuperAdmin {
		query = query.Where("admin_id = ?", adminID)
	}

	if name != nil {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}

	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Admin").Offset(offset).Limit(pageSize).Find(&mcps).Error; err != nil {
		return nil, 0, err
	}

	return mcps, total, nil
}

func (r *mcpRepository) Update(mcp *database.MCP) error {
	return r.db.Save(mcp).Error
}

func (r *mcpRepository) Delete(id uint) error {
	// Use transaction to ensure atomicity
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete all project relationships first
		if err := tx.Exec("DELETE FROM mcp_projects WHERE mcp_id = ?", id).Error; err != nil {
			return err
		}

		// Delete all environment relationships
		if err := tx.Exec("DELETE FROM mcp_environments WHERE mcp_id = ?", id).Error; err != nil {
			return err
		}

		// Hard delete the MCP record (not soft delete)
		return tx.Unscoped().Where("id = ?", id).Delete(&database.MCP{}).Error
	})
}

// Project association methods

func (r *mcpRepository) AddProject(mcpID, projectID uint) error {
	// Check if the relationship already exists
	var count int64
	err := r.db.Table("mcp_projects").
		Where("mcp_id = ? AND project_id = ?", mcpID, projectID).
		Count(&count).Error
	if err != nil {
		return err
	}

	// If relationship doesn't exist, create it
	if count == 0 {
		return r.db.Exec("INSERT INTO mcp_projects (mcp_id, project_id) VALUES (?, ?)", mcpID, projectID).Error
	}

	return nil // Already exists, no error
}

func (r *mcpRepository) RemoveProject(mcpID, projectID uint) error {
	return r.db.Exec("DELETE FROM mcp_projects WHERE mcp_id = ? AND project_id = ?", mcpID, projectID).Error
}

func (r *mcpRepository) GetProjects(mcpID uint) ([]database.Project, error) {
	var projects []database.Project
	err := r.db.Table("projects").
		Joins("INNER JOIN mcp_projects ON projects.id = mcp_projects.project_id").
		Where("mcp_projects.mcp_id = ?", mcpID).
		Find(&projects).Error
	return projects, err
}

func (r *mcpRepository) GetProjectMCPs(projectID uint) ([]database.MCP, error) {
	var mcps []database.MCP
	err := r.db.Table("mcps").
		Joins("INNER JOIN mcp_projects ON mcps.id = mcp_projects.mcp_id").
		Where("mcp_projects.project_id = ?", projectID).
		Find(&mcps).Error
	return mcps, err
}

func (r *mcpRepository) GetEnabledProjectMCPs(projectID uint) ([]database.MCP, error) {
	var mcps []database.MCP
	err := r.db.Table("mcps").
		Joins("INNER JOIN mcp_projects ON mcps.id = mcp_projects.mcp_id").
		Where("mcp_projects.project_id = ? AND mcps.enabled = ?", projectID, true).
		Find(&mcps).Error
	return mcps, err
}

func (r *mcpRepository) IsAssociatedWithProject(mcpID, projectID uint) (bool, error) {
	var count int64
	err := r.db.Table("mcp_projects").
		Where("mcp_id = ? AND project_id = ?", mcpID, projectID).
		Count(&count).Error
	return count > 0, err
}

// Environment association methods

func (r *mcpRepository) AddEnvironment(mcpID, devEnvID uint) error {
	// Check if the relationship already exists
	var count int64
	err := r.db.Table("mcp_environments").
		Where("mcp_id = ? AND dev_environment_id = ?", mcpID, devEnvID).
		Count(&count).Error
	if err != nil {
		return err
	}

	// If relationship doesn't exist, create it
	if count == 0 {
		return r.db.Exec("INSERT INTO mcp_environments (mcp_id, dev_environment_id) VALUES (?, ?)", mcpID, devEnvID).Error
	}

	return nil // Already exists, no error
}

func (r *mcpRepository) RemoveEnvironment(mcpID, devEnvID uint) error {
	return r.db.Exec("DELETE FROM mcp_environments WHERE mcp_id = ? AND dev_environment_id = ?", mcpID, devEnvID).Error
}

func (r *mcpRepository) GetEnvironments(mcpID uint) ([]database.DevEnvironment, error) {
	var environments []database.DevEnvironment
	err := r.db.Table("dev_environments").
		Joins("INNER JOIN mcp_environments ON dev_environments.id = mcp_environments.dev_environment_id").
		Where("mcp_environments.mcp_id = ?", mcpID).
		Find(&environments).Error
	return environments, err
}

func (r *mcpRepository) GetEnvironmentMCPs(devEnvID uint) ([]database.MCP, error) {
	var mcps []database.MCP
	err := r.db.Table("mcps").
		Joins("INNER JOIN mcp_environments ON mcps.id = mcp_environments.mcp_id").
		Where("mcp_environments.dev_environment_id = ?", devEnvID).
		Find(&mcps).Error
	return mcps, err
}

func (r *mcpRepository) GetEnabledEnvironmentMCPs(devEnvID uint) ([]database.MCP, error) {
	var mcps []database.MCP
	err := r.db.Table("mcps").
		Joins("INNER JOIN mcp_environments ON mcps.id = mcp_environments.mcp_id").
		Where("mcp_environments.dev_environment_id = ? AND mcps.enabled = ?", devEnvID, true).
		Find(&mcps).Error
	return mcps, err
}

func (r *mcpRepository) IsAssociatedWithEnvironment(mcpID, devEnvID uint) (bool, error) {
	var count int64
	err := r.db.Table("mcp_environments").
		Where("mcp_id = ? AND dev_environment_id = ?", mcpID, devEnvID).
		Count(&count).Error
	return count > 0, err
}

// Permission helper methods

func (r *mcpRepository) IsOwner(mcpID, adminID uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.MCP{}).
		Where("id = ? AND admin_id = ?", mcpID, adminID).
		Count(&count).Error
	return count > 0, err
}

func (r *mcpRepository) CountByAdminID(adminID uint) (int64, error) {
	var count int64
	err := r.db.Model(&database.MCP{}).
		Where("admin_id = ?", adminID).
		Count(&count).Error
	return count, err
}