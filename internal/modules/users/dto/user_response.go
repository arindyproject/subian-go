package dto

import (
	"time"

	authModels "subian_go/internal/modules/auth/models"
	rbacDto "subian_go/internal/modules/rbac/dto"
	"subian_go/internal/modules/users/models"
	"subian_go/internal/shared/types"
)

// ─── User Response (detail) ────────────────────────────────────────────────────

// UserResponse response lengkap untuk single user
// - roles      : tanpa permissions di dalamnya (flat)
// - permissions: semua permission dari role + direct, ditampilkan sebagai object lengkap
type UserResponse struct {
	ID                int64                        `json:"id"`
	Photo             *string                      `json:"photo"`
	PhotoThumbnail    *string                      `json:"photo_thumbnail"`
	Username          string                       `json:"username"`
	Email             string                       `json:"email"`
	Name              string                       `json:"name"`
	IsSuperadmin      bool                         `json:"is_superadmin"`
	IsActive          bool                         `json:"is_active"`
	IsStaff           bool                         `json:"is_staff"`
	IsVerified        bool                         `json:"is_verified"`
	PasswordChangedAt *time.Time                   `json:"password_changed_at"`
	LastLoginAt       *time.Time                   `json:"last_login_at"`
	Settings          []models.UserSetting         `json:"settings"`
	Roles             []rbacDto.RoleSimpleResponse `json:"roles"`       // ← tanpa permissions
	Permissions       []rbacDto.PermissionResponse `json:"permissions"` // ← object lengkap, deduplicated
	Histories         []authModels.LoginHistory    `json:"histories"`
	Creator           *models.UserCreator          `json:"creator"`
	CreatedAt         types.CustomTime             `json:"created_at"`
	UpdatedAt         types.CustomTime             `json:"updated_at"`
}

// ─── User Simple Response (list) ──────────────────────────────────────────────

// UserSimpleResponse response ringkas untuk list
type UserSimpleResponse struct {
	ID             int64                        `json:"id"`
	PhotoThumbnail *string                      `json:"photo_thumbnail"`
	Username       string                       `json:"username"`
	Email          string                       `json:"email"`
	Name           string                       `json:"name"`
	IsSuperadmin   bool                         `json:"is_superadmin"`
	IsActive       bool                         `json:"is_active"`
	IsStaff        bool                         `json:"is_staff"`
	IsVerified     bool                         `json:"is_verified"`
	Roles          []rbacDto.RoleSimpleResponse `json:"roles"`
	CreatedAt      types.CustomTime             `json:"created_at"`
	UpdatedAt      types.CustomTime             `json:"updated_at"`
}

// ─── Builders ──────────────────────────────────────────────────────────────────

// UserResponseParams parameter untuk ToUserResponse
type UserResponseParams struct {
	User        *models.User
	Roles       []rbacDto.RoleSimpleResponse // ← simple (tanpa permissions)
	Permissions []rbacDto.PermissionResponse // ← object lengkap
	Histories   []authModels.LoginHistory
	Creator     *models.UserCreator
}

type UserSimpleResponseParams struct {
	User  *models.User
	Roles []rbacDto.RoleSimpleResponse // ← simple (tanpa permissions)
}

type LoginHistoryResponse struct {
	CreatedAt types.CustomTime `json:"created_at"`
	Status    string           `json:"status"`
}

// ToUserResponse mengubah User + RBAC data menjadi UserResponse lengkap
func ToUserResponse(p UserResponseParams, is_allow bool) *UserResponse {
	roles := p.Roles
	if roles == nil {
		roles = []rbacDto.RoleSimpleResponse{}
	}

	histories := p.Histories
	if histories == nil {
		histories = []authModels.LoginHistory{}
	}

	result := make([]authModels.LoginHistory, 0, len(histories))

	for _, h := range histories {
		createdAt := h.CreatedAt

		result = append(result, authModels.LoginHistory{
			CreatedAt: createdAt,
			Status:    h.Status,
		})
	}

	if !is_allow {
		// Jika tidak diizinkan melihat detail, kembalikan response dengan data minimal
		return &UserResponse{
			ID:             p.User.ID,
			PhotoThumbnail: p.User.PhotoThumbnail,
			Username:       p.User.Username,
			Email:          p.User.Email,
			Name:           p.User.Name,
			IsSuperadmin:   p.User.IsSuperadmin,
			IsActive:       p.User.IsActive,
			IsStaff:        p.User.IsStaff,
			IsVerified:     p.User.IsVerified,
			Roles:          roles,
			CreatedAt:      types.CustomTime(p.User.CreatedAt),
			UpdatedAt:      types.CustomTime(p.User.UpdatedAt),
			Creator:        p.Creator,
			LastLoginAt:    p.User.LastLoginAt,
			Histories:      result,
		}
	}
	settings, _ := p.User.GetSettings()

	permissions := p.Permissions
	if permissions == nil {
		permissions = []rbacDto.PermissionResponse{}
	}

	histories = p.Histories
	if histories == nil {
		histories = []authModels.LoginHistory{}
	}

	return &UserResponse{
		ID:                p.User.ID,
		Photo:             p.User.Photo,
		PhotoThumbnail:    p.User.PhotoThumbnail,
		Username:          p.User.Username,
		Email:             p.User.Email,
		Name:              p.User.Name,
		IsSuperadmin:      p.User.IsSuperadmin,
		IsActive:          p.User.IsActive,
		IsStaff:           p.User.IsStaff,
		IsVerified:        p.User.IsVerified,
		PasswordChangedAt: p.User.PasswordChangedAt,
		LastLoginAt:       p.User.LastLoginAt,
		Settings:          settings,
		Roles:             roles,
		Permissions:       permissions,
		Histories:         histories,
		Creator:           p.Creator,
		CreatedAt:         types.CustomTime(p.User.CreatedAt),
		UpdatedAt:         types.CustomTime(p.User.UpdatedAt),
	}
}

