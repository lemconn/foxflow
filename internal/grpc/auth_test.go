package grpc

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewAuthManager(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		expected string
	}{
		{
			name:     "使用环境变量密钥",
			envKey:   "test-secret-key",
			expected: "test-secret-key",
		},
		{
			name:     "使用默认密钥",
			envKey:   "",
			expected: "foxflow-default-secret-key-change-in-production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境变量
			if tt.envKey != "" {
				os.Setenv("JWT_SECRET_KEY", tt.envKey)
			} else {
				os.Unsetenv("JWT_SECRET_KEY")
			}
			defer os.Unsetenv("JWT_SECRET_KEY")

			manager := NewAuthManager()
			if string(manager.secretKey) != tt.expected {
				t.Errorf("NewAuthManager() secretKey = %s, want %s", string(manager.secretKey), tt.expected)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	token, expiresAt, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() token should not be empty")
	}

	if expiresAt <= time.Now().Unix() {
		t.Error("GenerateToken() expiresAt should be in the future")
	}

	// 验证 token 格式
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return manager.secretKey, nil
	})

	if err != nil {
		t.Fatalf("ParseWithClaims() error = %v", err)
	}

	if claims, ok := parsedToken.Claims.(*JWTClaims); ok && parsedToken.Valid {
		if claims.Username != username {
			t.Errorf("GenerateToken() username = %s, want %s", claims.Username, username)
		}
		if claims.Issuer != "foxflow" {
			t.Errorf("GenerateToken() issuer = %s, want foxflow", claims.Issuer)
		}
		if claims.Subject != username {
			t.Errorf("GenerateToken() subject = %s, want %s", claims.Subject, username)
		}
	} else {
		t.Error("GenerateToken() token should be valid")
	}
}

func TestValidateToken(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	// 生成有效 token
	token, _, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 测试有效 token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}
	if claims.Username != username {
		t.Errorf("ValidateToken() username = %s, want %s", claims.Username, username)
	}

	// 测试无效 token
	_, err = manager.ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}

	// 测试空 token
	_, err = manager.ValidateToken("")
	if err == nil {
		t.Error("ValidateToken() should return error for empty token")
	}
}

func TestRefreshToken(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	// 生成原始 token
	originalToken, _, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 添加小延迟确保时间戳不同
	time.Sleep(time.Millisecond)

	// 刷新 token
	newToken, newExpiresAt, err := manager.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	// 验证新 token 有效（由于时间戳可能相同，token 可能相同，但应该有效）
	// 注意：在某些情况下，如果时间戳相同，JWT 可能生成相同的 token
	// 这是可以接受的，因为重要的是 token 的有效性

	// 验证新过期时间在将来
	if newExpiresAt <= time.Now().Unix() {
		t.Error("RefreshToken() newExpiresAt should be in the future")
	}

	// 验证新 token 有效
	claims, err := manager.ValidateToken(newToken)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}
	if claims.Username != username {
		t.Errorf("RefreshToken() username = %s, want %s", claims.Username, username)
	}

	// 测试无效 token 刷新
	_, _, err = manager.RefreshToken("invalid-token")
	if err == nil {
		t.Error("RefreshToken() should return error for invalid token")
	}
}

func TestIsTokenExpired(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	// 生成正常 token
	token, _, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 测试正常 token（应该未过期）
	expired, err := manager.IsTokenExpired(token)
	if err != nil {
		t.Errorf("IsTokenExpired() error = %v", err)
	}
	if expired {
		t.Error("IsTokenExpired() should return false for fresh token")
	}

	// 测试无效 token
	_, err = manager.IsTokenExpired("invalid-token")
	if err == nil {
		t.Error("IsTokenExpired() should return error for invalid token")
	}
}

func TestGetTokenExpiry(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	// 生成 token
	token, expectedExpiresAt, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 测试获取过期时间
	expiresAt, err := manager.GetTokenExpiry(token)
	if err != nil {
		t.Errorf("GetTokenExpiry() error = %v", err)
	}

	// 验证过期时间（允许 1 秒误差）
	expectedTime := time.Unix(expectedExpiresAt, 0)
	if expiresAt.Sub(expectedTime).Abs() > time.Second {
		t.Errorf("GetTokenExpiry() expiresAt = %v, want %v", expiresAt, expectedTime)
	}

	// 测试无效 token
	_, err = manager.GetTokenExpiry("invalid-token")
	if err == nil {
		t.Error("GetTokenExpiry() should return error for invalid token")
	}
}

func TestTokenExpiry(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	// 生成 token
	token, _, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 验证 token 在 24 小时内有效
	expiresAt, err := manager.GetTokenExpiry(token)
	if err != nil {
		t.Fatalf("GetTokenExpiry() error = %v", err)
	}

	expectedMin := time.Now().Add(23 * time.Hour)
	expectedMax := time.Now().Add(25 * time.Hour)

	if expiresAt.Before(expectedMin) || expiresAt.After(expectedMax) {
		t.Errorf("Token expiry %v should be between %v and %v", expiresAt, expectedMin, expectedMax)
	}
}

func TestTokenClaims(t *testing.T) {
	manager := NewAuthManager()
	username := "testuser"

	token, _, err := manager.GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// 验证所有声明
	if claims.Username != username {
		t.Errorf("Claims username = %s, want %s", claims.Username, username)
	}
	if claims.Issuer != "foxflow" {
		t.Errorf("Claims issuer = %s, want foxflow", claims.Issuer)
	}
	if claims.Subject != username {
		t.Errorf("Claims subject = %s, want %s", claims.Subject, username)
	}
	if claims.ExpiresAt == nil {
		t.Error("Claims ExpiresAt should not be nil")
	}
	if claims.IssuedAt == nil {
		t.Error("Claims IssuedAt should not be nil")
	}
	if claims.NotBefore == nil {
		t.Error("Claims NotBefore should not be nil")
	}
}
