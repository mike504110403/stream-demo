package utils

import (
	"testing"
	"time"
)

func TestNewJWTUtil(t *testing.T) {
	secret := "test-secret"
	jwtUtil := NewJWTUtil(secret)

	if jwtUtil == nil {
		t.Fatal("NewJWTUtil should not return nil")
	}

	if string(jwtUtil.secret) != secret {
		t.Errorf("Expected secret %s, got %s", secret, string(jwtUtil.secret))
	}
}

func TestGenerateToken(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret")
	userID := uint(123)
	role := "user"

	token, err := jwtUtil.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token should not be empty")
	}

	// 驗證生成的令牌
	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestValidateToken(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret")
	userID := uint(456)
	role := "admin"

	// 生成有效令牌
	token, err := jwtUtil.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// 驗證有效令牌
	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestValidateTokenWithInvalidToken(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret")

	// 測試無效令牌
	_, err := jwtUtil.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("ValidateToken should fail with invalid token")
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	jwtUtil1 := NewJWTUtil("secret1")
	jwtUtil2 := NewJWTUtil("secret2")

	userID := uint(789)
	role := "user"

	// 用 secret1 生成令牌
	token, err := jwtUtil1.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// 用 secret2 驗證令牌
	_, err = jwtUtil2.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken should fail with wrong secret")
	}
}

func TestTokenExpiration(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret")
	userID := uint(999)
	role := "user"

	// 生成令牌
	token, err := jwtUtil.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// 驗證令牌未過期
	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// 檢查過期時間是否在未來
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatal("Token should not be expired")
	}

	// 檢查發行時間是否在過去
	if claims.IssuedAt.Time.After(time.Now()) {
		t.Fatal("Token issued time should be in the past")
	}
}

func TestJWTClaimsStructure(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret")
	userID := uint(111)
	role := "moderator"

	token, err := jwtUtil.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// 檢查 claims 結構
	if claims == nil {
		t.Fatal("Claims should not be nil")
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}

	// 檢查 JWT 標準聲明
	if claims.IssuedAt.Time.IsZero() {
		t.Fatal("IssuedAt should not be zero")
	}

	if claims.ExpiresAt.Time.IsZero() {
		t.Fatal("ExpiresAt should not be zero")
	}
}
