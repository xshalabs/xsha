package services

import (
	"errors"
	"fmt"
	"strings"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type adminService struct {
	adminRepo       repository.AdminRepository
	authService     AuthService
	devEnvService   DevEnvironmentService
	gitCredService  GitCredentialService
	projectService  ProjectService
	taskService     TaskService
	taskConvService TaskConversationService
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{
		adminRepo: adminRepo,
	}
}

func (s *adminService) SetAuthService(authService AuthService) {
	s.authService = authService
}

func (s *adminService) SetDevEnvironmentService(devEnvService DevEnvironmentService) {
	s.devEnvService = devEnvService
}

func (s *adminService) SetGitCredentialService(gitCredService GitCredentialService) {
	s.gitCredService = gitCredService
}

func (s *adminService) SetProjectService(projectService ProjectService) {
	s.projectService = projectService
}

func (s *adminService) SetTaskService(taskService TaskService) {
	s.taskService = taskService
}

func (s *adminService) SetTaskConversationService(taskConvService TaskConversationService) {
	s.taskConvService = taskConvService
}

func (s *adminService) CreateAdmin(username, password, name, email, createdBy string) (*database.Admin, error) {
	// Default to admin role for backward compatibility
	return s.CreateAdminWithRole(username, password, name, email, database.AdminRoleAdmin, createdBy)
}

func (s *adminService) CreateAdminWithRole(username, password, name, email string, role database.AdminRole, createdBy string) (*database.Admin, error) {
	// Validate input
	if err := s.validateAdminData(username, password, name); err != nil {
		return nil, err
	}

	// Check if username already exists
	existingAdmin, err := s.adminRepo.GetByUsername(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing admin: %v", err)
	}
	if existingAdmin != nil {
		return nil, appErrors.ErrAdminUsernameExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	admin := &database.Admin{
		Username:     username,
		PasswordHash: string(passwordHash),
		Name:         name,
		Email:        email,
		Role:         role,
		IsActive:     true,
		CreatedBy:    createdBy,
	}

	if err := s.adminRepo.Create(admin); err != nil {
		// Check if it's already an internationalized error
		if _, ok := err.(*appErrors.I18nError); ok {
			return nil, err
		}
		// For other errors, wrap with generic message
		return nil, fmt.Errorf("failed to create admin: %v", err)
	}

	return admin, nil
}

func (s *adminService) GetAdmin(id uint) (*database.Admin, error) {
	return s.adminRepo.GetByID(id)
}

func (s *adminService) GetAdminByUsername(username string) (*database.Admin, error) {
	return s.adminRepo.GetByUsername(username)
}

func (s *adminService) ListAdmins(search *string, isActive *bool, roles []string, page, pageSize int) ([]database.Admin, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return s.adminRepo.List(search, isActive, roles, page, pageSize)
}

func (s *adminService) UpdateAdmin(id uint, updates map[string]interface{}) error {
	admin, err := s.adminRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrAdminNotFound
		}
		return fmt.Errorf("failed to get admin: %v", err)
	}

	// Apply updates
	if username, ok := updates["username"]; ok {
		if usernameStr, ok := username.(string); ok {
			if err := s.validateUsername(usernameStr); err != nil {
				return err
			}
			existingAdmin, err := s.adminRepo.GetByUsername(usernameStr)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to check existing admin: %v", err)
			}
			if existingAdmin != nil && existingAdmin.ID != id {
				return appErrors.ErrAdminUsernameExists
			}
			admin.Username = usernameStr
		}
	}

	if name, ok := updates["name"]; ok {
		if nameStr, ok := name.(string); ok {
			if err := s.validateName(nameStr); err != nil {
				return err
			}
			admin.Name = nameStr
		}
	}

	if email, ok := updates["email"]; ok {
		if emailStr, ok := email.(string); ok {
			admin.Email = emailStr
		}
	}

	if isActive, ok := updates["is_active"]; ok {
		if isActiveBool, ok := isActive.(bool); ok {
			admin.IsActive = isActiveBool
		}
	}

	if err := s.adminRepo.Update(admin.ID, updates); err != nil {
		return err
	}

	return nil
}

