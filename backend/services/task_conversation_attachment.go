package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
)

type taskConversationAttachmentService struct {
	repo   repository.TaskConversationAttachmentRepository
	config *config.Config
}

func NewTaskConversationAttachmentService(repo repository.TaskConversationAttachmentRepository, cfg *config.Config) TaskConversationAttachmentService {
	return &taskConversationAttachmentService{
		repo:   repo,
		config: cfg,
	}
}

func (s *taskConversationAttachmentService) UploadAttachment(fileName, originalName, contentType string, fileSize int64, filePath string, attachmentType database.AttachmentType, createdBy string) (*database.TaskConversationAttachment, error) {
	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, appErrors.NewI18nError("attachment.file_not_found", "File does not exist")
	}

	attachment := &database.TaskConversationAttachment{
		ConversationID: nil, // Will be set later when associating with conversation
		FileName:       fileName,
		OriginalName:   originalName,
		FilePath:       filePath,
		FileSize:       fileSize,
		ContentType:    contentType,
		Type:           attachmentType,
		SortOrder:      0, // Will be set when associating with conversation
		CreatedBy:      createdBy,
	}

	if err := s.repo.Create(attachment); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (s *taskConversationAttachmentService) AssociateWithConversation(attachmentID, conversationID uint) error {
	attachment, err := s.repo.GetByID(attachmentID)
	if err != nil {
		return err
	}

	// Get sort order (next sequence number for this conversation)
	existingAttachments, err := s.repo.GetByConversationID(conversationID)
	if err != nil {
		return err
	}
	sortOrder := len(existingAttachments) + 1

	attachment.ConversationID = &conversationID
	attachment.SortOrder = sortOrder

	return s.repo.Update(attachment)
}

func (s *taskConversationAttachmentService) GetAttachment(id uint) (*database.TaskConversationAttachment, error) {
	return s.repo.GetByID(id)
}

func (s *taskConversationAttachmentService) GetAttachmentsByConversation(conversationID uint) ([]database.TaskConversationAttachment, error) {
	return s.repo.GetByConversationID(conversationID)
}

func (s *taskConversationAttachmentService) UpdateAttachment(id uint, attachment *database.TaskConversationAttachment) error {
	return s.repo.Update(attachment)
}

func (s *taskConversationAttachmentService) DeleteAttachment(id uint) error {
	attachment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete physical file
	if err := os.Remove(attachment.FilePath); err != nil {
		// Log error but don't fail the operation
		// File might already be deleted or moved
		fmt.Printf("Warning: Failed to delete physical file %s: %v\n", attachment.FilePath, err)
	}

	return s.repo.Delete(id)
}

func (s *taskConversationAttachmentService) DeleteAttachmentsByConversation(conversationID uint) error {
	attachments, err := s.repo.GetByConversationID(conversationID)
	if err != nil {
		return err
	}

	// Delete physical files
	for _, attachment := range attachments {
		if err := os.Remove(attachment.FilePath); err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Failed to delete physical file %s: %v\n", attachment.FilePath, err)
		}
	}

	return s.repo.DeleteByConversationID(conversationID)
}

func (s *taskConversationAttachmentService) ProcessContentWithAttachments(content string, attachments []database.TaskConversationAttachment) string {
	if len(attachments) == 0 {
		return content
	}

	// Create maps for quick lookup
	imageCount := 1
	pdfCount := 1

	for _, attachment := range attachments {
		var tag string
		switch attachment.Type {
		case database.AttachmentTypeImage:
			tag = fmt.Sprintf("[image%d]", imageCount)
			imageCount++
		case database.AttachmentTypePDF:
			tag = fmt.Sprintf("[pdf%d]", pdfCount)
			pdfCount++
		}

		// Add tag to content if not already present
		if tag != "" && !strings.Contains(content, tag) {
			if content != "" {
				content += " "
			}
			content += tag
		}
	}

	return content
}

func (s *taskConversationAttachmentService) ParseAttachmentTags(content string) []string {
	// Parse [image1], [pdf1], etc. tags from content
	re := regexp.MustCompile(`\[(image|pdf)(\d+)\]`)
	matches := re.FindAllString(content, -1)

	tags := append([]string(nil), matches...)

	return tags
}

// GetAttachmentStorageDir returns the storage directory for attachments
func (s *taskConversationAttachmentService) GetAttachmentStorageDir() string {
	return s.config.AttachmentsDir
}

// GenerateAttachmentFileName generates a unique filename for an attachment
func GenerateAttachmentFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	// Generate unique timestamp-based filename without original name
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("attach_%d%s", timestamp, ext)
}
