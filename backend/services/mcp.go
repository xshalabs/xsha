package services

import (
	"encoding/json"
	"fmt"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
)

type mcpService struct {
	repo        repository.MCPRepository
	projectRepo repository.ProjectRepository
	devEnvRepo  repository.DevEnvironmentRepository
}

func NewMCPService(repo repository.MCPRepository, projectRepo repository.ProjectRepository, devEnvRepo repository.DevEnvironmentRepository) MCPService {
	return &mcpService{
		repo:        repo,
		projectRepo: projectRepo,
		devEnvRepo:  devEnvRepo,
	}
}

// CRUD operations with permission checks

func (s *mcpService) CreateMCP(name, description, config string, enabled bool, admin *database.Admin) (*database.MCP, error) {
	// Only admin and super_admin can create MCPs
	if admin.Role != database.AdminRoleAdmin && admin.Role != database.AdminRoleSuperAdmin {
		return nil, appErrors.ErrInsufficientPermissions
	}

	// Validate MCP name format
	if err := s.ValidateMCPName(name); err != nil {
		return nil, err
	}

	// Check if name already exists
	if existing, _ := s.repo.GetByName(name); existing != nil {
		return nil, fmt.Errorf("MCP with name '%s' already exists", name)
	}

	// Validate config JSON
	if err := s.ValidateMCPConfig(config); err != nil {
		return nil, err
	}

	mcp := &database.MCP{
		Name:        name,
		Description: description,
		Config:      config,
		Enabled:     enabled,
		AdminID:     &admin.ID,
		CreatedBy:   admin.Username,
	}

	if err := s.repo.Create(mcp); err != nil {
		return nil, err
	}

	return mcp, nil
}

func (s *mcpService) GetMCP(id uint, admin *database.Admin) (*database.MCP, error) {
	mcp, err := s.repo.GetByIDWithAdmin(id)
	if err != nil {
		return nil, err
	}

	// Check permissions
	canAccess, err := s.CanAdminAccessMCP(id, admin)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.ErrInsufficientPermissions
	}

	return mcp, nil
}

