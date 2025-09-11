package result_parser

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// 基本配置
	ParseTimeout time.Duration `json:"parse_timeout"`

	// 验证配置
	StrictValidation bool     `json:"strict_validation"`
	RequiredFields   []string `json:"required_fields"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ParseTimeout:     30 * time.Second,
		StrictValidation: false,
		RequiredFields: []string{
			"type",
			"subtype",
			"is_error",
			"session_id",
		},
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	if val := os.Getenv("XSHA_PARSER_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.ParseTimeout = duration
		}
	}

	if val := os.Getenv("XSHA_PARSER_STRICT_VALIDATION"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			c.StrictValidation = boolVal
		}
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.ParseTimeout <= 0 {
		c.ParseTimeout = 30 * time.Second
	}

	if len(c.RequiredFields) == 0 {
		c.RequiredFields = []string{"type", "subtype", "is_error", "session_id"}
	}

	return nil
}
