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
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{
		adminRepo: adminRepo,
	}
}

func (s *adminService) CreateAdmin(username, password, email, createdBy string) (*database.Admin, error) {
	// Validate input
	if err := s.validateAdminData(username, password); err != nil {
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
		Email:        email,
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

func (s *adminService) ListAdmins(username *string, isActive *bool, page, pageSize int) ([]database.Admin, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return s.adminRepo.List(username, isActive, page, pageSize)
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

	return s.adminRepo.Update(admin)
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

	admin.PasswordHash = string(passwordHash)
	return s.adminRepo.Update(admin)
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
func (s *adminService) validateAdminData(username, password string) error {
	if err := s.validateUsername(username); err != nil {
		return err
	}
	return s.validatePassword(password)
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
