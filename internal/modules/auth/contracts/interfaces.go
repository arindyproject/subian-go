package contracts

import (
	"subian_go/internal/modules/auth/dto"
	"subian_go/internal/modules/auth/models"
)

// ─── Repository ────────────────────────────────────────────────────────────────

// AuthRepository mendefinisikan operasi database untuk auth
type AuthRepository interface {
	// Auth Token
	SaveToken(token *models.AuthToken) error
	GetTokenByJTI(jti string) (*models.AuthToken, error)
	BlacklistToken(jti string) error
	BlacklistAllUserTokens(userID int64) error
	CountActiveTokens(userID int64) (int64, error)

	// Login History
	SaveLoginHistory(history *models.LoginHistory) error
	GetUserLoginHistories(userID int64, limit int) ([]models.LoginHistory, error)

	// Password History
	SavePasswordHistory(history *models.PasswordHistory) error
	GetPasswordHistories(userID int64, limit int) ([]models.PasswordHistory, error)
}

// ─── Service ───────────────────────────────────────────────────────────────────

// AuthService mendefinisikan business logic untuk auth
type AuthService interface {
	// Auth
	Login(req *dto.LoginRequest, ip, userAgent string) (*dto.TokenResponse, error)
	Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	RefreshToken(req *dto.RefreshTokenRequest) (*dto.TokenResponse, error)

	// Logout
	Logout(req *dto.LogoutRequest) error // logout device saat ini
	LogoutAll(userID int64) error        // logout semua device

	// Password
	ForgotPassword(req *dto.ForgotPasswordRequest) error
	ResetPassword(req *dto.ResetPasswordRequest) error
}
