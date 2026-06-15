package apps

import (
	"database/sql"
	"log"

	"subian_go/config"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ─── Interfaces ────────────────────────────────────────────────────────────────

type Module interface {
	Models() []interface{}
	SeedData(db *gorm.DB) error
	MigrateSQL(sqlDB *sql.DB) error
	InitRoutes(e *echo.Echo)
}

type DBInjectable interface {
	SetDB(db *gorm.DB)
}

type RedisInjectable interface {
	SetRedis(client *redis.Client)
}

type ConfigInjectable interface {
	SetConfig(cfg *config.Config)
}

// ─── Registry ──────────────────────────────────────────────────────────────────

var registeredModules []Module

func Register(m Module) {
	registeredModules = append(registeredModules, m)
}

// ─── Routes ────────────────────────────────────────────────────────────────────

func InitAllRoutes(e *echo.Echo) {
	for _, m := range registeredModules {
		m.InitRoutes(e)
	}
}

// ─── Injections ────────────────────────────────────────────────────────────────

func InjectDB(db *gorm.DB) {
	for _, m := range registeredModules {
		if injectable, ok := m.(DBInjectable); ok {
			injectable.SetDB(db)
		}
	}
}

func InjectRedis(client *redis.Client) {
	for _, m := range registeredModules {
		if injectable, ok := m.(RedisInjectable); ok {
			injectable.SetRedis(client)
		}
	}
}

func InjectConfig(cfg *config.Config) {
	for _, m := range registeredModules {
		if injectable, ok := m.(ConfigInjectable); ok {
			injectable.SetConfig(cfg)
		}
	}
}

// ─── Migration ─────────────────────────────────────────────────────────────────

func AllModels() []interface{} {
	var all []interface{}
	for _, m := range registeredModules {
		all = append(all, m.Models()...)
	}
	return all
}

func DropAll(db *gorm.DB) error {
	models := AllModels()
	reversed := make([]interface{}, len(models))
	for i, m := range models {
		reversed[len(models)-1-i] = m
	}
	return db.Migrator().DropTable(reversed...)
}

func MigrateAll(db *gorm.DB) error {
	for _, m := range registeredModules {
		if err := db.AutoMigrate(m.Models()...); err != nil {
			return err
		}
	}
	return nil
}

func SeedAll(db *gorm.DB) {
	for _, m := range registeredModules {
		if err := m.SeedData(db); err != nil {
			log.Printf("Warning: seed failed for %T: %v", m, err)
		}
	}
}

func MigrateAllSQL(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	for _, m := range registeredModules {
		if err := m.MigrateSQL(sqlDB); err != nil {
			return err
		}
	}
	return nil
}
