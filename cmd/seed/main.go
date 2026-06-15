package main

import (
	"flag"
	"log"

	"subian_go/config"

	// =====================================================================
	// import seeder di sini
	// =====================================================================
	// Menggunakan alias untuk membedakan kedua paket seeders

	rbacSeed "subian_go/internal/modules/rbac/tests/seeders"
	userSeed "subian_go/internal/modules/users/tests/seeders"
	// =====================================================================
)

func main() {
	cfg := config.LoadConfig()
	env := flag.String("env", cfg.EnvCode, "Environment (DEV atau PROD)")
	fresh := flag.Bool("fresh", false, "Hapus semua data lalu seed ulang")
	flag.Parse()

	if *env != "DEV" && *env != "PROD" {
		log.Fatal("❌ Environment tidak valid. Gunakan DEV atau PROD")
	}

	// Safety guard: fresh seed tidak boleh di PROD
	if *fresh && *env == "PROD" {
		log.Fatal("❌ Fresh seed TIDAK diizinkan di environment PROD!")
	}

	log.Printf("🚀 Menjalankan seeder untuk environment: %s", *env)

	db, err := cfg.ConnectDB()
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}
	defer config.CloseDB(db)

	// =====================================================================
	// TODO: Tambahkan seeder lain di sini
	// =====================================================================
	// Jalankan semua seeder di sini
	// Contoh: user seeder
	// =====================================================================
	userSeeder := userSeed.NewUserSeeder(db)

	if *fresh {
		if err := userSeeder.Fresh(); err != nil {
			log.Fatal("Gagal fresh seed users:", err)
		}
	} else {
		if err := userSeeder.Run(); err != nil {
			log.Fatal("Gagal seed users:", err)
		}
	}
	// =====================================================================
	// Contoh: RBAC seeder
	rbacSeeder := rbacSeed.NewRBACSeeder(db)

	if *fresh {
		if err := rbacSeeder.Fresh(); err != nil {
			log.Fatal("Gagal fresh seed RBAC:", err)
		}
	} else {
		if err := rbacSeeder.Run(); err != nil {
			log.Fatal("Gagal seed RBAC:", err)
		}
	}
	// =====================================================================

	// =====================================================================

	log.Println("✅ Seeding selesai!")
}
