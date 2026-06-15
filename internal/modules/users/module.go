package users

import (
	authContracts "subian_go/internal/modules/auth/contracts"
	rbacContracts "subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/modules/users/contracts"
	"subian_go/internal/modules/users/handlers"
	"subian_go/internal/modules/users/repositories"
	"subian_go/internal/modules/users/services"
	"subian_go/internal/shared/storage"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Module represents the users module
type Module struct {
	db             *gorm.DB
	handler        *handlers.Handler
	service        contracts.Service
	repo           contracts.Repository
	rbacRepo       rbacContracts.RBACRepository
	storageService storage.ImageStorage
	jwtManager     *utils.JWTManager
}

// NewModule membuat instance module dan wire semua layer
// Menambahkan authRepo ke parameter agar bisa disuntikkan ke NewUserService
func NewModule(
	db *gorm.DB,
	jwtManager *utils.JWTManager,
	rbacRepo rbacContracts.RBACRepository,
	authRepo authContracts.AuthRepository,
	storageService storage.ImageStorage,
) *Module {
	repo := repositories.NewRepository(db)

	// Pastikan NewUserService di internal/modules/users/services menerima 4 parameter ini:
	svc := services.NewUserService(
		repo,
		rbacRepo,
		authRepo,
		storageService)
	handler := handlers.NewHandler(svc)

	return &Module{
		db:             db,
		handler:        handler,
		service:        svc,
		repo:           repo,
		rbacRepo:       rbacRepo,
		jwtManager:     jwtManager,
		storageService: storageService,
	}
}

// InitRoutes mendaftarkan routes — db diteruskan ke JWTMiddleware untuk realtime check
func (m *Module) InitRoutes(e *echo.Echo) {
	RegisterRoutes(e, m.handler, m.rbacRepo, m.jwtManager, m.db)
}

func (m *Module) GetRepository() contracts.Repository { return m.repo }
func (m *Module) GetService() contracts.Service       { return m.service }
func (m *Module) GetHandler() *handlers.Handler       { return m.handler }
