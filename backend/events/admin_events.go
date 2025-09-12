package events

import (
	"time"
	"xsha-backend/database"
)

// 管理员事件类型常量
const (
	EventTypeAdminCreated      = "admin.created"
	EventTypeAdminUpdated      = "admin.updated"
	EventTypeAdminDeleted      = "admin.deleted"
	EventTypeAdminActivated    = "admin.activated"
	EventTypeAdminDeactivated  = "admin.deactivated"
	EventTypeAdminRoleChanged  = "admin.role.changed"
	EventTypeAdminLogin        = "admin.login"
	EventTypeAdminLogout       = "admin.logout"
	EventTypeAdminPasswordChanged = "admin.password.changed"
	
	// 权限相关事件
	EventTypePermissionGranted = "permission.granted"
	EventTypePermissionRevoked = "permission.revoked"
	
	// 资源访问事件
	EventTypeResourceAccessGranted = "resource.access.granted"
	EventTypeResourceAccessRevoked = "resource.access.revoked"
)

// AdminCreatedEvent 管理员创建事件
type AdminCreatedEvent struct {
	BaseEvent
	AdminID   uint                `json:"admin_id"`
	Username  string              `json:"username"`
	Name      string              `json:"name"`
	Email     string              `json:"email"`
	Role      database.AdminRole  `json:"role"`
	IsActive  bool                `json:"is_active"`
	CreatedBy string              `json:"created_by"`
}

// NewAdminCreatedEvent 创建管理员创建事件
func NewAdminCreatedEvent(admin *database.Admin) *AdminCreatedEvent {
	event := &AdminCreatedEvent{
		BaseEvent: NewBaseEvent(EventTypeAdminCreated),
		AdminID:   admin.ID,
		Username:  admin.Username,
		Name:      admin.Name,
		Email:     admin.Email,
		Role:      admin.Role,
		IsActive:  admin.IsActive,
		CreatedBy: admin.CreatedBy,
	}
	event.Payload = event
	return event
}

// AdminUpdatedEvent 管理员更新事件
type AdminUpdatedEvent struct {
	BaseEvent
	AdminID       uint                   `json:"admin_id"`
	Username      string                 `json:"username"`
	Changes       map[string]interface{} `json:"changes"`
	UpdatedBy     string                 `json:"updated_by"`
	UpdatedFields []string               `json:"updated_fields"`
	OldValues     map[string]interface{} `json:"old_values"`
}

// NewAdminUpdatedEvent 创建管理员更新事件
func NewAdminUpdatedEvent(adminID uint, username string, changes, oldValues map[string]interface{}, updatedBy string) *AdminUpdatedEvent {
	fields := make([]string, 0, len(changes))
	for field := range changes {
		fields = append(fields, field)
	}
	
	event := &AdminUpdatedEvent{
		BaseEvent:     NewBaseEvent(EventTypeAdminUpdated),
		AdminID:       adminID,
		Username:      username,
		Changes:       changes,
		UpdatedBy:     updatedBy,
		UpdatedFields: fields,
		OldValues:     oldValues,
	}
	event.Payload = event
	return event
}

// AdminDeletedEvent 管理员删除事件
type AdminDeletedEvent struct {
	BaseEvent
	AdminID            uint     `json:"admin_id"`
	Username           string   `json:"username"`
	Name               string   `json:"name"`
	Role               database.AdminRole `json:"role"`
	DeletedBy          string   `json:"deleted_by"`
	Reason             string   `json:"reason"`
	RelatedCredentials []uint   `json:"related_credentials"`
	RelatedProjects    []uint   `json:"related_projects"`
	RelatedTasks       []uint   `json:"related_tasks"`
	RelatedEnvironments []uint  `json:"related_environments"`
}

// NewAdminDeletedEvent 创建管理员删除事件
func NewAdminDeletedEvent(admin *database.Admin, deletedBy, reason string) *AdminDeletedEvent {
	event := &AdminDeletedEvent{
		BaseEvent:           NewBaseEvent(EventTypeAdminDeleted),
		AdminID:             admin.ID,
		Username:            admin.Username,
		Name:                admin.Name,
		Role:                admin.Role,
		DeletedBy:           deletedBy,
		Reason:              reason,
		RelatedCredentials:  []uint{},
		RelatedProjects:     []uint{},
		RelatedTasks:        []uint{},
		RelatedEnvironments: []uint{},
	}
	event.Payload = event
	return event
}

// SetRelatedResources 设置相关资源
func (e *AdminDeletedEvent) SetRelatedResources(credentials, projects, tasks, environments []uint) {
	e.RelatedCredentials = credentials
	e.RelatedProjects = projects
	e.RelatedTasks = tasks
	e.RelatedEnvironments = environments
}

