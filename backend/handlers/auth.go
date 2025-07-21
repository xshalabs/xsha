package handlers

import (
	"log"
	"net/http"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LoginHandler handles login
func LoginHandler(c *gin.Context) {
	cfg := config.Load()
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

	// 验证用户名和密码
	var loginSuccess bool
	var failureReason string

	if loginData.Username == cfg.AdminUser {
		if loginData.Password == cfg.AdminPass {
			loginSuccess = true
		} else {
			failureReason = "invalid_password"
		}
	} else {
		failureReason = "invalid_username"
	}

	// 异步记录登录日志（不阻塞登录流程）
	go func() {
		if err := database.AddLoginLog(loginData.Username, clientIP, userAgent, failureReason, loginSuccess); err != nil {
			log.Printf("记录登录日志失败: %v", err)
		}
	}()

	if loginSuccess {
		// 生成JWT token
		token, err := utils.GenerateJWT(loginData.Username, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": i18n.T(lang, "login.token_generate_error"),
			})
			return
		}

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
func LogoutHandler(c *gin.Context) {
	cfg := config.Load()
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
	claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": i18n.T(lang, "logout.invalid_token") + ": " + err.Error(),
		})
		return
	}

	// Get token expiration time
	expiresAt, err := utils.GetTokenExpiration(token, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "auth.get_token_exp_error"),
		})
		return
	}

	// Add token to blacklist
	if err := database.AddTokenToBlacklist(token, claims.Username, expiresAt, "logout"); err != nil {
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
func CurrentUserHandler(c *gin.Context) {
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
func GetLoginLogsHandler(c *gin.Context) {
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

	// 获取登录日志
	logs, total, err := database.GetLoginLogs(username, page, pageSize)
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
