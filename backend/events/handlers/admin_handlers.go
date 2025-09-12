package handlers

import (
	"fmt"
	"log"
	"xsha-backend/events"
	"xsha-backend/services"
)

// AdminEventHandlers 管理员事件处理器
type AdminEventHandlers struct {
	auditService             services.AdminOperationLogService
	authService              services.AuthService
	adminService             services.AdminService
	gitCredentialService     services.GitCredentialService
	projectService           services.ProjectService
	devEnvironmentService    services.DevEnvironmentService
	taskService              services.TaskService
	taskConversationService  services.TaskConversationService
}

// NewAdminEventHandlers 创建管理员事件处理器
func NewAdminEventHandlers(
	auditService services.AdminOperationLogService,
	authService services.AuthService,
	adminService services.AdminService,
	gitCredService services.GitCredentialService,
	projectService services.ProjectService,
	devEnvService services.DevEnvironmentService,
	taskService services.TaskService,
	taskConvService services.TaskConversationService,
) *AdminEventHandlers {
	return &AdminEventHandlers{
		auditService:            auditService,
		authService:             authService,
		adminService:            adminService,
		gitCredentialService:    gitCredService,
		projectService:          projectService,
		devEnvironmentService:   devEnvService,
		taskService:             taskService,
		taskConversationService: taskConvService,
	}
}

// HandleAdminCreated 处理管理员创建事件
func (h *AdminEventHandlers) HandleAdminCreated(event events.Event) error {
	adminEvent, ok := event.(*events.AdminCreatedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for admin created handler")
	}

	log.Printf("Processing admin created event: Admin ID %d, Username: %s", 
		adminEvent.AdminID, adminEvent.Username)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogCreate(
			adminEvent.CreatedBy,
			&adminEvent.AdminID,
			"admin",
			fmt.Sprintf("%d", adminEvent.AdminID),
			fmt.Sprintf("Created admin: %s (%s) with role %s", adminEvent.Username, adminEvent.Name, adminEvent.Role),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log admin creation audit: %v", err)
		}
	}()

	// 2. 初始化管理员统计
	go func() {
		h.initializeAdminStats(adminEvent)
	}()

	// 3. 发送欢迎通知
	go func() {
		h.sendWelcomeNotification(adminEvent)
	}()

	return nil
}

// HandleAdminUpdated 处理管理员更新事件
func (h *AdminEventHandlers) HandleAdminUpdated(event events.Event) error {
	updateEvent, ok := event.(*events.AdminUpdatedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for admin updated handler")
	}

	log.Printf("Processing admin updated event: Admin ID %d, Fields: %v", 
		updateEvent.AdminID, updateEvent.UpdatedFields)

	// 1. 记录审计日志
	go func() {
		details := fmt.Sprintf("Updated fields: %v", updateEvent.UpdatedFields)
		if err := h.auditService.LogUpdate(
			updateEvent.UpdatedBy,
			&updateEvent.AdminID,
			"admin",
			fmt.Sprintf("%d", updateEvent.AdminID),
			fmt.Sprintf("Updated admin: %s", updateEvent.Username),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log admin update audit: %v", err)
		}
		_ = details
	}()

	// 2. 处理敏感字段变更
	go func() {
		h.handleSensitiveFieldChanges(updateEvent)
	}()

	return nil
}

// HandleAdminDeleted 处理管理员删除事件
func (h *AdminEventHandlers) HandleAdminDeleted(event events.Event) error {
	deleteEvent, ok := event.(*events.AdminDeletedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for admin deleted handler")
	}

	log.Printf("Processing admin deleted event: Admin ID %d, Username: %s", 
		deleteEvent.AdminID, deleteEvent.Username)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogDelete(
			deleteEvent.DeletedBy,
			&deleteEvent.AdminID,
			"admin",
			fmt.Sprintf("%d", deleteEvent.AdminID),
			fmt.Sprintf("Deleted admin: %s (%s), Reason: %s", deleteEvent.Username, deleteEvent.Name, deleteEvent.Reason),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log admin deletion audit: %v", err)
		}
	}()

	// 2. 清理相关资源
	go func() {
		if err := h.cleanupAdminResources(deleteEvent); err != nil {
			log.Printf("Failed to cleanup resources for deleted admin %d: %v", deleteEvent.AdminID, err)
		}
	}()

	// 3. 撤销所有token
	go func() {
		h.revokeAdminTokens(deleteEvent.AdminID, deleteEvent.Username)
	}()

	// 4. 移除管理员访问权限
	go func() {
		h.revokeAdminAccess(deleteEvent)
	}()

	return nil
}

