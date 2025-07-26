package utils

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// LogLevel 日志级别类型
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogFormat 日志格式类型
type LogFormat string

const (
	FormatJSON LogFormat = "JSON"
	FormatText LogFormat = "TEXT"
)

// LogConfig 日志配置
type LogConfig struct {
	Level  LogLevel  `json:"level" env:"LOG_LEVEL" default:"INFO"`
	Format LogFormat `json:"format" env:"LOG_FORMAT" default:"JSON"`
	Output string    `json:"output" env:"LOG_OUTPUT" default:"stdout"` // stdout, stderr, file path
}

var defaultLogger *slog.Logger

// InitLogger 初始化全局日志记录器
func InitLogger(config LogConfig) error {
	// 解析日志级别
	var level slog.Level
	switch strings.ToUpper(string(config.Level)) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 配置选项
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // 添加源码位置信息
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format("2006-01-02T15:04:05.000Z07:00"))
			}
			// 简化源码路径
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// 只显示文件名和行号
					source.File = filepath.Base(source.File)
				}
			}
			return a
		},
	}

	// 确定输出目标
	var output *os.File
	switch config.Output {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// 文件输出
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}

	// 创建Handler
	var handler slog.Handler
	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	case FormatText:
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	// 设置全局日志记录器
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	return nil
}

// GetLogger 获取默认日志记录器
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		InitLogger(LogConfig{
			Level:  LevelInfo,
			Format: FormatJSON,
			Output: "stdout",
		})
	}
	return defaultLogger
}

// WithContext 创建带上下文的日志记录器
func WithContext(ctx context.Context) *slog.Logger {
	logger := GetLogger()

	// 从上下文中提取有用的信息
	if traceID := ctx.Value("trace_id"); traceID != nil {
		logger = logger.With("trace_id", traceID)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.With("user_id", userID)
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		logger = logger.With("request_id", requestID)
	}

	return logger
}

// WithFields 创建带字段的日志记录器
func WithFields(fields map[string]interface{}) *slog.Logger {
	logger := GetLogger()
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return logger.With(args...)
}

// 便利函数
func Debug(msg string, args ...interface{}) {
	GetLogger().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetLogger().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetLogger().Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...interface{}) {
	WithContext(ctx).Debug(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...interface{}) {
	WithContext(ctx).Info(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	WithContext(ctx).Warn(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	WithContext(ctx).Error(msg, args...)
}

// LogError 记录错误日志，如果err不为nil
func LogError(err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		GetLogger().Error(msg, allArgs...)
	}
}

// LogErrorContext 带上下文记录错误日志
func LogErrorContext(ctx context.Context, err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		WithContext(ctx).Error(msg, allArgs...)
	}
}
