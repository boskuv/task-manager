package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
)

func TestRegisterSuccess(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepo()
	tokens := &mockTokenIssuer{}
	svc := NewService(repo, tokens)

	token, err := svc.Register(context.Background(), "User@Example.com", "password1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if len(repo.users) != 1 {
		t.Fatalf("users count = %d, want 1", len(repo.users))
	}
	if _, ok := repo.users["user@example.com"]; !ok {
		t.Fatal("expected normalized email in repository")
	}
	if tokens.lastUserID != 1 {
		t.Errorf("lastUserID = %d, want 1", tokens.lastUserID)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepo()
	svc := NewService(repo, &mockTokenIssuer{})

	_, err := svc.Register(context.Background(), "user@example.com", "password1")
	if err != nil {
		t.Fatalf("first Register: %v", err)
	}

	_, err = svc.Register(context.Background(), "user@example.com", "password2")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestRegisterInvalidInput(t *testing.T) {
	t.Parallel()

	svc := NewService(newMockUserRepo(), &mockTokenIssuer{})

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "empty email", email: "", password: "password1"},
		{name: "invalid email", email: "not-email", password: "password1"},
		{name: "short password", email: "user@example.com", password: "short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.Register(context.Background(), tt.email, tt.password)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestLoginSuccess(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepo()
	tokens := &mockTokenIssuer{}
	svc := NewService(repo, tokens)

	if _, err := svc.Register(context.Background(), "user@example.com", "password1"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	token, err := svc.Login(context.Background(), "user@example.com", "password1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepo()
	svc := NewService(repo, &mockTokenIssuer{})

	if _, err := svc.Register(context.Background(), "user@example.com", "password1"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, err := svc.Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("wrong password: error = %v, want ErrUnauthorized", err)
	}

	_, err = svc.Login(context.Background(), "missing@example.com", "password1")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("missing user: error = %v, want ErrUnauthorized", err)
	}
}

type mockUserRepo struct {
	users  map[string]domain.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(_ context.Context, user domain.User) (domain.User, error) {
	if _, exists := m.users[user.Email]; exists {
		return domain.User{}, domain.ErrConflict
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.Email] = user
	return user, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	user, ok := m.users[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id int64) (domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

type mockTokenIssuer struct {
	lastUserID int64
}

func (m *mockTokenIssuer) GenerateAccessToken(userID int64) (string, error) {
	m.lastUserID = userID
	return "token", nil
}
