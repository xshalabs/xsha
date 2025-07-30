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
	"项目未关联Git凭据，请先关联Git Credential后再推送":                  "task.no_git_credential",

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
	"invalid repository URL format":                                                  "validation.invalid_format",
	"failed to get credential":                                                       "git_credential.use_failed",
	"failed to decrypt credential":                                                   "git_credential.use_failed",
	"failed to decrypt SSH private key":                                              "git_credential.use_failed",

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
	"failed to check conversation status":                                             "taskConversation.get_failed",
	"cannot create new conversation while there are pending or running conversations": "taskConversation.create_failed",
	"cannot create conversation for completed or cancelled task":                      "taskConversation.task_completed",
	"conversation content cannot be empty":                                            "validation.required",
	"cannot delete conversation while it is running":                                  "taskConversation.delete_failed",
	"failed to get latest conversation":                                               "taskConversation.get_failed",
	"only the latest conversation can be deleted":                                     "taskConversation.delete_latest_only",
	"failed to delete related execution logs":                                         "taskConversation.delete_failed",
	"conversation content is required":                                                "validation.required",
	"conversation content too long":                                                   "validation.too_long",
	"task ID is required":                                                             "validation.required",

	// Task execution related errors
	"task has no development environment configured, cannot execute": "task_execution.no_dev_environment",
	"failed to update conversation status to cancelled":              "task_execution.update_status_failed",
	"failed to update conversation status":                           "task_execution.update_status_failed",
	"failed to create execution log":                                 "task_execution.create_log_failed",
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
