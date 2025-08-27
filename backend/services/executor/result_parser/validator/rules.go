package validator

import (
	"fmt"
	"strings"
)

// RequiredRule 必需字段验证规则
type RequiredRule struct{}

func (r *RequiredRule) Name() string {
	return "required"
}

func (r *RequiredRule) IsApplicable(field string) bool {
	requiredFields := []string{"type", "subtype", "is_error", "session_id"}
	for _, required := range requiredFields {
		if field == required {
			return true
		}
	}
	return false
}

func (r *RequiredRule) Validate(field string, value interface{}) *ValidationError {
	if value == nil {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "non-null value",
			Message:  fmt.Sprintf("required field '%s' cannot be null", field),
			Code:     "FIELD_REQUIRED",
		}
	}
	
	if str, ok := value.(string); ok && str == "" {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "non-empty string",
			Message:  fmt.Sprintf("required field '%s' cannot be empty", field),
			Code:     "FIELD_EMPTY",
		}
	}
	
	return nil
}

// TypeRule 类型验证规则
type TypeRule struct{}

func (r *TypeRule) Name() string {
	return "type"
}

func (r *TypeRule) IsApplicable(field string) bool {
	return field == "type"
}

func (r *TypeRule) Validate(field string, value interface{}) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "string",
			Message:  "type field must be a string",
			Code:     "INVALID_TYPE",
		}
	}
	
	if str != "result" {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "result",
			Message:  "type field must be 'result'",
			Code:     "INVALID_VALUE",
		}
	}
	
	return nil
}

// StringRule 字符串验证规则
type StringRule struct{}

func (r *StringRule) Name() string {
	return "string"
}

func (r *StringRule) IsApplicable(field string) bool {
	stringFields := []string{"type", "subtype", "session_id", "result", "usage"}
	for _, strField := range stringFields {
		if field == strField {
			return true
		}
	}
	return false
}

func (r *StringRule) Validate(field string, value interface{}) *ValidationError {
	if _, ok := value.(string); !ok {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "string",
			Message:  fmt.Sprintf("field '%s' must be a string", field),
			Code:     "INVALID_TYPE",
		}
	}
	
	return nil
}

// BoolRule 布尔值验证规则
type BoolRule struct{}

func (r *BoolRule) Name() string {
	return "bool"
}

func (r *BoolRule) IsApplicable(field string) bool {
	return field == "is_error"
}

func (r *BoolRule) Validate(field string, value interface{}) *ValidationError {
	if _, ok := value.(bool); !ok {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "boolean",
			Message:  fmt.Sprintf("field '%s' must be a boolean", field),
			Code:     "INVALID_TYPE",
		}
	}
	
	return nil
}

// NumberRule 数字验证规则
type NumberRule struct{}

func (r *NumberRule) Name() string {
	return "number"
}

func (r *NumberRule) IsApplicable(field string) bool {
	numberFields := []string{"duration_ms", "duration_api_ms", "num_turns", "total_cost_usd"}
	for _, numField := range numberFields {
		if field == numField {
			return true
		}
	}
	return false
}

func (r *NumberRule) Validate(field string, value interface{}) *ValidationError {
	switch field {
	case "duration_ms", "duration_api_ms":
		return r.validateInt64(field, value)
	case "num_turns":
		return r.validateInt(field, value)
	case "total_cost_usd":
		return r.validateFloat(field, value)
	}
	
	return nil
}

func (r *NumberRule) validateInt64(field string, value interface{}) *ValidationError {
	switch v := value.(type) {
	case int64:
		return nil
	case int:
		return nil
	case float64:
		if v == float64(int64(v)) {
			return nil
		}
	}
	
	return &ValidationError{
		Field:    field,
		Value:    value,
		Expected: "integer",
		Message:  fmt.Sprintf("field '%s' must be an integer", field),
		Code:     "INVALID_TYPE",
	}
}

