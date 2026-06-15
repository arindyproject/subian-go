package factories

import (
	"time"

	"subian_go/internal/modules/rbac/models"
	userModels "subian_go/internal/modules/users/models"
)

// ─── Permission Factory ────────────────────────────────────────────────────────

func MakePermission(overrides ...func(*models.Permission)) *models.Permission {
	desc := "Permission untuk testing"
	p := &models.Permission{
		ID:          1,
		Name:        "users:read",
		DisplayName: "Read Users",
		Description: &desc,
		Resource:    "users",
		Action:      "read",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, fn := range overrides {
		fn(p)
	}
	return p
}

func MakePermissionList(count int) []models.Permission {
	perms := make([]models.Permission, count)
	names := [][]string{
		{"users:read", "users", "read"},
		{"users:create", "users", "create"},
		{"users:update", "users", "update"},
		{"users:delete", "users", "delete"},
		{"roles:read", "roles", "read"},
		{"roles:create", "roles", "create"},
	}
	for i := 0; i < count; i++ {
		idx := i % len(names)
		desc := "Permission " + names[idx][0]
		perms[i] = models.Permission{
			ID:          int64(i + 1),
			Name:        names[idx][0],
			DisplayName: names[idx][0],
			Description: &desc,
			Resource:    names[idx][1],
			Action:      names[idx][2],
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}
	return perms
}

// ─── Role Factory ──────────────────────────────────────────────────────────────

func MakeRole(overrides ...func(*models.Role)) *models.Role {
	desc := "Role untuk testing"
	r := &models.Role{
		ID:          1,
		Name:        "admin",
		DisplayName: "Administrator",
		Description: &desc,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, fn := range overrides {
		fn(r)
	}
	return r
}

func MakeSystemRole() *models.Role {
	return MakeRole(func(r *models.Role) {
		r.Name = "superadmin"
		r.DisplayName = "Super Administrator"
		r.IsSystem = true
	})
}

func MakeRoleWithPermissions() *models.Role {
	role := MakeRole()
	role.Permissions = []models.Permission{
		*MakePermission(),
		*MakePermission(func(p *models.Permission) {
			p.ID = 2
			p.Name = "users:create"
			p.Action = "create"
		}),
	}
	return role
}

func MakeRoleList(count int) []models.Role {
	roles := make([]models.Role, count)
	for i := 0; i < count; i++ {
		desc := "Role testing"
		roles[i] = models.Role{
			ID:          int64(i + 1),
			Name:        "role_" + itoa(i+1),
			DisplayName: "Role " + itoa(i+1),
			Description: &desc,
			IsSystem:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}
	return roles
}

// ─── UserPermission Factory ────────────────────────────────────────────────────

func MakeUserPermission(userID, permissionID int64, isGranted bool) models.UserPermission {
	return models.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
		IsGranted:    isGranted,
		CreatedAt:    time.Now(),
	}
}

// ─── User Factory ──────────────────────────────────────────────────────────────

func MakeUser(overrides ...func(*userModels.User)) *userModels.User {
	now := time.Now()
	createdBy := int64(1)
	u := &userModels.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		IsStaff:   false,
		CreatedBy: &createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	for _, fn := range overrides {
		fn(u)
	}
	return u
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	digits := []byte{}
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
