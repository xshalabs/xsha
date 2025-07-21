package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptAES 使用AES-GCM加密数据
func EncryptAES(plaintext, key string) (string, error) {
	// 确保密钥长度为32字节
	normalizedKey := normalizeAESKey(key)

	block, err := aes.NewCipher([]byte(normalizedKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密并合并nonce和密文
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES 使用AES-GCM解密数据
func DecryptAES(ciphertext, key string) (string, error) {
	normalizedKey := normalizeAESKey(key)

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(normalizedKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// normalizeAESKey 标准化AES密钥为32字节
func normalizeAESKey(key string) string {
	if len(key) >= 32 {
		return key[:32]
	}
	// 密钥不足32字节时用0填充
	normalized := make([]byte, 32)
	copy(normalized, []byte(key))
	return string(normalized)
}

// GenerateAESKey 生成32字节的AES密钥（工具函数）
func GenerateAESKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
