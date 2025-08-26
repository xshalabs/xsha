package result_parser

import (
	"context"
	"testing"
	"time"
)

func TestPlanModeStrategy(t *testing.T) {
	tests := []struct {
		name     string
		logs     string
		expected bool
	}{
		{
			name: "valid_plan_mode_result",
			logs: `{"type":"assistant","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-sonnet-4-20250514","content":[{"type":"tool_use","id":"toolu_123","name":"ExitPlanMode","input":{"plan":"## Plan to Remove PC Registration Functionality\n\nI'll remove the registration functionality from the PC version by:\n\n### 1. Remove RegisterDialog Component\n- Delete the entire /src/components/register-dialog/ folder"}}],"stop_reason":"tool_use","usage":{"input_tokens":0,"output_tokens":592}},"session_id":"f36b7001-6613-4ddf-b296-bd6d7d34c106"}`,
			expected: true,
		},
		{
			name: "regular_result",
			logs: `{"type":"result","subtype":"success","is_error":false,"duration_ms":21860,"session_id":"0f432ca5-25b5-4215-b15b-53bf6e94e648"}`,
			expected: false,
		},
		{
			name: "empty_logs",
			logs: "",
			expected: false,
		},
	}

	parser := NewDefaultParser(nil, nil, nil)
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			result, err := parser.ParseFromLogsWithContext(ctx, tt.logs)
			
			if tt.expected {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
				if result == nil {
					t.Error("Expected result but got nil")
					return
				}
				
				// Verify plan mode specific fields
				if typeVal, ok := result["type"].(string); !ok || typeVal != "result" {
					t.Errorf("Expected type 'result', got: %v", result["type"])
				}
				if subtypeVal, ok := result["subtype"].(string); !ok || subtypeVal != "plan_mode" {
					t.Errorf("Expected subtype 'plan_mode', got: %v", result["subtype"])
				}
				if planContent, ok := result["result"].(string); !ok || planContent == "" {
					t.Errorf("Expected plan content, got: %v", result["result"])
				}
				if sessionID, ok := result["session_id"].(string); !ok || sessionID == "" {
					t.Errorf("Expected session_id, got: %v", result["session_id"])
				}
			} else {
				// For non-plan mode logs, we might still get results, just not plan mode ones
				if result != nil {
					if subtypeVal, ok := result["subtype"].(string); ok && subtypeVal == "plan_mode" {
						t.Error("Should not detect plan mode for non-plan mode logs")
					}
				}
			}
		})
	}
}

func TestPlanModeDetection(t *testing.T) {
	tests := []struct {
		name     string
		logs     string
		expected bool
	}{
		{
			name: "contains_plan_mode_indicators_with_valid_structure",
			logs: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"ExitPlanMode","input":{"plan":"test plan"}}]}}`,
			expected: true,
		},
		{
			name: "regular_json_logs",
			logs: `{"type":"result","subtype":"success"}`,
			expected: false,
		},
		{
			name: "text_with_plan_mode_keywords",
			logs: `Task completed using ExitPlanMode tool`,
			expected: false, // Need at least 2 indicators
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the containsPlanMode helper function indirectly through strategy selection
			factory := NewDefaultStrategyFactory(DefaultConfig())
			strategy := factory.GetBestStrategy(tt.logs)
			
			isPlanMode := (strategy.Name() == "plan_mode")
			if isPlanMode != tt.expected {
				t.Errorf("Expected plan mode detection: %v, got: %v (strategy: %s)", tt.expected, isPlanMode, strategy.Name())
			}
		})
	}
}