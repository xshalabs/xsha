package services

import (
	"testing"
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"
)

// mockSystemConfigService implements SystemConfigService for testing
type mockSystemConfigService struct {
	configs map[string]string
}

func (m *mockSystemConfigService) GetValuesByKeys(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range keys {
		if value, exists := m.configs[key]; exists {
			result[key] = value
		} else {
			result[key] = ""
		}
	}
	return result, nil
}

func (m *mockSystemConfigService) ListAllConfigsWithTranslation(lang string) ([]database.SystemConfig, error) {
	return nil, nil
}

func (m *mockSystemConfigService) BatchUpdateConfigs(configs []ConfigUpdateItem) error {
	return nil
}

func (m *mockSystemConfigService) GetValue(key string) (string, error) {
	if value, exists := m.configs[key]; exists {
		return value, nil
	}
	return "", nil
}

func (m *mockSystemConfigService) InitializeDefaultConfigs() error {
	return nil
}

func (m *mockSystemConfigService) ValidateConfigData(key, value, category string) error {
	return nil
}

func (m *mockSystemConfigService) GetGitProxyConfig() (*utils.GitProxyConfig, error) {
	return nil, nil
}

func (m *mockSystemConfigService) GetGitCloneTimeout() (time.Duration, error) {
	return 5 * time.Minute, nil
}

func (m *mockSystemConfigService) GetGitSSLVerify() (bool, error) {
	return false, nil
}

func (m *mockSystemConfigService) GetDockerTimeout() (time.Duration, error) {
	return 120 * time.Minute, nil
}

func (m *mockSystemConfigService) GetSMTPEnabled() (bool, error) {
	return false, nil
}

func TestWeChatService_loadWeChatConfig(t *testing.T) {
	tests := []struct {
		name     string
		configs  map[string]string
		expected *WeChatConfig
		hasError bool
	}{
		{
			name: "enabled config",
			configs: map[string]string{
				"wechat_webhook_enabled": "true",
				"wechat_webhook_url":     "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				"wechat_webhook_timeout": "30s",
			},
			expected: &WeChatConfig{
				Enabled:    true,
				WebhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				Timeout:    30 * time.Second,
			},
			hasError: false,
		},
		{
			name: "disabled config",
			configs: map[string]string{
				"wechat_webhook_enabled": "false",
				"wechat_webhook_url":     "",
				"wechat_webhook_timeout": "30s",
			},
			expected: &WeChatConfig{
				Enabled: false,
			},
			hasError: false,
		},
		{
			name: "default timeout",
			configs: map[string]string{
				"wechat_webhook_enabled": "true",
				"wechat_webhook_url":     "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				"wechat_webhook_timeout": "",
			},
			expected: &WeChatConfig{
				Enabled:    true,
				WebhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				Timeout:    30 * time.Second,
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSysConfig := &mockSystemConfigService{configs: tt.configs}
			service := NewWeChatService(mockSysConfig).(*wechatService)

			config, err := service.loadWeChatConfig()

			if tt.hasError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config.Enabled != tt.expected.Enabled {
				t.Errorf("expected Enabled %v, got %v", tt.expected.Enabled, config.Enabled)
			}
			if config.WebhookURL != tt.expected.WebhookURL {
				t.Errorf("expected WebhookURL %v, got %v", tt.expected.WebhookURL, config.WebhookURL)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("expected Timeout %v, got %v", tt.expected.Timeout, config.Timeout)
			}
		})
	}
}

func TestWeChatService_isWeChatEnabled(t *testing.T) {
	tests := []struct {
		name     string
		configs  map[string]string
		expected bool
	}{
		{
			name: "fully configured and enabled",
			configs: map[string]string{
				"wechat_webhook_enabled": "true",
				"wechat_webhook_url":     "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
			},
			expected: true,
		},
		{
			name: "disabled",
			configs: map[string]string{
				"wechat_webhook_enabled": "false",
				"wechat_webhook_url":     "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
			},
			expected: false,
		},
		{
			name: "enabled but no webhook URL",
			configs: map[string]string{
				"wechat_webhook_enabled": "true",
				"wechat_webhook_url":     "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSysConfig := &mockSystemConfigService{configs: tt.configs}
			service := NewWeChatService(mockSysConfig).(*wechatService)

			enabled, err := service.isWeChatEnabled()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if enabled != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, enabled)
			}
		})
	}
}