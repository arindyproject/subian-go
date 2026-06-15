package services

import (
	"errors"
	"fmt"
	"time"

	"subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/modules/rbac/dto"
	"subian_go/internal/modules/rbac/models"

	userContracts "subian_go/internal/modules/users/contracts"
	appErrors "subian_go/internal/shared/errors"

	"gorm.io/gorm"
)

type rbacService struct {
	rbacRepo contracts.RBACRepository
	userRepo userContracts.Repository
}

func NewRBACService(rbacRepo contracts.RBACRepository, userRepo userContracts.Repository) contracts.RBACService {
	return &rbacService{rbacRepo: rbacRepo, userRepo: userRepo}
}

// ─── Permission CRUD ───────────────────────────────────────────────────────────

func (s *rbacService) CreatePermission(req *dto.CreatePermissionRequest, createdBy *int64) (*dto.PermissionResponse, error) {
	// Cek nama sudah ada
	existing, _ := s.rbacRepo.GetPermissionByName(req.Name)
	if existing != nil {
		return nil, errors.New("nama permission sudah digunakan")
	}

	p := &models.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}
	if err := s.rbacRepo.CreatePermission(p); err != nil {
		return nil, err
	}
	return dto.ToPermissionResponse(p), nil
}

func (s *rbacService) GetPermissionByID(id int64) (*dto.PermissionResponse, error) {
	p, err := s.rbacRepo.GetPermissionByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("permission tidak ditemukan")
	}
	return dto.ToPermissionResponse(p), nil
}

func (s *rbacService) ListPermissions(page, pageSize int) ([]dto.PermissionResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	items, total, err := s.rbacRepo.ListPermissions(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return dto.ToPermissionListResponse(items), total, nil
}

func (s *rbacService) UpdatePermission(id int64, req *dto.UpdatePermissionRequest, updatedBy *int64) (*dto.PermissionResponse, error) {
	p, err := s.rbacRepo.GetPermissionByID(id)
	if err != nil || p == nil {
		return nil, errors.New("permission tidak ditemukan")
	}

	if req.DisplayName != nil {
		p.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		p.Description = req.Description
	}
	p.UpdatedAt = time.Now()

	if err := s.rbacRepo.UpdatePermission(p); err != nil {
		return nil, err
	}
	return dto.ToPermissionResponse(p), nil
}

func (s *rbacService) DeletePermission(id int64) error {
	p, err := s.rbacRepo.GetPermissionByID(id)
	if err != nil || p == nil {
		return errors.New("permission tidak ditemukan")
	}
	return s.rbacRepo.DeletePermission(id)
}

// ─── Role CRUD ─────────────────────────────────────────────────────────────────

func (s *rbacService) CreateRole(req *dto.CreateRoleRequest, createdBy *int64) (*dto.RoleResponse, error) {
	existing, _ := s.rbacRepo.GetRoleByName(req.Name)
	if existing != nil {
		return nil, errors.New("nama role sudah digunakan")
	}

	role := &models.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}
	if err := s.rbacRepo.CreateRole(role); err != nil {
		return nil, err
	}
	return dto.ToRoleResponse(role), nil
}

func (s *rbacService) GetRoleByID(id int64) (*dto.RoleResponse, error) {
	role, err := s.rbacRepo.GetRoleByID(id)
	if err != nil || role == nil {
		return nil, errors.New("role tidak ditemukan")
	}
	return dto.ToRoleResponse(role), nil
}

func (s *rbacService) ListRoles(page, pageSize int) ([]dto.RoleResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	items, total, err := s.rbacRepo.ListRoles(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return dto.ToRoleListResponse(items), total, nil
}

func (s *rbacService) UpdateRole(id int64, req *dto.UpdateRoleRequest, updatedBy *int64) (*dto.RoleResponse, error) {
	role, err := s.rbacRepo.GetRoleByID(id)
	if err != nil || role == nil {
		return nil, errors.New("role tidak ditemukan")
	}
	if role.IsSystem {
		return nil, errors.New("system role tidak bisa diubah")
	}

	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = req.Description
	}
	role.UpdatedBy = updatedBy
	role.UpdatedAt = time.Now()

	if err := s.rbacRepo.UpdateRole(role); err != nil {
		return nil, err
	}
	return dto.ToRoleResponse(role), nil
}

func (s *rbacService) DeleteRole(id int64) error {
	role, err := s.rbacRepo.GetRoleByID(id)
	if err != nil || role == nil {
		return errors.New("role tidak ditemukan")
	}
	if role.IsSystem {
		return errors.New("system role tidak bisa dihapus")
	}
	return s.rbacRepo.DeleteRole(id)
}

