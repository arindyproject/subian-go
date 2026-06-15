package migrations

import (
	"database/sql"
	_ "embed"
	"log"

	"subian_go/internal/modules/auth/models"

	"gorm.io/gorm"
)

//go:embed 001_create_auth_tokens_table.sql
var authTokensSQL string

//go:embed 002_create_login_histories_table.sql
var loginHistoriesSQL string

//go:embed 003_create_password_histories_table.sql
var passwordHistoriesSQL string

// MigrateAuth menjalankan GORM auto-migration untuk semua tabel auth
func MigrateAuth(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AuthToken{},
		&models.LoginHistory{},
		&models.PasswordHistory{},
	)
}

// MigrateAuthWithSQL menjalankan migrasi via raw SQL
func MigrateAuthWithSQL(sqlDB *sql.DB) error {
	sqls := []struct {
		name  string
		query string
	}{
		{"auth_tokens", authTokensSQL},
		{"users_login_histories", loginHistoriesSQL},
		{"users_password_histories", passwordHistoriesSQL},
	}

	for _, s := range sqls {
		if _, err := sqlDB.Exec(s.query); err != nil {
			log.Printf("Error creating %s table: %v", s.name, err)
			return err
		}
		log.Printf("✅ Table %s migrated successfully", s.name)
	}

	return nil
}

// DropAuthTables menghapus semua tabel auth (gunakan dengan hati-hati!)
func DropAuthTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.PasswordHistory{},
		&models.LoginHistory{},
		&models.AuthToken{},
	)
}
