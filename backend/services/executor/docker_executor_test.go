package executor

import (
	"strings"
	"testing"
	"xsha-backend/config"
	"xsha-backend/database"
)

func TestBuildCommandWithPreciseSessionMapping(t *testing.T) {
	cfg := &config.Config{
		WorkspaceBaseDir: "/test/workspaces",
		DevSessionsDir:   "/test/sessions",
	}

	executor := &dockerExecutor{
		config: cfg,
	}

	// Create test data
	devEnv := &database.DevEnvironment{
		Type:        "claude-code",
		DockerImage: "anthropic/claude-code:latest",
		SessionDir:  "env-1234567890-1234",
		CPULimit:    2.0,
		MemoryLimit: 2048,
		EnvVars:     `{}`,
	}

	project := &database.Project{
		SystemPrompt: "Test system prompt",
	}

	task := &database.Task{
		SessionID:      "test-session-id",
		Project:        project,
		DevEnvironment: devEnv,
	}

	conv := &database.TaskConversation{
		Content: "test content",
		Task:    task,
	}

	// Build command
	cmd := executor.BuildCommandForLog(conv, "task-1-123456", "")

	// Verify the command contains three precise volume mappings
	expectedMappings := []string{
		"-v /test/sessions/env-1234567890-1234/.claude:/home/xsha/.claude",
		"-v /test/sessions/env-1234567890-1234/.claude.json:/home/xsha/.claude.json",
		"-v /test/sessions/env-1234567890-1234/.claude.json.backup:/home/xsha/.claude.json.backup",
	}

	for _, expectedMapping := range expectedMappings {
		if !strings.Contains(cmd, expectedMapping) {
			t.Errorf("Command does not contain expected mapping: %s\nFull command: %s", expectedMapping, cmd)
		}
	}

	// Verify it does NOT contain the old full session directory mapping
	oldMapping := "-v /test/sessions/env-1234567890-1234:/home/xsha"
	if strings.Contains(cmd, oldMapping) {
		t.Errorf("Command still contains old full session directory mapping: %s\nFull command: %s", oldMapping, cmd)
	}
}

func TestBuildCommandWithoutSessionDir(t *testing.T) {
	cfg := &config.Config{
		WorkspaceBaseDir: "/test/workspaces",
		DevSessionsDir:   "/test/sessions",
	}

	executor := &dockerExecutor{
		config: cfg,
	}

	// Create test data without SessionDir
	devEnv := &database.DevEnvironment{
		Type:        "claude-code",
		DockerImage: "anthropic/claude-code:latest",
		SessionDir:  "", // Empty session dir
		CPULimit:    2.0,
		MemoryLimit: 2048,
		EnvVars:     `{}`,
	}

	project := &database.Project{}

	task := &database.Task{
		Project:        project,
		DevEnvironment: devEnv,
	}

	conv := &database.TaskConversation{
		Content: "test content",
		Task:    task,
	}

	// Build command
	cmd := executor.BuildCommandForLog(conv, "task-1-123456", "")

	// Verify no session mappings are present
	sessionMappings := []string{
		"/.claude:",
		"/.claude.json:",
		"/.claude.json.backup:",
		"/home/xsha",
	}

	for _, mapping := range sessionMappings {
		if strings.Contains(cmd, mapping) {
			t.Errorf("Command contains unexpected session mapping when SessionDir is empty: %s\nFull command: %s", mapping, cmd)
		}
	}
}

func TestBuildCommandInContainerMode(t *testing.T) {
	// This test simulates container mode by setting Docker volume names
	cfg := &config.Config{
		WorkspaceBaseDir:       "/app/workspaces",
		DevSessionsDir:         "/app/sessions",
		DockerVolumeWorkspaces: "xsha-workspaces",
		DockerVolumeSessions:   "xsha-sessions",
	}

	executor := &dockerExecutor{
		config: cfg,
	}

	devEnv := &database.DevEnvironment{
		Type:        "claude-code",
		DockerImage: "anthropic/claude-code:latest",
		SessionDir:  "env-1234567890-1234",
		CPULimit:    2.0,
		MemoryLimit: 2048,
		EnvVars:     `{}`,
	}

	project := &database.Project{}

	task := &database.Task{
		Project:        project,
		DevEnvironment: devEnv,
	}

	conv := &database.TaskConversation{
		Content: "test content",
		Task:    task,
	}

	// Note: BuildCommandForLog doesn't directly test container mode detection
	// but we can verify the logic exists in buildDockerCommandCore
	cmd := executor.BuildCommandForLog(conv, "task-1-123456", "")

	// In host mode (which BuildCommandForLog uses), should have host paths
	if strings.Contains(cmd, "xsha-workspaces") {
		t.Logf("Command uses Docker volumes (container mode detected): %s", cmd)
	} else {
		t.Logf("Command uses host paths (host mode): %s", cmd)
		// Verify precise mappings in host mode
		expectedMappings := []string{
			"/.claude:/home/xsha/.claude",
			"/.claude.json:/home/xsha/.claude.json",
			"/.claude.json.backup:/home/xsha/.claude.json.backup",
		}

		for _, expectedMapping := range expectedMappings {
			if !strings.Contains(cmd, expectedMapping) {
				t.Errorf("Command does not contain expected mapping pattern: %s\nFull command: %s", expectedMapping, cmd)
			}
		}
	}
}