// ToUserSimpleResponse mengubah models.User menjadi UserSimpleResponse
func ToUserSimpleResponse(u UserSimpleResponseParams) *UserSimpleResponse {
	roles := u.Roles
	if roles == nil {
		roles = []rbacDto.RoleSimpleResponse{}
	}
	return &UserSimpleResponse{
		ID:             u.User.ID,
		PhotoThumbnail: u.User.PhotoThumbnail,
		Username:       u.User.Username,
		Email:          u.User.Email,
		Name:           u.User.Name,
		IsSuperadmin:   u.User.IsSuperadmin,
		IsActive:       u.User.IsActive,
		IsStaff:        u.User.IsStaff,
		IsVerified:     u.User.IsVerified,
		CreatedAt:      types.CustomTime(u.User.CreatedAt),
		UpdatedAt:      types.CustomTime(u.User.UpdatedAt),
		Roles:          roles,
	}
}

func ToUserListResponse(users []models.User, userRolesMap map[int64][]rbacDto.RoleSimpleResponse) []UserSimpleResponse {
	responses := make([]UserSimpleResponse, 0, len(users))
	for _, u := range users {
		// Ambil roles dari map berdasarkan ID user, jika tidak ada set slice kosong
		roles, exists := userRolesMap[u.ID]
		if !exists {
			roles = []rbacDto.RoleSimpleResponse{}
		}

		responses = append(responses, *ToUserSimpleResponse(UserSimpleResponseParams{
			User:  &u,
			Roles: roles,
		}))
	}
	return responses
}

// ─── User Deleted Response ──────────────────────────────────────────────
type UserDeletedResponse struct {
	ID             int64                        `json:"id"`
	PhotoThumbnail *string                      `json:"photo_thumbnail"`
	Username       string                       `json:"username"`
	Email          string                       `json:"email"`
	Name           string                       `json:"name"`
	IsSuperadmin   bool                         `json:"is_superadmin"`
	IsActive       bool                         `json:"is_active"`
	IsStaff        bool                         `json:"is_staff"`
	IsVerified     bool                         `json:"is_verified"`
	Roles          []rbacDto.RoleSimpleResponse `json:"roles"`
	DeletedAt      time.Time                    `json:"deleted_at"`
	Deleter        *models.UserCreator          `json:"deleter"`
	DeleteReason   *string                      `json:"delete_reason"`
	Creator        *models.UserCreator          `json:"creator"`
}

type UserDeletedResponseParams struct {
	User    *models.User
	Roles   []rbacDto.RoleSimpleResponse // ← simple (tanpa permissions)
	Creator *models.UserCreator
	Deleter *models.UserCreator
}

func ToUserDeletedResponse(u UserDeletedResponseParams) *UserDeletedResponse {
	roles := u.Roles
	if roles == nil {
		roles = []rbacDto.RoleSimpleResponse{}
	}
	return &UserDeletedResponse{
		ID:             u.User.ID,
		PhotoThumbnail: u.User.PhotoThumbnail,
		Username:       u.User.Username,
		Email:          u.User.Email,
		Name:           u.User.Name,
		IsSuperadmin:   u.User.IsSuperadmin,
		IsActive:       u.User.IsActive,
		IsStaff:        u.User.IsStaff,
		IsVerified:     u.User.IsVerified,
		Roles:          roles,
		DeletedAt:      u.User.DeletedAt.Time,
		DeleteReason:   u.User.DeleteReason,
		Creator:        u.Creator,
		Deleter:        u.Deleter,
	}
}

func ToUserDeletedListResponse(users []models.User, userRolesMap map[int64][]rbacDto.RoleSimpleResponse, creatorsMap map[int64]*models.UserCreator, deletersMap map[int64]*models.UserCreator) []UserDeletedResponse {
	responses := make([]UserDeletedResponse, 0, len(users))
	for _, u := range users {
		// Ambil roles dari map berdasarkan ID user, jika tidak ada set slice kosong
		roles, exists := userRolesMap[u.ID]
		if !exists {
			roles = []rbacDto.RoleSimpleResponse{}
		}

		creator := creatorsMap[u.ID]
		deleter := deletersMap[u.ID]

		responses = append(responses, *ToUserDeletedResponse(UserDeletedResponseParams{
			User:    &u,
			Roles:   roles,
			Creator: creator,
			Deleter: deleter,
		}))
	}
	return responses
}
