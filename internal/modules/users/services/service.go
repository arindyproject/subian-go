package services

import (
	authContracts "subian_go/internal/modules/auth/contracts"
	rbacContracts "subian_go/internal/modules/rbac/contracts"
	rbacDto "subian_go/internal/modules/rbac/dto"
	userContracts "subian_go/internal/modules/users/contracts"
	"subian_go/internal/modules/users/models"

	"subian_go/internal/shared/storage"
)

// ─── Service ───────────────────────────────────────────────────────────────────

type service struct {
	repo           userContracts.Repository
	rbacRepo       rbacContracts.RBACRepository
	authRepo       authContracts.AuthRepository
	storageService storage.ImageStorage
}

func NewUserService(
	repo userContracts.Repository,
	rbacRepo rbacContracts.RBACRepository,
	authRepo authContracts.AuthRepository,
	storageService storage.ImageStorage,
) userContracts.Service {
	return &service{
		repo:           repo,
		rbacRepo:       rbacRepo,
		authRepo:       authRepo,
		storageService: storageService,
	}
}

// ─── RBAC Data Builder ─────────────────────────────────────────────────────────
// Letakkan di internal/modules/users/services/user_service.go
// Ganti fungsi buildUserRBAC yang lama dengan ini

// buildUserRBAC mengambil roles (tanpa permissions) dan permissions (object lengkap, deduplicated)
func (s *service) buildUserRBAC(userID int64) ([]rbacDto.RoleSimpleResponse, []rbacDto.PermissionResponse) {
	// 1. Ambil roles dari DB — tanpa preload permissions agar ringan
	roles, err := s.rbacRepo.GetUserRoles(userID)
	var roleSimple []rbacDto.RoleSimpleResponse
	if err == nil {
		roleSimple = rbacDto.ToRoleSimpleListResponse(roles)
	} else {
		roleSimple = []rbacDto.RoleSimpleResponse{}
	}

	// 2. Ambil semua permissions (dari role + direct) sebagai object lengkap
	// Gunakan map untuk deduplication berdasarkan permission ID
	permMap := make(map[int64]rbacDto.PermissionResponse)

	// 2a. Permissions dari role
	for _, role := range roles {
		for _, p := range role.Permissions {
			if _, exists := permMap[p.ID]; !exists {
				permMap[p.ID] = *rbacDto.ToPermissionResponse(&p)
			}
		}
	}

	// 2b. Direct permissions yang di-grant — override/tambah ke map
	directPerms, err := s.rbacRepo.GetUserDirectPermissions(userID)
	if err == nil {
		for _, up := range directPerms {
			if !up.IsGranted {
				// Direct deny — hapus dari map jika ada
				delete(permMap, up.PermissionID)
				continue
			}
			// Direct grant — tambah jika belum ada
			if _, exists := permMap[up.PermissionID]; !exists {
				perm, err := s.rbacRepo.GetPermissionByID(up.PermissionID)
				if err == nil && perm != nil {
					permMap[perm.ID] = *rbacDto.ToPermissionResponse(perm)
				}
			}
		}
	}

	// 3. Convert map ke slice
	permList := make([]rbacDto.PermissionResponse, 0, len(permMap))
	for _, p := range permMap {
		permList = append(permList, p)
	}

	return roleSimple, permList
}

func (s *service) buildUsersRBAC(userIDs []int64) map[int64][]rbacDto.RoleSimpleResponse {
	userRolesMap := make(map[int64][]rbacDto.RoleSimpleResponse)
	if len(userIDs) == 0 {
		return userRolesMap
	}

	// Menggunakan method batch repository yang sudah ada
	dbRolesMap, err := s.rbacRepo.GetUsersRoles(userIDs)
	if err != nil {
		return userRolesMap
	}

	// Konversi dari map[int64][]models.Role ke map[int64][]rbacDto.RoleSimpleResponse
	for userID, roles := range dbRolesMap {
		userRolesMap[userID] = rbacDto.ToRoleSimpleListResponse(roles)
	}

	return userRolesMap
}

// buildCreator mengambil data creator user
func (s *service) buildCreator(createdBy *int64) *models.UserCreator {
	if createdBy == nil {
		return nil
	}
	creator, err := s.repo.GetByID(*createdBy)
	if err != nil || creator == nil {
		return nil
	}
	return &models.UserCreator{
		ID:       creator.ID,
		Username: creator.Username,
		Name:     creator.Name,
	}
}
