package auth

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/pkg/password"
	"github.com/boskuv/task-manager/internal/repository"
)

const minPasswordLength = 8

// TokenIssuer issues access tokens for authenticated users.
type TokenIssuer interface {
	GenerateAccessToken(userID int64) (string, error)
}

// Service handles registration and login.
type Service struct {
	users  repository.UserRepository
	tokens TokenIssuer
}

// NewService creates an auth use case service.
func NewService(users repository.UserRepository, tokens TokenIssuer) *Service {
	return &Service{
		users:  users,
		tokens: tokens,
	}
}

// Register creates a user account and returns an access token.
func (s *Service) Register(ctx context.Context, email, plainPassword string) (string, error) {
	email, plainPassword, err := validateCredentials(email, plainPassword)
	if err != nil {
		return "", err
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, domain.User{
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		return "", err
	}

	return s.tokens.GenerateAccessToken(user.ID)
}

// Login authenticates a user and returns an access token.
func (s *Service) Login(ctx context.Context, email, plainPassword string) (string, error) {
	email, plainPassword, err := validateCredentials(email, plainPassword)
	if err != nil {
		return "", err
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return "", domain.ErrUnauthorized
		}
		return "", err
	}

	if err := password.Compare(user.PasswordHash, plainPassword); err != nil {
		return "", domain.ErrUnauthorized
	}

	return s.tokens.GenerateAccessToken(user.ID)
}

func validateCredentials(email, plainPassword string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	password := strings.TrimSpace(plainPassword)

	if email == "" || password == "" {
		return "", "", domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", domain.ErrInvalidInput
	}
	if len(password) < minPasswordLength {
		return "", "", domain.ErrInvalidInput
	}

	return email, password, nil
}
