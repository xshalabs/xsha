package testdata

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"xsha-backend/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateTestAdmin creates a test admin user
func CreateTestAdmin(db *gorm.DB) *database.Admin {
	admin := &database.Admin{
		Username:     "testadmin",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password" hashed
		Name:         "Test Admin",
		Email:        "test@example.com",
		IsActive:     true,
		Role:         database.AdminRoleAdmin,
	}
	db.Create(admin)
	return admin
}

// CreateTestSuperAdmin creates a test super admin user
func CreateTestSuperAdmin(db *gorm.DB) *database.Admin {
	admin := &database.Admin{
		Username:     "superadmin",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password" hashed
		Name:         "Super Admin",
		Email:        "super@example.com",
		IsActive:     true,
		Role:         database.AdminRoleSuperAdmin,
	}
	db.Create(admin)
	return admin
}

// CreateTestEnvironment creates a test development environment
func CreateTestEnvironment(db *gorm.DB, adminID uint) *database.DevEnvironment {
	env := &database.DevEnvironment{
		Name:         "Test Environment",
		Description:  "Test environment for unit tests",
		SystemPrompt: "You are a test assistant",
		Type:         "development",
		DockerImage:  "ubuntu:20.04",
		CPULimit:     1.0,
		MemoryLimit:  512,
		CreatedBy:    "testadmin",
		AdminID:      &adminID,
	}
	db.Create(env)
	return env
}

// SetupTestGinContext creates a test Gin context with response recorder
func SetupTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Set language context
	c.Set("lang", "en-US")
	
	return c, w
}

// CreateJSONRequest creates an HTTP request with JSON body
func CreateJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// SetAuthContext sets authentication context for testing
func SetAuthContext(c *gin.Context, admin *database.Admin) {
	c.Set("admin_id", admin.ID)
	c.Set("username", admin.Username)
	c.Set("admin", admin)
}

// AssertJSONResponse asserts the JSON response structure
func AssertJSONResponse(w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	if w.Code != expectedStatus {
		panic("Expected status " + string(rune(expectedStatus)) + " but got " + string(rune(w.Code)))
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		panic("Failed to unmarshal JSON response: " + err.Error())
	}
	
	return response
}