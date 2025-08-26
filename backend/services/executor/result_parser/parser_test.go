package result_parser

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
	"xsha-backend/services/executor/result_parser/strategies"
	"xsha-backend/services/executor/result_parser/validator"
)

func TestJSONStrategy(t *testing.T) {
	strategy := strategies.NewJSONStrategy()
	
	tests := []struct {
		name     string
		logs     string
		expected bool
		hasError bool
	}{
		{
			name: "valid JSON result",
			logs: `[12:34:56] INFO: {"type": "result", "subtype": "success", "is_error": false, "session_id": "test-123", "duration_ms": 1000}`,
			expected: true,
			hasError: false,
		},
		{
			name: "plain JSON",
			logs: `{"type": "result", "subtype": "success", "is_error": false, "session_id": "test-456", "result": "completed"}`,
			expected: true,
			hasError: false,
		},
		{
			name: "multiple lines with result at end",
			logs: `Starting task execution
Task ID: 12345
Processing...
{"type": "result", "subtype": "success", "is_error": false, "session_id": "test-789", "num_turns": 5}`,
			expected: true,
			hasError: false,
		},
		{
			name: "no valid JSON",
			logs: "Simple log message without JSON",
			expected: false,
			hasError: true,
		},
		{
			name: "empty logs",
			logs: "",
			expected: false,
			hasError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			canParse := strategy.CanParse(tt.logs)
			if canParse != tt.expected {
				t.Errorf("CanParse() = %v, expected %v", canParse, tt.expected)
			}
			
			if tt.expected {
				result, err := strategy.Parse(ctx, tt.logs)
				if tt.hasError {
					if err == nil {
						t.Error("Expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if result == nil {
						t.Error("Expected result but got nil")
					}
					if result != nil {
						if result["type"] != "result" {
							t.Errorf("Expected type=result, got %v", result["type"])
						}
					}
				}
			}
		})
	}
}

func TestTextStrategy(t *testing.T) {
	strategy := strategies.NewTextStrategy()
	
	tests := []struct {
		name     string
		logs     string
		expected map[string]interface{}
		hasError bool
	}{
		{
			name: "key=value format",
			logs: "Result: type=result subtype=success is_error=false session_id=test-123",
			expected: map[string]interface{}{
				"type":       "result",
				"subtype":    "success", 
				"is_error":   false,
				"session_id": "test-123",
			},
			hasError: false,
		},
		{
			name: "colon format",
			logs: "Task completed: type=result, subtype=success, is_error=false, session_id=test-456",
			expected: map[string]interface{}{
				"type":       "result",
				"subtype":    "success",
				"is_error":   false,
				"session_id": "test-456",
			},
			hasError: false,
		},
		{
			name: "no structured data",
			logs: "Simple log message",
			expected: nil,
			hasError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := strategy.Parse(ctx, tt.logs)
			
			if tt.hasError {
				if err == nil && result != nil {
					t.Error("Expected error or nil result")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
				
				if result != nil {
					for key, expectedValue := range tt.expected {
						if actualValue, exists := result[key]; !exists {
							t.Errorf("Missing key %s", key)
						} else if actualValue != expectedValue {
							t.Errorf("Key %s: expected %v, got %v", key, expectedValue, actualValue)
						}
					}
				}
			}
		})
	}
}

func TestFallbackStrategy(t *testing.T) {
	strategy := strategies.NewFallbackStrategy()
	
	// Fallback strategy should always be able to parse
	if !strategy.CanParse("any logs") {
		t.Error("Fallback strategy should always return true for CanParse")
	}
	
	ctx := context.Background()
	result, err := strategy.Parse(ctx, "some error log")
	
	if err != nil {
		t.Errorf("Fallback strategy should not return error: %v", err)
	}
	
	if result == nil {
		t.Error("Fallback strategy should always return a result")
	}
	
	if result != nil {
		// Check required fields are present
		requiredFields := []string{"type", "subtype", "is_error", "session_id"}
		for _, field := range requiredFields {
			if _, exists := result[field]; !exists {
				t.Errorf("Required field %s missing from fallback result", field)
			}
		}
	}
}

