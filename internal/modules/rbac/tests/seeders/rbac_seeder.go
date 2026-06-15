package seeders

import (
	"log"

	"subian_go/internal/modules/rbac/models"
	"subian_go/internal/modules/rbac/tests/factories"

	userModels "subian_go/internal/modules/users/models"

	"gorm.io/gorm"
)

type RBACSeeder struct {
	db *gorm.DB
}

func NewRBACSeeder(db *gorm.DB) *RBACSeeder {
	return &RBACSeeder{db: db}
}

// ─── PUBLIC METHODS ────────────────────────────────────────────────────────────

func (s *RBACSeeder) Run() error {
	log.Println("🌱 Seeding RBAC...")

	perms, err := s.seedPermissions()
	if err != nil {
		return err
	}

	roles, err := s.seedRoles()
	if err != nil {
		return err
	}

	if err := s.assignPermissionsToRoles(perms, roles); err != nil {
		return err
	}

	if err := s.assignRolesToUsers(roles); err != nil {
		return err
	}

	if err := s.assignPermissionsToUsers(perms); err != nil {
		return err
	}

	log.Println("✅ RBAC seeding selesai!")
	return nil
}

func (s *RBACSeeder) Fresh() error {
	log.Println("🗑️  Reset data RBAC...")

	if err := s.db.Exec("DELETE FROM role_permissions").Error; err != nil {
		return err
	}
	if err := s.db.Exec("DELETE FROM roles").Error; err != nil {
		return err
	}
	if err := s.db.Exec("DELETE FROM permissions").Error; err != nil {
		return err
	}

	// reset sequence (optional, PostgreSQL)
	s.db.Exec("ALTER SEQUENCE role_permissions_role_id_seq RESTART WITH 1")
	s.db.Exec("ALTER SEQUENCE roles_id_seq RESTART WITH 1")
	s.db.Exec("ALTER SEQUENCE permissions_id_seq RESTART WITH 1")

	log.Println("✅ Data RBAC direset.")
	return s.Run()
}

// ─── PERMISSIONS ───────────────────────────────────────────────────────────────

func (s *RBACSeeder) seedPermissions() (map[string]*models.Permission, error) {
	log.Println("   🔑 Seeding permissions...")

	list := []*models.Permission{
		//users
		factories.MakeUserReadPermission("users"),
		factories.MakeUserWritePermission("users"),
		factories.MakeUserUpdatePermission("users"),
		factories.MakeUserDeletePermission("users"),
		factories.MakeUserAllPermission("users"),
		//roles
		factories.MakeUserReadPermission("roles"),
		factories.MakeUserWritePermission("roles"),
		factories.MakeUserUpdatePermission("roles"),
		factories.MakeUserDeletePermission("roles"),
		factories.MakeUserAllPermission("roles"),
		//permissions
		factories.MakeUserReadPermission("permissions"),
		factories.MakeUserWritePermission("permissions"),
		factories.MakeUserUpdatePermission("permissions"),
		factories.MakeUserDeletePermission("permissions"),
		factories.MakeUserAllPermission("permissions"),
		//any
		factories.MakeUserReadPermission("any"),
		factories.MakeUserWritePermission("any"),
		factories.MakeUserUpdatePermission("any"),
		factories.MakeUserDeletePermission("any"),
		factories.MakeUserAllPermission("any"),
	}

	result := make(map[string]*models.Permission)

	for _, perm := range list {
		var existing models.Permission

		err := s.db.Where("name = ?", perm.Name).First(&existing).Error
		if err == nil {
			log.Printf("   ⏭️  Permission '%s' sudah ada", perm.Name)
			result[perm.Name] = &existing
			continue
		}

		if err := s.db.Create(perm).Error; err != nil {
			return nil, err
		}

		log.Printf("   ✅ Permission '%s' dibuat", perm.Name)
		result[perm.Name] = perm
	}

	return result, nil
}

// ─── ROLES ─────────────────────────────────────────────────────────────────────

func (s *RBACSeeder) seedRoles() (map[string]*models.Role, error) {
	log.Println("   👥 Seeding roles...")

	list := []*models.Role{
		factories.MakeSuperuserRole(),
		factories.MakeAdminRole(),
		factories.MakeUserRole(),
		factories.MakeGuestRole(),
	}

	result := make(map[string]*models.Role)

	for _, role := range list {
		var existing models.Role

		err := s.db.Where("name = ?", role.Name).First(&existing).Error
		if err == nil {
			log.Printf("   ⏭️  Role '%s' sudah ada", role.Name)
			result[role.Name] = &existing
			continue
		}

		if err := s.db.Create(role).Error; err != nil {
			return nil, err
		}

		log.Printf("   ✅ Role '%s' dibuat", role.Name)
		result[role.Name] = role
	}

	return result, nil
}

// ─── ASSIGNMENT (CORE RBAC) ────────────────────────────────────────────────────

