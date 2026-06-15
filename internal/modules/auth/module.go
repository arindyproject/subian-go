package auth

import (
	"subian_go/config"
	"subian_go/internal/modules/auth/handlers"
	"subian_go/internal/modules/auth/repositories"
	"subian_go/internal/modules/auth/services"
	rbacContracts "subian_go/internal/modules/rbac/contracts" // ← Tambahkan import RBAC contract
	userContracts "subian_go/internal/modules/users/contracts"
	userRepositories "subian_go/internal/modules/users/repositories"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Module mewakili auth module
type Module struct {
	db         *gorm.DB
	redis      *redis.Client
	handler    *handlers.AuthHandler
	jwtManager *utils.JWTManager
}

// NewModule membuat instance module dan wire semua layer
// Tambahkan rbacRepo rbacContracts.RBACRepository pada parameter fungsi
func NewModule(db *gorm.DB, redisClient *redis.Client, rbacRepo rbacContracts.RBACRepository, cfg *config.Config) *Module {
	authRepo := repositories.NewAuthRepository(db)

	var userRepo userContracts.Repository = userRepositories.NewRepository(db)

	var mailer *utils.Mailer
	if cfg.SMTPHost != "" {
		mailer = utils.NewMailer(utils.MailConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
	}

	svcCfg := services.AuthServiceConfig{
		JWTManager: utils.NewJWTManager(
			cfg.JWTSecret,
			cfg.JWTIssuer,
			cfg.JWTAccessTokenExpMinutes,
			cfg.JWTRefreshTokenExpDays,
		),
		LoginMaxAttempts:             cfg.LoginMaxAttempts,
		LoginLockDurationMinutes:     cfg.LoginLockDurationMinutes,
		MaxConcurrentSessions:        cfg.MaxConcurrentSessions,
		RateLimitLoginPerIPPerMinute: cfg.RateLimitLoginPerIPPerMinute,
		PasswordPolicy: &utils.PasswordPolicy{
			MinLength:        cfg.PasswordMinLength,
			RequireUppercase: cfg.PasswordRequireUppercase,
			RequireNumber:    cfg.PasswordRequireNumber,
			RequireSymbol:    cfg.PasswordRequireSymbol,
		},
		PasswordHistoryCount:     cfg.PasswordHistoryCount,
		IsRegistrationActive:     cfg.IsRegistrationActive,
		AutoActiveUser:           cfg.AutoActiveUser,
		MailResetTokenExpMinutes: cfg.MailResetTokenExpMinutes,
		AppFrontendURL:           cfg.AppFrontendURL,
		Mailer:                   mailer,
	}

	jwtManager := utils.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAccessTokenExpMinutes,
		cfg.JWTRefreshTokenExpDays,
	)

	// SOLUSI: Masukkan rbacRepo ke dalam parameter NewAuthService sesuai urutan di servicenya
	svc := services.NewAuthService(authRepo, userRepo, rbacRepo, redisClient, svcCfg)
	handler := handlers.NewAuthHandler(svc)

	return &Module{
		db:         db,
		redis:      redisClient,
		handler:    handler,
		jwtManager: jwtManager,
	}
}

func (m *Module) InitRoutes(e *echo.Echo) {
	// Inject db ke RegisterRoutes untuk JWTMiddleware realtime
	RegisterRoutes(e, m.handler, m.jwtManager, m.db)
}
