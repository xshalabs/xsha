package notifiers

import (
	"fmt"
	"xsha-backend/database"
)

// TestI18nFormats tests the i18n formatting functions for different languages
func TestI18nFormats() {
	fmt.Println("=== Testing Notification I18n Formatting ===")

	// Test English
	fmt.Println("\n--- English (en-US) ---")
	englishMsg := FormatNotificationMessage(
		"Test Task",
		"This is a test task content for verification",
		database.ConversationStatusSuccess,
		"en-US",
	)
	fmt.Println(englishMsg)

	englishTest := FormatTestMessage("en-US")
	fmt.Println("Test message:", englishTest)

	// Test Chinese
	fmt.Println("\n--- Chinese (zh-CN) ---")
	chineseMsg := FormatNotificationMessage(
		"测试任务",
		"这是一个用于验证的测试任务内容",
		database.ConversationStatusFailed,
		"zh-CN",
	)
	fmt.Println(chineseMsg)

	chineseTest := FormatTestMessage("zh-CN")
	fmt.Println("测试消息:", chineseTest)

	fmt.Println("\n=== Test Complete ===")
}
