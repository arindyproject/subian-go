package seeders

import (
	"log"

	"subian_go/internal/modules/users/models"
	"subian_go/internal/modules/users/tests/factories"

	"gorm.io/gorm"
)

// UserSeeder mengelola seeding data user ke database
type UserSeeder struct {
	db *gorm.DB
}

// NewUserSeeder membuat instance UserSeeder baru
func NewUserSeeder(db *gorm.DB) *UserSeeder {
	return &UserSeeder{db: db}
}

// Run menjalankan semua seeder
func (s *UserSeeder) Run() error {
	log.Println("🌱 Seeding users...")

	if err := s.seedSuperuser(); err != nil {
		return err
	}

	if err := s.seedStaff(); err != nil {
		return err
	}

	if err := s.seedRegularUsers(); err != nil {
		return err
	}

	log.Println("✅ Users seeding selesai!")
	return nil
}

// Fresh menghapus semua data user lalu seed ulang
func (s *UserSeeder) Fresh() error {
	log.Println("🗑️  Menghapus semua data user...")

	if err := s.db.Exec("DELETE FROM users").Error; err != nil {
		return err
	}

	// Reset auto increment sequence (PostgreSQL)
	if err := s.db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1").Error; err != nil {
		log.Printf("Warning: Gagal reset sequence: %v", err)
	}

	log.Println("✅ Data user dihapus.")
	return s.Run()
}

// ─── Seeder Methods ────────────────────────────────────────────────────────────

func (s *UserSeeder) seedSuperuser() error {
	user := factories.MakeSuperadminUser()

	// Skip jika sudah ada
	var count int64
	s.db.Model(&models.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		log.Printf("   ⏭️  Superuser '%s' sudah ada, skip.", user.Username)
		return nil
	}

	if err := s.db.Create(user).Error; err != nil {
		return err
	}

	log.Printf("   ✅ Superuser '%s' dibuat.", user.Username)
	return nil
}

func (s *UserSeeder) seedStaff() error {
	staffCount := 3

	for i := 1; i <= staffCount; i++ {
		user := factories.MakeStaffsUser(i)

		// Skip jika sudah ada
		var count int64
		s.db.Model(&models.User{}).Where("username = ?", user.Username).Count(&count)
		if count > 0 {
			log.Printf("   ⏭️  Staff '%s' sudah ada, skip.", user.Username)
			continue
		}

		if err := s.db.Create(user).Error; err != nil {
			return err
		}

		log.Printf("   ✅ Staff '%s' dibuat.", user.Username)
	}

	return nil
}

func (s *UserSeeder) seedRegularUsers() error {
	regularCount := 10

	users := factories.NewUserFactory().MakeMany(regularCount)

	for _, user := range users {
		if err := s.db.Create(user).Error; err != nil {
			log.Printf("   ⚠️  Gagal membuat user '%s': %v", user.Username, err)
			continue
		}
		log.Printf("   ✅ User '%s' dibuat.", user.Username)
	}

	return nil
}
