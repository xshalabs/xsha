package utils

import (
	"context"
	"os"
	"strings"
	"time"
	"xsha-backend/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

var (
	defaultLogger *zap.Logger
	sugar         *zap.SugaredLogger
)

func InitLogger(config LogConfig) error {
	var level zapcore.Level
	switch strings.ToUpper(string(config.Level)) {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Format == FormatJSON {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// Customize time format
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
	}

	// Configure output
	var writer zapcore.WriteSyncer
	switch config.Output {
	case "stdout", "":
		writer = zapcore.AddSync(os.Stdout)
	case "stderr":
		writer = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		writer = zapcore.AddSync(file)
	}

	// Create encoder
	var encoder zapcore.Encoder
	if config.Format == FormatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core and logger
	core := zapcore.NewCore(encoder, writer, level)
	defaultLogger = zap.New(core)
	sugar = defaultLogger.Sugar()

	return nil
}

func GetLogger() *zap.Logger {
	if defaultLogger == nil {
		InitLogger(LogConfig{
			Level:  LevelInfo,
			Format: FormatJSON,
			Output: "stdout",
		})
	}
	return defaultLogger
}

func GetSugaredLogger() *zap.SugaredLogger {
	if sugar == nil {
		GetLogger() // This will initialize both
	}
	return sugar
}

func WithContext(ctx context.Context) *zap.Logger {
	logger := GetLogger()

	fields := make([]zap.Field, 0, 3)
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.Any("trace_id", traceID))
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.Any("request_id", requestID))
	}

	return logger.With(fields...)
}

func WithFields(fields map[string]interface{}) *zap.Logger {
	logger := GetLogger()
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return logger.With(zapFields...)
}

func Debug(msg string, args ...interface{}) {
	logWithCaller(zapcore.DebugLevel, msg, args...)
}

func Info(msg string, args ...interface{}) {
	logWithCaller(zapcore.InfoLevel, msg, args...)
}

func Warn(msg string, args ...interface{}) {
	logWithCaller(zapcore.WarnLevel, msg, args...)
}

func Error(msg string, args ...interface{}) {
	logWithCaller(zapcore.ErrorLevel, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, zapcore.DebugLevel, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, zapcore.InfoLevel, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, zapcore.WarnLevel, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	logWithCallerContext(ctx, zapcore.ErrorLevel, msg, args...)
}

func logWithCaller(level zapcore.Level, msg string, args ...interface{}) {
	logger := GetLogger()
	if !logger.Core().Enabled(level) {
		return
	}

	// Convert args to zap fields
	fields := argsToFields(args)
	logger.Log(level, msg, fields...)
}

func logWithCallerContext(ctx context.Context, level zapcore.Level, msg string, args ...interface{}) {
	logger := WithContext(ctx)
	if !logger.Core().Enabled(level) {
		return
	}

	// Convert args to zap fields
	fields := argsToFields(args)
	logger.Log(level, msg, fields...)
}

func argsToFields(args []interface{}) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			fields = append(fields, zap.Any(key, args[i+1]))
		}
	}
	return fields
}

func LogError(err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		logWithCaller(zapcore.ErrorLevel, msg, allArgs...)
	}
}

func LogErrorContext(ctx context.Context, err error, msg string, args ...interface{}) {
	if err != nil {
		allArgs := append([]interface{}{"error", err.Error()}, args...)
		logWithCallerContext(ctx, zapcore.ErrorLevel, msg, allArgs...)
	}
}

// Additional Zap-specific functions for enhanced functionality
func DPanic(msg string, fields ...zap.Field) {
	GetLogger().DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

func Sync() error {
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}
