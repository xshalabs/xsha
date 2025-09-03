package mocks

import (
	"xsha-backend/database"

	"github.com/stretchr/testify/mock"
)

// MockDevEnvironmentService is a mock implementation of services.DevEnvironmentService
type MockDevEnvironmentService struct {
	mock.Mock
}

func (m *MockDevEnvironmentService) CreateEnvironment(name, description, systemPrompt, envType, dockerImage string, cpuLimit float64, memoryLimit int64, envVars map[string]string, adminID uint, createdBy string) (*database.DevEnvironment, error) {
	args := m.Called(name, description, systemPrompt, envType, dockerImage, cpuLimit, memoryLimit, envVars, adminID, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.DevEnvironment), args.Error(1)
}

func (m *MockDevEnvironmentService) GetEnvironment(id uint) (*database.DevEnvironment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.DevEnvironment), args.Error(1)
}

func (m *MockDevEnvironmentService) GetEnvironmentWithAdmins(id uint) (*database.DevEnvironment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.DevEnvironment), args.Error(1)
}

func (m *MockDevEnvironmentService) ListEnvironments(name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	args := m.Called(name, dockerImage, page, pageSize)
	return args.Get(0).([]database.DevEnvironment), args.Get(1).(int64), args.Error(2)
}

func (m *MockDevEnvironmentService) ListEnvironmentsByAdminAccess(adminID uint, name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	args := m.Called(adminID, name, dockerImage, page, pageSize)
	return args.Get(0).([]database.DevEnvironment), args.Get(1).(int64), args.Error(2)
}

func (m *MockDevEnvironmentService) UpdateEnvironment(id uint, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) DeleteEnvironment(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) ValidateEnvVars(envVars map[string]string) error {
	args := m.Called(envVars)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) UpdateEnvironmentVars(id uint, envVars map[string]string) error {
	args := m.Called(id, envVars)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) ValidateResourceLimits(cpuLimit float64, memoryLimit int64) error {
	args := m.Called(cpuLimit, memoryLimit)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) GetAvailableEnvironmentImages() ([]map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDevEnvironmentService) AddAdminToEnvironment(envID, adminID uint) error {
	args := m.Called(envID, adminID)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) RemoveAdminFromEnvironment(envID, adminID uint) error {
	args := m.Called(envID, adminID)
	return args.Error(0)
}

func (m *MockDevEnvironmentService) GetEnvironmentAdmins(envID uint) ([]database.Admin, error) {
	args := m.Called(envID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Admin), args.Error(1)
}

func (m *MockDevEnvironmentService) CanAdminAccessEnvironment(envID, adminID uint) (bool, error) {
	args := m.Called(envID, adminID)
	return args.Bool(0), args.Error(1)
}