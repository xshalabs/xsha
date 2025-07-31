package utils

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"xsha-backend/config"
)

type LogLevel = config.LogLevel
type LogFormat = config.LogFormat

const (
	LevelDebug = config.LevelDebug
	LevelInfo  = config.LevelInfo
	LevelWarn  = config.LevelWarn
	LevelError = config.LevelError
)

const (
	FormatJSON = config.FormatJSON
	FormatText = config.FormatText
)

type LogConfig struct {
	Level  LogLevel  `json:"level" env:"LOG_LEVEL" default:"INFO"`
	Format LogFormat `json:"format" env:"LOG_FORMAT" default:"JSON"`
	Output string    `json:"output" env:"LOG_OUTPUT" default:"stdout"`
}

var defaultLogger *slog.Logger

func InitLogger(config LogConfig) error {
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

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format("2006-01-02T15:04:05.000Z07:00"))
			}
			return a
		},
	}

	var output *os.File
	switch config.Output {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}

	var handler slog.Handler
	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	case FormatText:
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	return nil
}

func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		InitLogger(LogConfig{
			Level:  LevelInfo,
			Format: FormatJSON,
			Output: "stdout",
		})
	}
	return defaultLogger
}

func WithContext(ctx context.Context) *slog.Logger {
	logger := GetLogger()

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

func WithFields(fields map[string]interface{}) *slog.Logger {
	logger := GetLogger()
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return logger.With(args...)
}

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

func logWithCaller(level slog.Level, msg string, args ...interface{}) {
	logger := GetLogger()
	if !logger.Enabled(context.Background(), level) {
		return
	}

	var pc uintptr
	var file string
	var line int
	var function string

	pc, file, line, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			function = fn.Name()
		}
		file = filepath.Base(file)
	}

	record := slog.NewRecord(time.Now(), level, msg, pc)

	if len(args) > 0 {
		attrs := make([]slog.Attr, 0, len(args)/2)
		for i := 0; i < len(args)-1; i += 2 {
			if key, ok := args[i].(string); ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
		record.AddAttrs(attrs...)
	}

	if ok {
		record.AddAttrs(slog.Group("source",
			slog.String("function", function),
			slog.String("file", file),
			slog.Int("line", line),
		))
	}

	logger.Handler().Handle(context.Background(), record)
}

func logWithCallerContext(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	logger := WithContext(ctx)
	if !logger.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var file string
	var line int
	var function string

	pc, file, line, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			function = fn.Name()
		}
		file = filepath.Base(file)
	}

	record := slog.NewRecord(time.Now(), level, msg, pc)

	if len(args) > 0 {
		attrs := make([]slog.Attr, 0, len(args)/2)
		for i := 0; i < len(args)-1; i += 2 {
			if key, ok := args[i].(string); ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
		record.AddAttrs(attrs...)
	}

	if ok {
		record.AddAttrs(slog.Group("source",
			slog.String("function", function),
			slog.String("file", file),
			slog.Int("line", line),
		))
	}

	logger.Handler().Handle(ctx, record)
}

func LogError(err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		logWithCaller(slog.LevelError, msg, allArgs...)
	}
}

func LogErrorContext(ctx context.Context, err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		logWithCallerContext(ctx, slog.LevelError, msg, allArgs...)
	}
}
