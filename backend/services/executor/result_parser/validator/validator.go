package validator

import (
	"fmt"
)

// Validator 验证器接口
type Validator interface {
	// Validate 验证数据
	Validate(data map[string]interface{}) error

	// ValidatePartial 部分验证（允许缺少某些字段）
	ValidatePartial(data map[string]interface{}) error

	// GetValidationErrors 获取详细的验证错误
	GetValidationErrors(data map[string]interface{}) []ValidationError

	// IsValid 快速检查是否有效
	IsValid(data map[string]interface{}) bool
}

// ValidationError 验证错误
type ValidationError struct {
	Field    string      `json:"field"`
	Value    interface{} `json:"value"`
	Expected string      `json:"expected"`
	Message  string      `json:"message"`
	Code     string      `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationRule 验证规则接口
type ValidationRule interface {
	// Name 规则名称
	Name() string

	// Validate 执行验证
	Validate(field string, value interface{}) *ValidationError

	// IsApplicable 检查规则是否适用于给定字段
	IsApplicable(field string) bool
}

// ResultValidator 结果验证器
type ResultValidator struct {
	rules          []ValidationRule
	requiredFields []string
	optionalFields []string
	strictMode     bool
}

// NewResultValidator 创建结果验证器
func NewResultValidator(strictMode bool) *ResultValidator {
	validator := &ResultValidator{
		strictMode: strictMode,
		requiredFields: []string{
			"type",
			"subtype",
			"is_error",
			"session_id",
		},
		optionalFields: []string{
			"duration_ms",
			"duration_api_ms",
			"num_turns",
			"result",
			"total_cost_usd",
			"usage",
		},
	}

	// 添加默认验证规则
	validator.addDefaultRules()

	return validator
}

// addDefaultRules 添加默认验证规则
func (v *ResultValidator) addDefaultRules() {
	v.rules = []ValidationRule{
		&RequiredRule{},
		&TypeRule{},
		&StringRule{},
		&BoolRule{},
		&NumberRule{},
		&RangeRule{},
		&EnumRule{},
	}
}

// AddRule 添加验证规则
func (v *ResultValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// SetRequiredFields 设置必需字段
func (v *ResultValidator) SetRequiredFields(fields []string) {
	v.requiredFields = fields
}

// SetOptionalFields 设置可选字段
func (v *ResultValidator) SetOptionalFields(fields []string) {
	v.optionalFields = fields
}

// Validate 验证数据
func (v *ResultValidator) Validate(data map[string]interface{}) error {
	errors := v.GetValidationErrors(data)
	if len(errors) == 0 {
		return nil
	}

	// 返回第一个错误
	return errors[0]
}

// ValidatePartial 部分验证
func (v *ResultValidator) ValidatePartial(data map[string]interface{}) error {
	errors := v.getValidationErrors(data, true)
	if len(errors) == 0 {
		return nil
	}

	return errors[0]
}

// GetValidationErrors 获取所有验证错误
func (v *ResultValidator) GetValidationErrors(data map[string]interface{}) []ValidationError {
	return v.getValidationErrors(data, false)
}

// IsValid 快速检查是否有效
func (v *ResultValidator) IsValid(data map[string]interface{}) bool {
	return len(v.GetValidationErrors(data)) == 0
}

// getValidationErrors 获取验证错误（内部方法）
func (v *ResultValidator) getValidationErrors(data map[string]interface{}, partial bool) []ValidationError {
	var errors []ValidationError

	// 检查必需字段
	if !partial {
		for _, field := range v.requiredFields {
			if _, exists := data[field]; !exists {
				errors = append(errors, ValidationError{
					Field:    field,
					Value:    nil,
					Expected: "required field",
					Message:  fmt.Sprintf("required field '%s' is missing", field),
					Code:     "FIELD_REQUIRED",
				})
			}
		}
	}

	// 验证存在的字段
	for field, value := range data {
		// 检查字段是否被允许
		if v.strictMode && !v.isAllowedField(field) {
			errors = append(errors, ValidationError{
				Field:    field,
				Value:    value,
				Expected: "allowed field",
				Message:  fmt.Sprintf("field '%s' is not allowed", field),
				Code:     "FIELD_NOT_ALLOWED",
			})
			continue
		}

		// 应用验证规则
		for _, rule := range v.rules {
			if rule.IsApplicable(field) {
				if err := rule.Validate(field, value); err != nil {
					errors = append(errors, *err)
				}
			}
		}
	}

	return errors
}

// isAllowedField 检查字段是否被允许
func (v *ResultValidator) isAllowedField(field string) bool {
	for _, allowedField := range v.requiredFields {
		if field == allowedField {
			return true
		}
	}

	for _, allowedField := range v.optionalFields {
		if field == allowedField {
			return true
		}
	}

	return false
}

// ChainValidator 链式验证器
type ChainValidator struct {
	validators  []Validator
	stopOnFirst bool
}

// NewChainValidator 创建链式验证器
func NewChainValidator(stopOnFirst bool, validators ...Validator) *ChainValidator {
	return &ChainValidator{
		validators:  validators,
		stopOnFirst: stopOnFirst,
	}
}

// AddValidator 添加验证器
func (c *ChainValidator) AddValidator(validator Validator) {
	c.validators = append(c.validators, validator)
}

// Validate 执行链式验证
func (c *ChainValidator) Validate(data map[string]interface{}) error {
	for _, validator := range c.validators {
		if err := validator.Validate(data); err != nil {
			if c.stopOnFirst {
				return err
			}
		}
	}

	return nil
}

// ValidatePartial 执行链式部分验证
func (c *ChainValidator) ValidatePartial(data map[string]interface{}) error {
	for _, validator := range c.validators {
		if err := validator.ValidatePartial(data); err != nil {
			if c.stopOnFirst {
				return err
			}
		}
	}

	return nil
}

// GetValidationErrors 获取所有验证错误
func (c *ChainValidator) GetValidationErrors(data map[string]interface{}) []ValidationError {
	var allErrors []ValidationError

	for _, validator := range c.validators {
		errors := validator.GetValidationErrors(data)
		allErrors = append(allErrors, errors...)

		if c.stopOnFirst && len(errors) > 0 {
			break
		}
	}

	return allErrors
}

// IsValid 检查是否有效
func (c *ChainValidator) IsValid(data map[string]interface{}) bool {
	for _, validator := range c.validators {
		if !validator.IsValid(data) {
			return false
		}
	}

	return true
}

// ConditionalValidator 条件验证器
type ConditionalValidator struct {
	condition func(data map[string]interface{}) bool
	validator Validator
}

// NewConditionalValidator 创建条件验证器
func NewConditionalValidator(condition func(data map[string]interface{}) bool, validator Validator) *ConditionalValidator {
	return &ConditionalValidator{
		condition: condition,
		validator: validator,
	}
}

// Validate 条件验证
func (c *ConditionalValidator) Validate(data map[string]interface{}) error {
	if !c.condition(data) {
		return nil
	}

	return c.validator.Validate(data)
}

// ValidatePartial 条件部分验证
func (c *ConditionalValidator) ValidatePartial(data map[string]interface{}) error {
	if !c.condition(data) {
		return nil
	}

	return c.validator.ValidatePartial(data)
}

// GetValidationErrors 获取验证错误
func (c *ConditionalValidator) GetValidationErrors(data map[string]interface{}) []ValidationError {
	if !c.condition(data) {
		return []ValidationError{}
	}

	return c.validator.GetValidationErrors(data)
}

// IsValid 检查是否有效
func (c *ConditionalValidator) IsValid(data map[string]interface{}) bool {
	if !c.condition(data) {
		return true
	}

	return c.validator.IsValid(data)
}

// ValidationContext 验证上下文
type ValidationContext struct {
	StrictMode    bool
	AllowPartial  bool
	CustomRules   map[string]ValidationRule
	FieldAliases  map[string]string
	DefaultValues map[string]interface{}
}

// ContextualValidator 上下文验证器
type ContextualValidator struct {
	baseValidator Validator
	context       *ValidationContext
}

// NewContextualValidator 创建上下文验证器
func NewContextualValidator(baseValidator Validator, context *ValidationContext) *ContextualValidator {
	return &ContextualValidator{
		baseValidator: baseValidator,
		context:       context,
	}
}

// Validate 上下文验证
func (c *ContextualValidator) Validate(data map[string]interface{}) error {
	processedData := c.preprocessData(data)

	if c.context.AllowPartial {
		return c.baseValidator.ValidatePartial(processedData)
	}

	return c.baseValidator.Validate(processedData)
}

// ValidatePartial 上下文部分验证
func (c *ContextualValidator) ValidatePartial(data map[string]interface{}) error {
	processedData := c.preprocessData(data)
	return c.baseValidator.ValidatePartial(processedData)
}

// GetValidationErrors 获取验证错误
func (c *ContextualValidator) GetValidationErrors(data map[string]interface{}) []ValidationError {
	processedData := c.preprocessData(data)
	return c.baseValidator.GetValidationErrors(processedData)
}

// IsValid 检查是否有效
func (c *ContextualValidator) IsValid(data map[string]interface{}) bool {
	processedData := c.preprocessData(data)
	return c.baseValidator.IsValid(processedData)
}

// preprocessData 预处理数据
func (c *ContextualValidator) preprocessData(data map[string]interface{}) map[string]interface{} {
	if c.context == nil {
		return data
	}

	result := make(map[string]interface{})

	// 复制原始数据
	for k, v := range data {
		result[k] = v
	}

	// 应用字段别名
	if c.context.FieldAliases != nil {
		for alias, realField := range c.context.FieldAliases {
			if value, exists := result[alias]; exists {
				result[realField] = value
				delete(result, alias)
			}
		}
	}

	// 应用默认值
	if c.context.DefaultValues != nil {
		for field, defaultValue := range c.context.DefaultValues {
			if _, exists := result[field]; !exists {
				result[field] = defaultValue
			}
		}
	}

	return result
}
