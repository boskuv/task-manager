package dto

// RegisterRequest is the body for POST /api/v1/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body for POST /api/v1/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after successful register or login.
type AuthResponse struct {
	Token string `json:"token"`
}

// UserResponse exposes public user fields.
type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}
