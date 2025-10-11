package utils

import (
	"fmt"
	"strings"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	return string(l)
}

// IsValid checks if the log level is valid
func (l LogLevel) IsValid() bool {
	switch l {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return true
	default:
		return false
	}
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (l *LogLevel) UnmarshalText(text []byte) error {
	level := LogLevel(strings.ToUpper(string(text)))
	if !level.IsValid() {
		return fmt.Errorf("invalid log level: %s", text)
	}
	*l = level
	return nil
}

// LogFormat represents the logging format
type LogFormat string

const (
	FormatJSON LogFormat = "JSON"
	FormatText LogFormat = "TEXT"
)

// String returns the string representation of LogFormat
func (f LogFormat) String() string {
	return string(f)
}

// IsValid checks if the log format is valid
func (f LogFormat) IsValid() bool {
	switch f {
	case FormatJSON, FormatText:
		return true
	default:
		return false
	}
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (f *LogFormat) UnmarshalText(text []byte) error {
	format := LogFormat(strings.ToUpper(string(text)))
	if !format.IsValid() {
		return fmt.Errorf("invalid log format: %s", text)
	}
	*f = format
	return nil
}
