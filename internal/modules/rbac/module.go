package rbac

import (
	"subian_go/config"
	"subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/modules/rbac/handlers"
	"subian_go/internal/modules/rbac/repositories"
	"subian_go/internal/modules/rbac/services"

	// PERBAIKAN: Import interface (contracts), BUKAN implementasi konkret (repositories).
	// Ini mencegah circular dependency jika module users juga membutuhkan rbac.
	userContracts "subian_go/internal/modules/users/contracts"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Module mewakili rbac module
type Module struct {
	db         *gorm.DB
	rbacRepo   contracts.RBACRepository
	handler    *handlers.RBACHandler
	jwtManager *utils.JWTManager
}

// NewModule sekarang menerima userRepo sebagai interface (Dependency Injection).
func NewModule(db *gorm.DB, cfg *config.Config, userRepo userContracts.Repository) *Module {
	rbacRepo := repositories.NewRBACRepository(db)

	// Langsung gunakan userRepo yang di-inject, tidak perlu instantiate di sini
	svc := services.NewRBACService(rbacRepo, userRepo)

	handler := handlers.NewRBACHandler(svc)
	jwtManager := utils.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAccessTokenExpMinutes,
		cfg.JWTRefreshTokenExpDays,
	)

	return &Module{
		db:         db,
		rbacRepo:   rbacRepo,
		handler:    handler,
		jwtManager: jwtManager,
	}
}

func (m *Module) InitRoutes(e *echo.Echo) {
	RegisterRoutes(e, m.handler, m.rbacRepo, m.jwtManager, m.db) // ← tambah m.db
}

func (m *Module) GetRepository() contracts.RBACRepository {
	return m.rbacRepo
}
