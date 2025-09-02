package database

// ToAdminListResponse converts Admin to AdminListResponse with minimal avatar
func ToAdminListResponse(admin Admin) AdminListResponse {
	response := AdminListResponse{
		ID:          admin.ID,
		CreatedAt:   admin.CreatedAt,
		UpdatedAt:   admin.UpdatedAt,
		Username:    admin.Username,
		Name:        admin.Name,
		Email:       admin.Email,
		Role:        admin.Role,
		IsActive:    admin.IsActive,
		LastLoginAt: admin.LastLoginAt,
		LastLoginIP: admin.LastLoginIP,
		AvatarID:    admin.AvatarID,
		CreatedBy:   admin.CreatedBy,
	}

	// Convert avatar to minimal version
	if admin.Avatar != nil {
		response.Avatar = &AdminAvatarMinimal{
			UUID:         admin.Avatar.UUID,
			OriginalName: admin.Avatar.OriginalName,
		}
	}

	return response
}

// ToAdminListResponses converts slice of Admin to slice of AdminListResponse
func ToAdminListResponses(admins []Admin) []AdminListResponse {
	responses := make([]AdminListResponse, len(admins))
	for i, admin := range admins {
		responses[i] = ToAdminListResponse(admin)
	}
	return responses
}

// ToMinimalAdminResponse converts Admin to MinimalAdminResponse with minimal avatar
func ToMinimalAdminResponse(admin Admin) MinimalAdminResponse {
	response := MinimalAdminResponse{
		ID:       admin.ID,
		Username: admin.Username,
		Name:     admin.Name,
		Email:    admin.Email,
	}

	// Convert avatar to minimal version
	if admin.Avatar != nil {
		response.Avatar = &AdminAvatarMinimal{
			UUID:         admin.Avatar.UUID,
			OriginalName: admin.Avatar.OriginalName,
		}
	}

	return response
}

// ToMinimalAdminResponses converts slice of Admin to slice of MinimalAdminResponse
func ToMinimalAdminResponses(admins []Admin) []MinimalAdminResponse {
	responses := make([]MinimalAdminResponse, len(admins))
	for i, admin := range admins {
		responses[i] = ToMinimalAdminResponse(admin)
	}
	return responses
}

// ToEnvironmentListItemResponse converts DevEnvironment to EnvironmentListItemResponse with minimal admin data
func ToEnvironmentListItemResponse(env DevEnvironment) EnvironmentListItemResponse {
	response := EnvironmentListItemResponse{
		ID:           env.ID,
		CreatedAt:    env.CreatedAt,
		UpdatedAt:    env.UpdatedAt,
		Name:         env.Name,
		Description:  env.Description,
		SystemPrompt: env.SystemPrompt,
		Type:         env.Type,
		DockerImage:  env.DockerImage,
		CPULimit:     env.CPULimit,
		MemoryLimit:  env.MemoryLimit,
		EnvVars:      env.EnvVars,
		SessionDir:   env.SessionDir,
		AdminID:      env.AdminID,
		CreatedBy:    env.CreatedBy,
	}

	// Convert legacy single admin to minimal version
	if env.Admin != nil {
		minimalAdmin := ToMinimalAdminResponse(*env.Admin)
		response.Admin = &minimalAdmin
	}

	// Convert many-to-many admins to minimal versions
	if len(env.Admins) > 0 {
		response.Admins = ToMinimalAdminResponses(env.Admins)
	}

	return response
}

// ToEnvironmentListItemResponses converts slice of DevEnvironment to slice of EnvironmentListItemResponse
func ToEnvironmentListItemResponses(environments []DevEnvironment) []EnvironmentListItemResponse {
	responses := make([]EnvironmentListItemResponse, len(environments))
	for i, env := range environments {
		responses[i] = ToEnvironmentListItemResponse(env)
	}
	return responses
}