func (r *NumberRule) validateInt(field string, value interface{}) *ValidationError {
	switch v := value.(type) {
	case int:
		return nil
	case int64:
		if v <= int64(int(^uint(0)>>1)) && v >= int64(-int(^uint(0)>>1)-1) {
			return nil
		}
	case float64:
		if v == float64(int(v)) && v <= float64(int(^uint(0)>>1)) && v >= float64(-int(^uint(0)>>1)-1) {
			return nil
		}
	}
	
	return &ValidationError{
		Field:    field,
		Value:    value,
		Expected: "integer",
		Message:  fmt.Sprintf("field '%s' must be an integer", field),
		Code:     "INVALID_TYPE",
	}
}

func (r *NumberRule) validateFloat(field string, value interface{}) *ValidationError {
	switch value.(type) {
	case float64, float32:
		return nil
	case int, int64:
		return nil
	}
	
	return &ValidationError{
		Field:    field,
		Value:    value,
		Expected: "number",
		Message:  fmt.Sprintf("field '%s' must be a number", field),
		Code:     "INVALID_TYPE",
	}
}

// RangeRule 范围验证规则
type RangeRule struct{}

func (r *RangeRule) Name() string {
	return "range"
}

func (r *RangeRule) IsApplicable(field string) bool {
	rangeFields := []string{"duration_ms", "duration_api_ms", "num_turns", "total_cost_usd"}
	for _, rangeField := range rangeFields {
		if field == rangeField {
			return true
		}
	}
	return false
}

func (r *RangeRule) Validate(field string, value interface{}) *ValidationError {
	switch field {
	case "duration_ms", "duration_api_ms":
		return r.validateDurationRange(field, value)
	case "num_turns":
		return r.validateTurnsRange(field, value)
	case "total_cost_usd":
		return r.validateCostRange(field, value)
	}
	
	return nil
}

func (r *RangeRule) validateDurationRange(field string, value interface{}) *ValidationError {
	var num int64
	
	switch v := value.(type) {
	case int64:
		num = v
	case int:
		num = int64(v)
	case float64:
		num = int64(v)
	default:
		return nil // 类型验证由其他规则处理
	}
	
	if num < 0 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: ">= 0",
			Message:  fmt.Sprintf("field '%s' must be non-negative", field),
			Code:     "INVALID_RANGE",
		}
	}
	
	if num > 24*60*60*1000 { // 24小时
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "<= 86400000",
			Message:  fmt.Sprintf("field '%s' exceeds maximum duration (24 hours)", field),
			Code:     "INVALID_RANGE",
		}
	}
	
	return nil
}

func (r *RangeRule) validateTurnsRange(field string, value interface{}) *ValidationError {
	var num int
	
	switch v := value.(type) {
	case int:
		num = v
	case int64:
		num = int(v)
	case float64:
		num = int(v)
	default:
		return nil
	}
	
	if num < 1 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: ">= 1",
			Message:  fmt.Sprintf("field '%s' must be at least 1", field),
			Code:     "INVALID_RANGE",
		}
	}
	
	if num > 1000 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "<= 1000",
			Message:  fmt.Sprintf("field '%s' exceeds maximum turns (1000)", field),
			Code:     "INVALID_RANGE",
		}
	}
	
	return nil
}

func (r *RangeRule) validateCostRange(field string, value interface{}) *ValidationError {
	var num float64
	
	switch v := value.(type) {
	case float64:
		num = v
	case float32:
		num = float64(v)
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	default:
		return nil
	}
	
	if num < 0 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: ">= 0",
			Message:  fmt.Sprintf("field '%s' must be non-negative", field),
			Code:     "INVALID_RANGE",
		}
	}
	
	if num > 10000 { // $10,000 上限
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "<= 10000",
			Message:  fmt.Sprintf("field '%s' exceeds maximum cost ($10,000)", field),
			Code:     "INVALID_RANGE",
		}
	}
	
	return nil
}

// EnumRule 枚举验证规则
type EnumRule struct{}

func (r *EnumRule) Name() string {
	return "enum"
}

func (r *EnumRule) IsApplicable(field string) bool {
	return field == "subtype"
}

func (r *EnumRule) Validate(field string, value interface{}) *ValidationError {
	if field == "subtype" {
		return r.validateSubtype(field, value)
	}
	
	return nil
}

