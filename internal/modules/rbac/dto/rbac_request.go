package dto

// ═══════════════════════════════════════════════════════════════
// REQUEST DTOs
// ═══════════════════════════════════════════════════════════════

// ─── Permission ────────────────────────────────────────────────

type CreatePermissionRequest struct {
	Name        string  `json:"name"         validate:"required,min=3,max=100"`
	DisplayName string  `json:"display_name" validate:"required,min=3,max=255"`
	Description *string `json:"description"  validate:"omitempty,max=500"`
	Resource    string  `json:"resource"     validate:"required,min=2,max=100"`
	Action      string  `json:"action"       validate:"required,min=2,max=100"`
}

type UpdatePermissionRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description"  validate:"omitempty,max=500"`
}

// ─── Role ──────────────────────────────────────────────────────

type CreateRoleRequest struct {
	Name        string  `json:"name"         validate:"required,min=3,max=100"`
	DisplayName string  `json:"display_name" validate:"required,min=3,max=255"`
	Description *string `json:"description"  validate:"omitempty,max=500"`
}

type UpdateRoleRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description"  validate:"omitempty,max=500"`
}

// ─── Assignment ────────────────────────────────────────────────

type AssignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids" validate:"required,min=1"`
}

type AssignRolesRequest struct {
	RoleIDs []int64 `json:"role_ids" validate:"required,min=1"`
}

type AssignDirectPermissionRequest struct {
	PermissionID int64 `json:"permission_id" validate:"required"`
	IsGranted    bool  `json:"is_granted"` // true = grant, false = deny
}
