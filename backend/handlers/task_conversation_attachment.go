package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

type TaskConversationAttachmentHandlers struct {
	attachmentService services.TaskConversationAttachmentService
}

func NewTaskConversationAttachmentHandlers(attachmentService services.TaskConversationAttachmentService) *TaskConversationAttachmentHandlers {
	return &TaskConversationAttachmentHandlers{
		attachmentService: attachmentService,
	}
}

// UploadAttachment uploads a file attachment (not yet associated with any conversation)
// @Summary Upload attachment
// @Description Upload an image or PDF file that will be associated with a conversation later
// @Tags Task Conversation Attachments
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Success 200 {object} object{message=string,data=object} "Attachment uploaded successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 413 {object} object{error=string} "File too large"
// @Router /attachments/upload [post]
func (h *TaskConversationAttachmentHandlers) UploadAttachment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "attachment.file_required")})
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": i18n.T(lang, "attachment.file_too_large"),
		})
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	attachmentType, err := h.getAttachmentType(contentType, header.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "attachment.unsupported_file_type")})
		return
	}

	// Create storage directory
	storageDir := h.attachmentService.GetAttachmentStorageDir()
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		utils.Error("Failed to create attachment storage directory", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	// Generate unique filename
	fileName := services.GenerateAttachmentFileName(header.Filename)
	filePath := filepath.Join(storageDir, fileName)

	// Save file
	if err := h.saveUploadedFile(file, filePath); err != nil {
		utils.Error("Failed to save uploaded file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "attachment.save_failed")})
		return
	}

	// Create attachment record
	attachment, err := h.attachmentService.UploadAttachment(
		fileName,
		header.Filename,
		contentType,
		header.Size,
		filePath,
		attachmentType,
		username.(string),
	)
	if err != nil {
		// Clean up file if database operation failed
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "attachment.upload_success"),
		"data":    attachment,
	})
}

// GetAttachment retrieves attachment information
// @Summary Get attachment info
// @Description Get attachment information by ID
// @Tags Task Conversation Attachments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Attachment ID"
// @Success 200 {object} object{message=string,data=object} "Attachment retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid attachment ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Attachment not found"
// @Router /attachments/{id} [get]
func (h *TaskConversationAttachmentHandlers) GetAttachment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	attachment, err := h.attachmentService.GetAttachment(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "attachment.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "attachment.get_success"),
		"data":    attachment,
	})
}

// DownloadAttachment downloads an attachment file
// @Summary Download attachment
// @Description Download attachment file by ID
// @Tags Task Conversation Attachments
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path int true "Attachment ID"
// @Success 200 {file} binary "Attachment file"
// @Failure 400 {object} object{error=string} "Invalid attachment ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Attachment not found"
// @Router /attachments/{id}/download [get]
func (h *TaskConversationAttachmentHandlers) DownloadAttachment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	attachment, err := h.attachmentService.GetAttachment(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "attachment.not_found")})
		return
	}

	// Check if file exists
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "attachment.file_not_found")})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.OriginalName))
	c.Header("Content-Type", attachment.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", attachment.FileSize))

	// Serve file
	c.File(attachment.FilePath)
}

// PreviewAttachment previews an attachment (for images)
// @Summary Preview attachment
// @Description Preview attachment file (mainly for images)
// @Tags Task Conversation Attachments
// @Accept json
// @Produce image/*
// @Security BearerAuth
// @Param id path int true "Attachment ID"
// @Success 200 {file} binary "Attachment file for preview"
// @Failure 400 {object} object{error=string} "Invalid attachment ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Attachment not found"
// @Router /attachments/{id}/preview [get]
func (h *TaskConversationAttachmentHandlers) PreviewAttachment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	attachment, err := h.attachmentService.GetAttachment(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "attachment.not_found")})
		return
	}

	// Check if file exists
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "attachment.file_not_found")})
		return
	}

	// Set headers for inline display
	c.Header("Content-Type", attachment.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", attachment.FileSize))

	// Serve file inline
	c.File(attachment.FilePath)
}

// DeleteAttachment deletes an attachment
// @Summary Delete attachment
// @Description Delete an attachment by ID
// @Tags Task Conversation Attachments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Attachment ID"
// @Success 200 {object} object{message=string} "Attachment deleted successfully"
// @Failure 400 {object} object{error=string} "Invalid attachment ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 404 {object} object{error=string} "Attachment not found"
// @Router /attachments/{id} [delete]
func (h *TaskConversationAttachmentHandlers) DeleteAttachment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	if err := h.attachmentService.DeleteAttachment(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "attachment.delete_success")})
}

// GetConversationAttachments gets all attachments for a conversation
// @Summary Get conversation attachments
// @Description Get all attachments for a specific conversation
// @Tags Task Conversation Attachments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param conversation_id query int true "Conversation ID"
// @Success 200 {object} object{message=string,data=[]object} "Attachments retrieved successfully"
// @Failure 400 {object} object{error=string} "Invalid conversation ID"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Router /attachments [get]
func (h *TaskConversationAttachmentHandlers) GetConversationAttachments(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Query("conversation_id")
	if conversationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.conversation_id_required")})
		return
	}

	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	attachments, err := h.attachmentService.GetAttachmentsByConversation(uint(conversationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "attachment.get_success"),
		"data":    attachments,
	})
}

// Helper functions

func (h *TaskConversationAttachmentHandlers) getAttachmentType(contentType, filename string) (database.AttachmentType, error) {
	// Check by content type first
	if strings.HasPrefix(contentType, "image/") {
		return database.AttachmentTypeImage, nil
	}
	if contentType == "application/pdf" {
		return database.AttachmentTypePDF, nil
	}

	// Check by file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return database.AttachmentTypeImage, nil
	case ".pdf":
		return database.AttachmentTypePDF, nil
	}

	return "", fmt.Errorf("unsupported file type")
}

func (h *TaskConversationAttachmentHandlers) saveUploadedFile(src multipart.File, dstPath string) error {
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
