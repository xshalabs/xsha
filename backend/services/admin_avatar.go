package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"

	"github.com/google/uuid"
)

type adminAvatarService struct {
	repo       repository.AdminAvatarRepository
	adminRepo  repository.AdminRepository
	config     *config.Config
}

func NewAdminAvatarService(repo repository.AdminAvatarRepository, adminRepo repository.AdminRepository, cfg *config.Config) AdminAvatarService {
	return &adminAvatarService{
		repo:       repo,
		adminRepo:  adminRepo,
		config:     cfg,
	}
}

func (s *adminAvatarService) UploadAvatar(fileName, originalName, contentType string, fileSize int64, filePath string, adminID uint, createdBy string) (*database.AdminAvatar, error) {
	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, appErrors.NewI18nError("avatar.file_not_found", "File does not exist")
	}

	// Generate UUID for secure access
	avatarUUID := uuid.New().String()

	// Check if admin already has an avatar
	existingAvatar, err := s.repo.GetByAdminID(adminID)
	if err == nil && existingAvatar != nil {
		// Delete old avatar file
		if err := os.Remove(existingAvatar.FilePath); err != nil {
			utils.Warn("Failed to delete old avatar file", "filePath", existingAvatar.FilePath, "error", err)
		}
		// Delete old avatar record
		if err := s.repo.Delete(existingAvatar.ID); err != nil {
			utils.Warn("Failed to delete old avatar record", "avatarID", existingAvatar.ID, "error", err)
		}
	}

	avatar := &database.AdminAvatar{
		UUID:         avatarUUID,
		FileName:     fileName,
		OriginalName: originalName,
		FilePath:     filePath,
		FileSize:     fileSize,
		ContentType:  contentType,
		AdminID:      &adminID,
		CreatedBy:    createdBy,
	}

	if err := s.repo.Create(avatar); err != nil {
		return nil, err
	}

	return avatar, nil
}

func (s *adminAvatarService) GetAvatar(id uint) (*database.AdminAvatar, error) {
	return s.repo.GetByID(id)
}

func (s *adminAvatarService) GetAvatarByUUID(uuid string) (*database.AdminAvatar, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *adminAvatarService) GetAvatarByAdminID(adminID uint) (*database.AdminAvatar, error) {
	return s.repo.GetByAdminID(adminID)
}

func (s *adminAvatarService) UpdateAdminAvatar(adminID uint, avatarID uint) error {
	// Get admin
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	// Update admin's avatar_id
	admin.AvatarID = &avatarID
	return s.adminRepo.Update(admin)
}

func (s *adminAvatarService) GetAvatarStorageDir() string {
	return s.config.AvatarsDir
}

func (s *adminAvatarService) GenerateAvatarFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	// Generate unique timestamp-based filename
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("avatar_%d%s", timestamp, ext)
}