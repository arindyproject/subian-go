package models

import (
	"time"

	"gorm.io/gorm"
)

// ─── Permission ────────────────────────────────────────────────────────────────

// Permission merepresentasikan satu aksi yang bisa dilakukan
// Format: "resource:action" contoh: "users:read", "users:create", "roles:manage"
type Permission struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string         `gorm:"column:name;type:varchar(100);uniqueIndex;not null" json:"name"`
	DisplayName string         `gorm:"column:display_name;type:varchar(255);not null" json:"display_name"`
	Description *string        `gorm:"column:description;type:text" json:"description"`
	Resource    string         `gorm:"column:resource;type:varchar(100);not null;index" json:"resource"`
	Action      string         `gorm:"column:action;type:varchar(100);not null" json:"action"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:NOW()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"`
}

func (Permission) TableName() string { return "permissions" }

// ─── Role ──────────────────────────────────────────────────────────────────────

// Role merepresentasikan kumpulan permission
type Role struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string         `gorm:"column:name;type:varchar(100);uniqueIndex;not null" json:"name"`
	DisplayName string         `gorm:"column:display_name;type:varchar(255);not null" json:"display_name"`
	Description *string        `gorm:"column:description;type:text" json:"description"`
	IsSystem    bool           `gorm:"column:is_system;not null;default:false" json:"is_system"` // role bawaan sistem, tidak bisa dihapus
	CreatedBy   *int64         `gorm:"column:created_by" json:"created_by"`
	UpdatedBy   *int64         `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:NOW()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"`

	// Relasi (lazy load)
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

// ─── RolePermission (pivot) ────────────────────────────────────────────────────

// RolePermission adalah pivot table antara Role dan Permission
type RolePermission struct {
	RoleID       int64     `gorm:"primaryKey;column:role_id" json:"role_id"`
	PermissionID int64     `gorm:"primaryKey;column:permission_id" json:"permission_id"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// ─── UserRole (pivot) ──────────────────────────────────────────────────────────

// UserRole adalah pivot table antara User dan Role
type UserRole struct {
	UserID    int64     `gorm:"primaryKey;column:user_id" json:"user_id"`
	RoleID    int64     `gorm:"primaryKey;column:role_id" json:"role_id"`
	CreatedBy *int64    `gorm:"column:created_by" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
}

func (UserRole) TableName() string { return "user_roles" }

// ─── UserPermission (pivot) ────────────────────────────────────────────────────

// UserPermission adalah permission langsung ke user (override role)
type UserPermission struct {
	UserID       int64     `gorm:"primaryKey;column:user_id" json:"user_id"`
	PermissionID int64     `gorm:"primaryKey;column:permission_id" json:"permission_id"`
	IsGranted    bool      `gorm:"column:is_granted;not null;default:true" json:"is_granted"` // false = deny
	CreatedBy    *int64    `gorm:"column:created_by" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
}

func (UserPermission) TableName() string { return "user_permissions" }

// ─── Constants ─────────────────────────────────────────────────────────────────

// System Roles — tidak bisa dihapus
const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
	RoleStaff      = "staff"
	RoleUser       = "user"
)

// Permission format: "resource:action"
const (
	// Users
	PermUsersRead   = "users:read"
	PermUsersCreate = "users:create"
	PermUsersUpdate = "users:update"
	PermUsersDelete = "users:delete"
	PermUsersManage = "users:manage" // semua aksi users

	// Roles
	PermRolesRead   = "roles:read"
	PermRolesCreate = "roles:create"
	PermRolesUpdate = "roles:update"
	PermRolesDelete = "roles:delete"
	PermRolesManage = "roles:manage"

	// Permissions
	PermPermissionsRead   = "permissions:read"
	PermPermissionsCreate = "permissions:create"
	PermPermissionsUpdate = "permissions:update"
	PermPermissionsDelete = "permissions:delete"
	PermPermissionsManage = "permissions:manage"

	// Any
	PermAnyRead   = "any:read"
	PermAnyCreate = "any:create"
	PermAnyUpdate = "any:update"
	PermAnyDelete = "any:delete"
	PermAnyManage = "any:manage"

	// Master
	PermMasterCreate = "master:create"
	PermMasterUpdate = "master:update"
	PermMasterDelete = "master:delete"
	PermMasterManage = "master:manage"
)
