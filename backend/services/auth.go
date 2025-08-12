package services

import (
	"xsha-backend/config"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type authService struct {
	tokenRepo           repository.TokenBlacklistRepository
	loginLogRepo        repository.LoginLogRepository
	operationLogService AdminOperationLogService
	systemConfigRepo    repository.SystemConfigRepository
	config              *config.Config
}

func NewAuthService(tokenRepo repository.TokenBlacklistRepository, loginLogRepo repository.LoginLogRepository, operationLogService AdminOperationLogService, systemConfigRepo repository.SystemConfigRepository, cfg *config.Config) AuthService {
	return &authService{
		tokenRepo:           tokenRepo,
		loginLogRepo:        loginLogRepo,
		operationLogService: operationLogService,
		systemConfigRepo:    systemConfigRepo,
		config:              cfg,
	}
}

func (s *authService) Login(username, password, clientIP, userAgent string) (bool, string, error) {
	var loginSuccess bool
	var failureReason string
	var token string

	// Get admin username and password from system config
	adminUser, err := s.systemConfigRepo.GetValue("admin_user")
	if err != nil {
		utils.Error("Failed to get admin username from system config",
			"error", err.Error(),
		)
	}

	adminPassword, err := s.systemConfigRepo.GetValue("admin_password")
	if err != nil {
		utils.Error("Failed to get admin password from system config",
			"error", err.Error(),
		)
	}

	if username == adminUser && password == adminPassword {
		loginSuccess = true

		var err error
		token, err = utils.GenerateJWT(username, s.config.JWTSecret)
		if err != nil {
			go func() {
				if logErr := s.operationLogService.LogLogin(username, clientIP, userAgent, false, "token_generation_failed"); logErr != nil {
					utils.Error("Failed to record admin operation log",
						"username", username,
						"client_ip", clientIP,
						"error", logErr.Error(),
					)
				}
			}()

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
		if username != adminUser {
			failureReason = "invalid_username"
		} else {
			failureReason = "invalid_password"
		}
	}

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

func (s *authService) Logout(token, username, clientIP, userAgent string) error {
	expiresAt, err := utils.GetTokenExpiration(token, s.config.JWTSecret)
	if err != nil {
		go func() {
			if logErr := s.operationLogService.LogLogout(username, clientIP, userAgent, false, err.Error()); logErr != nil {
				utils.Error("Failed to record admin operation log for logout failure",
					"username", username,
					"client_ip", clientIP,
					"error", logErr.Error(),
				)
			}
		}()
		return err
	}

	err = s.tokenRepo.Add(token, username, expiresAt, "logout")

	go func() {
		logoutSuccess := err == nil
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}

		if logErr := s.operationLogService.LogLogout(username, clientIP, userAgent, logoutSuccess, errorMsg); logErr != nil {
			utils.Error("Failed to record admin operation log for logout",
				"username", username,
				"client_ip", clientIP,
				"success", logoutSuccess,
				"error", logErr.Error(),
			)
		} else if logoutSuccess {
			utils.Info("User logged out successfully",
				"username", username,
				"client_ip", clientIP,
				"user_agent", userAgent,
			)
		}
	}()

	return err
}

func (s *authService) IsTokenBlacklisted(token string) (bool, error) {
	return s.tokenRepo.IsBlacklisted(token)
}

func (s *authService) CleanExpiredTokens() error {
	return s.tokenRepo.CleanExpired()
}