func (s *adminService) DeleteAdmin(id uint) error {
	admin, err := s.adminRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrAdminNotFound
		}
		return fmt.Errorf("failed to get admin: %v", err)
	}

	// Check if this is a system-created admin
	if admin.CreatedBy == "system" {
		return appErrors.ErrCannotDeleteSystemAdmin
	}

	// Check if this is the last active admin
	activeCount, err := s.adminRepo.CountAdmins()
	if err != nil {
		return fmt.Errorf("failed to count active admins: %v", err)
	}

	if activeCount <= 1 && admin.IsActive {
		return appErrors.ErrCannotDeleteLastAdmin
	}

	// Check if admin has created any dev environments
	if s.devEnvService != nil {
		envCount, err := s.devEnvService.CountByAdminID(id)
		if err != nil {
			return fmt.Errorf("failed to count admin environments: %v", err)
		}
		if envCount > 0 {
			return appErrors.ErrAdminHasEnvironments
		}
	}

	// Check if admin has created any git credentials
	if s.gitCredService != nil {
		credCount, err := s.gitCredService.CountByAdminID(id)
		if err != nil {
			return fmt.Errorf("failed to count admin credentials: %v", err)
		}
		if credCount > 0 {
			return appErrors.ErrAdminHasCredentials
		}
	}

	// Check if admin has created any tasks
	if s.taskService != nil {
		taskCount, err := s.taskService.CountByAdminID(id)
		if err != nil {
			return fmt.Errorf("failed to count admin tasks: %v", err)
		}
		if taskCount > 0 {
			return appErrors.ErrAdminHasTasks
		}
	}

	// Delete admin associations first, then delete the admin
	if err := s.adminRepo.DeleteAdminAssociations(id); err != nil {
		return fmt.Errorf("failed to delete admin associations: %v", err)
	}

	return s.adminRepo.Delete(id)
}

func (s *adminService) ChangePassword(id uint, newPassword string) error {
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	admin, err := s.adminRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrAdminNotFound
		}
		return fmt.Errorf("failed to get admin: %v", err)
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	updates := map[string]interface{}{
		"password_hash": string(passwordHash),
	}
	return s.adminRepo.Update(admin.ID, updates)
}

func (s *adminService) ValidateCredentials(username, password string) (*database.Admin, error) {
	admin, err := s.adminRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get admin: %v", err)
	}

	if !admin.IsActive {
		return nil, appErrors.ErrAdminInactive
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
	if err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	return admin, nil
}

func (s *adminService) InitializeDefaultAdmin() error {
	return s.adminRepo.InitializeDefaultAdmin()
}

// validateAdminData validates admin creation data
func (s *adminService) validateAdminData(username, password, name string) error {
	if err := s.validateUsername(username); err != nil {
		return err
	}
	if err := s.validatePassword(password); err != nil {
		return err
	}
	return s.validateName(name)
}

// validateUsername validates admin username
func (s *adminService) validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return appErrors.ErrAdminUsernameRequired
	}
	if len(username) < 3 {
		return appErrors.ErrAdminUsernameInvalid
	}
	if len(username) > 50 {
		return appErrors.ErrAdminUsernameInvalid
	}
	// Check for valid characters (alphanumeric and underscore)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return appErrors.ErrAdminUsernameInvalid
		}
	}
	return nil
}

// validatePassword validates admin password
func (s *adminService) validatePassword(password string) error {
	if password == "" {
		return appErrors.ErrAdminPasswordRequired
	}
	if len(password) < 6 {
		return appErrors.ErrAdminPasswordInvalid
	}
	if len(password) > 128 {
		return appErrors.ErrAdminPasswordInvalid
	}
	return nil
}

// validateName validates admin name
func (s *adminService) validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return appErrors.ErrAdminNameRequired
	}
	if len(name) < 2 {
		return appErrors.ErrAdminNameInvalid
	}
	if len(name) > 100 {
		return appErrors.ErrAdminNameInvalid
	}
	return nil
}

// UpdateAdminRole updates admin role
func (s *adminService) UpdateAdminRole(id uint, role database.AdminRole) error {
	admin, err := s.adminRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrAdminNotFound
		}
		return fmt.Errorf("failed to get admin: %v", err)
	}

	// Check if this is a system-created admin
	if admin.CreatedBy == "system" {
		return appErrors.ErrCannotModifySystemAdminRole
	}

	// Check if there would be no super_admin left
	if admin.Role == database.AdminRoleSuperAdmin && role != database.AdminRoleSuperAdmin {
		count, err := s.adminRepo.CountActiveAdminsByRole(database.AdminRoleSuperAdmin)
		if err != nil {
			return fmt.Errorf("failed to count super admins: %v", err)
		}
		if count <= 1 {
			return appErrors.ErrCannotRemoveLastSuperAdmin
		}
	}

	updates := map[string]interface{}{
		"role": role,
	}

	return s.adminRepo.Update(id, updates)
}

