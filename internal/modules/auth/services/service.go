package services

import (
	"subian_go/internal/modules/auth/contracts"
	rbacContracts "subian_go/internal/modules/rbac/contracts"
	userContracts "subian_go/internal/modules/users/contracts"
	"subian_go/internal/shared/utils"

	"github.com/redis/go-redis/v9"
)

// ─── Config ────────────────────────────────────────────────────────────────────

type AuthServiceConfig struct {
	JWTManager                   *utils.JWTManager
	LoginMaxAttempts             int
	LoginLockDurationMinutes     int
	MaxConcurrentSessions        int
	RateLimitLoginPerIPPerMinute int
	PasswordPolicy               *utils.PasswordPolicy
	PasswordHistoryCount         int
	IsRegistrationActive         bool
	AutoActiveUser               bool
	MailResetTokenExpMinutes     int
	AppFrontendURL               string
	Mailer                       *utils.Mailer
}

// ─── Init ──────────────────────────────────────────────────────────────────────
// ─── Init ──────────────────────────────────────────────────────────────────────
type authService struct {
	repo     contracts.AuthRepository
	userRepo userContracts.Repository
	rbacRepo rbacContracts.RBACRepository // ← Tambahkan ini
	redis    *redis.Client
	cfg      AuthServiceConfig
}

func NewAuthService(
	repo contracts.AuthRepository,
	userRepo userContracts.Repository,
	rbacRepo rbacContracts.RBACRepository, // ← Tambahkan ini
	redisClient *redis.Client,
	cfg AuthServiceConfig,
) contracts.AuthService {
	return &authService{
		repo:     repo,
		userRepo: userRepo,
		rbacRepo: rbacRepo, // ← Set di sini
		redis:    redisClient,
		cfg:      cfg,
	}
}

// ─── End Init ──────────────────────────────────────────────────────────────────