func (s *mcpService) ListMCPs(admin *database.Admin, name *string, enabled *bool, page, pageSize int) ([]database.MCPListItemResponse, int64, error) {
	mcps, total, err := s.repo.ListByAdminAccess(admin.ID, admin.Role, name, enabled, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	var result []database.MCPListItemResponse
	for _, mcp := range mcps {
		var adminInfo *database.MinimalAdminResponse
		if mcp.Admin != nil {
			adminInfo = &database.MinimalAdminResponse{
				ID:       mcp.Admin.ID,
				Username: mcp.Admin.Username,
				Name:     mcp.Admin.Name,
				Email:    mcp.Admin.Email,
			}
		}

		item := database.MCPListItemResponse{
			ID:          mcp.ID,
			CreatedAt:   mcp.CreatedAt,
			UpdatedAt:   mcp.UpdatedAt,
			Name:        mcp.Name,
			Description: mcp.Description,
			Config:      mcp.Config,
			Enabled:     mcp.Enabled,
			AdminID:     mcp.AdminID,
			Admin:       adminInfo,
			CreatedBy:   mcp.CreatedBy,
		}
		result = append(result, item)
	}

	return result, total, nil
}

func (s *mcpService) UpdateMCP(id uint, updates map[string]interface{}, admin *database.Admin) error {
	// Check if MCP exists and get it
	mcp, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Check permissions
	canAccess, err := s.CanAdminAccessMCP(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.ErrInsufficientPermissions
	}

	// Validate config if being updated
	if config, exists := updates["config"]; exists {
		configStr, ok := config.(string)
		if !ok {
			return appErrors.ErrInvalidInput
		}
		if err := s.ValidateMCPConfig(configStr); err != nil {
			return err
		}
	}

	// Check if name is being updated and validate format and uniqueness
	if name, exists := updates["name"]; exists {
		nameStr := name.(string)

		// Validate name format
		if err := s.ValidateMCPName(nameStr); err != nil {
			return err
		}

		// Check uniqueness if name is different
		if nameStr != mcp.Name {
			if existing, _ := s.repo.GetByName(nameStr); existing != nil && existing.ID != id {
				return fmt.Errorf("MCP with name '%s' already exists", nameStr)
			}
		}
	}

	// Update the fields
	for field, value := range updates {
		switch field {
		case "name":
			mcp.Name = value.(string)
		case "description":
			mcp.Description = value.(string)
		case "config":
			mcp.Config = value.(string)
		case "enabled":
			mcp.Enabled = value.(bool)
		}
	}

	return s.repo.Update(mcp)
}

func (s *mcpService) DeleteMCP(id uint, admin *database.Admin) error {
	// Check permissions
	canAccess, err := s.CanAdminAccessMCP(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.ErrInsufficientPermissions
	}

	return s.repo.Delete(id)
}

// Project association methods

func (s *mcpService) AddMCPToProject(projectID, mcpID uint, admin *database.Admin) error {
	// Check if project exists and admin has access
	canAccess, err := s.projectRepo.IsAdminForProject(projectID, admin.ID)
	if err != nil {
		return err
	}
	if !canAccess && admin.Role != database.AdminRoleSuperAdmin {
		return appErrors.ErrInsufficientPermissions
	}

	// Check if MCP exists and admin has access
	canAccessMCP, err := s.CanAdminAccessMCP(mcpID, admin)
	if err != nil {
		return err
	}
	if !canAccessMCP {
		return appErrors.ErrInsufficientPermissions
	}

	return s.repo.AddProject(mcpID, projectID)
}

func (s *mcpService) RemoveMCPFromProject(projectID, mcpID uint, admin *database.Admin) error {
	// Check if project exists and admin has access
	canAccess, err := s.projectRepo.IsAdminForProject(projectID, admin.ID)
	if err != nil {
		return err
	}
	if !canAccess && admin.Role != database.AdminRoleSuperAdmin {
		return appErrors.ErrInsufficientPermissions
	}

	return s.repo.RemoveProject(mcpID, projectID)
}

func (s *mcpService) GetProjectMCPs(projectID uint) ([]database.MCPListItemResponse, error) {
	mcps, err := s.repo.GetProjectMCPs(projectID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	var result []database.MCPListItemResponse
	for _, mcp := range mcps {
		item := database.MCPListItemResponse{
			ID:          mcp.ID,
			CreatedAt:   mcp.CreatedAt,
			UpdatedAt:   mcp.UpdatedAt,
			Name:        mcp.Name,
			Description: mcp.Description,
			Config:      mcp.Config,
			Enabled:     mcp.Enabled,
			AdminID:     mcp.AdminID,
			CreatedBy:   mcp.CreatedBy,
		}
		result = append(result, item)
	}

	return result, nil
}

// Environment association methods

func (s *mcpService) AddMCPToEnvironment(devEnvID, mcpID uint, admin *database.Admin) error {
	// Check if environment exists and admin has access
	isOwner, err := s.devEnvRepo.IsOwner(devEnvID, admin.ID)
	if err != nil {
		return err
	}
	if !isOwner && admin.Role != database.AdminRoleSuperAdmin {
		return appErrors.ErrInsufficientPermissions
	}

	// Check if MCP exists and admin has access
	canAccessMCP, err := s.CanAdminAccessMCP(mcpID, admin)
	if err != nil {
		return err
	}
	if !canAccessMCP {
		return appErrors.ErrInsufficientPermissions
	}

	return s.repo.AddEnvironment(mcpID, devEnvID)
}

func (s *mcpService) RemoveMCPFromEnvironment(devEnvID, mcpID uint, admin *database.Admin) error {
	// Check if environment exists and admin has access
	isOwner, err := s.devEnvRepo.IsOwner(devEnvID, admin.ID)
	if err != nil {
		return err
	}
	if !isOwner && admin.Role != database.AdminRoleSuperAdmin {
		return appErrors.ErrInsufficientPermissions
	}

	return s.repo.RemoveEnvironment(mcpID, devEnvID)
}

func (s *mcpService) GetEnvironmentMCPs(devEnvID uint) ([]database.MCPListItemResponse, error) {
	mcps, err := s.repo.GetEnvironmentMCPs(devEnvID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	var result []database.MCPListItemResponse
	for _, mcp := range mcps {
		item := database.MCPListItemResponse{
			ID:          mcp.ID,
			CreatedAt:   mcp.CreatedAt,
			UpdatedAt:   mcp.UpdatedAt,
			Name:        mcp.Name,
			Description: mcp.Description,
			Config:      mcp.Config,
			Enabled:     mcp.Enabled,
			AdminID:     mcp.AdminID,
			CreatedBy:   mcp.CreatedBy,
		}
		result = append(result, item)
	}

	return result, nil
}

// MCP-specific methods

func (s *mcpService) ValidateMCPConfig(config string) error {
	if config == "" {
		return appErrors.ErrInvalidInput
	}

	// Try to parse as JSON to ensure it's valid JSON
	var configData map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configData); err != nil {
		return fmt.Errorf("invalid JSON config: %v", err)
	}

	// Additional validation can be added here based on MCP requirements
	return nil
}

// ValidateMCPName validates the format of MCP name
func (s *mcpService) ValidateMCPName(name string) error {
	if name == "" {
		return appErrors.ErrInvalidInput
	}

	// Validate name format: only allow English letters, numbers, underscore, and hyphen
	// Pattern: ^[a-zA-Z0-9_-]+$
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' ||
			char == '-') {
			return appErrors.ErrMCPNameInvalidFormat
		}
	}

	return nil
}