// ─── Role ↔ Permission ─────────────────────────────────────────────────────────

func (s *rbacService) AssignPermissionsToRole(roleID int64, req *dto.AssignPermissionsRequest) error {
	role, err := s.rbacRepo.GetRoleByID(roleID)
	if err != nil || role == nil {
		return errors.New("role tidak ditemukan")
	}
	return s.rbacRepo.AssignPermissionsToRole(roleID, req.PermissionIDs)
}

func (s *rbacService) RevokePermissionsFromRole(roleID int64, req *dto.AssignPermissionsRequest) error {
	role, err := s.rbacRepo.GetRoleByID(roleID)
	if err != nil || role == nil {
		return errors.New("role tidak ditemukan")
	}
	return s.rbacRepo.RevokePermissionsFromRole(roleID, req.PermissionIDs)
}

func (s *rbacService) SyncRolePermissions(roleID int64, req *dto.AssignPermissionsRequest) error {
	role, err := s.rbacRepo.GetRoleByID(roleID)
	if err != nil || role == nil {
		return errors.New("role tidak ditemukan")
	}
	return s.rbacRepo.SyncRolePermissions(roleID, req.PermissionIDs)
}

// ─── User ↔ Role ───────────────────────────────────────────────────────────────

func (s *rbacService) AssignRolesToUser(userID int64, req *dto.AssignRolesRequest, assignedBy *int64) error {
	return s.rbacRepo.AssignRolesToUser(userID, req.RoleIDs, assignedBy)
}

func (s *rbacService) RevokeRolesFromUser(userID int64, req *dto.AssignRolesRequest) error {
	return s.rbacRepo.RevokeRolesFromUser(userID, req.RoleIDs)
}

func (s *rbacService) SyncUserRoles(userID int64, req *dto.AssignRolesRequest, assignedBy *int64) error {
	return s.rbacRepo.SyncUserRoles(userID, req.RoleIDs, assignedBy)
}

// ─── User Permissions (direct) ───────────────────────────────────────────────
func (s *rbacService) GetUserRoles(userID int64) ([]dto.RoleResponse, error) {
	// 1. Validasi keberadaan user di database
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		// Cek apakah error karena record memang tidak ditemukan (GORM)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user tidak ditemukan")
		}
		// Jika error lain (misal: koneksi database putus, query salah)
		return nil, appErrors.Internal("gagal memeriksa keberadaan user: " + err.Error())
	}

	// Fallback jika repository mengembalikan nil, nil saat user tidak ada
	if user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	// 2. Ambil roles user
	roles, err := s.rbacRepo.GetUserRoles(userID)
	if err != nil {
		// Sertakan err.Error() agar pesan error asli dari database terlihat di log
		return nil, appErrors.Internal("gagal mengambil roles untuk user: " + err.Error())
	}

	// 3. PENGECEKAN BARU: Jika roles kosong (tidak ada role yang terassign)
	if len(roles) == 0 {
		// Mengembalikan error 404 dengan pesan spesifik menyertakan ID user
		return nil, appErrors.NotFound(fmt.Sprintf("roles pada user %d tidak ditemukan", user.ID))
	}

	// 4. Kembalikan response DTO
	return dto.ToRoleListResponse(roles), nil
}

// ─── User ↔ Permission (direct) ───────────────────────────────────────────────

func (s *rbacService) AssignDirectPermission(userID int64, req *dto.AssignDirectPermissionRequest, assignedBy *int64) error {
	return s.rbacRepo.AssignDirectPermission(userID, req.PermissionID, req.IsGranted, assignedBy)
}

func (s *rbacService) RevokeDirectPermission(userID, permissionID int64) error {
	return s.rbacRepo.RevokeDirectPermission(userID, permissionID)
}

func (s *rbacService) GetUserDirectPermissions(userID int64) ([]dto.DirectPermissionResponse, error) {
	items, err := s.rbacRepo.GetUserDirectPermissions(userID)
	if err != nil {
		return nil, err
	}
	var result []dto.DirectPermissionResponse
	for _, up := range items {
		perm, _ := s.rbacRepo.GetPermissionByID(up.PermissionID)
		if perm == nil {
			continue
		}
		result = append(result, dto.DirectPermissionResponse{
			Permission: *dto.ToPermissionResponse(perm),
			IsGranted:  up.IsGranted,
		})
	}
	return result, nil
}

// ─── Check ─────────────────────────────────────────────────────────────────────

func (s *rbacService) GetUserAllPermissions(userID int64) ([]string, error) {
	return s.rbacRepo.GetUserAllPermissions(userID)
}

func (s *rbacService) HasPermission(userID int64, permission string) (bool, error) {
	return s.rbacRepo.HasPermission(userID, permission)
}
