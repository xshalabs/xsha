package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

type AdminAvatarHandlers struct {
	avatarService services.AdminAvatarService
	adminService  services.AdminService
}

func NewAdminAvatarHandlers(avatarService services.AdminAvatarService, adminService services.AdminService) *AdminAvatarHandlers {
	return &AdminAvatarHandlers{
		avatarService: avatarService,
		adminService:  adminService,
	}
}

// UploadAvatarHandler uploads an avatar for admin
// @Summary Upload avatar
// @Description Upload an avatar image for administrator
// @Tags Admin Avatar
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Avatar image to upload"
// @Success 200 {object} object{message=string,data=object} "Avatar uploaded successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 401 {object} object{error=string} "Authentication failed"
// @Failure 413 {object} object{error=string} "File too large"
// @Router /avatar/upload [post]
func (h *AdminAvatarHandlers) UploadAvatarHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "avatar.file_required")})
		return
	}
	defer file.Close()

	// Validate file size (max 5MB)
	const maxFileSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": i18n.T(lang, "avatar.file_too_large"),
		})
		return
	}

	// Validate file type (images only)
	contentType := header.Header.Get("Content-Type")
	if !h.isValidImageType(contentType, header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "avatar.unsupported_file_type")})
		return
	}

	// Create storage directory
	storageDir := h.avatarService.GetAvatarStorageDir()
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		utils.Error("Failed to create avatar storage directory", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	// Generate unique filename
	fileName := h.avatarService.GenerateAvatarFileName(header.Filename)
	filePath := filepath.Join(storageDir, fileName)

	// Save file
	if err := h.saveUploadedFile(file, filePath); err != nil {
		utils.Error("Failed to save uploaded avatar file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "avatar.save_failed")})
		return
	}

	// Create avatar record with relative path (just the filename)
	avatar, err := h.avatarService.UploadAvatar(
		fileName,
		header.Filename,
		contentType,
		header.Size,
		fileName,
		admin.ID,
		admin.Username,
	)
	if err != nil {
		// Clean up file if database operation failed
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	// Return avatar info with preview URL
	avatarResponse := map[string]interface{}{
		"id":            avatar.ID,
		"uuid":          avatar.UUID,
		"original_name": avatar.OriginalName,
		"file_size":     avatar.FileSize,
		"content_type":  avatar.ContentType,
		"preview_url":   fmt.Sprintf("/api/v1/admin/avatar/preview/%s", avatar.UUID),
		"created_at":    avatar.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "avatar.upload_success"),
		"data":    avatarResponse,
	})
}

// PreviewAvatarHandler previews an avatar by UUID
// @Summary Preview avatar
// @Description Preview avatar image by UUID
// @Tags Admin Avatar
// @Accept json
// @Produce image/*
// @Param uuid path string true "Avatar UUID"
// @Success 200 {file} binary "Avatar image for preview"
// @Failure 400 {object} object{error=string} "Invalid UUID"
// @Failure 404 {object} object{error=string} "Avatar not found"
// @Router /admin/avatar/preview/{uuid} [get]
func (h *AdminAvatarHandlers) PreviewAvatarHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "avatar.invalid_uuid")})
		return
	}

	avatar, err := h.avatarService.GetAvatarByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "avatar.not_found")})
		return
	}

	// Construct full file path from relative path
	fullFilePath := h.avatarService.GetFullAvatarPath(avatar.FilePath)

	// Check if file exists
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "avatar.file_not_found")})
		return
	}

	// Set headers for inline display
	c.Header("Content-Type", avatar.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", avatar.FileSize))
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Serve file inline
	c.File(fullFilePath)
}

// UpdateAdminAvatarHandler updates admin's avatar by avatar UUID
// @Summary Update admin avatar
// @Description Update administrator's avatar by avatar UUID
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Avatar UUID"
// @Param adminData body object{admin_id=uint} true "Admin ID information"
// @Success 200 {object} object{message=string} "Avatar updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Admin or avatar not found"
// @Router /admin/avatar/{uuid} [put]
func (h *AdminAvatarHandlers) UpdateAdminAvatarHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "avatar.invalid_uuid"),
		})
		return
	}

	var adminData struct {
		AdminID uint `json:"admin_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&adminData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Update admin's avatar using UUID
	if err := h.avatarService.UpdateAdminAvatarByUUID(uuid, adminData.AdminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "avatar.update_success"),
	})
}

// isValidImageType checks if the content type and filename indicate a valid image
func (h *AdminAvatarHandlers) isValidImageType(contentType, filename string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if validTypes[contentType] {
		return true
	}

	// Check file extension as fallback
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	return validExtensions[ext]
}

// saveUploadedFile saves an uploaded file to the specified path
func (h *AdminAvatarHandlers) saveUploadedFile(file multipart.File, filePath string) error {
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}
