package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 定义JWT声明结构
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成JWT token
func GenerateJWT(username, secret string) (string, error) {
	// 设置token过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用指定的签名方法创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT 验证JWT token并返回claims
func ValidateJWT(tokenString, secret string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 检查token是否有效
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的token")
}

// GetTokenExpiration 获取token的过期时间
func GetTokenExpiration(tokenString, secret string) (time.Time, error) {
	claims, err := ValidateJWT(tokenString, secret)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, fmt.Errorf("token没有过期时间")
	}

	return claims.ExpiresAt.Time, nil
}

// ExtractTokenFromAuthHeader 从Authorization header中提取token
func ExtractTokenFromAuthHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("缺少Authorization header")
	}

	// 检查是否以"Bearer "开头
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("Authorization header格式错误，应该以'Bearer '开头")
	}

	// 提取token部分
	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("token为空")
	}

	return token, nil
}
