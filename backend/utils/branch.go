package utils

import (
	"fmt"
	"strings"
	"time"
)

func GenerateWorkBranchName(title, createdBy string) string {
	cleanTitle := strings.ToLower(strings.TrimSpace(title))

	cleanTitle = strings.ReplaceAll(cleanTitle, " ", "-")
	cleanTitle = strings.ReplaceAll(cleanTitle, "_", "-")

	var result strings.Builder
	for _, r := range cleanTitle {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	cleanTitle = result.String()

	if len(cleanTitle) > 30 {
		cleanTitle = cleanTitle[:30]
	}

	cleanTitle = strings.Trim(cleanTitle, "-")

	if cleanTitle == "" {
		cleanTitle = "task"
	}

	timestamp := time.Now().Format("20060102-150405")

	return fmt.Sprintf("xsha/%s/%s-%s", createdBy, cleanTitle, timestamp)
}