func (r *EnumRule) validateSubtype(field string, value interface{}) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return nil // 类型验证由其他规则处理
	}
	
	validSubtypes := []string{"success", "error", "timeout", "cancelled", "fallback"}
	for _, valid := range validSubtypes {
		if str == valid {
			return nil
		}
	}
	
	return &ValidationError{
		Field:    field,
		Value:    value,
		Expected: strings.Join(validSubtypes, ", "),
		Message:  fmt.Sprintf("field '%s' must be one of: %s", field, strings.Join(validSubtypes, ", ")),
		Code:     "INVALID_ENUM",
	}
}

// LengthRule 长度验证规则
type LengthRule struct{}

func (r *LengthRule) Name() string {
	return "length"
}

func (r *LengthRule) IsApplicable(field string) bool {
	lengthFields := []string{"session_id", "result", "usage"}
	for _, lengthField := range lengthFields {
		if field == lengthField {
			return true
		}
	}
	return false
}

func (r *LengthRule) Validate(field string, value interface{}) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return nil // 类型验证由其他规则处理
	}
	
	switch field {
	case "session_id":
		return r.validateSessionIDLength(field, str)
	case "result":
		return r.validateResultLength(field, str)
	case "usage":
		return r.validateUsageLength(field, str)
	}
	
	return nil
}

func (r *LengthRule) validateSessionIDLength(field, value string) *ValidationError {
	if len(value) < 1 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "length >= 1",
			Message:  "session_id cannot be empty",
			Code:     "INVALID_LENGTH",
		}
	}
	
	if len(value) > 100 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "length <= 100",
			Message:  "session_id is too long (max 100 characters)",
			Code:     "INVALID_LENGTH",
		}
	}
	
	return nil
}

func (r *LengthRule) validateResultLength(field, value string) *ValidationError {
	if len(value) > 1000000 { // 1MB
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "length <= 1000000",
			Message:  "result content is too long (max 1MB)",
			Code:     "INVALID_LENGTH",
		}
	}
	
	return nil
}

func (r *LengthRule) validateUsageLength(field, value string) *ValidationError {
	if len(value) > 10000 {
		return &ValidationError{
			Field:    field,
			Value:    value,
			Expected: "length <= 10000",
			Message:  "usage data is too long (max 10,000 characters)",
			Code:     "INVALID_LENGTH",
		}
	}
	
	return nil
}

// FormatRule 格式验证规则
type FormatRule struct{}

func (r *FormatRule) Name() string {
	return "format"
}

func (r *FormatRule) IsApplicable(field string) bool {
	formatFields := []string{"session_id"}
	for _, formatField := range formatFields {
		if field == formatField {
			return true
		}
	}
	return false
}

func (r *FormatRule) Validate(field string, value interface{}) *ValidationError {
	if field == "session_id" {
		return r.validateSessionIDFormat(field, value)
	}
	
	return nil
}

func (r *FormatRule) validateSessionIDFormat(field string, value interface{}) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return nil // 类型验证由其他规则处理
	}
	
	// 检查是否包含有效字符（字母、数字、连字符、下划线）
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return &ValidationError{
				Field:    field,
				Value:    value,
				Expected: "alphanumeric characters, hyphens, or underscores",
				Message:  "session_id contains invalid characters",
				Code:     "INVALID_FORMAT",
			}
		}
	}
	
	return nil
}

// CustomRule 自定义验证规则
type CustomRule struct {
	name       string
	fields     []string
	validator  func(field string, value interface{}) *ValidationError
}

// NewCustomRule 创建自定义规则
func NewCustomRule(name string, fields []string, validator func(field string, value interface{}) *ValidationError) *CustomRule {
	return &CustomRule{
		name:      name,
		fields:    fields,
		validator: validator,
	}
}

func (r *CustomRule) Name() string {
	return r.name
}

func (r *CustomRule) IsApplicable(field string) bool {
	for _, f := range r.fields {
		if field == f {
			return true
		}
	}
	return false
}

func (r *CustomRule) Validate(field string, value interface{}) *ValidationError {
	if r.validator != nil {
		return r.validator(field, value)
	}
	return nil
}