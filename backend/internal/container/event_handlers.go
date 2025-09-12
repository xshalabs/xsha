package container

import (
	eventHandlers "xsha-backend/events/handlers"
)

type EventHandlers struct {
	Task  *eventHandlers.TaskEventHandlers
	Admin *eventHandlers.AdminEventHandlers
}

func (c *Container) createEventHandlers() *EventHandlers {
	taskEventHandlers := eventHandlers.NewTaskEventHandlers(
		c.AdminOperationLogService,
		c.WorkspaceManager,
		c.TaskService,
		c.ProjectService,
		c.GitCredService,
		c.SystemConfigService,
	)

	adminEventHandlers := eventHandlers.NewAdminEventHandlers(
		c.AdminOperationLogService,
		c.AuthService,
		c.AdminService,
		c.GitCredService,
		c.ProjectService,
		c.DevEnvService,
		c.TaskService,
		c.TaskConvService,
	)

	return &EventHandlers{
		Task:  taskEventHandlers,
		Admin: adminEventHandlers,
	}
}