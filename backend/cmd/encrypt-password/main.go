package main

import (
	"fmt"
	"log"
	"os"
	"sleep0-backend/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  go run cmd/encrypt-password/main.go <password> [aes-key]")
		fmt.Println("示例:")
		fmt.Println("  go run cmd/encrypt-password/main.go admin123")
		fmt.Println("  go run cmd/encrypt-password/main.go admin123 my-custom-32-byte-key-here-123")
		os.Exit(1)
	}

	password := os.Args[1]

	var aesKey string
	if len(os.Args) >= 3 {
		aesKey = os.Args[2]
	} else {
		// 生成随机密钥
		var err error
		aesKey, err = utils.GenerateAESKey()
		if err != nil {
			log.Fatalf("生成AES密钥失败: %v", err)
		}
		fmt.Printf("自动生成AES密钥: %s\n", aesKey)
	}

	encrypted, err := utils.EncryptAES(password, aesKey)
	if err != nil {
		log.Fatalf("密码加密失败: %v", err)
	}

	fmt.Printf("\n=== 加密结果 ===\n")
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("加密密码: %s\n", encrypted)
	fmt.Printf("\n=== 环境变量设置 ===\n")
	fmt.Printf("export SLEEP0_ADMIN_PASS=\"%s\"\n", encrypted)
	fmt.Printf("export SLEEP0_AES_KEY=\"%s\"\n", aesKey)
	fmt.Printf("\n=== Docker环境变量 ===\n")
	fmt.Printf("- SLEEP0_ADMIN_PASS=%s\n", encrypted)
	fmt.Printf("- SLEEP0_AES_KEY=%s\n", aesKey)
}
