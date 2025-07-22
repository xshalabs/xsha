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
// @Description 创建Git凭据的请求参数
type CreateCredentialRequest struct {
	Name        string            `json:"name" binding:"required" example:"我的GitHub凭据"`
	Description string            `json:"description" example:"用于GitHub项目的凭据"`
	Type        string            `json:"type" binding:"required,oneof=password token ssh_key" example:"password"`
	Username    string            `json:"username" example:"myusername"`
	SecretData  map[string]string `json:"secret_data" binding:"required" example:"{\"password\":\"mypassword\"}"`
}

// UpdateCredentialRequest 更新凭据请求结构
// @Description 更新Git凭据的请求参数
type UpdateCredentialRequest struct {
	Name        string            `json:"name" example:"更新的凭据名称"`
	Description string            `json:"description" example:"更新的描述"`
	Username    string            `json:"username" example:"newusername"`
	SecretData  map[string]string `json:"secret_data" example:"{\"password\":\"newpassword\"}"`
}

// CreateCredential 创建Git凭据
// @Summary 创建Git凭据
// @Description 创建新的Git凭据，支持密码、令牌、SSH密钥类型
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credential body CreateCredentialRequest true "凭据信息"
// @Success 201 {object} object{message=string,credential=object} "凭据创建成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 500 {object} object{error=string} "创建凭据失败"
// @Router /git-credentials [post]
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
// @Summary 获取Git凭据详情
// @Description 根据ID获取指定Git凭据的详细信息
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "凭据ID"
// @Success 200 {object} object{credential=object} "凭据详情"
// @Failure 400 {object} object{error=string} "无效的凭据ID"
// @Failure 404 {object} object{error=string} "凭据不存在"
// @Router /git-credentials/{id} [get]
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
// @Summary 获取Git凭据列表
// @Description 获取当前用户的Git凭据列表，支持按类型筛选和分页
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "凭据类型筛选 (password/token/ssh_key)"
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20，最大100"
// @Success 200 {object} object{message=string,credentials=[]object,total=number,page=number,page_size=number,total_pages=number} "凭据列表"
// @Failure 500 {object} object{error=string} "获取凭据列表失败"
// @Router /git-credentials [get]
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
// @Summary 更新Git凭据
// @Description 更新指定Git凭据的信息
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "凭据ID"
// @Param credential body UpdateCredentialRequest true "凭据更新信息"
// @Success 200 {object} object{message=string} "凭据更新成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "凭据不存在"
// @Router /git-credentials/{id} [put]
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
// @Summary 删除Git凭据
// @Description 删除指定的Git凭据
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "凭据ID"
// @Success 200 {object} object{message=string} "凭据删除成功"
// @Failure 400 {object} object{error=string} "无效的凭据ID"
// @Failure 404 {object} object{error=string} "凭据不存在"
// @Router /git-credentials/{id} [delete]
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
// @Summary 切换凭据状态
// @Description 切换Git凭据的激活/停用状态
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "凭据ID"
// @Param toggleData body object{is_active=bool} true "切换状态信息"
// @Success 200 {object} object{message=string} "凭据状态切换成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "凭据不存在"
// @Router /git-credentials/{id}/toggle [patch]
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
// @Summary 使用Git凭据
// @Description 获取解密后的Git凭据信息用于认证
// @Tags Git凭据
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "凭据ID"
// @Success 200 {object} object{credential=object,type=string,username=string,secret=string,private_key=string,public_key=string} "凭据使用成功"
// @Failure 400 {object} object{error=string} "无效的凭据ID"
// @Failure 404 {object} object{error=string} "凭据不存在"
// @Router /git-credentials/{id}/use [post]
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
