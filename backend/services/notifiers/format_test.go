package notifiers

import (
	"strings"
	"testing"
	"xsha-backend/database"
)

func TestFormatNotificationMessage(t *testing.T) {
	// Test with project name
	result := FormatNotificationMessage(
		"Test Task",
		"Test content",
		"Test Project",
		database.ConversationStatusSuccess,
		"en-US",
	)

	// Check if the result contains expected elements
	if !strings.Contains(result, "Test Task") {
		t.Errorf("Expected result to contain task title")
	}
	if !strings.Contains(result, "Test Project") {
		t.Errorf("Expected result to contain project name")
	}
	if !strings.Contains(result, "Project:") {
		t.Errorf("Expected result to contain project label")
	}

	// Test without project name
	result2 := FormatNotificationMessage(
		"Test Task",
		"Test content",
		"",
		database.ConversationStatusSuccess,
		"en-US",
	)

	// Should not contain project line when project name is empty
	if strings.Contains(result2, "Project:") {
		t.Errorf("Expected result to not contain project label when project name is empty")
	}
}

func TestFormatNotificationMessageChinese(t *testing.T) {
	result := FormatNotificationMessage(
		"测试任务",
		"测试内容",
		"测试项目",
		database.ConversationStatusFailed,
		"zh-CN",
	)

	if !strings.Contains(result, "测试任务") {
		t.Errorf("Expected result to contain Chinese task title")
	}
	if !strings.Contains(result, "测试项目") {
		t.Errorf("Expected result to contain Chinese project name")
	}
	if !strings.Contains(result, "项目:") {
		t.Errorf("Expected result to contain Chinese project label")
	}
}
