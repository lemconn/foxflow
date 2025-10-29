package grpc

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT 声明结构
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthManager 认证管理器
type AuthManager struct {
	secretKey []byte
}

// NewAuthManager 创建认证管理器
func NewAuthManager() *AuthManager {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		// 默认密钥，生产环境应该使用环境变量
		secretKey = "foxflow-default-secret-key-change-in-production"
	}
	return &AuthManager{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken 生成 JWT token
func (am *AuthManager) GenerateToken(username string) (string, int64, error) {
	// 设置过期时间为 24 小时
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "foxflow",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken 验证 JWT token
func (am *AuthManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新 token
func (am *AuthManager) RefreshToken(tokenString string) (string, int64, error) {
	// 验证当前 token
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return "", 0, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// 生成新的 token
	return am.GenerateToken(claims.Username)
}

// IsTokenExpired 检查 token 是否即将过期（1小时内）
func (am *AuthManager) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return true, err
	}

	// 检查是否在 1 小时内过期
	expiresAt := claims.ExpiresAt.Time
	oneHourFromNow := time.Now().Add(1 * time.Hour)

	return expiresAt.Before(oneHourFromNow), nil
}

// GetTokenExpiry 获取 token 过期时间
func (am *AuthManager) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}
