package handlers

import (
	"subian_go/internal/modules/auth/contracts"
)

// ─── Init ──────────────────────────────────────────────────────────────────────
// AuthHandler menangani HTTP request untuk auth
type AuthHandler struct {
	service contracts.AuthService
}

// NewAuthHandler membuat instance handler baru
func NewAuthHandler(service contracts.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}