// AdminActivatedEvent 管理员激活事件
type AdminActivatedEvent struct {
	BaseEvent
	AdminID     uint   `json:"admin_id"`
	Username    string `json:"username"`
	ActivatedBy string `json:"activated_by"`
	Reason      string `json:"reason"`
}

// NewAdminActivatedEvent 创建管理员激活事件
func NewAdminActivatedEvent(adminID uint, username, activatedBy, reason string) *AdminActivatedEvent {
	event := &AdminActivatedEvent{
		BaseEvent:   NewBaseEvent(EventTypeAdminActivated),
		AdminID:     adminID,
		Username:    username,
		ActivatedBy: activatedBy,
		Reason:      reason,
	}
	event.Payload = event
	return event
}

// AdminDeactivatedEvent 管理员停用事件
type AdminDeactivatedEvent struct {
	BaseEvent
	AdminID       uint   `json:"admin_id"`
	Username      string `json:"username"`
	DeactivatedBy string `json:"deactivated_by"`
	Reason        string `json:"reason"`
}

// NewAdminDeactivatedEvent 创建管理员停用事件
func NewAdminDeactivatedEvent(adminID uint, username, deactivatedBy, reason string) *AdminDeactivatedEvent {
	event := &AdminDeactivatedEvent{
		BaseEvent:     NewBaseEvent(EventTypeAdminDeactivated),
		AdminID:       adminID,
		Username:      username,
		DeactivatedBy: deactivatedBy,
		Reason:        reason,
	}
	event.Payload = event
	return event
}

// AdminRoleChangedEvent 管理员角色变更事件
type AdminRoleChangedEvent struct {
	BaseEvent
	AdminID   uint               `json:"admin_id"`
	Username  string             `json:"username"`
	OldRole   database.AdminRole `json:"old_role"`
	NewRole   database.AdminRole `json:"new_role"`
	ChangedBy string             `json:"changed_by"`
	Reason    string             `json:"reason"`
}

// NewAdminRoleChangedEvent 创建管理员角色变更事件
func NewAdminRoleChangedEvent(adminID uint, username string, oldRole, newRole database.AdminRole, changedBy, reason string) *AdminRoleChangedEvent {
	event := &AdminRoleChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeAdminRoleChanged),
		AdminID:   adminID,
		Username:  username,
		OldRole:   oldRole,
		NewRole:   newRole,
		ChangedBy: changedBy,
		Reason:    reason,
	}
	event.Payload = event
	return event
}

// AdminLoginEvent 管理员登录事件
type AdminLoginEvent struct {
	BaseEvent
	AdminID    uint      `json:"admin_id"`
	Username   string    `json:"username"`
	ClientIP   string    `json:"client_ip"`
	UserAgent  string    `json:"user_agent"`
	LoginTime  time.Time `json:"login_time"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
	SessionID  string    `json:"session_id,omitempty"`
}

// NewAdminLoginEvent 创建管理员登录事件
func NewAdminLoginEvent(adminID uint, username, clientIP, userAgent, errorMsg, sessionID string, success bool) *AdminLoginEvent {
	event := &AdminLoginEvent{
		BaseEvent: NewBaseEvent(EventTypeAdminLogin),
		AdminID:   adminID,
		Username:  username,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		LoginTime: time.Now(),
		Success:   success,
		ErrorMsg:  errorMsg,
		SessionID: sessionID,
	}
	event.Payload = event
	return event
}

// AdminLogoutEvent 管理员登出事件
type AdminLogoutEvent struct {
	BaseEvent
	AdminID    uint      `json:"admin_id"`
	Username   string    `json:"username"`
	ClientIP   string    `json:"client_ip"`
	UserAgent  string    `json:"user_agent"`
	LogoutTime time.Time `json:"logout_time"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
	SessionID  string    `json:"session_id,omitempty"`
	Reason     string    `json:"reason"` // manual, timeout, force_logout
}

// NewAdminLogoutEvent 创建管理员登出事件
func NewAdminLogoutEvent(adminID uint, username, clientIP, userAgent, errorMsg, sessionID, reason string, success bool) *AdminLogoutEvent {
	event := &AdminLogoutEvent{
		BaseEvent:  NewBaseEvent(EventTypeAdminLogout),
		AdminID:    adminID,
		Username:   username,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
		LogoutTime: time.Now(),
		Success:    success,
		ErrorMsg:   errorMsg,
		SessionID:  sessionID,
		Reason:     reason,
	}
	event.Payload = event
	return event
}

// AdminPasswordChangedEvent 管理员密码变更事件
type AdminPasswordChangedEvent struct {
	BaseEvent
	AdminID   uint   `json:"admin_id"`
	Username  string `json:"username"`
	ChangedBy string `json:"changed_by"`
	IsSelfChange bool `json:"is_self_change"`
	Reason    string `json:"reason"`
}

