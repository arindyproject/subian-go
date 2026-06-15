package main

import (
	"flag"
	"log"

	"subian_go/config"
	"subian_go/internal/apps"

	// =====================================================================
	// import module di sini
	// =====================================================================

	_ "subian_go/internal/modules/auth"
	_ "subian_go/internal/modules/rbac"
	_ "subian_go/internal/modules/users"
	// _ "subian_go/internal/modules/roles"
	// =====================================================================
)

func main() {
	cfg := config.LoadConfig()
	env := flag.String("env", cfg.EnvCode, "Environment (DEV atau PROD)")
	migrationType := flag.String("type", "gorm", "Tipe migrasi: gorm atau sql")
	fresh := flag.Bool("fresh", false, "Drop semua tabel sebelum migrasi")
	flag.Parse()

	// Validasi environment
	if *env != "DEV" && *env != "PROD" {
		log.Fatal("❌ Environment tidak valid. Gunakan DEV atau PROD")
	}

	// Safety guard: fresh tidak diizinkan di PROD
	if *fresh && *env == "PROD" {
		log.Fatal("❌ Fresh migration TIDAK diizinkan di environment PROD!")
	}

	log.Printf("🚀 Memulai migrasi untuk environment: %s", *env)

	// Load config & connect DB

	db, err := cfg.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer config.CloseDB(db)

	// Fresh: drop semua tabel terlebih dahulu
	if *fresh {
		log.Println("🗑️  Menghapus semua tabel...")
		if err := apps.DropAll(db); err != nil {
			log.Fatal("Gagal menghapus tabel:", err)
		}
		log.Println("✅ Semua tabel berhasil dihapus.")
	}

	// Jalankan migrasi sesuai tipe
	switch *migrationType {
	case "sql":
		log.Println("⚙️  Menjalankan SQL-based migrations...")
		if err := apps.MigrateAllSQL(db); err != nil {
			log.Fatal("SQL Migration gagal:", err)
		}
	default:
		log.Println("⚙️  Menjalankan GORM auto-migrations...")
		if err := apps.MigrateAll(db); err != nil {
			log.Fatal("GORM Migration gagal:", err)
		}
	}

	// Jalankan seed data
	apps.SeedAll(db)

	log.Println("✅ Semua migrasi selesai!")
}
