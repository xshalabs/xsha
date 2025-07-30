package executor

import (
	"fmt"
	"xsha-backend/utils"
)

type workspaceCleaner struct {
	workspaceManager *utils.WorkspaceManager
}

// NewWorkspaceCleaner 创建工作空间清理器
func NewWorkspaceCleaner(workspaceManager *utils.WorkspaceManager) WorkspaceCleaner {
	return &workspaceCleaner{
		workspaceManager: workspaceManager,
	}
}

// CleanupOnFailure 在任务执行失败时清理工作空间
func (w *workspaceCleaner) CleanupOnFailure(taskID uint, workspacePath string) error {
	if workspacePath == "" {
		utils.Warn("Workspace path is empty, skipping cleanup", "task_id", taskID)
		return nil
	}

	utils.Info("Starting to clean failed task workspace", "task_id", taskID, "workspace", workspacePath)

	// 检查工作空间是否为脏状态
	isDirty, err := w.workspaceManager.CheckWorkspaceIsDirty(workspacePath)
	if err != nil {
		utils.Error("Failed to check workspace status", "task_id", taskID, "workspace", workspacePath, "error", err)
		// 即使检查失败，也尝试清理
	}

	if isDirty || err != nil {
		// 重置工作空间到干净状态
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

// CleanupOnCancel 在任务被取消时清理工作空间
func (w *workspaceCleaner) CleanupOnCancel(taskID uint, workspacePath string) error {
	if workspacePath == "" {
		utils.Warn("Workspace path is empty, skipping cleanup", "task_id", taskID)
		return nil
	}

	utils.Info("Starting to clean cancelled task workspace", "task_id", taskID, "workspace", workspacePath)

	// 检查工作空间是否为脏状态
	isDirty, err := w.workspaceManager.CheckWorkspaceIsDirty(workspacePath)
	if err != nil {
		utils.Error("Failed to check workspace status", "task_id", taskID, "workspace", workspacePath, "error", err)
		// 即使检查失败，也尝试清理
	}

	if isDirty || err != nil {
		// 重置工作空间到干净状态
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
