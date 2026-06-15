package migrations

import (
	"database/sql"
	_ "embed"
	"log"

	"subian_go/internal/modules/users/models"

	"gorm.io/gorm"
)

// ================================================ Users Migration ================================================
//
//go:generate go run cmd/migrate/main.go -env=DEV -type=sql > 001_create_users_table.sql
//go:embed 001_create_users_table.sql
var usersSQL string

// MigrateUsers runs the users migration
func MigrateUsers(db *gorm.DB) error {
	return db.Migrator().CreateTable(&models.User{})
}

// MigrateUsersWithSQL creates users table using raw SQL from embedded file
func MigrateUsersWithSQL(sqlDB *sql.DB) error {
	_, err := sqlDB.Exec(usersSQL)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	log.Println("Users table migrated successfully")
	return nil
}

// SeedDefaultSettings creates default user settings
func SeedDefaultSettings(db *gorm.DB) error {
	defaultSettings := []models.UserSetting{
		{
			Key:         "is_dark_mode",
			Type:        "boolean",
			Value:       false,
			Description: "Aktifkan tema gelap pada antarmuka",
		},
	}

	log.Println("Default settings seeded (used when creating new users)")
	_ = defaultSettings // Used in service when creating users
	return nil
}

// DropUsersTable drops the users table (use with caution!)
func DropUsersTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.User{})
}

// RollbackUsers rolls back the users migration
func RollbackUsers(db *gorm.DB) error {
	return DropUsersTable(db)
}

// ================================================ Users Migration ================================================