func (s *RBACSeeder) assignPermissionsToRoles(
	perms map[string]*models.Permission,
	roles map[string]*models.Role,
) error {

	log.Println("   🔗 Assign permissions ke roles...")

	assign := map[string][]string{
		"superuser": {
			"users:manage",
		},
		"admin": {
			"users:read",
			"users:write",
			"users:delete",
		},
		"user": {
			"users:read",
		},
		"guest": {},
	}

	for roleName, permNames := range assign {
		role := roles[roleName]

		for _, permName := range permNames {
			perm := perms[permName]

			var count int64
			s.db.Model(&models.RolePermission{}).
				Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).
				Count(&count)

			if count > 0 {
				continue
			}

			err := s.db.Create(&models.RolePermission{
				RoleID:       role.ID,
				PermissionID: perm.ID,
			}).Error

			if err != nil {
				return err
			}

			log.Printf("   ✅ %s -> %s", role.Name, perm.Name)
		}
	}

	return nil
}

// ─── ASSIGNMENT Role To User ────────────────────────────────────────────────────
func (s *RBACSeeder) assignRolesToUsers(roles map[string]*models.Role) error {
	log.Println("   🔗 Assign roles ke users...")
	// Ambil user superadmin
	var superadmin userModels.User
	if err := s.db.Where("username = ?", "superadmin").First(&superadmin).Error; err != nil {
		log.Printf("   ⚠️  superadmin tidak ditemukan, skip assign role: %v", err)
		return nil
	}

	// Assign semua role ke superuser
	for _, role := range roles {
		var count int64
		s.db.Model(&models.UserRole{}).
			Where("user_id = ? AND role_id = ?", superadmin.ID, role.ID).
			Count(&count)

		if count > 0 {
			continue
		}

		err := s.db.Create(&models.UserRole{
			UserID: superadmin.ID,
			RoleID: role.ID,
		}).Error

		if err != nil {
			return err
		}

		log.Printf("   ✅ User '%s' diberikan role '%s'", superadmin.Username, role.Name)
	}

	//buat user random untuk role random
	//-----------------------------------------------------------------------------------
	users := []userModels.User{}
	if err := s.db.Find(&users).Error; err != nil {
		log.Printf("   ⚠️  Gagal mengambil users untuk assign role: %v", err)
		return nil
	}

	for _, user := range users {
		if user.Username == "superadmin" {
			continue
		}

		for _, role := range roles {
			var count int64
			s.db.Model(&models.UserRole{}).
				Where("user_id = ? AND role_id = ?", user.ID, role.ID).
				Count(&count)

			if count > 0 {
				continue
			}

			err := s.db.Create(&models.UserRole{
				UserID: user.ID,
				RoleID: role.ID,
			}).Error

			if err != nil {
				return err
			}

			log.Printf("   ✅ User '%s' diberikan role '%s'", user.Username, role.Name)
			break // assign 1 role saja untuk user selain superuser
		}
	}

	return nil
}

// ─── ASSIGNMENT Permission To User ────────────────────────────────────────────────
func (s *RBACSeeder) assignPermissionsToUsers(perms map[string]*models.Permission) error {
	log.Println("   🔗 Assign permissions langsung ke users...")

	// Ambil user superuser
	var superadmin userModels.User
	if err := s.db.Where("username = ?", "superadmin").First(&superadmin).Error; err != nil {
		log.Printf("   ⚠️  superadmin tidak ditemukan, skip assign direct permission: %v", err)
		return nil
	}

	// Assign semua permission ke superadmin
	for _, perm := range perms {
		var count int64
		s.db.Model(&models.UserPermission{}).
			Where("user_id = ? AND permission_id = ?", superadmin.ID, perm.ID).
			Count(&count)

		if count > 0 {
			continue
		}

		err := s.db.Create(&models.UserPermission{
			UserID:       superadmin.ID,
			PermissionID: perm.ID,
			IsGranted:    true,
		}).Error

		if err != nil {
			return err
		}

		log.Printf("   ✅ User '%s' diberikan direct permission '%s'", superadmin.Username, perm.Name)
	}

	// ambil user random untuk permission
	// -----------------------------------------------------------------------------------
	users := []userModels.User{}

	if err := s.db.Find(&users).Error; err != nil {
		log.Printf("   ⚠️  Gagal mengambil users untuk assign permission: %v", err)
		return nil
	}

	for _, user := range users {
		if user.Username == "superadmin" {
			continue
		}

		for _, prm := range perms {
			var count int64
			s.db.Model(&models.UserRole{}).
				Where("user_id = ? AND role_id = ?", user.ID, prm.ID).
				Count(&count)

			if count > 0 {
				continue
			}

			err := s.db.Create(&models.UserPermission{
				UserID:       user.ID,
				PermissionID: prm.ID,
			}).Error

			if err != nil {
				return err
			}

			log.Printf("   ✅ User '%s' diberikan Permission '%s'", user.Username, prm.Name)
			break // assign 1 Permission saja untuk user selain superuser
		}
	}

	return nil
}
