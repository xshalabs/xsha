package handlers

import (
	"net/http"
	"sleep0-backend/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LoginHandler 登录处理
func LoginHandler(c *gin.Context) {
	cfg := config.Load()

	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求数据格式错误",
		})
		return
	}

	// 验证用户名和密码
	if loginData.Username == cfg.AdminUser && loginData.Password == cfg.AdminPass {
		session := sessions.Default(c)
		session.Set("authenticated", true)
		session.Set("username", loginData.Username)
		session.Save()

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"user":    loginData.Username,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
	}
}

// LogoutHandler 登出处理
func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("authenticated")
	session.Delete("username")
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// CurrentUserHandler 获取当前用户信息
func CurrentUserHandler(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"user":          username,
		"authenticated": true,
	})
}
