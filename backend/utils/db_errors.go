package utils

import (
	"strings"
	appErrors "xsha-backend/errors"
)

// ParseDBError converts database constraint errors to appropriate i18n errors
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Handle SQLite UNIQUE constraint failures
	if strings.Contains(errMsg, "UNIQUE constraint failed") {
		if strings.Contains(errMsg, "admins.username") {
			return appErrors.ErrAdminUsernameExists
		}
		if strings.Contains(errMsg, "projects.name") {
			return appErrors.ErrProjectNameExists
		}
		if strings.Contains(errMsg, "git_credentials.name") {
			return appErrors.ErrCredentialNameExists
		}
		if strings.Contains(errMsg, "dev_environments.name") {
			return appErrors.ErrEnvironmentNameExists
		}
	}

	// Handle MySQL duplicate entry errors
	if strings.Contains(errMsg, "Duplicate entry") {
		if strings.Contains(errMsg, "admins.username") || strings.Contains(errMsg, "for key 'admins.username'") {
			return appErrors.ErrAdminUsernameExists
		}
		if strings.Contains(errMsg, "projects.name") || strings.Contains(errMsg, "for key 'projects.name'") {
			return appErrors.ErrProjectNameExists
		}
		if strings.Contains(errMsg, "git_credentials.name") || strings.Contains(errMsg, "for key 'git_credentials.name'") {
			return appErrors.ErrCredentialNameExists
		}
		if strings.Contains(errMsg, "dev_environments.name") || strings.Contains(errMsg, "for key 'dev_environments.name'") {
			return appErrors.ErrEnvironmentNameExists
		}
	}

	// Return original error if no specific mapping found
	return err
}

// IsConstraintError checks if the error is a database constraint error
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "constraint failed")
}

// IsUniqueConstraintError checks if the error is specifically a unique constraint violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "Duplicate entry")
}
