package helpers

import (
	"log"

	"subian_go/config"
	"subian_go/internal/modules/users/models"

	"gorm.io/gorm"
)

// SetupTestDB membuat koneksi DB untuk keperluan test
func SetupTestDB() *gorm.DB {
	cfg := config.LoadConfig()

	db, err := cfg.ConnectDB()
	if err != nil {
		log.Fatal("Gagal koneksi DB untuk test:", err)
	}

	return db
}

// MigrateTestDB menjalankan migrasi untuk test DB
func MigrateTestDB(db *gorm.DB) {
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Gagal migrasi test DB:", err)
	}
}

// CleanTable menghapus semua record dari tabel tertentu
func CleanTable(db *gorm.DB, tables ...string) {
	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			log.Printf("Warning: Gagal clean table %s: %v", table, err)
		}
	}
}

// TruncateTable menghapus semua record dan reset sequence
func TruncateTable(db *gorm.DB, tables ...string) {
	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
			log.Printf("Warning: Gagal truncate table %s: %v", table, err)
		}
	}
}
