package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

type systemConfigService struct {
	repo repository.SystemConfigRepository
}

func NewSystemConfigService(repo repository.SystemConfigRepository) SystemConfigService {
	return &systemConfigService{
		repo: repo,
	}
}

func (s *systemConfigService) ListAllConfigs() ([]database.SystemConfig, error) {
	return s.repo.ListAll()
}

func (s *systemConfigService) BatchUpdateConfigs(configItems []ConfigUpdateItem) error {
	for _, item := range configItems {
		existingConfig, err := s.repo.GetByKey(item.ConfigKey)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("configuration %s does not exist", item.ConfigKey)
			}
			return fmt.Errorf("failed to check existing config for key %s: %v", item.ConfigKey, err)
		}

		if !existingConfig.IsEditable {
			return fmt.Errorf("configuration %s is not editable", item.ConfigKey)
		}

		if err := s.ValidateConfigData(item.ConfigKey, item.ConfigValue, existingConfig.Category); err != nil {
			return fmt.Errorf("validation failed for key %s: %v", item.ConfigKey, err)
		}

		existingConfig.ConfigValue = item.ConfigValue
		if err := s.repo.Update(existingConfig); err != nil {
			return fmt.Errorf("failed to update config %s: %v", item.ConfigKey, err)
		}
	}

	return nil
}

func (s *systemConfigService) GetValue(key string) (string, error) {
	return s.repo.GetValue(key)
}

func (s *systemConfigService) GetValuesByKeys(keys []string) (map[string]string, error) {
	configs, err := s.repo.GetByKeys(keys)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, key := range keys {
		if config, exists := configs[key]; exists {
			result[key] = config.ConfigValue
		} else {
			result[key] = ""
		}
	}

	return result, nil
}

func (s *systemConfigService) SetValue(key, value string) error {
	return s.repo.SetValue(key, value)
}

func (s *systemConfigService) InitializeDefaultConfigs() error {
	return s.repo.InitializeDefaultConfigs()
}

func (s *systemConfigService) ValidateConfigData(key, value, category string) error {
	if strings.TrimSpace(key) == "" {
		return appErrors.ErrSystemConfigKeyRequired
	}

	allowEmptyValue := s.isOptionalConfig(key)
	if !allowEmptyValue && strings.TrimSpace(value) == "" {
		return appErrors.ErrSystemConfigValueRequired
	}

	if strings.TrimSpace(category) == "" {
		return appErrors.ErrSystemConfigCategoryRequired
	}

	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-') {
			return appErrors.ErrSystemConfigInvalidKeyFormat
		}
	}

	return nil
}

func (s *systemConfigService) isOptionalConfig(key string) bool {
	optionalConfigs := []string{
		"git_proxy_http",
		"git_proxy_https",
		"git_proxy_no_proxy",
		"smtp_host",
		"smtp_username",
		"smtp_password",
		"smtp_from",
	}

	for _, optionalKey := range optionalConfigs {
		if key == optionalKey {
			return true
		}
	}

	return false
}

func (s *systemConfigService) GetGitProxyConfig() (*utils.GitProxyConfig, error) {
	enabled, err := s.repo.GetValue("git_proxy_enabled")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get git_proxy_enabled: %v", err)
	}

	isEnabled := false
	if enabled != "" {
		isEnabled, _ = strconv.ParseBool(enabled)
	}

	httpProxy, err := s.repo.GetValue("git_proxy_http")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get git_proxy_http: %v", err)
	}

	httpsProxy, err := s.repo.GetValue("git_proxy_https")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get git_proxy_https: %v", err)
	}

	noProxy, err := s.repo.GetValue("git_proxy_no_proxy")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get git_proxy_no_proxy: %v", err)
	}

	return &utils.GitProxyConfig{
		Enabled:    isEnabled,
		HttpProxy:  httpProxy,
		HttpsProxy: httpsProxy,
		NoProxy:    noProxy,
	}, nil
}

func (s *systemConfigService) GetGitCloneTimeout() (time.Duration, error) {
	timeoutStr, err := s.repo.GetValue("git_clone_timeout")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 5 * time.Minute, nil
		}
		return 0, fmt.Errorf("failed to get git_clone_timeout: %v", err)
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		utils.Error("Failed to parse git clone timeout, using default 5 minutes", "timeout", timeoutStr, "error", err)
		return 5 * time.Minute, nil
	}

	return timeout, nil
}

func (s *systemConfigService) GetGitSSLVerify() (bool, error) {
	verifyStr, err := s.repo.GetValue("git_ssl_verify")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get git_ssl_verify: %v", err)
	}

	verify, err := strconv.ParseBool(verifyStr)
	if err != nil {
		utils.Error("Failed to parse git SSL verify, using default false", "value", verifyStr, "error", err)
		return false, nil
	}

	return verify, nil
}

func (s *systemConfigService) GetDockerTimeout() (time.Duration, error) {
	timeoutStr, err := s.repo.GetValue("docker_timeout")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 120 * time.Minute, nil
		}
		return 0, fmt.Errorf("failed to get docker_timeout: %v", err)
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		utils.Error("Failed to parse docker timeout, using default 120 minutes", "timeout", timeoutStr, "error", err)
		return 120 * time.Minute, nil
	}

	return timeout, nil
}

// GetSMTPEnabled returns whether SMTP email service is enabled
func (s *systemConfigService) GetSMTPEnabled() (bool, error) {
	enabled, err := s.repo.GetValue("smtp_enabled")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get smtp_enabled: %v", err)
	}

	return enabled == "true", nil
}

// IsEmailServiceConfigured checks if all required email configurations are set
func (s *systemConfigService) IsEmailServiceConfigured() (bool, error) {
	enabled, err := s.GetSMTPEnabled()
	if err != nil {
		return false, err
	}

	if !enabled {
		return false, nil
	}

	// Check required fields
	requiredFields := []string{"smtp_host", "smtp_username", "smtp_password", "smtp_from"}
	for _, field := range requiredFields {
		value, err := s.repo.GetValue(field)
		if err != nil && err != gorm.ErrRecordNotFound {
			return false, fmt.Errorf("failed to get %s: %v", field, err)
		}
		if value == "" {
			return false, nil
		}
	}

	return true, nil
}
