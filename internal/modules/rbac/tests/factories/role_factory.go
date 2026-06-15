package factories

import (
	"fmt"
	"math/rand"
	"time"

	"subian_go/internal/modules/rbac/models"
)

var roleRng = rand.New(rand.NewSource(time.Now().UnixNano()))

type RoleFactory struct {
	overrides map[string]interface{}
}

func NewRoleFactory() *RoleFactory {
	return &RoleFactory{
		overrides: make(map[string]interface{}),
	}
}

func (f *RoleFactory) With(field string, value interface{}) *RoleFactory {
	f.overrides[field] = value
	return f
}

func (f *RoleFactory) Make() *models.Role {
	idx := roleRng.Intn(999999)

	name := fmt.Sprintf("role_%d", idx)
	displayName := fmt.Sprintf("Role %d", idx)

	if v, ok := f.overrides["name"]; ok {
		name = v.(string)
	}
	if v, ok := f.overrides["display_name"]; ok {
		displayName = v.(string)
	}

	var description *string
	if v, ok := f.overrides["description"]; ok {
		val := v.(string)
		description = &val
	}

	isSystem := false
	if v, ok := f.overrides["is_system"]; ok {
		isSystem = v.(bool)
	}

	return &models.Role{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		IsSystem:    isSystem,
	}
}

func (f *RoleFactory) MakeMany(n int) []*models.Role {
	items := make([]*models.Role, n)
	for i := 0; i < n; i++ {
		items[i] = NewRoleFactory().Make()
	}
	return items
}

// Make Roles
func MakeSuperuserRole() *models.Role {
	return &models.Role{
		Name:        "superuser",
		DisplayName: "Superuser",
		Description: nil,
		IsSystem:    true,
	}
}

func MakeAdminRole() *models.Role {
	return &models.Role{
		Name:        "admin",
		DisplayName: "Administrator",
		Description: nil,
		IsSystem:    true,
	}
}

func MakeUserRole() *models.Role {
	return &models.Role{
		Name:        "user",
		DisplayName: "User",
		Description: nil,
		IsSystem:    true,
	}
}

func MakeGuestRole() *models.Role {
	return &models.Role{
		Name:        "guest",
		DisplayName: "Guest",
		Description: nil,
		IsSystem:    true,
	}
}

// makse assing permission ke role UserPermission ke superuser, admin
func MakeSuperuserRoleWithPermissions(perms []models.Permission) *models.Role {
	role := MakeSuperuserRole()
	role.Permissions = perms
	return role
}

func MakeAdminRoleWithPermissions(perms []models.Permission) *models.Role {
	role := MakeAdminRole()
	role.Permissions = perms
	return role
}

func MakeUserRoleWithPermissions(perms []models.Permission) *models.Role {
	role := MakeUserRole()
	role.Permissions = perms
	return role
}

func MakeGuestRoleWithPermissions(perms []models.Permission) *models.Role {
	role := MakeGuestRole()
	role.Permissions = perms
	return role
}
