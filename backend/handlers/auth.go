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

	// Validate username and password
	if loginData.Username == cfg.AdminUser && loginData.Password == cfg.AdminPass {
		// Generate JWT token
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
