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
		Lang:        admin.Lang,
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

// ToCredentialListItemResponse converts GitCredential to CredentialListItemResponse with minimal admin data
func ToCredentialListItemResponse(cred GitCredential) CredentialListItemResponse {
	response := CredentialListItemResponse{
		ID:          cred.ID,
		CreatedAt:   cred.CreatedAt,
		UpdatedAt:   cred.UpdatedAt,
		Name:        cred.Name,
		Description: cred.Description,
		Type:        cred.Type,
		Username:    cred.Username,
		AdminID:     cred.AdminID,
		CreatedBy:   cred.CreatedBy,
	}

	// Convert legacy single admin to minimal version
	if cred.Admin != nil {
		minimalAdmin := ToMinimalAdminResponse(*cred.Admin)
		response.Admin = &minimalAdmin
	}

	// Convert many-to-many admins to minimal versions
	if len(cred.Admins) > 0 {
		response.Admins = ToMinimalAdminResponses(cred.Admins)
	}

	return response
}

// ToCredentialListItemResponses converts slice of GitCredential to slice of CredentialListItemResponse
func ToCredentialListItemResponses(credentials []GitCredential) []CredentialListItemResponse {
	responses := make([]CredentialListItemResponse, len(credentials))
	for i, cred := range credentials {
		responses[i] = ToCredentialListItemResponse(cred)
	}
	return responses
}

// ToProjectListItemResponse converts Project to ProjectListItemResponse with minimal admin data
func ToProjectListItemResponse(project Project) ProjectListItemResponse {
	response := ProjectListItemResponse{
		ID:          project.ID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		Name:        project.Name,
		Description: project.Description,
		RepoURL:     project.RepoURL,
		Protocol:    project.Protocol,
		AdminID:     project.AdminID,
		AdminCount:  0, // Will be set by service layer
		CreatedBy:   project.CreatedBy,
	}

	// Convert legacy single admin to minimal version
	if project.Admin != nil {
		minimalAdmin := ToMinimalAdminResponse(*project.Admin)
		response.Admin = &minimalAdmin
	}

	// Convert many-to-many admins to minimal versions
	if len(project.Admins) > 0 {
		response.Admins = ToMinimalAdminResponses(project.Admins)
	}

	return response
}

// ToProjectListItemResponses converts slice of Project to slice of ProjectListItemResponse
func ToProjectListItemResponses(projects []Project) []ProjectListItemResponse {
	responses := make([]ProjectListItemResponse, len(projects))
	for i, project := range projects {
		responses[i] = ToProjectListItemResponse(project)
	}
	return responses
}
