package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"xsha-backend/config"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/utils"
)

// MCPManager handles MCP configuration management for task execution
type MCPManager struct {
	mcpService   services.MCPService
	taskConvRepo repository.TaskConversationRepository
}

// NewMCPManager creates a new MCP manager instance
func NewMCPManager(mcpService services.MCPService, taskConvRepo repository.TaskConversationRepository) *MCPManager {
	return &MCPManager{
		mcpService:   mcpService,
		taskConvRepo: taskConvRepo,
	}
}

// escapeJSONForShell escapes a JSON string for safe inclusion in a shell script
// It validates the JSON and ensures it can be safely wrapped in single quotes
func (m *MCPManager) escapeJSONForShell(jsonStr string) (string, error) {
	// First validate that it's valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	// Re-marshal to ensure consistent formatting and escape any special characters
	compactJSON, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Escape single quotes in the JSON by replacing ' with '\''
	escaped := strings.ReplaceAll(string(compactJSON), "'", "'\\''")

	return escaped, nil
}

// escapeShellArg escapes a string for safe use as a shell argument
func (m *MCPManager) escapeShellArg(arg string) string {
	// For simple alphanumeric strings with hyphens and underscores, no escaping needed
	// For others, we'll wrap in double quotes and escape special characters
	needsEscape := false
	for _, char := range arg {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' ||
			char == '-') {
			needsEscape = true
			break
		}
	}

	if !needsEscape {
		return arg
	}

	// Escape special characters and wrap in double quotes
	escaped := strings.ReplaceAll(arg, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "$", "\\$")
	escaped = strings.ReplaceAll(escaped, "`", "\\`")

	return fmt.Sprintf("\"%s\"", escaped)
}

// GenerateMCPSetupScriptFile generates a shell script file for MCP configuration
// The script file will be saved to __xsha_workspace/mcp_setup_{conversationID}.sh
// Returns the relative path to the script file (relative to workspace root)
func (m *MCPManager) GenerateMCPSetupScriptFile(conversationID uint, workspacePath string, cfg *config.Config) (string, error) {
	// Get MCPs for this conversation
	mcps, err := m.mcpService.GetMCPsForTaskConversation(conversationID, m.taskConvRepo)
	if err != nil {
		return "", fmt.Errorf("failed to get MCPs for conversation: %v", err)
	}

	// If no MCPs are configured, return empty string
	if len(mcps) == 0 {
		return "", nil
	}

	// Convert relative workspace path to absolute if needed
	absoluteWorkspacePath := workspacePath
	if !filepath.IsAbs(workspacePath) {
		absoluteWorkspacePath = filepath.Join(cfg.WorkspaceBaseDir, workspacePath)
	}

	// Create __xsha_workspace directory in workspace
	xshaDir := filepath.Join(absoluteWorkspacePath, "__xsha_workspace")
	if err := os.MkdirAll(xshaDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create __xsha_workspace directory: %v", err)
	}

	// Build the script content
	var scriptLines []string

	// Add shebang
	scriptLines = append(scriptLines, "#!/bin/sh")
	scriptLines = append(scriptLines, "")

	// Add header comment
	scriptLines = append(scriptLines, "# MCP Configuration Setup Script")
	scriptLines = append(scriptLines, "# Generated for conversation ID: "+fmt.Sprintf("%d", conversationID))
	scriptLines = append(scriptLines, "")
	scriptLines = append(scriptLines, "echo '=== Configuring MCPs ==='")

	// Add each MCP
	for _, mcp := range mcps {
		// Validate and escape the config JSON
		configJSON, err := m.escapeJSONForShell(mcp.Config)
		if err != nil {
			utils.Error("Failed to escape MCP config JSON, skipping",
				"mcp_name", mcp.Name,
				"conversation_id", conversationID,
				"error", err)
			continue
		}

		// Check if MCP exists and remove it
		scriptLines = append(scriptLines, "")
		scriptLines = append(scriptLines, fmt.Sprintf("# Check and remove existing MCP: %s", mcp.Name))
		scriptLines = append(scriptLines, fmt.Sprintf("if claude mcp get %s >/dev/null 2>&1; then", m.escapeShellArg(mcp.Name)))
		scriptLines = append(scriptLines, fmt.Sprintf("  echo 'Removing existing MCP: %s'", mcp.Name))
		scriptLines = append(scriptLines, fmt.Sprintf("  claude mcp remove %s --scope local", m.escapeShellArg(mcp.Name)))
		scriptLines = append(scriptLines, "fi")
		scriptLines = append(scriptLines, "")

		// Generate add command
		scriptLines = append(scriptLines, fmt.Sprintf("echo 'Adding MCP: %s'", mcp.Name))
		addCommand := fmt.Sprintf(
			"claude mcp add-json %s '%s' --scope local",
			m.escapeShellArg(mcp.Name),
			configJSON,
		)
		scriptLines = append(scriptLines, addCommand)
	}

	scriptLines = append(scriptLines, "")
	scriptLines = append(scriptLines, "echo '=== MCP Configuration Complete ==='")

	// Join all lines with newlines
	scriptContent := strings.Join(scriptLines, "\n")

	// Generate script filename
	scriptFileName := fmt.Sprintf("mcp_setup_%d.sh", conversationID)
	scriptFilePath := filepath.Join(xshaDir, scriptFileName)

	// Write script to file
	if err := os.WriteFile(scriptFilePath, []byte(scriptContent), 0755); err != nil {
		return "", fmt.Errorf("failed to write MCP setup script file: %v", err)
	}

	// Return relative path (relative to workspace root)
	relativePath := filepath.Join("__xsha_workspace", scriptFileName)

	return relativePath, nil
}
