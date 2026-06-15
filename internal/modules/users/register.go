package users

import (
	"database/sql"

	"subian_go/config"
	"subian_go/internal/apps"
	authContracts "subian_go/internal/modules/auth/contracts"
	authRepositories "subian_go/internal/modules/auth/repositories"
	rbacContracts "subian_go/internal/modules/rbac/contracts"
	rbacRepositories "subian_go/internal/modules/rbac/repositories"
	"subian_go/internal/modules/users/migrations"
	"subian_go/internal/modules/users/models"
	"subian_go/internal/shared/storage"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type registryModule struct {
	db             *gorm.DB
	cfg            *config.Config
	rbacRepo       rbacContracts.RBACRepository
	authRepo       authContracts.AuthRepository
	storageService storage.ImageStorage
}

func init() {
	apps.Register(&registryModule{})
}

// ─── Injections ────────────────────────────────────────────────────────────────

func (r *registryModule) SetDB(db *gorm.DB) {
	r.db = db
	r.rbacRepo = rbacRepositories.NewRBACRepository(db)
	r.authRepo = authRepositories.NewAuthRepository(db)
}

func (r *registryModule) SetConfig(cfg *config.Config) {
	r.cfg = cfg

	// ─── INISIALISASI STORAGE DI SINI ──────────────────────────────────────────
	// Sesuaikan properti cfg dengan field yang ada pada struct config.Config Anda
	storageCfg := storage.LocalImageConfig{
		BasePath:   "./uploads",                // atau ambil dari env jika ada: cfg.StorageBasePath
		BaseURL:    r.cfg.BaseURL + "/uploads", // atau ambil dari env jika ada: cfg.StorageBaseURL
		MaxSizeMB:  5,
		ThumbnailW: 150,
		Quality:    70,
	}

	// Assign ke properti storageService milik registryModule
	r.storageService = storage.NewLocalImageStorage(storageCfg)
}

// ─── Routes ────────────────────────────────────────────────────────────────────

func (r *registryModule) InitRoutes(e *echo.Echo) {
	jwtManager := utils.NewJWTManager(
		r.cfg.JWTSecret,
		r.cfg.JWTIssuer,
		r.cfg.JWTAccessTokenExpMinutes,
		r.cfg.JWTRefreshTokenExpDays,
	)
	// ← hapus r.authRepo, NewModule hanya butuh 3 argumen
	NewModule(r.db, jwtManager, r.rbacRepo, r.authRepo, r.storageService).InitRoutes(e)
}

// ─── Migration ─────────────────────────────────────────────────────────────────

func (r *registryModule) Models() []interface{} {
	return []interface{}{
		&models.User{},
	}
}

func (r *registryModule) SeedData(db *gorm.DB) error {
	return migrations.SeedDefaultSettings(db)
}

func (r *registryModule) MigrateSQL(sqlDB *sql.DB) error {
	return migrations.MigrateUsersWithSQL(sqlDB)
}
