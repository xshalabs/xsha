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
		utils.Warn("工作空间路径为空，跳过清理", "task_id", taskID)
		return nil
	}

	utils.Info("开始清理失败任务的工作空间", "task_id", taskID, "workspace", workspacePath)

	// 检查工作空间是否为脏状态
	isDirty, err := w.workspaceManager.CheckWorkspaceIsDirty(workspacePath)
	if err != nil {
		utils.Error("检查工作空间状态失败", "task_id", taskID, "workspace", workspacePath, "error", err)
		// 即使检查失败，也尝试清理
	}

	if isDirty || err != nil {
		// 重置工作空间到干净状态
		if resetErr := w.workspaceManager.ResetWorkspaceToCleanState(workspacePath); resetErr != nil {
			utils.Error("重置工作空间失败", "task_id", taskID, "workspace", workspacePath, "error", resetErr)
			return fmt.Errorf("清理失败任务工作空间失败: %v", resetErr)
		}
		utils.Info("已清理失败任务的工作空间文件变动", "task_id", taskID, "workspace", workspacePath)
	} else {
		utils.Info("工作空间已处于干净状态，无需清理", "task_id", taskID, "workspace", workspacePath)
	}

	return nil
}

// CleanupOnCancel 在任务被取消时清理工作空间
func (w *workspaceCleaner) CleanupOnCancel(taskID uint, workspacePath string) error {
	if workspacePath == "" {
		utils.Warn("工作空间路径为空，跳过清理", "task_id", taskID)
		return nil
	}

	utils.Info("开始清理被取消任务的工作空间", "task_id", taskID, "workspace", workspacePath)

	// 检查工作空间是否为脏状态
	isDirty, err := w.workspaceManager.CheckWorkspaceIsDirty(workspacePath)
	if err != nil {
		utils.Error("检查工作空间状态失败", "task_id", taskID, "workspace", workspacePath, "error", err)
		// 即使检查失败，也尝试清理
	}

	if isDirty || err != nil {
		// 重置工作空间到干净状态
		if resetErr := w.workspaceManager.ResetWorkspaceToCleanState(workspacePath); resetErr != nil {
			utils.Error("重置工作空间失败", "task_id", taskID, "workspace", workspacePath, "error", resetErr)
			return fmt.Errorf("清理取消任务工作空间失败: %v", resetErr)
		}
		utils.Info("已清理被取消任务的工作空间文件变动", "task_id", taskID, "workspace", workspacePath)
	} else {
		utils.Info("工作空间已处于干净状态，无需清理", "task_id", taskID, "workspace", workspacePath)
	}

	return nil
}