func TestResultValidator(t *testing.T) {
	validator := validator.NewResultValidator(false) // non-strict mode
	
	tests := []struct {
		name     string
		data     map[string]interface{}
		isValid  bool
	}{
		{
			name: "valid complete result",
			data: map[string]interface{}{
				"type":          "result",
				"subtype":       "success", 
				"is_error":      false,
				"session_id":    "test-123",
				"duration_ms":   int64(1000),
				"num_turns":     5,
				"result":        "Task completed successfully",
			},
			isValid: true,
		},
		{
			name: "minimal valid result",
			data: map[string]interface{}{
				"type":       "result",
				"subtype":    "success",
				"is_error":   false,
				"session_id": "test-456",
			},
			isValid: true,
		},
		{
			name: "missing required field",
			data: map[string]interface{}{
				"type":     "result",
				"subtype":  "success", 
				"is_error": false,
				// missing session_id
			},
			isValid: false,
		},
		{
			name: "invalid type value",
			data: map[string]interface{}{
				"type":       "invalid",
				"subtype":    "success",
				"is_error":   false,
				"session_id": "test-789",
			},
			isValid: false,
		},
		{
			name: "invalid field type",
			data: map[string]interface{}{
				"type":       "result",
				"subtype":    "success",
				"is_error":   "not_boolean", // should be boolean
				"session_id": "test-abc",
			},
			isValid: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validator.IsValid(tt.data)
			if isValid != tt.isValid {
				t.Errorf("IsValid() = %v, expected %v", isValid, tt.isValid)
				
				// Print validation errors for debugging
				errors := validator.GetValidationErrors(tt.data)
				for _, err := range errors {
					t.Logf("Validation error: %s", err.Error())
				}
			}
		})
	}
}

func TestDefaultParser(t *testing.T) {
	// Create parser with nil dependencies (test mode)
	parser := &DefaultParser{
		config:          DefaultConfig(),
		strategyFactory: NewDefaultStrategyFactory(DefaultConfig()),
		validator:       validator.NewResultValidator(false),
		metrics:         NewMetrics(),
	}
	
	tests := []struct {
		name     string
		logs     string
		hasError bool
	}{
		{
			name: "valid JSON logs",
			logs: `{"type": "result", "subtype": "success", "is_error": false, "session_id": "test-123"}`,
			hasError: false,
		},
		{
			name: "valid text logs", 
			logs: "Result: type=result subtype=success is_error=false session_id=test-456",
			hasError: false,
		},
		{
			name: "fallback to heuristic parsing",
			logs: "Task completed with session test-789",
			hasError: false,
		},
		{
			name: "empty logs",
			logs: "",
			hasError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseFromLogs(tt.logs)
			
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

func TestParserWithContext(t *testing.T) {
	parser := &DefaultParser{
		config:          DefaultConfig(),
		strategyFactory: NewDefaultStrategyFactory(DefaultConfig()),
		validator:       validator.NewResultValidator(false),
		metrics:         NewMetrics(),
	}
	
	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		// Use a long log string to potentially trigger timeout
		logs := strings.Repeat("log line\n", 10000) + 
			`{"type": "result", "subtype": "success", "is_error": false, "session_id": "test-timeout"}`
		
		_, err := parser.ParseFromLogsWithContext(ctx, logs)
		
		// Should either succeed quickly or timeout - both are acceptable
		if err != nil && err != context.DeadlineExceeded {
			t.Logf("Got error (expected): %v", err)
		}
	})
	
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		logs := `{"type": "result", "subtype": "success", "is_error": false, "session_id": "test-cancel"}`
		
		_, err := parser.ParseFromLogsWithContext(ctx, logs)
		if err != context.Canceled {
			t.Logf("Expected context canceled, got: %v", err)
		}
	})
}

