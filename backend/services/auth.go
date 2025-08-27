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
	adminService        AdminService
	adminRepo           repository.AdminRepository
	config              *config.Config
}

func NewAuthService(tokenRepo repository.TokenBlacklistRepository, loginLogRepo repository.LoginLogRepository, operationLogService AdminOperationLogService, adminService AdminService, adminRepo repository.AdminRepository, cfg *config.Config) AuthService {
	return &authService{
		tokenRepo:           tokenRepo,
		loginLogRepo:        loginLogRepo,
		operationLogService: operationLogService,
		adminService:        adminService,
		adminRepo:           adminRepo,
		config:              cfg,
	}
}

func (s *authService) Login(username, password, clientIP, userAgent string) (bool, string, error) {
	var loginSuccess bool
	var failureReason string
	var token string

	// Validate admin credentials using AdminService
	_, err := s.adminService.ValidateCredentials(username, password)
	if err != nil {
		loginSuccess = false
		if err.Error() == "admin.invalid_credentials" {
			failureReason = "invalid_credentials"
		} else if err.Error() == "admin.inactive" {
			failureReason = "account_inactive"
		} else {
			failureReason = "validation_error"
			utils.Error("Failed to validate admin credentials",
				"username", username,
				"error", err.Error(),
			)
		}
	} else {
		loginSuccess = true

		// Generate JWT token
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

		// Update admin's last login information
		go func() {
			if err := s.adminRepo.UpdateLastLogin(username, clientIP); err != nil {
				utils.Error("Failed to update admin last login info",
					"username", username,
					"client_ip", clientIP,
					"error", err.Error(),
				)
			}
		}()
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

func (s *authService) InvalidateUserSessions(username string) error {
	return s.tokenRepo.InvalidateAllUserTokens(username, "admin_deactivated")
}

func (s *authService) IsUserDeactivated(username string) (bool, error) {
	return s.tokenRepo.IsUserDeactivated(username)
}