// NewAdminPasswordChangedEvent 创建管理员密码变更事件
func NewAdminPasswordChangedEvent(adminID uint, username, changedBy, reason string, isSelfChange bool) *AdminPasswordChangedEvent {
	event := &AdminPasswordChangedEvent{
		BaseEvent:    NewBaseEvent(EventTypeAdminPasswordChanged),
		AdminID:      adminID,
		Username:     username,
		ChangedBy:    changedBy,
		IsSelfChange: isSelfChange,
		Reason:       reason,
	}
	event.Payload = event
	return event
}

// PermissionGrantedEvent 权限授予事件
type PermissionGrantedEvent struct {
	BaseEvent
	AdminID      uint   `json:"admin_id"`
	Username     string `json:"username"`
	Resource     string `json:"resource"`
	ResourceID   uint   `json:"resource_id"`
	Action       string `json:"action"`
	GrantedBy    string `json:"granted_by"`
	GrantType    string `json:"grant_type"` // direct, role_based, inherited
	Reason       string `json:"reason"`
}

// NewPermissionGrantedEvent 创建权限授予事件
func NewPermissionGrantedEvent(adminID uint, username, resource string, resourceID uint, action, grantedBy, grantType, reason string) *PermissionGrantedEvent {
	event := &PermissionGrantedEvent{
		BaseEvent:  NewBaseEvent(EventTypePermissionGranted),
		AdminID:    adminID,
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
		GrantedBy:  grantedBy,
		GrantType:  grantType,
		Reason:     reason,
	}
	event.Payload = event
	return event
}

// PermissionRevokedEvent 权限撤销事件
type PermissionRevokedEvent struct {
	BaseEvent
	AdminID      uint   `json:"admin_id"`
	Username     string `json:"username"`
	Resource     string `json:"resource"`
	ResourceID   uint   `json:"resource_id"`
	Action       string `json:"action"`
	RevokedBy    string `json:"revoked_by"`
	RevokeType   string `json:"revoke_type"` // direct, cascading, policy_change
	Reason       string `json:"reason"`
}

// NewPermissionRevokedEvent 创建权限撤销事件
func NewPermissionRevokedEvent(adminID uint, username, resource string, resourceID uint, action, revokedBy, revokeType, reason string) *PermissionRevokedEvent {
	event := &PermissionRevokedEvent{
		BaseEvent:  NewBaseEvent(EventTypePermissionRevoked),
		AdminID:    adminID,
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
		RevokedBy:  revokedBy,
		RevokeType: revokeType,
		Reason:     reason,
	}
	event.Payload = event
	return event
}

// ResourceAccessGrantedEvent 资源访问授予事件
type ResourceAccessGrantedEvent struct {
	BaseEvent
	AdminID      uint   `json:"admin_id"`
	Username     string `json:"username"`
	ResourceType string `json:"resource_type"` // project, credential, environment
	ResourceID   uint   `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	AccessLevel  string `json:"access_level"` // owner, member, viewer
	GrantedBy    string `json:"granted_by"`
	Reason       string `json:"reason"`
}

// NewResourceAccessGrantedEvent 创建资源访问授予事件
func NewResourceAccessGrantedEvent(adminID uint, username, resourceType string, resourceID uint, resourceName, accessLevel, grantedBy, reason string) *ResourceAccessGrantedEvent {
	event := &ResourceAccessGrantedEvent{
		BaseEvent:    NewBaseEvent(EventTypeResourceAccessGranted),
		AdminID:      adminID,
		Username:     username,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		AccessLevel:  accessLevel,
		GrantedBy:    grantedBy,
		Reason:       reason,
	}
	event.Payload = event
	return event
}

// ResourceAccessRevokedEvent 资源访问撤销事件
type ResourceAccessRevokedEvent struct {
	BaseEvent
	AdminID      uint   `json:"admin_id"`
	Username     string `json:"username"`
	ResourceType string `json:"resource_type"` // project, credential, environment
	ResourceID   uint   `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	AccessLevel  string `json:"access_level"`
	RevokedBy    string `json:"revoked_by"`
	Reason       string `json:"reason"`
}

// NewResourceAccessRevokedEvent 创建资源访问撤销事件
func NewResourceAccessRevokedEvent(adminID uint, username, resourceType string, resourceID uint, resourceName, accessLevel, revokedBy, reason string) *ResourceAccessRevokedEvent {
	event := &ResourceAccessRevokedEvent{
		BaseEvent:    NewBaseEvent(EventTypeResourceAccessRevoked),
		AdminID:      adminID,
		Username:     username,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		AccessLevel:  accessLevel,
		RevokedBy:    revokedBy,
		Reason:       reason,
	}
	event.Payload = event
	return event
}