package result_parser

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// 日志解析相关配置
	MaxLogLines   int           `json:"max_log_lines"`
	ParseTimeout  time.Duration `json:"parse_timeout"`
	RetryAttempts int           `json:"retry_attempts"`

	// JSON解析配置
	JSONLogPatterns []string `json:"json_log_patterns"`
	RequiredFields  []string `json:"required_fields"`
	OptionalFields  []string `json:"optional_fields"`

	// 验证配置
	StrictValidation bool `json:"strict_validation"`
	AllowPartialData bool `json:"allow_partial_data"`

	// 性能配置
	BufferSize    int  `json:"buffer_size"`
	EnableMetrics bool `json:"enable_metrics"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxLogLines:   10000,
		ParseTimeout:  30 * time.Second,
		RetryAttempts: 3,
		JSONLogPatterns: []string{
			`^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`,
			`^(\{.*\})$`,
		},
		RequiredFields: []string{
			"type",
			"subtype",
			"is_error",
			"session_id",
		},
		OptionalFields: []string{
			"duration_ms",
			"duration_api_ms",
			"num_turns",
			"result",
			"total_cost_usd",
			"usage",
		},
		StrictValidation: false,
		AllowPartialData: true,
		BufferSize:       4096,
		EnableMetrics:    true,
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	if val := os.Getenv("XSHA_PARSER_MAX_LOG_LINES"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			c.MaxLogLines = intVal
		}
	}

	if val := os.Getenv("XSHA_PARSER_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.ParseTimeout = duration
		}
	}

	if val := os.Getenv("XSHA_PARSER_RETRY_ATTEMPTS"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			c.RetryAttempts = intVal
		}
	}

	if val := os.Getenv("XSHA_PARSER_STRICT_VALIDATION"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			c.StrictValidation = boolVal
		}
	}

	if val := os.Getenv("XSHA_PARSER_ALLOW_PARTIAL_DATA"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			c.AllowPartialData = boolVal
		}
	}

	if val := os.Getenv("XSHA_PARSER_BUFFER_SIZE"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			c.BufferSize = intVal
		}
	}

	if val := os.Getenv("XSHA_PARSER_ENABLE_METRICS"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			c.EnableMetrics = boolVal
		}
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.MaxLogLines <= 0 {
		c.MaxLogLines = 1000
	}

	if c.ParseTimeout <= 0 {
		c.ParseTimeout = 30 * time.Second
	}

	if c.RetryAttempts < 0 {
		c.RetryAttempts = 0
	}

	if c.BufferSize <= 0 {
		c.BufferSize = 4096
	}

	if len(c.RequiredFields) == 0 {
		c.RequiredFields = []string{"type", "subtype", "is_error", "session_id"}
	}

	if len(c.JSONLogPatterns) == 0 {
		c.JSONLogPatterns = []string{
			`^(?:\[\d{2}:\d{2}:\d{2}\]\s*)?(?:\w+:\s*)?(\{.*\})\s*$`,
			`^(\{.*\})$`,
		}
	}

	return nil
}
