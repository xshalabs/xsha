package services

import (
	"sleep0-backend/config"
	"sleep0-backend/repository"
	"sleep0-backend/utils"
)

type authService struct {
	tokenRepo           repository.TokenBlacklistRepository
	loginLogRepo        repository.LoginLogRepository
	operationLogService AdminOperationLogService
	config              *config.Config
}

// NewAuthService 创建认证服务实例
func NewAuthService(tokenRepo repository.TokenBlacklistRepository, loginLogRepo repository.LoginLogRepository, operationLogService AdminOperationLogService, cfg *config.Config) AuthService {
	return &authService{
		tokenRepo:           tokenRepo,
		loginLogRepo:        loginLogRepo,
		operationLogService: operationLogService,
		config:              cfg,
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
			// 记录失败的管理员操作日志
			go func() {
				if logErr := s.operationLogService.LogLogin(username, clientIP, userAgent, false, "token_generation_failed"); logErr != nil {
					utils.Error("Failed to record admin operation log",
						"username", username,
						"client_ip", clientIP,
						"error", logErr.Error(),
					)
				}
			}()

			// 记录常规登录日志
			go func() {
				if logErr := s.loginLogRepo.Add(username, clientIP, userAgent, "token_generation_failed", false); logErr != nil {
					utils.Error("Failed to record login log",
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

	// 记录管理员操作日志
	go func() {
		if err := s.operationLogService.LogLogin(username, clientIP, userAgent, loginSuccess, failureReason); err != nil {
			utils.Error("Failed to record admin operation log",
				"username", username,
				"client_ip", clientIP,
				"success", loginSuccess,
				"error", err.Error(),
			)
		}
	}()

	// 异步记录登录日志（不阻塞登录流程）
	go func() {
		if err := s.loginLogRepo.Add(username, clientIP, userAgent, failureReason, loginSuccess); err != nil {
			utils.Error("Failed to record login log",
				"username", username,
				"client_ip", clientIP,
				"success", loginSuccess,
				"error", err.Error(),
			)
		} else if loginSuccess {
			utils.Info("User logged in successfully",
				"username", username,
				"client_ip", clientIP,
				"user_agent", userAgent,
			)
		} else {
			utils.Warn("Login attempt failed",
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
		// 记录失败的管理员操作日志
		go func() {
			if logErr := s.operationLogService.LogLogout(username, "", "", false, err.Error()); logErr != nil {
				utils.Error("Failed to record admin operation log for logout failure",
					"username", username,
					"error", logErr.Error(),
				)
			}
		}()
		return err
	}

	// 将token添加到黑名单
	err = s.tokenRepo.Add(token, username, expiresAt, "logout")

	// 记录管理员操作日志
	go func() {
		logoutSuccess := err == nil
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}

		if logErr := s.operationLogService.LogLogout(username, "", "", logoutSuccess, errorMsg); logErr != nil {
			utils.Error("Failed to record admin operation log for logout",
				"username", username,
				"success", logoutSuccess,
				"error", logErr.Error(),
			)
		} else if logoutSuccess {
			utils.Info("User logged out successfully",
				"username", username,
			)
		}
	}()

	return err
}

// IsTokenBlacklisted 检查Token是否在黑名单
func (s *authService) IsTokenBlacklisted(token string) (bool, error) {
	return s.tokenRepo.IsBlacklisted(token)
}

// CleanExpiredTokens 清理过期Token
func (s *authService) CleanExpiredTokens() error {
	return s.tokenRepo.CleanExpired()
}
