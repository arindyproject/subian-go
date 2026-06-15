package apps

import (
	"subian_go/config"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	// Blank import untuk trigger init() tiap module
	// Tambah baris di sini saat ada module baru
	//_ "subian_go/internal/modules/auth"
	//_ "subian_go/internal/modules/users"
)

// RegisterModules adalah satu-satunya entry point yang dipanggil dari main.go
func RegisterModules(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, e *echo.Echo) {
	InjectConfig(cfg)
	InjectDB(db)
	InjectRedis(redisClient)
	InitAllRoutes(e)
}