// HasPermission checks if admin has permission for specific resource and action
func (s *adminService) HasPermission(admin *database.Admin, resource, action string, resourceId uint) bool {
	if admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	switch resource {
	case "project":
		return s.checkProjectPermission(admin, action, resourceId)
	case "task":
		return s.checkTaskPermission(admin, action, resourceId)
	case "conversation":
		return s.checkConversationPermission(admin, action, resourceId)
	case "credential":
		return s.checkCredentialPermission(admin, action, resourceId)
	case "environment":
		return s.checkEnvironmentPermission(admin, action, resourceId)
	default:
		return false
	}
}

// Helper methods for specific resource permissions
func (s *adminService) checkProjectPermission(admin *database.Admin, action string, projectID uint) bool {
	if admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	switch action {
	case "create":
		return admin.Role == database.AdminRoleAdmin
	case "read":
		if s.projectService != nil {
			canAccess, err := s.projectService.CanAdminAccessProject(projectID, admin.ID)
			if err != nil {
				utils.Error("Failed to check project access", "projectID", projectID, "adminID", admin.ID, "error", err)
				return false
			}
			return canAccess
		}
		return false
	case "update", "delete":
		if admin.Role == database.AdminRoleAdmin && s.projectService != nil {
			canAccess, err := s.projectService.IsOwner(projectID, admin.ID)
			if err != nil {
				utils.Error("Failed to check project ownership", "projectID", projectID, "adminID", admin.ID, "error", err)
				return false
			}
			return canAccess
		}
		return false
	case "tasks":
		if s.projectService != nil {
			if admin.Role == database.AdminRoleAdmin {
				canAccess, err := s.projectService.IsOwner(projectID, admin.ID)
				if err != nil {
					utils.Error("Failed to check project ownership for tasks", "projectID", projectID, "adminID", admin.ID, "error", err)
					return false
				}
				return canAccess
			}
			if admin.Role == database.AdminRoleDeveloper {
				canAccess, err := s.projectService.CanAdminAccessProject(projectID, admin.ID)
				if err != nil {
					utils.Error("Failed to check project access for tasks", "projectID", projectID, "adminID", admin.ID, "error", err)
					return false
				}
				return canAccess
			}
		}
		return false
	default:
		return false
	}
}

func (s *adminService) checkTaskPermission(admin *database.Admin, action string, taskId uint) bool {
	if admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	if admin.Role == database.AdminRoleDeveloper && s.taskService != nil {
		switch action {
		case "delete":
			task, err := s.taskService.GetTask(taskId)
			if err != nil {
				utils.Error("Failed to get task for permission check", "taskID", taskId, "adminID", admin.ID, "error", err)
				return false
			}
			return task.AdminID != nil && *task.AdminID == admin.ID
		default:
			return false
		}
	}

	return false
}

func (s *adminService) checkConversationPermission(admin *database.Admin, action string, convId uint) bool {
	if admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	if admin.Role == database.AdminRoleDeveloper && s.taskConvService != nil {
		switch action {
		case "delete":
			conversation, err := s.taskConvService.GetConversation(convId)
			if err != nil {
				utils.Error("Failed to get conversation for permission check", "conversationID", convId, "adminID", admin.ID, "error", err)
				return false
			}
			return conversation.AdminID != nil && *conversation.AdminID == admin.ID
		default:
			return false
		}
	}

	return false
}

func (s *adminService) checkCredentialPermission(admin *database.Admin, action string, resourceId uint) bool {
	if admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	if s.gitCredService == nil {
		utils.Error("GitCredentialService not initialized for permission check", "adminID", admin.ID)
		return false
	}

	switch action {
	case "read", "create":
		return true
	case "update", "delete":
		if admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleDeveloper {
			canAccess, err := s.gitCredService.IsOwner(resourceId, admin.ID)
			if err != nil {
				utils.Error("Failed to check credential access permission", "credentialID", resourceId, "adminID", admin.ID, "error", err)
				return false
			}
			return canAccess
		}
		return false
	default:
		return false
	}
}

func (s *adminService) checkEnvironmentPermission(admin *database.Admin, action string, resourceId uint) bool {
	if admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	if s.devEnvService == nil {
		utils.Error("DevEnvironmentService not initialized for permission check", "adminID", admin.ID)
		return false
	}

	switch action {
	case "read", "create":
		return true
	case "update", "delete":
		if admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleDeveloper {
			canAccess, err := s.devEnvService.IsOwner(resourceId, admin.ID)
			if err != nil {
				utils.Error("Failed to check environment access permission", "environmentID", resourceId, "adminID", admin.ID, "error", err)
				return false
			}
			return canAccess
		}
		return false
	default:
		return false
	}
}
