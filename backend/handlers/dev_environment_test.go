package handlers

import (
	"errors"
	"net/http"
	"testing"
	"xsha-backend/database"
	"xsha-backend/testdata"
	"xsha-backend/testdata/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupAuth      bool
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name: "successful_creation",
			requestBody: CreateEnvironmentRequest{
				Name:        "Test Environment",
				Description: "Test description",
				Type:        "development",
				DockerImage: "ubuntu:20.04",
				CPULimit:    1.0,
				MemoryLimit: 512,
				EnvVars:     map[string]string{"KEY": "value"},
			},
			setupAuth: true,
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				env := &database.DevEnvironment{
					ID:          1,
					Name:        "Test Environment",
					Description: "Test description",
					Type:        "development",
					DockerImage: "ubuntu:20.04",
					CPULimit:    1.0,
					MemoryLimit: 512,
				}
				mockService.On("CreateEnvironment", 
					"Test Environment", "Test description", "", "development", "ubuntu:20.04", 
					1.0, int64(512), map[string]string{"KEY": "value"}, uint(1), "testadmin").
					Return(env, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedKeys:   []string{"message", "environment"},
		},
		{
			name:           "missing_required_fields",
			requestBody:    map[string]string{"description": "incomplete"},
			setupAuth:      true,
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name: "service_error",
			requestBody: CreateEnvironmentRequest{
				Name:        "Test Environment",
				Type:        "development",
				DockerImage: "ubuntu:20.04",
				CPULimit:    1.0,
				MemoryLimit: 512,
			},
			setupAuth: true,
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("CreateEnvironment", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 
					mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
					mock.AnythingOfType("float64"), mock.AnythingOfType("int64"), mock.AnythingOfType("map[string]string"), 
					mock.AnythingOfType("uint"), mock.AnythingOfType("string")).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()

			// Setup authentication context
			if tt.setupAuth {
				testdata.SetAuthContext(c, &database.Admin{
					ID:       1,
					Username: "testadmin",
				})
			}

			// Create request
			req, err := testdata.CreateJSONRequest("POST", "/environments", tt.requestBody)
			assert.NoError(t, err)
			c.Request = req

			// Execute handler
			handler.CreateEnvironment(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		envID          string
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:  "successful_get",
			envID: "1",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				env := &database.DevEnvironment{
					ID:          1,
					Name:        "Test Environment",
					Description: "Test description",
				}
				mockService.On("GetEnvironment", uint(1)).Return(env, nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"environment"},
		},
		{
			name:           "invalid_id_format",
			envID:          "invalid",
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:  "environment_not_found",
			envID: "999",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("GetEnvironment", uint(999)).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.envID}}

			// Execute handler
			handler.GetEnvironment(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestListEnvironments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		admin          *database.Admin
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name: "successful_list_super_admin",
			queryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
				"name":      "test",
			},
			admin: &database.Admin{
				ID:   1,
				Role: database.AdminRoleSuperAdmin,
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				envs := []database.DevEnvironment{
					{ID: 1, Name: "Test Environment 1"},
					{ID: 2, Name: "Test Environment 2"},
				}
				nameFilter := "test"
				mockService.On("ListEnvironments", &nameFilter, (*string)(nil), 1, 10).
					Return(envs, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"environments", "total", "page", "page_size", "total_pages"},
		},
		{
			name: "successful_list_regular_admin",
			queryParams: map[string]string{
				"page":      "1",
				"page_size": "5",
			},
			admin: &database.Admin{
				ID:   2,
				Role: database.AdminRoleAdmin,
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				envs := []database.DevEnvironment{
					{ID: 1, Name: "User Environment"},
				}
				mockService.On("ListEnvironmentsByAdminAccess", uint(2), (*string)(nil), (*string)(nil), 1, 5).
					Return(envs, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"environments", "total"},
		},
		{
			name: "service_error",
			queryParams: map[string]string{},
			admin: &database.Admin{
				ID:   1,
				Role: database.AdminRoleSuperAdmin,
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("ListEnvironments", (*string)(nil), (*string)(nil), 1, 10).
					Return([]database.DevEnvironment{}, int64(0), errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()

			// Build query string
			queryString := ""
			for key, value := range tt.queryParams {
				if queryString != "" {
					queryString += "&"
				}
				queryString += key + "=" + value
			}
			
			// Create a mock request with query parameters
			url := "/environments"
			if queryString != "" {
				url += "?" + queryString
			}
			req, err := http.NewRequest("GET", url, nil)
			assert.NoError(t, err)
			c.Request = req

			// Setup admin context
			c.Set("admin", tt.admin)

			// Execute handler
			handler.ListEnvironments(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		envID          string
		requestBody    interface{}
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:  "successful_update",
			envID: "1",
			requestBody: UpdateEnvironmentRequest{
				Name:        "Updated Environment",
				Description: "Updated description",
				CPULimit:    2.0,
				MemoryLimit: 1024,
				EnvVars:     map[string]string{"NEW_KEY": "new_value"},
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				expectedUpdates := map[string]interface{}{
					"name":          "Updated Environment",
					"description":   "Updated description",
					"system_prompt": "",
					"cpu_limit":     2.0,
					"memory_limit":  int64(1024),
				}
				mockService.On("UpdateEnvironment", uint(1), expectedUpdates).Return(nil)
				mockService.On("UpdateEnvironmentVars", uint(1), map[string]string{"NEW_KEY": "new_value"}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"message"},
		},
		{
			name:           "invalid_id_format",
			envID:          "invalid",
			requestBody:    UpdateEnvironmentRequest{},
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:  "invalid_request_body",
			envID: "1",
			requestBody: map[string]interface{}{
				"cpu_limit": "invalid_number",
			},
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:  "service_error",
			envID: "1",
			requestBody: UpdateEnvironmentRequest{
				Name: "Updated Environment",
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("UpdateEnvironment", uint(1), mock.AnythingOfType("map[string]interface {}")).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.envID}}

			// Create request
			req, err := testdata.CreateJSONRequest("PUT", "/environments/"+tt.envID, tt.requestBody)
			assert.NoError(t, err)
			c.Request = req

			// Execute handler
			handler.UpdateEnvironment(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		envID          string
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:  "successful_delete",
			envID: "1",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("DeleteEnvironment", uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"message"},
		},
		{
			name:           "invalid_id_format",
			envID:          "invalid",
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:  "service_error",
			envID: "1",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("DeleteEnvironment", uint(1)).Return(errors.New("cannot delete"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.envID}}

			// Execute handler
			handler.DeleteEnvironment(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetAvailableImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name: "successful_get_images",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				images := []map[string]interface{}{
					{"name": "Ubuntu", "image": "ubuntu:20.04"},
					{"name": "Node.js", "image": "node:16"},
				}
				mockService.On("GetAvailableEnvironmentImages").Return(images, nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"images"},
		},
		{
			name: "service_error",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("GetAvailableEnvironmentImages").Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()

			// Execute handler
			handler.GetAvailableImages(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetEnvironmentAdmins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		envID          string
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:  "successful_get_admins",
			envID: "1",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				admins := []database.Admin{
					{ID: 1, Username: "admin1", Name: "Admin 1"},
					{ID: 2, Username: "admin2", Name: "Admin 2"},
				}
				mockService.On("GetEnvironmentAdmins", uint(1)).Return(admins, nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"admins"},
		},
		{
			name:           "invalid_id_format",
			envID:          "invalid",
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:  "environment_not_found",
			envID: "999",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("GetEnvironmentAdmins", uint(999)).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.envID}}

			// Execute handler
			handler.GetEnvironmentAdmins(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAddAdminToEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		envID          string
		requestBody    interface{}
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:  "successful_add_admin",
			envID: "1",
			requestBody: AddAdminToEnvironmentRequest{
				AdminID: 2,
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("AddAdminToEnvironment", uint(1), uint(2)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"message"},
		},
		{
			name:           "invalid_env_id_format",
			envID:          "invalid",
			requestBody:    AddAdminToEnvironmentRequest{AdminID: 2},
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:           "missing_admin_id",
			envID:          "1",
			requestBody:    map[string]interface{}{},
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:  "service_error",
			envID: "1",
			requestBody: AddAdminToEnvironmentRequest{
				AdminID: 2,
			},
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("AddAdminToEnvironment", uint(1), uint(2)).
					Return(errors.New("admin already exists"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.envID}}

			// Create request
			req, err := testdata.CreateJSONRequest("POST", "/environments/"+tt.envID+"/admins", tt.requestBody)
			assert.NoError(t, err)
			c.Request = req

			// Execute handler
			handler.AddAdminToEnvironment(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestRemoveAdminFromEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		envID          string
		adminID        string
		setupMock      func(*mocks.MockDevEnvironmentService)
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:    "successful_remove_admin",
			envID:   "1",
			adminID: "2",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("RemoveAdminFromEnvironment", uint(1), uint(2)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"message"},
		},
		{
			name:           "invalid_env_id_format",
			envID:          "invalid",
			adminID:        "2",
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:           "invalid_admin_id_format",
			envID:          "1",
			adminID:        "invalid",
			setupMock:      func(mockService *mocks.MockDevEnvironmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
		{
			name:    "service_error",
			envID:   "1",
			adminID: "2",
			setupMock: func(mockService *mocks.MockDevEnvironmentService) {
				mockService.On("RemoveAdminFromEnvironment", uint(1), uint(2)).
					Return(errors.New("admin not found in environment"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockDevEnvironmentService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewDevEnvironmentHandlers(mockService)

			// Setup Gin context
			c, w := testdata.SetupTestGinContext()
			c.Params = gin.Params{
				gin.Param{Key: "id", Value: tt.envID},
				gin.Param{Key: "admin_id", Value: tt.adminID},
			}

			// Execute handler
			handler.RemoveAdminFromEnvironment(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			response := testdata.AssertJSONResponse(w, tt.expectedStatus)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, response, key)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}