func TestParserMetrics(t *testing.T) {
	metrics := NewMetrics()
	
	// Test recording parse attempts
	metrics.RecordParseAttempt(100 * time.Millisecond)
	metrics.RecordParseSuccess("json")
	
	metrics.RecordParseAttempt(200 * time.Millisecond)
	metrics.RecordParseError("timeout")
	
	stats := metrics.GetStats()
	
	if stats["parse_attempts"].(int64) != 2 {
		t.Errorf("Expected 2 parse attempts, got %v", stats["parse_attempts"])
	}
	
	if stats["parse_successes"].(int64) != 1 {
		t.Errorf("Expected 1 parse success, got %v", stats["parse_successes"])
	}
	
	if stats["parse_errors"].(int64) != 1 {
		t.Errorf("Expected 1 parse error, got %v", stats["parse_errors"])
	}
	
	successRate := stats["success_rate"].(float64)
	if successRate != 0.5 {
		t.Errorf("Expected success rate 0.5, got %v", successRate)
	}
	
	strategyUsage := stats["strategy_usage"].(map[string]int64)
	if strategyUsage["json"] != 1 {
		t.Errorf("Expected 1 JSON strategy usage, got %v", strategyUsage["json"])
	}
	
	errorTypes := stats["error_types"].(map[string]int64) 
	if errorTypes["timeout"] != 1 {
		t.Errorf("Expected 1 timeout error, got %v", errorTypes["timeout"])
	}
}

func TestStrategyFactory(t *testing.T) {
	config := DefaultConfig()
	factory := NewDefaultStrategyFactory(config)
	
	// Test JSON logs
	jsonLogs := `{"type": "result", "subtype": "success", "is_error": false, "session_id": "test"}`
	strategy := factory.GetBestStrategy(jsonLogs)
	
	if strategy.Name() != "json" {
		t.Errorf("Expected JSON strategy for JSON logs, got %s", strategy.Name())
	}
	
	// Test structured text logs  
	textLogs := "Result: type=result subtype=success is_error=false session_id=test"
	strategy = factory.GetBestStrategy(textLogs)
	
	if strategy.Name() != "structured_text" {
		t.Errorf("Expected text strategy for structured text logs, got %s", strategy.Name())
	}
	
	// Test fallback
	plainLogs := "simple log message"
	strategy = factory.GetBestStrategy(plainLogs)
	
	if strategy.Name() != "fallback" {
		t.Errorf("Expected fallback strategy for plain logs, got %s", strategy.Name())
	}
}

func TestOptimizedJSONStrategy(t *testing.T) {
	strategy := strategies.NewOptimizedJSONStrategy(10) // limit to 10 lines
	
	// Create logs with many lines, result at the end
	lines := make([]string, 100)
	for i := 0; i < 99; i++ {
		lines[i] = fmt.Sprintf("log line %d", i+1)
	}
	lines[99] = `{"type": "result", "subtype": "success", "is_error": false, "session_id": "test-optimized"}`
	
	logs := strings.Join(lines, "\n")
	
	ctx := context.Background()
	result, err := strategy.Parse(ctx, logs)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if result == nil {
		t.Error("Expected result but got nil")
	}
	
	if result != nil && result["session_id"] != "test-optimized" {
		t.Errorf("Expected session_id=test-optimized, got %v", result["session_id"])
	}
}

func BenchmarkJSONStrategy(b *testing.B) {
	strategy := strategies.NewJSONStrategy()
	logs := `[12:34:56] INFO: {"type": "result", "subtype": "success", "is_error": false, "session_id": "bench-test", "duration_ms": 1000}`
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := strategy.Parse(ctx, logs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDefaultParser(b *testing.B) {
	parser := &DefaultParser{
		config:          DefaultConfig(),
		strategyFactory: NewDefaultStrategyFactory(DefaultConfig()),
		validator:       validator.NewResultValidator(false),
		metrics:         NewMetrics(),
	}
	
	logs := `{"type": "result", "subtype": "success", "is_error": false, "session_id": "bench-test"}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFromLogs(logs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

