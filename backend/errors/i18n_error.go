package errors

import "fmt"

type I18nError struct {
	Key     string
	Details string
	Params  map[string]interface{}
}

func (e *I18nError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Key, e.Details)
	}
	return e.Key
}

func NewI18nError(key string, details ...string) *I18nError {
	err := &I18nError{Key: key}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

func NewI18nErrorWithParams(key string, params map[string]interface{}, details ...string) *I18nError {
	err := &I18nError{Key: key, Params: params}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

var (
	ErrRequired      = &I18nError{Key: "validation.required"}
	ErrInvalidFormat = &I18nError{Key: "validation.invalid_format"}
	ErrTooLong       = &I18nError{Key: "validation.too_long"}

	ErrTaskTitleRequired   = &I18nError{Key: "task.title_required"}
	ErrTaskTitleTooLong    = &I18nError{Key: "task.title_too_long"}
	ErrStartBranchRequired = &I18nError{Key: "task.start_branch_required"}
	ErrProjectIDRequired   = &I18nError{Key: "task.project_id_required"}
	ErrProjectNotFound     = &I18nError{Key: "task.project_not_found"}
	ErrTaskNotFound        = &I18nError{Key: "task.not_found"}
	ErrNoGitCredential     = &I18nError{Key: "task.no_git_credential"}

	ErrProjectNameExists      = &I18nError{Key: "project.name_exists"}
	ErrIncompatibleCredential = &I18nError{Key: "project.incompatible_credential"}
	ErrInvalidProtocol        = &I18nError{Key: "project.invalid_protocol"}

	ErrCredentialNameExists  = &I18nError{Key: "git_credential.name_exists"}
	ErrCredentialUseFailed   = &I18nError{Key: "git_credential.use_failed"}
	ErrInvalidCredentialType = &I18nError{Key: "git_credential.invalid_type"}

	ErrEnvironmentCreateFailed = &I18nError{Key: "dev_environment.create_failed"}
	ErrDevEnvironmentNotFound  = &I18nError{Key: "dev_environment.not_found"}

	ErrConversationGetFailed        = &I18nError{Key: "taskConversation.get_failed"}
	ErrConversationCreateFailed     = &I18nError{Key: "taskConversation.create_failed"}
	ErrConversationTaskCompleted    = &I18nError{Key: "taskConversation.task_completed"}
	ErrConversationDeleteFailed     = &I18nError{Key: "taskConversation.delete_failed"}
	ErrConversationDeleteLatestOnly = &I18nError{Key: "taskConversation.delete_latest_only"}

	ErrNoDevEnvironment   = &I18nError{Key: "task_execution.no_dev_environment"}
	ErrUpdateStatusFailed = &I18nError{Key: "task_execution.update_status_failed"}

	ErrProjectHasInProgressTasks = &I18nError{Key: "project.delete_has_in_progress_tasks"}
	ErrCredentialUsedByProjects  = &I18nError{Key: "git_credential.delete_used_by_projects"}
	ErrEnvironmentUsedByTasks    = &I18nError{Key: "dev_environment.delete_used_by_tasks"}

	ErrTaskIDsEmpty         = &I18nError{Key: "validation.required"}
	ErrTooManyTasksForBatch = &I18nError{Key: "validation.too_many"}

	ErrFilePathEmpty      = &I18nError{Key: "validation.required"}
	ErrWorkspacePathEmpty = &I18nError{Key: "task.workspace_path_empty"}
	ErrNoCommitHash       = &I18nError{Key: "taskConversation.no_commit_hash"}
)
