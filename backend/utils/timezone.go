package utils

import (
	"fmt"
	"time"
)

// UTC 强制使用UTC时区
var UTC = time.UTC

// Now 返回当前UTC时间
func Now() time.Time {
	return time.Now().UTC()
}

// ParseTime 解析时间字符串为UTC时间
// 支持多种常见格式，包括带时区信息的ISO格式
func ParseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// 尝试解析常见的时间格式
	formats := []string{
		time.RFC3339,               // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,           // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02T15:04:05",      // "2006-01-02T15:04:05"
		"2006-01-02T15:04:05.000Z", // "2006-01-02T15:04:05.000Z"
		"2006-01-02 15:04:05",      // "2006-01-02 15:04:05"
		"2006-01-02",               // "2006-01-02"
	}

	var parsedTime time.Time
	var err error

	for _, format := range formats {
		parsedTime, err = time.Parse(format, timeStr)
		if err == nil {
			// 如果解析成功但没有时区信息，视为UTC时间
			if parsedTime.Location() == time.UTC || parsedTime.Location().String() == "UTC" {
				return parsedTime.UTC(), nil
			}
			// 如果有时区信息，转换为UTC
			return parsedTime.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string '%s': %v", timeStr, err)
}

// ParseDateRange 解析日期范围字符串为UTC时间
// startTime: 开始日期字符串，如 "2024-01-01"
// endTime: 结束日期字符串，如 "2024-01-31"
// 返回开始时间（当天00:00:00 UTC）和结束时间（当天23:59:59 UTC）
func ParseDateRange(startTime, endTime string) (*time.Time, *time.Time, error) {
	var start, end *time.Time

	if startTime != "" {
		t, err := ParseTime(startTime + " 00:00:00")
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start time: %v", err)
		}
		start = &t
	}

	if endTime != "" {
		t, err := ParseTime(endTime + " 23:59:59")
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end time: %v", err)
		}
		end = &t
	}

	return start, end, nil
}

// FormatTime 格式化时间为UTC的ISO格式字符串
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ToUTC 确保时间为UTC时区
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// TimePtr 返回时间指针，确保为UTC时区
func TimePtr(t time.Time) *time.Time {
	utc := t.UTC()
	return &utc
}

// NowPtr 返回当前UTC时间指针
func NowPtr() *time.Time {
	now := Now()
	return &now
}

// ParseTimeCompatible 兼容解析时间字符串，支持多种格式
// 优先尝试解析完整的ISO格式，如果失败则尝试简单日期格式
// isEndOfDay: 如果为true且输入是简单日期格式，则设置为当天23:59:59，否则为00:00:00
func ParseTimeCompatible(timeStr string, isEndOfDay bool) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// 优先尝试使用完整的时间解析函数
	if parsed, err := ParseTime(timeStr); err == nil {
		return ToUTC(parsed), nil
	}

	// 如果失败，尝试解析简单日期格式 "YYYY-MM-DD"
	if parsed, err := time.Parse("2006-01-02", timeStr); err == nil {
		if isEndOfDay {
			// 设置为当天结束时间 23:59:59 UTC
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			return ToUTC(endOfDay), nil
		} else {
			// 设置为当天开始时间 00:00:00 UTC
			return ToUTC(parsed), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string '%s'", timeStr)
}

// ParseStartTimeCompatible 解析开始时间，兼容多种格式
func ParseStartTimeCompatible(timeStr string) (time.Time, error) {
	return ParseTimeCompatible(timeStr, false)
}

// ParseEndTimeCompatible 解析结束时间，兼容多种格式
func ParseEndTimeCompatible(timeStr string) (time.Time, error) {
	return ParseTimeCompatible(timeStr, true)
}
