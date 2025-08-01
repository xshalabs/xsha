package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xsha-backend/database"
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

func (s *systemConfigService) SetValue(key, value string) error {
	return s.repo.SetValue(key, value)
}

func (s *systemConfigService) InitializeDefaultConfigs() error {
	return s.repo.InitializeDefaultConfigs()
}

func (s *systemConfigService) ValidateConfigData(key, value, category string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("configuration key is required")
	}

	allowEmptyValue := s.isOptionalConfig(key)
	if !allowEmptyValue && strings.TrimSpace(value) == "" {
		return errors.New("configuration value is required")
	}

	if strings.TrimSpace(category) == "" {
		return errors.New("configuration category is required")
	}

	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-') {
			return errors.New("configuration key can only contain letters, numbers, underscores, and hyphens")
		}
	}

	return nil
}

func (s *systemConfigService) isOptionalConfig(key string) bool {
	optionalConfigs := []string{
		"git_proxy_http",
		"git_proxy_https",
		"git_proxy_no_proxy",
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
