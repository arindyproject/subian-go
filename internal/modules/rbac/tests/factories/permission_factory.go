package factories

import (
	"fmt"
	"math/rand"
	"time"

	"subian_go/internal/modules/rbac/models"
)

var permRng = rand.New(rand.NewSource(time.Now().UnixNano()))

type PermissionFactory struct {
	overrides map[string]interface{}
}

func NewPermissionFactory() *PermissionFactory {
	return &PermissionFactory{
		overrides: make(map[string]interface{}),
	}
}

func (f *PermissionFactory) With(field string, value interface{}) *PermissionFactory {
	f.overrides[field] = value
	return f
}

func (f *PermissionFactory) Make() *models.Permission {
	idx := permRng.Intn(999999)

	name := fmt.Sprintf("resource_%d:read", idx)
	displayName := fmt.Sprintf("Read Resource %d", idx)
	resource := fmt.Sprintf("resource_%d", idx)
	action := "read"

	if v, ok := f.overrides["name"]; ok {
		name = v.(string)
	}
	if v, ok := f.overrides["display_name"]; ok {
		displayName = v.(string)
	}
	if v, ok := f.overrides["resource"]; ok {
		resource = v.(string)
	}
	if v, ok := f.overrides["action"]; ok {
		action = v.(string)
	}

	var description *string
	if v, ok := f.overrides["description"]; ok {
		val := v.(string)
		description = &val
	}

	return &models.Permission{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Resource:    resource,
		Action:      action,
	}
}

func (f *PermissionFactory) MakeMany(n int) []*models.Permission {
	items := make([]*models.Permission, n)
	for i := 0; i < n; i++ {
		items[i] = NewPermissionFactory().Make()
	}
	return items
}

// Make Permmission untuk Users
// ---- Resource: users, Actions: read, write, delete, all
func MakeUserReadPermission(resourceName string) *models.Permission {
	return &models.Permission{
		Name:        fmt.Sprintf("%s:read", resourceName),
		DisplayName: fmt.Sprintf("Read %s", resourceName),
		Description: nil,
		Resource:    resourceName,
		Action:      "read",
	}
}

func MakeUserWritePermission(resourceName string) *models.Permission {
	return &models.Permission{
		Name:        fmt.Sprintf("%s:write", resourceName),
		DisplayName: fmt.Sprintf("Write %s", resourceName),
		Description: nil,
		Resource:    resourceName,
		Action:      "write",
	}
}

func MakeUserUpdatePermission(resourceName string) *models.Permission {
	return &models.Permission{
		Name:        fmt.Sprintf("%s:update", resourceName),
		DisplayName: fmt.Sprintf("Update %s", resourceName),
		Description: nil,
		Resource:    resourceName,
		Action:      "update",
	}
}

func MakeUserDeletePermission(resourceName string) *models.Permission {
	return &models.Permission{
		Name:        fmt.Sprintf("%s:delete", resourceName),
		DisplayName: fmt.Sprintf("Delete %s", resourceName),
		Description: nil,
		Resource:    resourceName,
		Action:      "delete",
	}
}

func MakeUserAllPermission(resourceName string) *models.Permission {
	return &models.Permission{
		Name:        fmt.Sprintf("%s:manage", resourceName),
		DisplayName: fmt.Sprintf("All permissions for %s", resourceName),
		Description: nil,
		Resource:    resourceName,
		Action:      "manage",
	}
}
