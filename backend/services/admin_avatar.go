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

	"github.com/google/uuid"
)

type adminAvatarService struct {
	repo      repository.AdminAvatarRepository
	adminRepo repository.AdminRepository
	config    *config.Config
}

func NewAdminAvatarService(repo repository.AdminAvatarRepository, adminRepo repository.AdminRepository, cfg *config.Config) AdminAvatarService {
	return &adminAvatarService{
		repo:      repo,
		adminRepo: adminRepo,
		config:    cfg,
	}
}

func (s *adminAvatarService) UploadAvatar(fileName, originalName, contentType string, fileSize int64, filePath string, adminID uint, createdBy string) (*database.AdminAvatar, error) {
	// Validate file exists - construct full path from relative path
	fullFilePath := s.GetFullAvatarPath(filePath)
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		return nil, appErrors.NewI18nError("avatar.file_not_found", "File does not exist")
	}

	// Generate UUID for secure access
	avatarUUID := uuid.New().String()

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

func (s *adminAvatarService) GetAvatarByUUID(uuid string) (*database.AdminAvatar, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *adminAvatarService) UpdateAdminAvatarByUUID(avatarUUID string, adminID uint) error {
	// Get avatar by UUID
	avatar, err := s.repo.GetByUUID(avatarUUID)
	if err != nil {
		return appErrors.NewI18nError("avatar.not_found", "Avatar not found")
	}

	// Get admin
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	// Update admin's avatar_id
	admin.AvatarID = &avatar.ID
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

func (s *adminAvatarService) GetFullAvatarPath(relativePath string) string {
	return filepath.Join(s.config.AvatarsDir, relativePath)
}
