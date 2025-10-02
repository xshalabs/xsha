package executor

import (
	"encoding/json"
	"fmt"
	"strings"
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

// GenerateMCPSetupScript generates a shell script to configure MCPs for a task conversation
// The script will:
// 1. List current MCPs
// 2. Add missing MCPs that should be enabled
// 3. Optionally remove MCPs that are no longer needed
func (m *MCPManager) GenerateMCPSetupScript(conversationID uint) (string, error) {
	// Get MCPs for this conversation
	mcps, err := m.mcpService.GetMCPsForTaskConversation(conversationID, m.taskConvRepo)
	if err != nil {
		return "", fmt.Errorf("failed to get MCPs for conversation: %v", err)
	}

	// If no MCPs are configured, return empty script
	if len(mcps) == 0 {
		utils.Info("No MCPs configured for conversation", "conversation_id", conversationID)
		return "", nil
	}

	// Build the script
	var scriptParts []string

	// Add header comment
	scriptParts = append(scriptParts, "# MCP Configuration Setup")
	scriptParts = append(scriptParts, "echo '=== Configuring MCPs ==='")

	// Add each MCP
	for _, mcp := range mcps {
		// Validate and escape the config JSON
		configJSON, err := m.escapeJSONForShell(mcp.Config)
		if err != nil {
			utils.Warn("Failed to escape MCP config JSON, skipping",
				"mcp_name", mcp.Name,
				"conversation_id", conversationID,
				"error", err)
			continue
		}

		// Generate add command
		// Format: claude mcp add-json <name> '<json>' --scope local
		addCommand := fmt.Sprintf(
			"claude mcp add-json %s '%s' --scope local 2>&1 | grep -v 'already exists' || true",
			m.escapeShellArg(mcp.Name),
			configJSON,
		)

		scriptParts = append(scriptParts, fmt.Sprintf("echo 'Adding MCP: %s'", mcp.Name))
		scriptParts = append(scriptParts, addCommand)
	}

	scriptParts = append(scriptParts, "echo '=== MCP Configuration Complete ==='")

	// Join all parts with newlines and semicolons for sequential execution
	script := strings.Join(scriptParts, " && ")

	utils.Info("Generated MCP setup script",
		"conversation_id", conversationID,
		"mcp_count", len(mcps))

	return script, nil
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
