package services

import (
	"os"
	"path/filepath"
	"testing"
	"xsha-backend/config"
	"xsha-backend/repository"
)

func TestInitializeClaudeSessionStructure(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	sessionDir := filepath.Join(tempDir, "test-session")

	// Create session directory
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create test session directory: %v", err)
	}

	// Create a minimal service instance for testing
	cfg := &config.Config{
		DevSessionsDir: tempDir,
	}
	service := &devEnvironmentService{
		config: cfg,
	}

	// Test initializeClaudeSessionStructure
	err := service.initializeClaudeSessionStructure(sessionDir)
	if err != nil {
		t.Fatalf("initializeClaudeSessionStructure failed: %v", err)
	}

	// Verify .claude directory exists
	claudeDir := filepath.Join(sessionDir, ".claude")
	if stat, err := os.Stat(claudeDir); err != nil || !stat.IsDir() {
		t.Errorf(".claude directory was not created or is not a directory")
	}

	// Verify .claude.json file exists
	claudeJSON := filepath.Join(sessionDir, ".claude.json")
	if stat, err := os.Stat(claudeJSON); err != nil || stat.IsDir() {
		t.Errorf(".claude.json file was not created or is a directory")
	}

	// Verify .claude.json.backup file exists
	claudeBackup := filepath.Join(sessionDir, ".claude.json.backup")
	if stat, err := os.Stat(claudeBackup); err != nil || stat.IsDir() {
		t.Errorf(".claude.json.backup file was not created or is a directory")
	}

	// Verify .claude.json content is valid JSON
	content, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Errorf("Failed to read .claude.json: %v", err)
	}
	if string(content) != "{}" {
		t.Errorf("Expected .claude.json to contain '{}', got: %s", string(content))
	}

	// Verify .claude.json.backup content is valid JSON
	backupContent, err := os.ReadFile(claudeBackup)
	if err != nil {
		t.Errorf("Failed to read .claude.json.backup: %v", err)
	}
	if string(backupContent) != "{}" {
		t.Errorf("Expected .claude.json.backup to contain '{}', got: %s", string(backupContent))
	}
}

func TestGenerateSessionDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	cfg := &config.Config{
		DevSessionsDir: tempDir,
	}

	// Create a minimal service instance
	service := &devEnvironmentService{
		repo:     &mockDevEnvRepo{},
		taskRepo: &mockTaskRepo{},
		config:   cfg,
	}

	// Generate session directory
	dirName, err := service.generateSessionDir()
	if err != nil {
		t.Fatalf("generateSessionDir failed: %v", err)
	}

	// Verify directory name is not empty
	if dirName == "" {
		t.Error("generateSessionDir returned empty directory name")
	}

	// Verify absolute path exists
	absolutePath := filepath.Join(tempDir, dirName)
	if stat, err := os.Stat(absolutePath); err != nil || !stat.IsDir() {
		t.Errorf("Session directory was not created: %v", err)
	}

	// Verify Claude session structure was initialized
	claudeDir := filepath.Join(absolutePath, ".claude")
	if _, err := os.Stat(claudeDir); err != nil {
		t.Errorf(".claude directory was not created: %v", err)
	}

	claudeJSON := filepath.Join(absolutePath, ".claude.json")
	if _, err := os.Stat(claudeJSON); err != nil {
		t.Errorf(".claude.json file was not created: %v", err)
	}

	claudeBackup := filepath.Join(absolutePath, ".claude.json.backup")
	if _, err := os.Stat(claudeBackup); err != nil {
		t.Errorf(".claude.json.backup file was not created: %v", err)
	}
}

// Mock repositories for testing
type mockDevEnvRepo struct {
	repository.DevEnvironmentRepository
}

type mockTaskRepo struct {
	repository.TaskRepository
}