// HandleAdminRoleChanged 处理管理员角色变更事件
func (h *AdminEventHandlers) HandleAdminRoleChanged(event events.Event) error {
	roleEvent, ok := event.(*events.AdminRoleChangedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for admin role changed handler")
	}

	log.Printf("Processing admin role changed event: Admin ID %d, %s -> %s", 
		roleEvent.AdminID, roleEvent.OldRole, roleEvent.NewRole)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogUpdate(
			roleEvent.ChangedBy,
			&roleEvent.AdminID,
			"admin",
			fmt.Sprintf("%d", roleEvent.AdminID),
			fmt.Sprintf("Changed role from %s to %s for admin: %s", roleEvent.OldRole, roleEvent.NewRole, roleEvent.Username),
			"", "", "", true, "",
		); err != nil {
			log.Printf("Failed to log admin role change audit: %v", err)
		}
	}()

	// 2. 更新权限
	go func() {
		h.updateAdminPermissions(roleEvent)
	}()

	// 3. 撤销现有token（强制重新登录以应用新权限）
	go func() {
		h.revokeAdminTokens(roleEvent.AdminID, roleEvent.Username)
	}()

	return nil
}

// HandleAdminLogin 处理管理员登录事件
func (h *AdminEventHandlers) HandleAdminLogin(event events.Event) error {
	loginEvent, ok := event.(*events.AdminLoginEvent)
	if !ok {
		return fmt.Errorf("invalid event type for admin login handler")
	}

	log.Printf("Processing admin login event: Admin ID %d, Success: %v", 
		loginEvent.AdminID, loginEvent.Success)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogLogin(
			loginEvent.Username,
			&loginEvent.AdminID,
			loginEvent.ClientIP,
			loginEvent.UserAgent,
			loginEvent.Success,
			loginEvent.ErrorMsg,
		); err != nil {
			log.Printf("Failed to log admin login audit: %v", err)
		}
	}()

	// 2. 更新登录统计
	go func() {
		h.updateLoginStats(loginEvent)
	}()

	// 3. 安全检查
	go func() {
		h.performSecurityChecks(loginEvent)
	}()

	return nil
}

// HandleAdminLogout 处理管理员登出事件
func (h *AdminEventHandlers) HandleAdminLogout(event events.Event) error {
	logoutEvent, ok := event.(*events.AdminLogoutEvent)
	if !ok {
		return fmt.Errorf("invalid event type for admin logout handler")
	}

	log.Printf("Processing admin logout event: Admin ID %d, Reason: %s", 
		logoutEvent.AdminID, logoutEvent.Reason)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogLogout(
			logoutEvent.Username,
			&logoutEvent.AdminID,
			logoutEvent.ClientIP,
			logoutEvent.UserAgent,
			logoutEvent.Success,
			logoutEvent.ErrorMsg,
		); err != nil {
			log.Printf("Failed to log admin logout audit: %v", err)
		}
	}()

	// 2. 清理会话数据
	go func() {
		h.cleanupSessionData(logoutEvent)
	}()

	return nil
}

// HandlePermissionGranted 处理权限授予事件
func (h *AdminEventHandlers) HandlePermissionGranted(event events.Event) error {
	permEvent, ok := event.(*events.PermissionGrantedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for permission granted handler")
	}

	log.Printf("Processing permission granted event: Admin %s, Resource: %s:%d, Action: %s", 
		permEvent.Username, permEvent.Resource, permEvent.ResourceID, permEvent.Action)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogOperation(
			permEvent.GrantedBy,
			&permEvent.AdminID,
			"grant_permission",
			permEvent.Resource,
			fmt.Sprintf("%d", permEvent.ResourceID),
			fmt.Sprintf("Granted %s permission on %s to %s", permEvent.Action, permEvent.Resource, permEvent.Username),
			"",
			true,
			"",
			"", "", "", "",
		); err != nil {
			log.Printf("Failed to log permission granted audit: %v", err)
		}
	}()

	return nil
}

// HandlePermissionRevoked 处理权限撤销事件
func (h *AdminEventHandlers) HandlePermissionRevoked(event events.Event) error {
	permEvent, ok := event.(*events.PermissionRevokedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for permission revoked handler")
	}

	log.Printf("Processing permission revoked event: Admin %s, Resource: %s:%d, Action: %s", 
		permEvent.Username, permEvent.Resource, permEvent.ResourceID, permEvent.Action)

	// 1. 记录审计日志
	go func() {
		if err := h.auditService.LogOperation(
			permEvent.RevokedBy,
			&permEvent.AdminID,
			"revoke_permission",
			permEvent.Resource,
			fmt.Sprintf("%d", permEvent.ResourceID),
			fmt.Sprintf("Revoked %s permission on %s from %s", permEvent.Action, permEvent.Resource, permEvent.Username),
			"",
			true,
			"",
			"", "", "", "",
		); err != nil {
			log.Printf("Failed to log permission revoked audit: %v", err)
		}
	}()

	// 2. 撤销相关token（如果影响当前权限）
	go func() {
		if permEvent.RevokeType == "direct" || permEvent.RevokeType == "cascading" {
			h.revokeAdminTokens(permEvent.AdminID, permEvent.Username)
		}
	}()

	return nil
}

