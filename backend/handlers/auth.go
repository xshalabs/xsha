package handlers

import (
	"net/http"
	"sleep0-backend/config"
	"sleep0-backend/utils"

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
		// 生成JWT token
		token, err := utils.GenerateJWT(loginData.Username, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "生成token失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"user":    loginData.Username,
			"token":   token,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
	}
}

// LogoutHandler 登出处理
func LogoutHandler(c *gin.Context) {
	// JWT是无状态的，客户端需要删除本地存储的token
	// 服务端只需要返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// CurrentUserHandler 获取当前用户信息
func CurrentUserHandler(c *gin.Context) {
	// 从context中获取用户信息（由认证中间件设置）
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法获取用户信息",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          username,
		"authenticated": true,
	})
}
