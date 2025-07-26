package i18n

import (
	"strings"
)

// ErrorMapping 错误映射表
var ErrorMapping = map[string]string{
	// Task related errors
	"task title is required":                             "task.title_required",
	"task title too long":                                "task.title_too_long",
	"start branch is required":                           "task.start_branch_required",
	"project ID is required":                             "task.project_id_required",
	"project not found or access denied":                 "task.project_not_found",
	"development environment not found or access denied": "dev_environment.not_found",
	"no updates provided":                                "validation.required",
	"invalid title format":                               "validation.invalid_format",
	"task title cannot be empty":                         "task.title_required",

	// Project related errors
	"project name already exists":                                                    "project.name_exists",
	"credential is not active":                                                       "git_credential.use_failed",
	"HTTPS protocol only supports password or token credentials":                     "project.incompatible_credential",
	"SSH protocol only supports SSH key credentials":                                 "project.incompatible_credential",
	"unsupported protocol type":                                                      "project.invalid_protocol",
	"project name is required":                                                       "validation.required",
	"repository URL is required":                                                     "validation.required",
	"HTTPS protocol requires URL to start with 'https://'":                           "validation.invalid_format",
	"SSH protocol requires URL in format 'user@host:path' or 'ssh://user@host/path'": "validation.invalid_format",

	// Development environment related errors
	"environment name already exists":                       "dev_environment.create_failed",
	"environment variable key cannot be empty":              "validation.required",
	"environment variable key cannot contain '=' character": "validation.invalid_format",
	"CPU limit must be between 0 and 16 cores":              "validation.invalid_format",
	"memory limit must be between 0 and 32GB (32768MB)":     "validation.invalid_format",
	"environment name is required":                          "validation.required",
	"unsupported environment type":                          "validation.invalid_format",

	// Git credential related errors
	"credential name already exists":           "git_credential.name_exists",
	"password not set":                         "git_credential.use_failed",
	"private key not set":                      "git_credential.use_failed",
	"unsupported secret type":                  "git_credential.invalid_type",
	"password is required for password type":   "validation.required",
	"token is required for token type":         "validation.required",
	"private key is required for SSH key type": "validation.required",
	"invalid private key format":               "validation.invalid_format",
	"unsupported credential type":              "git_credential.invalid_type",

	// Task conversation related errors
	"task not found or access denied":                                                 "task.not_found",
	"failed to check conversation status":                                             "task_conversation.get_failed",
	"cannot create new conversation while there are pending or running conversations": "task_conversation.create_failed",
	"conversation content cannot be empty":                                            "validation.required",
	"cannot delete conversation while it is running":                                  "task_conversation.create_failed",
	"failed to delete related execution logs":                                         "task_conversation.create_failed",
	"conversation content is required":                                                "validation.required",
	"conversation content too long":                                                   "validation.too_long",
	"task ID is required":                                                             "validation.required",
}

// MapErrorToI18nKey 将错误消息映射到国际化键值
func MapErrorToI18nKey(err error, lang string) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// 直接匹配
	if key, exists := ErrorMapping[errMsg]; exists {
		return T(lang, key)
	}

	// 模糊匹配（处理包含额外信息的错误）
	for errorPattern, key := range ErrorMapping {
		if strings.Contains(errMsg, errorPattern) {
			return T(lang, key)
		}
	}

	// 如果没有找到映射，返回原始错误消息
	return errMsg
}

// MapErrorWithDetails 将错误消息映射到国际化键值，并保留详细信息
func MapErrorWithDetails(err error, lang string) (string, string) {
	if err == nil {
		return "", ""
	}

	errMsg := err.Error()

	// 直接匹配
	if key, exists := ErrorMapping[errMsg]; exists {
		return T(lang, key), ""
	}

	// 模糊匹配（处理包含额外信息的错误）
	for errorPattern, key := range ErrorMapping {
		if strings.Contains(errMsg, errorPattern) {
			// 提取额外的详细信息
			details := strings.TrimSpace(strings.Replace(errMsg, errorPattern, "", 1))
			details = strings.TrimPrefix(details, ":")
			details = strings.TrimSpace(details)
			return T(lang, key), details
		}
	}

	// 如果没有找到映射，返回原始错误消息
	return errMsg, ""
}
