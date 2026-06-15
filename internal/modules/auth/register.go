package auth

import (
	"database/sql"

	"subian_go/config"

	"subian_go/internal/apps"
	"subian_go/internal/modules/auth/models"
	rbacContracts "subian_go/internal/modules/rbac/contracts"       // ← Tambahkan import contract RBAC
	rbacRepositories "subian_go/internal/modules/rbac/repositories" // ← Tambahkan import repository RBAC

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type registryModule struct {
	db       *gorm.DB
	redis    *redis.Client
	cfg      *config.Config
	rbacRepo rbacContracts.RBACRepository // ← Tambahkan field rbacRepo di struct registry
}

func init() {
	apps.Register(&registryModule{})
}

// ─── Injections ────────────────────────────────────────────────────────────────

func (r *registryModule) SetDB(db *gorm.DB) {
	r.db = db
	// SOLUSI: Inisialisasi rbacRepo di sini agar tidak nil saat aplikasi dijalankan
	r.rbacRepo = rbacRepositories.NewRBACRepository(db)
}

func (r *registryModule) SetRedis(client *redis.Client) {
	r.redis = client
}

func (r *registryModule) SetConfig(cfg *config.Config) {
	r.cfg = cfg
}

// ─── Routes ────────────────────────────────────────────────────────────────────

func (r *registryModule) InitRoutes(e *echo.Echo) {
	// SOLUSI: Sisipkan rbacRepo sebagai parameter ke-3 sesuai perubahan pada fungsi NewModule sebelumnya
	NewModule(r.db, r.redis, r.rbacRepo, r.cfg).InitRoutes(e)
}

// ─── Migration ─────────────────────────────────────────────────────────────────

func (r *registryModule) Models() []interface{} {
	return []interface{}{
		&models.AuthToken{},
		&models.LoginHistory{},
		&models.PasswordHistory{},
	}
}

func (r *registryModule) SeedData(db *gorm.DB) error {
	return nil
}

func (r *registryModule) MigrateSQL(sqlDB *sql.DB) error {
	return nil // Gunakan GORM auto-migrate
}
