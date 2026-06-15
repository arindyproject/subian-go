package dto

import (
	userDto "subian_go/internal/modules/users/dto"
	"time"
)

// ─── Token Response ────────────────────────────────────────────────────────────

// TokenResponse response setelah login / refresh berhasil
type TokenResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	TokenType    string               `json:"token_type"` // Bearer
	ExpiresIn    int                  `json:"expires_in"` // detik
	User         userDto.UserResponse `json:"user"`
}

// UserInfo data user ringkas untuk response token
type UserInfo struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	IsSuperadmin bool   `json:"is_superadmin"`
	IsStaff      bool   `json:"is_staff"`
	IsVerified   bool   `json:"is_verified"`
}

// ─── Register Response ─────────────────────────────────────────────────────────

// RegisterResponse response setelah register berhasil
type RegisterResponse struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	IsActive   bool      `json:"is_active"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}
