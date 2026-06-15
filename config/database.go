package config

import (
	"fmt"

	"log"

	"time"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"

	"gorm.io/gorm/logger"
)

// ConnectDB establishes a connection to PostgreSQL database using GORM

func (c *Config) ConnectDB() (*gorm.DB, error) {
	// 1. Rakit DSN dari variabel env
	dsn := c.DatabaseURL
	if dsn == "" {
		// Format DSN untuk PostgreSQL
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			c.DatabaseHost,
			c.DatabaseUser,
			c.DatabasePass,
			c.DatabaseName,
			c.DatabasePort,
			c.DatabaseSSL,
			c.TimeZone,
		)
	}

	// Konfigurasi GORM logger (tetap sama seperti kode kamu)
	var gormLogger logger.Interface
	if c.Env == "DEV" {
		gormLogger = logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
	} else {
		gormLogger = logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	}

	// 2. Connect ke database menggunakan driver postgres
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// ... (Sisa kode connection pool dan ping tetap sama)

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database (env: %s)", c.Env)
	return db, nil
}

// CloseDB closes the database connection

func CloseDB(db *gorm.DB) error {

	sqlDB, err := db.DB()

	if err != nil {

		return err

	}

	return sqlDB.Close()

}

// Migrate runs auto-migration for all models

func (c *Config) Migrate(db *gorm.DB, models ...interface{}) error {

	if len(models) == 0 {

		return fmt.Errorf("no models provided for migration")

	}

	log.Printf("Running auto-migration for %d models...", len(models))

	for _, model := range models {

		if err := db.AutoMigrate(model); err != nil {

			return fmt.Errorf("failed to migrate model %T: %w", model, err)

		}

	}

	log.Println("Auto-migration completed successfully")

	return nil

}

// SeedRolesAndPermissions is a placeholder for seeding initial data

func (c *Config) SeedRolesAndPermissions(db *gorm.DB) error {

	// This is a placeholder - implement actual seeding logic here

	log.Println("Seeding roles and permissions (placeholder)")

	return nil

}

// GetDBStats returns database connection statistics

func GetDBStats(db *gorm.DB) (map[string]interface{}, error) {

	sqlDB, err := db.DB()

	if err != nil {

		return nil, err

	}

	stats := sqlDB.Stats()

	return map[string]interface{}{

		"open_connections": stats.OpenConnections,

		"in_use": stats.InUse,

		"idle": stats.Idle,

		"wait_count": stats.WaitCount,

		"wait_duration": stats.WaitDuration.String(),

		"max_idle_closed": stats.MaxIdleClosed,

		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}, nil

}
