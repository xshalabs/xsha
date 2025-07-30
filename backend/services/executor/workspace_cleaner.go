package executor

import (
	"fmt"
	"xsha-backend/utils"
)

type workspaceCleaner struct {
	workspaceManager *utils.WorkspaceManager
}

func NewWorkspaceCleaner(workspaceManager *utils.WorkspaceManager) WorkspaceCleaner {
	return &workspaceCleaner{
		workspaceManager: workspaceManager,
	}
}

func (w *workspaceCleaner) CleanupOnFailure(taskID uint, workspacePath string) error {
	if workspacePath == "" {
		utils.Warn("Workspace path is empty, skipping cleanup", "task_id", taskID)
		return nil
	}

	utils.Info("Starting to clean failed task workspace", "task_id", taskID, "workspace", workspacePath)

	isDirty, err := w.workspaceManager.CheckWorkspaceIsDirty(workspacePath)
	if err != nil {
		utils.Error("Failed to check workspace status", "task_id", taskID, "workspace", workspacePath, "error", err)
	}

	if isDirty || err != nil {
		if resetErr := w.workspaceManager.ResetWorkspaceToCleanState(workspacePath); resetErr != nil {
			utils.Error("Failed to reset workspace", "task_id", taskID, "workspace", workspacePath, "error", resetErr)
			return fmt.Errorf("failed to cleanup failed task workspace: %v", resetErr)
		}
		utils.Info("Cleaned failed task workspace file changes", "task_id", taskID, "workspace", workspacePath)
	} else {
		utils.Info("Workspace is already clean, no cleanup needed", "task_id", taskID, "workspace", workspacePath)
	}

	return nil
}

func (w *workspaceCleaner) CleanupOnCancel(taskID uint, workspacePath string) error {
	if workspacePath == "" {
		utils.Warn("Workspace path is empty, skipping cleanup", "task_id", taskID)
		return nil
	}

	utils.Info("Starting to clean cancelled task workspace", "task_id", taskID, "workspace", workspacePath)

	isDirty, err := w.workspaceManager.CheckWorkspaceIsDirty(workspacePath)
	if err != nil {
		utils.Error("Failed to check workspace status", "task_id", taskID, "workspace", workspacePath, "error", err)
	}

	if isDirty || err != nil {
		if resetErr := w.workspaceManager.ResetWorkspaceToCleanState(workspacePath); resetErr != nil {
			utils.Error("Failed to reset workspace", "task_id", taskID, "workspace", workspacePath, "error", resetErr)
			return fmt.Errorf("failed to cleanup cancelled task workspace: %v", resetErr)
		}
		utils.Info("Cleaned cancelled task workspace file changes", "task_id", taskID, "workspace", workspacePath)
	} else {
		utils.Info("Workspace is already clean, no cleanup needed", "task_id", taskID, "workspace", workspacePath)
	}

	return nil
}
