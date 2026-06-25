package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

// Claims holds access token payload.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwtlib.RegisteredClaims
}

// Manager issues and validates access tokens.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewManager creates a JWT manager for access tokens.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

// GenerateAccessToken creates a signed JWT for the given user id.
func (m *Manager) GenerateAccessToken(userID int64) (string, error) {
	if len(m.secret) == 0 {
		return "", fmt.Errorf("jwt secret is required")
	}

	now := m.now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken validates the token and returns the user id.
func (m *Manager) ParseAccessToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwtlib.ParseWithClaims(tokenString, claims, func(token *jwtlib.Token) (any, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}
	if token == nil || !token.Valid {
		return 0, ErrInvalidToken
	}
	if claims.UserID == 0 {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
