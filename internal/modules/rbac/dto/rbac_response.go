package dto

import (
	"subian_go/internal/modules/rbac/models"
)

// ─── Permission ────────────────────────────────────────────────────────────────

type PermissionResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description *string `json:"description"`
	Resource    string  `json:"resource"`
	Action      string  `json:"action"`
	//CreatedAt   time.Time `json:"created_at"`
	//UpdatedAt   time.Time `json:"updated_at"`
}

func ToPermissionResponse(p *models.Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Description: p.Description,
		Resource:    p.Resource,
		Action:      p.Action,
	}
}

func ToPermissionListResponse(items []models.Permission) []PermissionResponse {
	result := make([]PermissionResponse, 0, len(items))
	for _, p := range items {
		result = append(result, *ToPermissionResponse(&p))
	}
	return result
}

// ─── Role (dengan permissions) ─────────────────────────────────────────────────

// RoleResponse dipakai untuk endpoint CRUD role — include permissions di dalamnya
type RoleResponse struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"display_name"`
	Description *string              `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	//CreatedAt   time.Time            `json:"created_at"`
	//UpdatedAt   time.Time            `json:"updated_at"`
}

func ToRoleResponse(r *models.Role) *RoleResponse {
	resp := &RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
	}
	if r.Permissions != nil {
		resp.Permissions = ToPermissionListResponse(r.Permissions)
	}
	return resp
}

func ToRoleListResponse(items []models.Role) []RoleResponse {
	result := make([]RoleResponse, 0, len(items))
	for _, r := range items {
		result = append(result, *ToRoleResponse(&r))
	}
	return result
}

// ─── Role Simple (tanpa permissions) ──────────────────────────────────────────

// RoleSimpleResponse dipakai di UserResponse — tanpa permissions
// karena permissions sudah ditampilkan flat di field "permissions" UserResponse
type RoleSimpleResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description *string `json:"description"`
	IsSystem    bool    `json:"is_system"`
	//CreatedAt   time.Time `json:"created_at"`
	//UpdatedAt   time.Time `json:"updated_at"`
}

func ToRoleSimpleResponse(r *models.Role) *RoleSimpleResponse {
	return &RoleSimpleResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
	}
}

func ToRoleSimpleListResponse(items []models.Role) []RoleSimpleResponse {
	result := make([]RoleSimpleResponse, 0, len(items))
	for _, r := range items {
		result = append(result, *ToRoleSimpleResponse(&r))
	}
	return result
}

// ─── Direct Permission ─────────────────────────────────────────────────────────

type DirectPermissionResponse struct {
	Permission PermissionResponse `json:"permission"`
	IsGranted  bool               `json:"is_granted"`
}

// ─── User RBAC Summary ─────────────────────────────────────────────────────────

type UserRBACResponse struct {
	UserID         int64                      `json:"user_id"`
	Roles          []RoleSimpleResponse       `json:"roles"`
	Permissions    []DirectPermissionResponse `json:"direct_permissions"`
	AllPermissions []string                   `json:"all_permissions"`
}
