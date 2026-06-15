package dto

// ─── Login ─────────────────────────────────────────────────────────────────────

// LoginRequest request body untuk login
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // username atau email
	Password   string `json:"password"   validate:"required"`
}

// ─── Register ──────────────────────────────────────────────────────────────────

// RegisterRequest request body untuk registrasi
type RegisterRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	Name     *string `json:"name,omitempty"`
}

// ─── Logout ────────────────────────────────────────────────────────────────────

// LogoutRequest request body untuk logout dari device saat ini
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ─── Refresh Token ─────────────────────────────────────────────────────────────

// RefreshTokenRequest request body untuk refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ─── Forgot Password ───────────────────────────────────────────────────────────

// ForgotPasswordRequest request body untuk forgot password
type ForgotPasswordRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email atau username
}

// ─── Reset Password ────────────────────────────────────────────────────────────

// ResetPasswordRequest request body untuk reset password
type ResetPasswordRequest struct {
	Token           string `json:"token"            validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}