func (s *mcpService) GetMCPProjects(mcpID uint, admin *database.Admin) ([]database.Project, error) {
	// Check if admin can access the MCP
	canAccess, err := s.CanAdminAccessMCP(mcpID, admin)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.ErrInsufficientPermissions
	}

	return s.repo.GetProjects(mcpID)
}

func (s *mcpService) GetMCPEnvironments(mcpID uint, admin *database.Admin) ([]database.DevEnvironment, error) {
	// Check if admin can access the MCP
	canAccess, err := s.CanAdminAccessMCP(mcpID, admin)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.ErrInsufficientPermissions
	}

	return s.repo.GetEnvironments(mcpID)
}

// Permission helpers

func (s *mcpService) CanAdminAccessMCP(mcpID uint, admin *database.Admin) (bool, error) {
	if admin.Role == database.AdminRoleSuperAdmin {
		return true, nil
	}

	if admin.Role == database.AdminRoleAdmin {
		return s.repo.IsOwner(mcpID, admin.ID)
	}

	// Developers can only read (if this method is used for read operations)
	return false, nil
}

func (s *mcpService) IsMCPOwner(mcpID uint, admin *database.Admin) (bool, error) {
	if admin.Role == database.AdminRoleSuperAdmin {
		return true, nil
	}

	return s.repo.IsOwner(mcpID, admin.ID)
}

// GetMCPsForTaskConversation gets all enabled MCPs associated with a task conversation
// It combines MCPs from both the project and development environment
func (s *mcpService) GetMCPsForTaskConversation(conversationID uint, taskConvRepo repository.TaskConversationRepository) ([]database.MCP, error) {
	// Get conversation with preloaded task, project, and dev environment
	conv, err := taskConvRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %v", err)
	}

	if conv.Task == nil {
		return nil, fmt.Errorf("conversation task is nil")
	}

	// Use a map to store unique MCPs (by ID)
	mcpMap := make(map[uint]database.MCP)

	// Get MCPs from project
	if conv.Task.ProjectID > 0 {
		projectMCPs, err := s.repo.GetEnabledProjectMCPs(conv.Task.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get project MCPs: %v", err)
		}
		for _, mcp := range projectMCPs {
			mcpMap[mcp.ID] = mcp
		}
	}

	// Get MCPs from dev environment
	if conv.Task.DevEnvironmentID != nil && *conv.Task.DevEnvironmentID > 0 {
		envMCPs, err := s.repo.GetEnabledEnvironmentMCPs(*conv.Task.DevEnvironmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get environment MCPs: %v", err)
		}
		for _, mcp := range envMCPs {
			mcpMap[mcp.ID] = mcp
		}
	}

	// Convert map to slice
	result := make([]database.MCP, 0, len(mcpMap))
	for _, mcp := range mcpMap {
		result = append(result, mcp)
	}

	return result, nil
}
