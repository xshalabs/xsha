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
// @Router /admin/avatar/upload [post]
func (h *AdminAvatarHandlers) UploadAvatarHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	adminIDInterface, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}
	adminID, ok := adminIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
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

	// Create avatar record
	avatar, err := h.avatarService.UploadAvatar(
		fileName,
		header.Filename,
		contentType,
		header.Size,
		filePath,
		adminID,
		username.(string),
	)
	if err != nil {
		// Clean up file if database operation failed
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	// Update admin's avatar_id
	if err := h.avatarService.UpdateAdminAvatar(adminID, avatar.ID); err != nil {
		utils.Error("Failed to update admin avatar", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
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

	// Check if file exists
	if _, err := os.Stat(avatar.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "avatar.file_not_found")})
		return
	}

	// Set headers for inline display
	c.Header("Content-Type", avatar.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", avatar.FileSize))
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Serve file inline
	c.File(avatar.FilePath)
}

// UpdateAdminAvatarHandler updates admin's avatar
// @Summary Update admin avatar
// @Description Update administrator's avatar by setting avatar ID
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Admin ID"
// @Param avatarData body object{avatar_id=uint} true "Avatar ID information"
// @Success 200 {object} object{message=string} "Avatar updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Admin or avatar not found"
// @Router /admin/users/{id}/avatar [put]
func (h *AdminAvatarHandlers) UpdateAdminAvatarHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idParam := c.Param("id")
	adminID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var avatarData struct {
		AvatarID uint `json:"avatar_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&avatarData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Verify avatar exists
	_, err = h.avatarService.GetAvatar(avatarData.AvatarID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "avatar.not_found")})
		return
	}

	// Update admin's avatar
	if err := h.avatarService.UpdateAdminAvatar(uint(adminID), avatarData.AvatarID); err != nil {
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