package contracts

import (
	"subian_go/internal/modules/rbac/dto"
	"subian_go/internal/modules/rbac/models"
)

// ─── Repository ────────────────────────────────────────────────────────────────

type RBACRepository interface {
	IsSuperadmin(userID int64) (bool, error)
	// Permission
	CreatePermission(p *models.Permission) error
	GetPermissionByID(id int64) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	ListPermissions(page, pageSize int) ([]models.Permission, int64, error)
	UpdatePermission(p *models.Permission) error
	DeletePermission(id int64) error

	// Role
	CreateRole(r *models.Role) error
	GetRoleByID(id int64) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	ListRoles(page, pageSize int) ([]models.Role, int64, error)
	UpdateRole(r *models.Role) error
	DeleteRole(id int64) error
	GetUsersRoles(userIDs []int64) (map[int64][]models.Role, error)

	// Role ↔ Permission
	AssignPermissionsToRole(roleID int64, permissionIDs []int64) error
	RevokePermissionsFromRole(roleID int64, permissionIDs []int64) error
	GetRolePermissions(roleID int64) ([]models.Permission, error)
	SyncRolePermissions(roleID int64, permissionIDs []int64) error

	// User ↔ Role
	AssignRolesToUser(userID int64, roleIDs []int64, assignedBy *int64) error
	RevokeRolesFromUser(userID int64, roleIDs []int64) error
	GetUserRoles(userID int64) ([]models.Role, error)
	SyncUserRoles(userID int64, roleIDs []int64, assignedBy *int64) error

	// User ↔ Permission (direct)
	AssignDirectPermission(userID, permissionID int64, isGranted bool, assignedBy *int64) error
	RevokeDirectPermission(userID, permissionID int64) error
	GetUserDirectPermissions(userID int64) ([]models.UserPermission, error)

	// Check
	GetUserAllPermissions(userID int64) ([]string, error) // gabungan dari role + direct
	HasPermission(userID int64, permission string) (bool, error)
}

// ─── Service ───────────────────────────────────────────────────────────────────

type RBACService interface {
	// Permission CRUD
	CreatePermission(req *dto.CreatePermissionRequest, createdBy *int64) (*dto.PermissionResponse, error)
	GetPermissionByID(id int64) (*dto.PermissionResponse, error)
	ListPermissions(page, pageSize int) ([]dto.PermissionResponse, int64, error)
	UpdatePermission(id int64, req *dto.UpdatePermissionRequest, updatedBy *int64) (*dto.PermissionResponse, error)
	DeletePermission(id int64) error

	// Role CRUD
	CreateRole(req *dto.CreateRoleRequest, createdBy *int64) (*dto.RoleResponse, error)
	GetRoleByID(id int64) (*dto.RoleResponse, error)
	ListRoles(page, pageSize int) ([]dto.RoleResponse, int64, error)
	UpdateRole(id int64, req *dto.UpdateRoleRequest, updatedBy *int64) (*dto.RoleResponse, error)
	DeleteRole(id int64) error

	// Role ↔ Permission
	AssignPermissionsToRole(roleID int64, req *dto.AssignPermissionsRequest) error
	RevokePermissionsFromRole(roleID int64, req *dto.AssignPermissionsRequest) error
	SyncRolePermissions(roleID int64, req *dto.AssignPermissionsRequest) error

	// User ↔ Role
	AssignRolesToUser(userID int64, req *dto.AssignRolesRequest, assignedBy *int64) error
	RevokeRolesFromUser(userID int64, req *dto.AssignRolesRequest) error
	SyncUserRoles(userID int64, req *dto.AssignRolesRequest, assignedBy *int64) error
	GetUserRoles(userID int64) ([]dto.RoleResponse, error)

	// User ↔ Permission (direct)
	AssignDirectPermission(userID int64, req *dto.AssignDirectPermissionRequest, assignedBy *int64) error
	RevokeDirectPermission(userID, permissionID int64) error
	GetUserDirectPermissions(userID int64) ([]dto.DirectPermissionResponse, error)

	// Check
	GetUserAllPermissions(userID int64) ([]string, error)
	HasPermission(userID int64, permission string) (bool, error)
}
