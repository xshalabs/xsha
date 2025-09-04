package services

import (
	"errors"
	"fmt"
	"strings"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type adminService struct {
	adminRepo      repository.AdminRepository
	authService    AuthService
	devEnvService  DevEnvironmentService
	gitCredService GitCredentialService
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

func (s *adminService) ListAdmins(search *string, isActive *bool, page, pageSize int) ([]database.Admin, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return s.adminRepo.List(search, isActive, page, pageSize)
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
			// Check if new username already exists (and it's not the same admin)
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

	// Update the admin first
	if err := s.adminRepo.Update(admin.ID, updates); err != nil {
		return err
	}

	// Note: When an admin is deactivated (is_active = false), their existing tokens
	// remain valid until expiration, but they cannot login to get new tokens.
	// This provides a grace period and avoids the complexity of user-based blacklisting.

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
func (s *adminService) HasPermission(admin *database.Admin, resource, action string, resourceOwnerID uint) bool {
	// Super admin has all permissions
	if admin.Role == database.AdminRoleSuperAdmin {
		return true
	}

	// Check permission based on role and resource
	switch resource {
	case "admin":
		return admin.Role == database.AdminRoleSuperAdmin
	case "system_config":
		return admin.Role == database.AdminRoleSuperAdmin
	case "operation_log":
		return admin.Role == database.AdminRoleSuperAdmin
	case "project":
		return s.checkProjectPermission(admin, action, resourceOwnerID)
	case "task":
		return s.checkTaskPermission(admin, action, resourceOwnerID)
	case "conversation":
		return s.checkConversationPermission(admin, action, resourceOwnerID)
	case "credential":
		return s.checkCredentialPermission(admin, action, resourceOwnerID)
	case "environment":
		return s.checkEnvironmentPermission(admin, action, resourceOwnerID)
	default:
		return false
	}
}

// CanAccessResource is an alias for HasPermission
func (s *adminService) CanAccessResource(admin *database.Admin, resource string, action string, resourceOwnerID uint) bool {
	return s.HasPermission(admin, resource, action, resourceOwnerID)
}

// GetAvailableRoles returns all available admin roles
func (s *adminService) GetAvailableRoles() []database.AdminRole {
	return []database.AdminRole{
		database.AdminRoleSuperAdmin,
		database.AdminRoleAdmin,
		database.AdminRoleDeveloper,
	}
}

// Helper methods for specific resource permissions

func (s *adminService) checkProjectPermission(admin *database.Admin, action string, resourceOwnerID uint) bool {
	switch action {
	case "create":
		return admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleSuperAdmin
	case "read":
		return true // All roles can read projects
	case "update", "delete":
		if admin.Role == database.AdminRoleSuperAdmin {
			return true
		}
		return admin.Role == database.AdminRoleAdmin && admin.ID == resourceOwnerID
	default:
		return false
	}
}

func (s *adminService) checkTaskPermission(admin *database.Admin, action string, resourceOwnerID uint) bool {
	switch action {
	case "create", "execute":
		return true // All roles can create and execute tasks
	case "read":
		return true // All roles can read tasks
	case "update":
		if admin.Role == database.AdminRoleSuperAdmin {
			return true
		}
		return admin.ID == resourceOwnerID // Only owner can update
	case "delete":
		if admin.Role == database.AdminRoleSuperAdmin {
			return true
		}
		return admin.Role == database.AdminRoleAdmin && admin.ID == resourceOwnerID
	default:
		return false
	}
}

func (s *adminService) checkConversationPermission(admin *database.Admin, action string, resourceOwnerID uint) bool {
	switch action {
	case "create", "execute":
		return true // All roles can create and execute conversations
	case "read":
		return true // All roles can read conversations
	case "update", "delete":
		if admin.Role == database.AdminRoleSuperAdmin {
			return true
		}
		return admin.ID == resourceOwnerID // Only owner can update/delete
	default:
		return false
	}
}

func (s *adminService) checkCredentialPermission(admin *database.Admin, action string, resourceOwnerID uint) bool {
	switch action {
	case "create":
		// Developer, Admin and SuperAdmin can all create credentials
		return admin.Role == database.AdminRoleDeveloper || admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleSuperAdmin
	case "read":
		return true // All roles can read credentials (masked)
	case "update", "delete":
		if admin.Role == database.AdminRoleSuperAdmin {
			return true
		}
		// Admin and Developer can update/delete credentials they have admin access to
		if admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleDeveloper {
			// resourceOwnerID is the credential ID, check if admin has access to this credential
			if s.gitCredService != nil {
				canAccess, err := s.gitCredService.CanAdminAccessCredential(resourceOwnerID, admin.ID)
				if err != nil {
					return false
				}
				return canAccess
			}
			// Fallback: if gitCredService is not set, deny access to prevent security issues
			return false
		}
		return false
	default:
		return false
	}
}

func (s *adminService) checkEnvironmentPermission(admin *database.Admin, action string, resourceOwnerID uint) bool {
	switch action {
	case "create":
		// Developer, Admin and SuperAdmin can all create environments
		return admin.Role == database.AdminRoleDeveloper || admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleSuperAdmin
	case "read":
		return true // All roles can read environments
	case "update", "delete":
		if admin.Role == database.AdminRoleSuperAdmin {
			return true
		}
		// Admin and Developer can update/delete environments they have admin access to
		if admin.Role == database.AdminRoleAdmin || admin.Role == database.AdminRoleDeveloper {
			// resourceOwnerID is the environment ID, check if admin has access to this environment
			if s.devEnvService != nil {
				canAccess, err := s.devEnvService.CanAdminAccessEnvironment(resourceOwnerID, admin.ID)
				if err != nil {
					return false
				}
				return canAccess
			}
			// Fallback: if devEnvService is not set, deny access to prevent security issues
			return false
		}
		return false
	default:
		return false
	}
}
