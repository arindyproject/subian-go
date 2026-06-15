package migrations

import (
	"database/sql"
	_ "embed"
	"log"

	"subian_go/internal/modules/rbac/models"

	"gorm.io/gorm"
)

// ================================================ Users Migration ================================================
//

//go:embed 001_create_permissions_table.sql
var permissionsSQL string

//go:embed 002_create_roles_table.sql
var rolesSQL string

//go:embed 003_create_role_permissions_table.sql
var rolePermissionsSQL string

//go:embed 004_create_user_roles_table.sql
var userRolesSQL string

// MigrateRBAC menjalankan GORM auto-migration untuk semua tabel RBAC
func MigrateRBAC(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},
		&models.UserRole{},
	)
}

// MigrateRBACWithSQL menjalankan migrasi via raw SQL
func MigrateRBACWithSQL(sqlDB *sql.DB) error {
	sqls := []struct {
		name  string
		query string
	}{
		{"permissions", permissionsSQL},
		{"roles", rolesSQL},
		{"role_permissions", rolePermissionsSQL},
		{"user_roles", userRolesSQL},
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

// DropRBACTables menghapus semua tabel RBAC (gunakan dengan hati-hati!)
func DropRBACTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.UserRole{},
		&models.RolePermission{},
		&models.Role{},
		&models.Permission{},
	)
}

// RollbackRBAC rolls back the RBAC migration
func RollbackRBAC(db *gorm.DB) error {
	return DropRBACTables(db)
}

// ================================================ Models Migration ================================================
