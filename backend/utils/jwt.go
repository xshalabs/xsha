package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	AdminID uint `json:"admin_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(adminID uint, secret string) (string, error) {
	expirationTime := Now().Add(24 * time.Hour)
	jti := uuid.New().String()

	claims := &Claims{
		AdminID: adminID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GetTokenExpiration(tokenString, secret string) (time.Time, error) {
	claims, err := ValidateJWT(tokenString, secret)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, fmt.Errorf("token has no expiration time")
	}

	return claims.ExpiresAt.Time, nil
}

func ExtractTokenFromAuthHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header format error, should start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

func GetTokenID(tokenString, secret string) (string, error) {
	claims, err := ValidateJWT(tokenString, secret)
	if err != nil {
		return "", err
	}

	if claims.ID == "" {
		return "", fmt.Errorf("token has no ID")
	}

	return claims.ID, nil
}
