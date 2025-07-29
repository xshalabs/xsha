package utils

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel log level type
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogFormat log format type
type LogFormat string

const (
	FormatJSON LogFormat = "JSON"
	FormatText LogFormat = "TEXT"
)

// LogConfig log configuration
type LogConfig struct {
	Level  LogLevel  `json:"level" env:"LOG_LEVEL" default:"INFO"`
	Format LogFormat `json:"format" env:"LOG_FORMAT" default:"JSON"`
	Output string    `json:"output" env:"LOG_OUTPUT" default:"stdout"` // stdout, stderr, file path
}

var defaultLogger *slog.Logger

// InitLogger initializes the global logger
func InitLogger(config LogConfig) error {
	// Parse log level
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

	// Configure options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false, // We manually add source location information
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Custom time format
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format("2006-01-02T15:04:05.000Z07:00"))
			}
			return a
		},
	}

	// Determine output target
	var output *os.File
	switch config.Output {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// File output
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}

	// Create Handler
	var handler slog.Handler
	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	case FormatText:
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	// Set global logger
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	return nil
}

// GetLogger gets the default logger
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		// If not initialized, use default configuration
		InitLogger(LogConfig{
			Level:  LevelInfo,
			Format: FormatJSON,
			Output: "stdout",
		})
	}
	return defaultLogger
}

// WithContext creates a logger with context
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

// WithFields creates a logger with fields
func WithFields(fields map[string]interface{}) *slog.Logger {
	logger := GetLogger()
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return logger.With(args...)
}

// 便利函数 - 修改以正确获取调用者信息
func Debug(msg string, args ...interface{}) {
	logWithCaller(slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...interface{}) {
	logWithCaller(slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...interface{}) {
	logWithCaller(slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...interface{}) {
	logWithCaller(slog.LevelError, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, slog.LevelDebug, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, slog.LevelInfo, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, slog.LevelWarn, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, slog.LevelError, msg, args...)
}

// logWithCaller 记录日志并正确获取调用者信息
func logWithCaller(level slog.Level, msg string, args ...interface{}) {
	logger := GetLogger()
	if !logger.Enabled(context.Background(), level) {
		return
	}

	var pc uintptr
	var file string
	var line int
	var function string

	// 获取调用者信息，跳过2层：logWithCaller -> 便利函数 -> 实际调用者
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			function = fn.Name()
		}
		file = filepath.Base(file)
	}

	// 创建带源码信息的记录
	record := slog.NewRecord(time.Now(), level, msg, pc)

	// 转换args为slog.Attr格式
	if len(args) > 0 {
		attrs := make([]slog.Attr, 0, len(args)/2)
		for i := 0; i < len(args)-1; i += 2 {
			if key, ok := args[i].(string); ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
		record.AddAttrs(attrs...)
	}

	// 手动设置源码信息
	if ok {
		record.AddAttrs(slog.Group("source",
			slog.String("function", function),
			slog.String("file", file),
			slog.Int("line", line),
		))
	}

	logger.Handler().Handle(context.Background(), record)
}

// logWithCallerContext 带上下文记录日志并正确获取调用者信息
func logWithCallerContext(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	logger := WithContext(ctx)
	if !logger.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var file string
	var line int
	var function string

	// 获取调用者信息，跳过2层：logWithCallerContext -> 便利函数 -> 实际调用者
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			function = fn.Name()
		}
		file = filepath.Base(file)
	}

	// 创建带源码信息的记录
	record := slog.NewRecord(time.Now(), level, msg, pc)

	// 转换args为slog.Attr格式
	if len(args) > 0 {
		attrs := make([]slog.Attr, 0, len(args)/2)
		for i := 0; i < len(args)-1; i += 2 {
			if key, ok := args[i].(string); ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
		record.AddAttrs(attrs...)
	}

	// 手动设置源码信息
	if ok {
		record.AddAttrs(slog.Group("source",
			slog.String("function", function),
			slog.String("file", file),
			slog.Int("line", line),
		))
	}

	logger.Handler().Handle(ctx, record)
}

// LogError 记录错误日志，如果err不为nil
func LogError(err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		logWithCaller(slog.LevelError, msg, allArgs...)
	}
}

// LogErrorContext 带上下文记录错误日志
func LogErrorContext(ctx context.Context, err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		logWithCallerContext(ctx, slog.LevelError, msg, allArgs...)
	}
}
