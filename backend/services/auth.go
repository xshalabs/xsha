package services

import (
	"log/slog"
	"sleep0-backend/config"
	"sleep0-backend/repository"
	"sleep0-backend/utils"
)

type authService struct {
	tokenRepo    repository.TokenBlacklistRepository
	loginLogRepo repository.LoginLogRepository
	config       *config.Config
	logger       *slog.Logger
}

// NewAuthService 创建认证服务实例
func NewAuthService(tokenRepo repository.TokenBlacklistRepository, loginLogRepo repository.LoginLogRepository, cfg *config.Config) AuthService {
	logger := utils.WithFields(map[string]interface{}{
		"component": "auth_service",
	})

	return &authService{
		tokenRepo:    tokenRepo,
		loginLogRepo: loginLogRepo,
		config:       cfg,
		logger:       logger,
	}
}

// Login 用户登录
func (s *authService) Login(username, password, clientIP, userAgent string) (bool, string, error) {
	var loginSuccess bool
	var failureReason string
	var token string

	// 验证用户名和密码
	if username == s.config.AdminUser && password == s.config.AdminPass {
		loginSuccess = true

		// 生成JWT token
		var err error
		token, err = utils.GenerateJWT(username, s.config.JWTSecret)
		if err != nil {
			// 记录失败日志
			go func() {
				if logErr := s.loginLogRepo.Add(username, clientIP, userAgent, "token_generation_failed", false); logErr != nil {
					s.logger.Error("Failed to record login log",
						"username", username,
						"client_ip", clientIP,
						"error", logErr.Error(),
					)
				}
			}()
			return false, "", err
		}
	} else {
		loginSuccess = false
		if username != s.config.AdminUser {
			failureReason = "invalid_username"
		} else {
			failureReason = "invalid_password"
		}
	}

	// 异步记录登录日志（不阻塞登录流程）
	go func() {
		if err := s.loginLogRepo.Add(username, clientIP, userAgent, failureReason, loginSuccess); err != nil {
			s.logger.Error("Failed to record login log",
				"username", username,
				"client_ip", clientIP,
				"success", loginSuccess,
				"error", err.Error(),
			)
		} else if loginSuccess {
			s.logger.Info("User logged in successfully",
				"username", username,
				"client_ip", clientIP,
				"user_agent", userAgent,
			)
		} else {
			s.logger.Warn("Login attempt failed",
				"username", username,
				"client_ip", clientIP,
				"user_agent", userAgent,
				"reason", failureReason,
			)
		}
	}()

	return loginSuccess, token, nil
}

// Logout 用户登出
func (s *authService) Logout(token, username string) error {
	// 获取token过期时间
	expiresAt, err := utils.GetTokenExpiration(token, s.config.JWTSecret)
	if err != nil {
		return err
	}

	// 将token添加到黑名单
	return s.tokenRepo.Add(token, username, expiresAt, "logout")
}

// IsTokenBlacklisted 检查Token是否在黑名单
func (s *authService) IsTokenBlacklisted(token string) (bool, error) {
	return s.tokenRepo.IsBlacklisted(token)
}

// CleanExpiredTokens 清理过期Token
func (s *authService) CleanExpiredTokens() error {
	return s.tokenRepo.CleanExpired()
}
