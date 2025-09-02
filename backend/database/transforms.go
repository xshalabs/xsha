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