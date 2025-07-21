package handlers

import (
	"net/http"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/utils"

	"github.com/gin-gonic/gin"
)

// LoginHandler 登录处理
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

	// 验证用户名和密码
	if loginData.Username == cfg.AdminUser && loginData.Password == cfg.AdminPass {
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

// LogoutHandler 登出处理
func LogoutHandler(c *gin.Context) {
	cfg := config.Load()
	lang := middleware.GetLangFromContext(c)

	// 从Authorization header中获取token
	authHeader := c.GetHeader("Authorization")
	token, err := utils.ExtractTokenFromAuthHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "logout.invalid_token") + "：" + err.Error(),
		})
		return
	}

	// 验证token并获取用户信息
	claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": i18n.T(lang, "logout.invalid_token") + "：" + err.Error(),
		})
		return
	}

	// 获取token过期时间
	expiresAt, err := utils.GetTokenExpiration(token, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "auth.get_token_exp_error"),
		})
		return
	}

	// 将token添加到黑名单
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

// CurrentUserHandler 获取当前用户信息
func CurrentUserHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 从context中获取用户信息（由认证中间件设置）
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