// 辅助方法实现

func (h *AdminEventHandlers) initializeAdminStats(adminEvent *events.AdminCreatedEvent) {
	log.Printf("Initializing stats for admin %d", adminEvent.AdminID)
	// 实现管理员统计初始化
}

func (h *AdminEventHandlers) sendWelcomeNotification(adminEvent *events.AdminCreatedEvent) {
	log.Printf("Sending welcome notification to admin %d", adminEvent.AdminID)
	// 实现欢迎通知发送
}

func (h *AdminEventHandlers) handleSensitiveFieldChanges(updateEvent *events.AdminUpdatedEvent) {
	log.Printf("Handling sensitive field changes for admin %d", updateEvent.AdminID)
	
	// 检查是否有敏感字段变更
	sensitiveFields := []string{"username", "email", "role", "is_active"}
	for _, field := range updateEvent.UpdatedFields {
		for _, sensitive := range sensitiveFields {
			if field == sensitive {
				// 撤销token以强制重新认证
				h.revokeAdminTokens(updateEvent.AdminID, updateEvent.Username)
				return
			}
		}
	}
}

func (h *AdminEventHandlers) cleanupAdminResources(deleteEvent *events.AdminDeletedEvent) error {
	// 1. 移除管理员与Git凭据的关联
	for _, credID := range deleteEvent.RelatedCredentials {
		if err := h.gitCredentialService.RemoveAdminFromCredential(credID, deleteEvent.AdminID); err != nil {
			log.Printf("Failed to remove admin %d from credential %d: %v", deleteEvent.AdminID, credID, err)
		}
	}

	// 2. 移除管理员与项目的关联
	for _, projectID := range deleteEvent.RelatedProjects {
		if err := h.projectService.RemoveAdminFromProject(projectID, deleteEvent.AdminID); err != nil {
			log.Printf("Failed to remove admin %d from project %d: %v", deleteEvent.AdminID, projectID, err)
		}
	}

	// 3. 移除管理员与开发环境的关联
	for _, envID := range deleteEvent.RelatedEnvironments {
		if err := h.devEnvironmentService.RemoveAdminFromEnvironment(envID, deleteEvent.AdminID); err != nil {
			log.Printf("Failed to remove admin %d from environment %d: %v", deleteEvent.AdminID, envID, err)
		}
	}

	return nil
}

func (h *AdminEventHandlers) revokeAdminTokens(adminID uint, username string) {
	log.Printf("Revoking all tokens for admin %d (%s)", adminID, username)
	// 这里需要实现token撤销逻辑
	// 可能需要查询所有该管理员的活跃token并加入黑名单
}

func (h *AdminEventHandlers) revokeAdminAccess(deleteEvent *events.AdminDeletedEvent) {
	log.Printf("Revoking access for deleted admin %d", deleteEvent.AdminID)
	// 实现管理员访问权限撤销逻辑
}

func (h *AdminEventHandlers) updateAdminPermissions(roleEvent *events.AdminRoleChangedEvent) {
	log.Printf("Updating permissions for admin %d role change", roleEvent.AdminID)
	// 实现权限更新逻辑
}

func (h *AdminEventHandlers) updateLoginStats(loginEvent *events.AdminLoginEvent) {
	log.Printf("Updating login stats for admin %d", loginEvent.AdminID)
	// 实现登录统计更新
}

func (h *AdminEventHandlers) performSecurityChecks(loginEvent *events.AdminLoginEvent) {
	log.Printf("Performing security checks for admin login %d", loginEvent.AdminID)
	
	if !loginEvent.Success {
		// 检查失败登录次数，可能需要锁定账户
		log.Printf("Failed login attempt for admin %d from IP %s", loginEvent.AdminID, loginEvent.ClientIP)
	}
	
	// 其他安全检查...
}

func (h *AdminEventHandlers) cleanupSessionData(logoutEvent *events.AdminLogoutEvent) {
	log.Printf("Cleaning up session data for admin %d", logoutEvent.AdminID)
	// 实现会话数据清理
}