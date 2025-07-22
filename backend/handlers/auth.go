package handlers

import (
	"net/http"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/services"
	"sleep0-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AuthHandlers 认证处理器结构体
type AuthHandlers struct {
	authService     services.AuthService
	loginLogService services.LoginLogService
}

// NewAuthHandlers 创建认证处理器实例
func NewAuthHandlers(authService services.AuthService, loginLogService services.LoginLogService) *AuthHandlers {
	return &AuthHandlers{
		authService:     authService,
		loginLogService: loginLogService,
	}
}

// LoginHandler handles login
// @Summary 用户登录
// @Description 使用用户名和密码进行登录认证
// @Tags 认证
// @Accept json
// @Produce json
// @Param loginData body object{username=string,password=string} true "登录信息"
// @Success 200 {object} object{token=string,expires_at=string} "登录成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败"
// @Failure 429 {object} object{error=string} "请求过于频繁"
// @Router /auth/login [post]
func (h *AuthHandlers) LoginHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "login.invalid_request"),
		})
		return
	}

	// 获取客户端信息用于日志记录
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 使用authService进行登录验证和日志记录
	loginSuccess, token, err := h.authService.Login(loginData.Username, loginData.Password, clientIP, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "login.token_generate_error"),
		})
		return
	}

	if loginSuccess {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.T(lang, "login.success"),
			"user":    loginData.Username,
			"token":   token,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": i18n.T(lang, "login.failed"),
		})
	}
}

// LogoutHandler handles logout
// @Summary 用户登出
// @Description 登出当前用户，将token加入黑名单
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{message=string} "登出成功"
// @Failure 400 {object} object{error=string} "无效的token"
// @Failure 401 {object} object{error=string} "认证失败"
// @Failure 500 {object} object{error=string} "登出失败"
// @Router /auth/logout [post]
func (h *AuthHandlers) LogoutHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	token, err := utils.ExtractTokenFromAuthHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "logout.invalid_token") + ": " + err.Error(),
		})
		return
	}

	// Validate token and get user information
	claims, err := utils.ValidateJWT(token, "")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": i18n.T(lang, "logout.invalid_token") + ": " + err.Error(),
		})
		return
	}

	// 使用认证服务进行登出
	if err := h.authService.Logout(token, claims.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "logout.failed"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "logout.success"),
	})
}

// CurrentUserHandler gets current user information
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{username=string} "用户信息"
// @Failure 500 {object} object{error=string} "获取用户信息失败"
// @Router /user/current [get]
func (h *AuthHandlers) CurrentUserHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// Get user information from context (set by auth middleware)
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "user.get_info_error"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          username,
		"authenticated": true,
		"message":       i18n.T(lang, "user.authenticated"),
	})
}

// GetLoginLogsHandler 获取登录日志（需要管理员权限）
// @Summary 获取登录日志
// @Description 获取系统的登录日志记录，支持按用户名筛选和分页
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string false "用户名筛选"
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20，最大100"
// @Success 200 {object} object{message=string,logs=[]object,total=number,page=number,page_size=number,total_pages=number} "登录日志列表"
// @Failure 500 {object} object{error=string} "获取登录日志失败"
// @Router /admin/login-logs [get]
func (h *AuthHandlers) GetLoginLogsHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取查询参数
	username := c.Query("username")
	page := 1
	pageSize := 20

	// 解析分页参数
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

	// 使用登录日志服务获取日志
	logs, total, err := h.loginLogService.GetLogs(username, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"logs":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// 保持向后兼容的全局函数
var globalAuthHandlers *AuthHandlers

// SetAuthHandlers 设置全局认证处理器实例
func SetAuthHandlers(handlers *AuthHandlers) {
	globalAuthHandlers = handlers
}

// LoginHandler 全局登录处理器（向后兼容）
func LoginHandler(c *gin.Context) {
	globalAuthHandlers.LoginHandler(c)
}

// LogoutHandler 全局登出处理器（向后兼容）
func LogoutHandler(c *gin.Context) {
	globalAuthHandlers.LogoutHandler(c)
}

// CurrentUserHandler 全局当前用户处理器（向后兼容）
func CurrentUserHandler(c *gin.Context) {
	globalAuthHandlers.CurrentUserHandler(c)
}

// GetLoginLogsHandler 全局登录日志处理器（向后兼容）
func GetLoginLogsHandler(c *gin.Context) {
	globalAuthHandlers.GetLoginLogsHandler(c)
}
