package handlers

import (
	"net/http"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GitCredentialHandlers Git凭据处理器结构体
type GitCredentialHandlers struct {
	gitCredService services.GitCredentialService
}

// NewGitCredentialHandlers 创建Git凭据处理器实例
func NewGitCredentialHandlers(gitCredService services.GitCredentialService) *GitCredentialHandlers {
	return &GitCredentialHandlers{
		gitCredService: gitCredService,
	}
}

// CreateCredentialRequest 创建凭据请求结构
type CreateCredentialRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Type        string            `json:"type" binding:"required,oneof=password token ssh_key"`
	Username    string            `json:"username"`
	SecretData  map[string]string `json:"secret_data" binding:"required"`
}

// UpdateCredentialRequest 更新凭据请求结构
type UpdateCredentialRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Username    string            `json:"username"`
	SecretData  map[string]string `json:"secret_data"`
}

// CreateCredential 创建Git凭据
func (h *GitCredentialHandlers) CreateCredential(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	var req CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	credential, err := h.gitCredService.CreateCredential(
		req.Name, req.Description, req.Type, req.Username,
		username.(string), req.SecretData,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "git_credential.create_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    i18n.T(lang, "git_credential.create_success"),
		"credential": credential,
	})
}

// GetCredential 获取单个Git凭据
func (h *GitCredentialHandlers) GetCredential(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	credential, err := h.gitCredService.GetCredential(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "git_credential.not_found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credential": credential,
	})
}

// ListCredentials 获取Git凭据列表
func (h *GitCredentialHandlers) ListCredentials(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	// 解析查询参数
	page := 1
	pageSize := 20
	var credType *database.GitCredentialType

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	if t := c.Query("type"); t != "" {
		credTypeValue := database.GitCredentialType(t)
		credType = &credTypeValue
	}

	credentials, total, err := h.gitCredService.ListCredentials(username.(string), credType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"credentials": credentials,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateCredential 更新Git凭据
func (h *GitCredentialHandlers) UpdateCredential(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	var req UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}

	err = h.gitCredService.UpdateCredential(uint(id), username.(string), updates, req.SecretData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "git_credential.update_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "git_credential.update_success"),
	})
}

// DeleteCredential 删除Git凭据
func (h *GitCredentialHandlers) DeleteCredential(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	err = h.gitCredService.DeleteCredential(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "git_credential.delete_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "git_credential.delete_success"),
	})
}

// ToggleCredential 切换凭据激活状态
func (h *GitCredentialHandlers) ToggleCredential(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	err = h.gitCredService.ToggleCredential(uint(id), username.(string), req.IsActive)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "git_credential.toggle_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "git_credential.toggle_success"),
	})
}

// UseCredential 使用凭据（获取解密后的凭据信息）
func (h *GitCredentialHandlers) UseCredential(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	credential, err := h.gitCredService.UseCredential(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "git_credential.use_failed") + ": " + err.Error(),
		})
		return
	}

	// 根据凭据类型返回解密后的信息
	result := gin.H{
		"credential": credential,
		"type":       credential.Type,
		"username":   credential.Username,
	}

	switch credential.Type {
	case database.GitCredentialTypePassword, database.GitCredentialTypeToken:
		if password, err := h.gitCredService.DecryptCredentialSecret(credential, "password"); err == nil {
			result["secret"] = password
		}
	case database.GitCredentialTypeSSHKey:
		if privateKey, err := h.gitCredService.DecryptCredentialSecret(credential, "private_key"); err == nil {
			result["private_key"] = privateKey
			result["public_key"] = credential.PublicKey
		}
	}

	c.JSON(http.StatusOK, result)
}
