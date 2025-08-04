package errors

import (
	"testing"
)

func TestI18nError(t *testing.T) {
	err := NewI18nError("test.key", "detail info")
	if err.Key != "test.key" {
		t.Errorf("Expected key 'test.key', got '%s'", err.Key)
	}
	if err.Details != "detail info" {
		t.Errorf("Expected details 'detail info', got '%s'", err.Details)
	}

	expected := "test.key: detail info"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestI18nErrorWithoutDetails(t *testing.T) {
	err := NewI18nError("test.key")
	if err.Key != "test.key" {
		t.Errorf("Expected key 'test.key', got '%s'", err.Key)
	}
	if err.Details != "" {
		t.Errorf("Expected empty details, got '%s'", err.Details)
	}

	expected := "test.key"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestI18nErrorWithParams(t *testing.T) {
	params := map[string]interface{}{
		"field": "title",
		"max":   200,
	}
	err := NewI18nErrorWithParams("validation.too_long", params, "extra details")

	if err.Key != "validation.too_long" {
		t.Errorf("Expected key 'validation.too_long', got '%s'", err.Key)
	}
	if err.Details != "extra details" {
		t.Errorf("Expected details 'extra details', got '%s'", err.Details)
	}
	if err.Params == nil || err.Params["field"] != "title" {
		t.Errorf("Expected params with field='title', got %v", err.Params)
	}
}

func TestPredefinedErrors(t *testing.T) {
	testCases := []struct {
		err      *I18nError
		expected string
	}{
		{ErrTaskTitleRequired, "task.title_required"},
		{ErrTaskTitleTooLong, "task.title_too_long"},
		{ErrStartBranchRequired, "task.start_branch_required"},
		{ErrProjectIDRequired, "task.project_id_required"},
		{ErrProjectNotFound, "task.project_not_found"},
		{ErrTaskNotFound, "task.not_found"},
		{ErrNoGitCredential, "task.no_git_credential"},
		{ErrProjectNameExists, "project.name_exists"},
		{ErrIncompatibleCredential, "project.incompatible_credential"},
		{ErrInvalidProtocol, "project.invalid_protocol"},
		{ErrCredentialNameExists, "git_credential.name_exists"},
		{ErrCredentialUseFailed, "git_credential.use_failed"},
		{ErrInvalidCredentialType, "git_credential.invalid_type"},
		{ErrEnvironmentCreateFailed, "dev_environment.create_failed"},
		{ErrDevEnvironmentNotFound, "dev_environment.not_found"},
		{ErrConversationGetFailed, "taskConversation.get_failed"},
		{ErrConversationCreateFailed, "taskConversation.create_failed"},
		{ErrConversationTaskCompleted, "taskConversation.task_completed"},
		{ErrConversationDeleteFailed, "taskConversation.delete_failed"},
		{ErrConversationDeleteLatestOnly, "taskConversation.delete_latest_only"},
		{ErrNoDevEnvironment, "task_execution.no_dev_environment"},
		{ErrUpdateStatusFailed, "task_execution.update_status_failed"},
		{ErrProjectHasInProgressTasks, "project.delete_has_in_progress_tasks"},
		{ErrCredentialUsedByProjects, "git_credential.delete_used_by_projects"},
		{ErrEnvironmentUsedByTasks, "dev_environment.delete_used_by_tasks"},
	}

	for _, tc := range testCases {
		if tc.err.Key != tc.expected {
			t.Errorf("Expected key '%s', got '%s'", tc.expected, tc.err.Key)
		}
	}
}

func TestErrorInterface(t *testing.T) {
	var err error = ErrTaskTitleRequired
	if err.Error() != "task.title_required" {
		t.Errorf("Expected 'task.title_required', got '%s'", err.Error())
	}
}

func TestErrorTypeDetection(t *testing.T) {
	var err error = ErrTaskTitleRequired

	if i18nErr, ok := err.(*I18nError); ok {
		if i18nErr.Key != "task.title_required" {
			t.Errorf("Expected 'task.title_required', got '%s'", i18nErr.Key)
		}
	} else {
		t.Error("Failed to detect I18nError type")
	}
}
