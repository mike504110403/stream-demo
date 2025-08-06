package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hashedPassword == "" {
		t.Fatal("Hashed password should not be empty")
	}

	if hashedPassword == password {
		t.Fatal("Hashed password should not be the same as original password")
	}

	// 測試相同密碼會產生不同的雜湊值（因為 salt）
	hashedPassword2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hashedPassword == hashedPassword2 {
		t.Fatal("Same password should produce different hashes due to salt")
	}
}

func TestComparePassword(t *testing.T) {
	password := "mypassword456"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// 測試正確密碼
	err = ComparePassword(hashedPassword, password)
	if err != nil {
		t.Fatalf("ComparePassword failed with correct password: %v", err)
	}

	// 測試錯誤密碼
	err = ComparePassword(hashedPassword, "wrongpassword")
	if err == nil {
		t.Fatal("ComparePassword should fail with wrong password")
	}
}

func TestComparePasswordWithEmptyPassword(t *testing.T) {
	password := "testpassword"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// 測試空密碼
	err = ComparePassword(hashedPassword, "")
	if err == nil {
		t.Fatal("ComparePassword should fail with empty password")
	}
}

func TestHashPasswordWithEmptyPassword(t *testing.T) {
	hashedPassword, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword should not fail with empty password: %v", err)
	}

	if hashedPassword == "" {
		t.Fatal("Hashed password should not be empty even for empty input")
	}
}

func TestHashPasswordWithSpecialCharacters(t *testing.T) {
	passwords := []string{
		"password123!@#",
		"密碼測試",
		"p@ssw0rd",
		"1234567890",
		"a",
		"verylongpasswordwithlotsofcharactersandnumbers1234567890!@#$%^&*()",
	}

	for _, password := range passwords {
		t.Run("password_"+password, func(t *testing.T) {
			hashedPassword, err := HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword failed for password '%s': %v", password, err)
			}

			if hashedPassword == "" {
				t.Fatal("Hashed password should not be empty")
			}

			// 驗證密碼
			err = ComparePassword(hashedPassword, password)
			if err != nil {
				t.Fatalf("ComparePassword failed for password '%s': %v", password, err)
			}
		})
	}
}

func TestPasswordSecurity(t *testing.T) {
	password := "securepassword"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// 檢查雜湊值長度（bcrypt 雜湊通常是 60 字元）
	if len(hashedPassword) != 60 {
		t.Errorf("Expected hash length 60, got %d", len(hashedPassword))
	}

	// 檢查雜湊值格式（bcrypt 雜湊以 $2a$ 或 $2b$ 開頭）
	if len(hashedPassword) >= 4 && (hashedPassword[:4] != "$2a$" && hashedPassword[:4] != "$2b$") {
		t.Errorf("Invalid bcrypt hash format: %s", hashedPassword[:4])
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword"

	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
	}
}

func BenchmarkComparePassword(b *testing.B) {
	password := "benchmarkpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		b.Fatalf("HashPassword failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ComparePassword(hashedPassword, password)
		if err != nil {
			b.Fatalf("ComparePassword failed: %v", err)
		}
	}
}
