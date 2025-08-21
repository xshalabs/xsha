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

func (s *taskConversationAttachmentService) ProcessContentWithAttachments(content string, attachments []database.TaskConversationAttachment, conversationID uint) string {
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
			tag = fmt.Sprintf("__xsha_workspace/%d_image%d%s", conversationID, imageCount, filepath.Ext(attachment.FileName))
			imageCount++
		case database.AttachmentTypePDF:
			tag = fmt.Sprintf("__xsha_workspace/%d_pdf%d.pdf", conversationID, pdfCount)
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
	// Parse __xsha_workspace/{conversation_id}_image{n}.ext or __xsha_workspace/{conversation_id}_pdf{n}.pdf paths from content
	re := regexp.MustCompile(`__xsha_workspace/\d+_(image|pdf)\d+\.\w+`)
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

// CopyAttachmentsToWorkspace copies conversation attachments to workspace xsha directory
func (s *taskConversationAttachmentService) CopyAttachmentsToWorkspace(conversationID uint, workspacePath string) ([]database.TaskConversationAttachment, error) {
	// Get attachments for the conversation
	attachments, err := s.repo.GetByConversationID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments: %v", err)
	}

	if len(attachments) == 0 {
		return []database.TaskConversationAttachment{}, nil
	}

	// Create __xsha_workspace directory in workspace
	xshaDir := filepath.Join(workspacePath, "__xsha_workspace")
	if err := os.MkdirAll(xshaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create __xsha_workspace directory: %v", err)
	}

	// Track counters for each attachment type
	imageCount := 1
	pdfCount := 1
	
	var copiedAttachments []database.TaskConversationAttachment

	// Sort attachments by SortOrder to maintain consistent numbering
	for i := 0; i < len(attachments)-1; i++ {
		for j := i + 1; j < len(attachments); j++ {
			if attachments[i].SortOrder > attachments[j].SortOrder {
				attachments[i], attachments[j] = attachments[j], attachments[i]
			}
		}
	}

	for _, attachment := range attachments {
		var workspaceFileName string
		var ext string

		switch attachment.Type {
		case database.AttachmentTypeImage:
			ext = filepath.Ext(attachment.FileName)
			workspaceFileName = fmt.Sprintf("%d_image%d%s", conversationID, imageCount, ext)
			imageCount++
		case database.AttachmentTypePDF:
			workspaceFileName = fmt.Sprintf("%d_pdf%d.pdf", conversationID, pdfCount)
			pdfCount++
		default:
			continue // Skip unknown types
		}

		// Source and destination paths
		srcPath := attachment.FilePath
		destPath := filepath.Join(xshaDir, workspaceFileName)

		// Copy file
		if err := s.copyFile(srcPath, destPath); err != nil {
			return nil, fmt.Errorf("failed to copy attachment %s to workspace: %v", attachment.FileName, err)
		}

		// Create a copy of the attachment with workspace path
		workspaceAttachment := attachment
		workspaceAttachment.FilePath = destPath
		workspaceAttachment.FileName = workspaceFileName
		
		copiedAttachments = append(copiedAttachments, workspaceAttachment)
	}

	return copiedAttachments, nil
}

// copyFile copies a file from src to dst
func (s *taskConversationAttachmentService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// ReplaceAttachmentTagsWithPaths handles backward compatibility for old format tags and ensures new format paths are correctly set
func (s *taskConversationAttachmentService) ReplaceAttachmentTagsWithPaths(content string, attachments []database.TaskConversationAttachment, workspacePath string) string {
	if len(attachments) == 0 {
		return content
	}

	processedContent := content

	// Track counters for each attachment type
	imageCount := 1
	pdfCount := 1

	// Sort attachments by SortOrder to maintain consistent numbering
	sortedAttachments := make([]database.TaskConversationAttachment, len(attachments))
	copy(sortedAttachments, attachments)
	
	for i := 0; i < len(sortedAttachments)-1; i++ {
		for j := i + 1; j < len(sortedAttachments); j++ {
			if sortedAttachments[i].SortOrder > sortedAttachments[j].SortOrder {
				sortedAttachments[i], sortedAttachments[j] = sortedAttachments[j], sortedAttachments[i]
			}
		}
	}

	for _, attachment := range sortedAttachments {
		var oldTag, newPath string

		switch attachment.Type {
		case database.AttachmentTypeImage:
			// Handle old format tags for backward compatibility
			oldTag = fmt.Sprintf("[image%d]", imageCount)
			newPath = fmt.Sprintf("__xsha_workspace/%s", attachment.FileName)
			imageCount++
		case database.AttachmentTypePDF:
			// Handle old format tags for backward compatibility
			oldTag = fmt.Sprintf("[pdf%d]", pdfCount)
			newPath = fmt.Sprintf("__xsha_workspace/%s", attachment.FileName)
			pdfCount++
		default:
			continue // Skip unknown types
		}

		// Replace the old tag format with the new path format for backward compatibility
		if strings.Contains(processedContent, oldTag) {
			processedContent = strings.Replace(processedContent, oldTag, newPath, -1)
		}
	}

	return processedContent
}

// CleanupWorkspaceAttachments removes attachment files from workspace
func (s *taskConversationAttachmentService) CleanupWorkspaceAttachments(workspacePath string) error {
	xshaDir := filepath.Join(workspacePath, "__xsha_workspace")
	
	// Check if __xsha_workspace directory exists
	if _, err := os.Stat(xshaDir); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to clean
		return nil
	}

	// Remove the entire __xsha_workspace directory and its contents
	if err := os.RemoveAll(xshaDir); err != nil {
		return fmt.Errorf("failed to remove __xsha_workspace directory: %v", err)
	}

	return nil
}
