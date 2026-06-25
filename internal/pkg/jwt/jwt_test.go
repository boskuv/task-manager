package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", time.Hour)

	token, err := manager.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	userID, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if userID != 42 {
		t.Errorf("userID = %d, want 42", userID)
	}
}

func TestParseAccessTokenInvalid(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", time.Hour)

	_, err := manager.ParseAccessToken("not-a-jwt")
	if !errorsIsInvalid(err) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func TestParseAccessTokenExpired(t *testing.T) {
	manager := &Manager{
		secret: []byte("test-secret"),
		ttl:    time.Hour,
		now: func() time.Time {
			return time.Now().Add(-2 * time.Hour)
		},
	}

	token, err := manager.GenerateAccessToken(7)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = manager.ParseAccessToken(token)
	if !errorsIsInvalid(err) {
		t.Fatalf("error = %v, want ErrInvalidToken for expired token", err)
	}
}

func TestGenerateAccessTokenEmptySecret(t *testing.T) {
	t.Parallel()

	manager := NewManager("", time.Hour)
	_, err := manager.GenerateAccessToken(1)
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func errorsIsInvalid(err error) bool {
	return err == ErrInvalidToken
}
