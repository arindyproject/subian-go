package tests

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"subian_go/internal/modules/rbac/dto"
	"subian_go/internal/modules/rbac/models"
	"subian_go/internal/modules/rbac/services"
	"subian_go/internal/modules/rbac/tests/factories"
	"subian_go/internal/modules/rbac/tests/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ─── Suite ─────────────────────────────────────────────────────────────────────

// Entrypoint suite

func TestMain(m *testing.M) {
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")
	fmt.Println("\033[35m  RBAC Service Test Suite\033[0m")
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")

	code := m.Run()

	if code == 0 {
		fmt.Println("\n\033[32m✓  PASS\033[0m  subian_go/internal/modules/rbac")
	} else {
		fmt.Println("\n\033[31m✗  FAIL\033[0m  subian_go/internal/modules/rbac")
	}

	os.Exit(code)
}

type RBACServiceTestSuite struct {
	suite.Suite
	rbacRepo *mocks.MockRBACRepository
	userRepo *mocks.MockUserRepository
	service  interface {
		// Permission
		CreatePermission(req *dto.CreatePermissionRequest, createdBy *int64) (*dto.PermissionResponse, error)
		GetPermissionByID(id int64) (*dto.PermissionResponse, error)
		ListPermissions(page, pageSize int) ([]dto.PermissionResponse, int64, error)
		UpdatePermission(id int64, req *dto.UpdatePermissionRequest, updatedBy *int64) (*dto.PermissionResponse, error)
		DeletePermission(id int64) error
		// Role
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
		GetUserRoles(userID int64) ([]dto.RoleResponse, error)
		AssignRolesToUser(userID int64, req *dto.AssignRolesRequest, assignedBy *int64) error
		RevokeRolesFromUser(userID int64, req *dto.AssignRolesRequest) error
		SyncUserRoles(userID int64, req *dto.AssignRolesRequest, assignedBy *int64) error
		// User ↔ Permission
		AssignDirectPermission(userID int64, req *dto.AssignDirectPermissionRequest, assignedBy *int64) error
		RevokeDirectPermission(userID, permissionID int64) error
		GetUserDirectPermissions(userID int64) ([]dto.DirectPermissionResponse, error)
		// Check
		GetUserAllPermissions(userID int64) ([]string, error)
		HasPermission(userID int64, permission string) (bool, error)
	}
	actorID *int64
}

func (s *RBACServiceTestSuite) SetupTest() {
	s.rbacRepo = new(mocks.MockRBACRepository)
	s.userRepo = new(mocks.MockUserRepository)
	s.service = services.NewRBACService(s.rbacRepo, s.userRepo)
	id := int64(1)
	s.actorID = &id
}

func (s *RBACServiceTestSuite) TearDownTest() {
	s.rbacRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func TestRBACService(t *testing.T) {
	suite.Run(t, new(RBACServiceTestSuite))
}

// ─── Helper ────────────────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) requireAppError(err error, wantStatus int) {
	s.Require().Error(err)
	appErr, ok := err.(interface{ StatusCode() int })
	s.Require().True(ok, "expected AppError dengan StatusCode()")
	s.Equal(wantStatus, appErr.StatusCode())
}

func dbError() error { return errors.New("db error") }

// ═══════════════════════════════════════════════════════════════════════════════
// PERMISSION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

// ─── CreatePermission ─────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestCreatePermission_Success() {
	req := &dto.CreatePermissionRequest{
		Name:        "users:read",
		DisplayName: "Read Users",
		Resource:    "users",
		Action:      "read",
	}

	// 1. Mock GetPermissionByName (return nil, nil menandakan nama belum ada)
	s.rbacRepo.On("GetPermissionByName", "users:read").Return(nil, nil)

	// 2. HAPUS baris factories.MakePermission() karena field waktu/ID tidak akan pernah cocok
	// s.rbacRepo.On("CreatePermission", factories.MakePermission()).Return(nil)

	// 3. Gunakan mock.MatchedBy untuk memvalidasi isi struct yang dikirim ke repository.
	// Kita hanya peduli pada field bisnis (Name, DisplayName, Resource, Action)
	// dan mengabaikan ID/CreatedAt/UpdatedAt yang di-generate oleh DB/GORM.
	s.rbacRepo.On("CreatePermission", mock.MatchedBy(func(p *models.Permission) bool {
		return p.Name == req.Name &&
			p.DisplayName == req.DisplayName &&
			p.Resource == req.Resource &&
			p.Action == req.Action
	})).Return(nil)

	/*
	   ATAU PENDEKATAN ALTERNATIF (Jika tidak ingin repot validasi isi struct):
	   Gunakan mock.Anything bawaan testify

	   s.rbacRepo.On("CreatePermission", mock.Anything).Return(nil)
	*/

	// Eksekusi Service
	result, err := s.service.CreatePermission(req, s.actorID)

	// Assertion
	s.NoError(err)
	s.NotNil(result)
	s.Equal("users:read", result.Name)
}

func (s *RBACServiceTestSuite) TestCreatePermission_DuplicateName() {
	req := &dto.CreatePermissionRequest{
		Name:        "users:read",
		DisplayName: "Read Users",
		Resource:    "users",
		Action:      "read",
	}

	s.rbacRepo.On("GetPermissionByName", "users:read").Return(factories.MakePermission(), nil)

	result, err := s.service.CreatePermission(req, s.actorID)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "sudah digunakan")
}

// ─── GetPermissionByID ────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestGetPermissionByID_Success() {
	perm := factories.MakePermission()
	s.rbacRepo.On("GetPermissionByID", int64(1)).Return(perm, nil)

	result, err := s.service.GetPermissionByID(1)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(int64(1), result.ID)
	s.Equal("users:read", result.Name)
}

func (s *RBACServiceTestSuite) TestGetPermissionByID_NotFound() {
	s.rbacRepo.On("GetPermissionByID", int64(99)).Return(nil, nil)

	result, err := s.service.GetPermissionByID(99)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

func (s *RBACServiceTestSuite) TestGetPermissionByID_DBError() {
	s.rbacRepo.On("GetPermissionByID", int64(1)).Return(nil, dbError())

	result, err := s.service.GetPermissionByID(1)

	s.Nil(result)
	s.Error(err)
}

// ─── ListPermissions ──────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestListPermissions_Success() {
	perms := factories.MakePermissionList(3)
	s.rbacRepo.On("ListPermissions", 1, 10).Return(perms, int64(3), nil)

	result, total, err := s.service.ListPermissions(1, 10)

	s.NoError(err)
	s.Equal(int64(3), total)
	s.Len(result, 3)
}

func (s *RBACServiceTestSuite) TestListPermissions_PageNormalized() {
	perms := factories.MakePermissionList(2)
	// page=0 harus dinormalisasi ke 1
	s.rbacRepo.On("ListPermissions", 1, 10).Return(perms, int64(2), nil)

	result, total, err := s.service.ListPermissions(0, 10)

	s.NoError(err)
	s.Equal(int64(2), total)
	s.Len(result, 2)
}

func (s *RBACServiceTestSuite) TestListPermissions_PageSizeNormalized() {
	perms := factories.MakePermissionList(5)
	// pageSize=0 harus dinormalisasi ke 10
	s.rbacRepo.On("ListPermissions", 1, 10).Return(perms, int64(5), nil)

	result, _, err := s.service.ListPermissions(1, 0)

	s.NoError(err)
	s.Len(result, 5)
}

// ─── UpdatePermission ─────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestUpdatePermission_Success() {
	perm := factories.MakePermission()
	newName := "Read All Users"
	req := &dto.UpdatePermissionRequest{DisplayName: &newName}

	s.rbacRepo.On("GetPermissionByID", int64(1)).Return(perm, nil)
	s.rbacRepo.On("UpdatePermission", perm).Return(nil)

	result, err := s.service.UpdatePermission(1, req, s.actorID)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("Read All Users", result.DisplayName)
}

func (s *RBACServiceTestSuite) TestUpdatePermission_UpdateDescription() {
	perm := factories.MakePermission()
	newDesc := "Deskripsi baru"
	req := &dto.UpdatePermissionRequest{Description: &newDesc}

	s.rbacRepo.On("GetPermissionByID", int64(1)).Return(perm, nil)
	s.rbacRepo.On("UpdatePermission", perm).Return(nil)

	result, err := s.service.UpdatePermission(1, req, s.actorID)

	s.NoError(err)
	s.Equal("Deskripsi baru", *result.Description)
}

func (s *RBACServiceTestSuite) TestUpdatePermission_NotFound() {
	req := &dto.UpdatePermissionRequest{}
	s.rbacRepo.On("GetPermissionByID", int64(99)).Return(nil, nil)

	result, err := s.service.UpdatePermission(99, req, s.actorID)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ─── DeletePermission ─────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestDeletePermission_Success() {
	perm := factories.MakePermission()
	s.rbacRepo.On("GetPermissionByID", int64(1)).Return(perm, nil)
	s.rbacRepo.On("DeletePermission", int64(1)).Return(nil)

	err := s.service.DeletePermission(1)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestDeletePermission_NotFound() {
	s.rbacRepo.On("GetPermissionByID", int64(99)).Return(nil, nil)

	err := s.service.DeletePermission(99)

	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ═══════════════════════════════════════════════════════════════════════════════
// ROLE TESTS
// ═══════════════════════════════════════════════════════════════════════════════

// ─── CreateRole ───────────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestCreateRole_Success() {
	req := &dto.CreateRoleRequest{
		Name:        "manager",
		DisplayName: "Manager",
	}

	// 1. Mock GetRoleByName (return nil, nil menandakan nama belum ada)
	s.rbacRepo.On("GetRoleByName", "manager").Return(nil, nil)

	// 2. HAPUS penggunaan mock_anything()
	// s.rbacRepo.On("CreateRole", mock_anything()).Return(nil)

	// 3. GUNAKAN mock.MatchedBy (Direkomendasikan untuk memvalidasi mapping data)
	s.rbacRepo.On("CreateRole", mock.MatchedBy(func(r *models.Role) bool {
		return r.Name == req.Name &&
			r.DisplayName == req.DisplayName
	})).Return(nil)

	/*
	   ALTERNATIF: Jika Anda hanya ingin melewati pengecekan argumen tanpa peduli isinya,
	   gunakan mock.Anything bawaan testify (JANGAN pakai mock_anything() buatan sendiri).

	   s.rbacRepo.On("CreateRole", mock.Anything).Return(nil)
	*/

	// Eksekusi Service
	result, err := s.service.CreateRole(req, s.actorID)

	// Assertion
	s.NoError(err)
	s.NotNil(result)
	s.Equal("manager", result.Name)
}

func (s *RBACServiceTestSuite) TestCreateRole_DuplicateName() {
	req := &dto.CreateRoleRequest{
		Name:        "admin",
		DisplayName: "Admin",
	}

	s.rbacRepo.On("GetRoleByName", "admin").Return(factories.MakeRole(), nil)

	result, err := s.service.CreateRole(req, s.actorID)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "sudah digunakan")
}

// ─── GetRoleByID ──────────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestGetRoleByID_Success() {
	role := factories.MakeRoleWithPermissions()
	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)

	result, err := s.service.GetRoleByID(1)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(int64(1), result.ID)
	s.Len(result.Permissions, 2)
}

func (s *RBACServiceTestSuite) TestGetRoleByID_NotFound() {
	s.rbacRepo.On("GetRoleByID", int64(99)).Return(nil, nil)

	result, err := s.service.GetRoleByID(99)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ─── ListRoles ────────────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestListRoles_Success() {
	roles := factories.MakeRoleList(3)
	s.rbacRepo.On("ListRoles", 1, 10).Return(roles, int64(3), nil)

	result, total, err := s.service.ListRoles(1, 10)

	s.NoError(err)
	s.Equal(int64(3), total)
	s.Len(result, 3)
}

func (s *RBACServiceTestSuite) TestListRoles_PageSizeCappedAt100() {
	roles := factories.MakeRoleList(2)
	// pageSize > 100 harus dinormalisasi ke 10
	s.rbacRepo.On("ListRoles", 1, 10).Return(roles, int64(2), nil)

	result, _, err := s.service.ListRoles(1, 999)

	s.NoError(err)
	s.Len(result, 2)
}

// ─── UpdateRole ───────────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestUpdateRole_Success() {
	role := factories.MakeRole()
	newDisplay := "Super Admin Updated"
	req := &dto.UpdateRoleRequest{DisplayName: &newDisplay}

	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)
	s.rbacRepo.On("UpdateRole", role).Return(nil)

	result, err := s.service.UpdateRole(1, req, s.actorID)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("Super Admin Updated", result.DisplayName)
}

func (s *RBACServiceTestSuite) TestUpdateRole_SystemRoleCannotBeUpdated() {
	role := factories.MakeSystemRole()
	newDisplay := "Hacked"
	req := &dto.UpdateRoleRequest{DisplayName: &newDisplay}

	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)

	result, err := s.service.UpdateRole(1, req, s.actorID)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak bisa diubah")
}

func (s *RBACServiceTestSuite) TestUpdateRole_NotFound() {
	req := &dto.UpdateRoleRequest{}
	s.rbacRepo.On("GetRoleByID", int64(99)).Return(nil, nil)

	result, err := s.service.UpdateRole(99, req, s.actorID)

	s.Nil(result)
	s.Error(err)
}

// ─── DeleteRole ───────────────────────────────────────────────────────────────

func (s *RBACServiceTestSuite) TestDeleteRole_Success() {
	role := factories.MakeRole() // bukan system role
	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)
	s.rbacRepo.On("DeleteRole", int64(1)).Return(nil)

	err := s.service.DeleteRole(1)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestDeleteRole_SystemRoleCannotBeDeleted() {
	role := factories.MakeSystemRole()
	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)

	err := s.service.DeleteRole(1)

	s.Error(err)
	s.Contains(err.Error(), "tidak bisa dihapus")
}

func (s *RBACServiceTestSuite) TestDeleteRole_NotFound() {
	s.rbacRepo.On("GetRoleByID", int64(99)).Return(nil, nil)

	err := s.service.DeleteRole(99)

	s.Error(err)
}

// ═══════════════════════════════════════════════════════════════════════════════
// ROLE ↔ PERMISSION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func (s *RBACServiceTestSuite) TestAssignPermissionsToRole_Success() {
	role := factories.MakeRole()
	req := &dto.AssignPermissionsRequest{PermissionIDs: []int64{1, 2, 3}}

	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)
	s.rbacRepo.On("AssignPermissionsToRole", int64(1), []int64{1, 2, 3}).Return(nil)

	err := s.service.AssignPermissionsToRole(1, req)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestAssignPermissionsToRole_RoleNotFound() {
	req := &dto.AssignPermissionsRequest{PermissionIDs: []int64{1}}
	s.rbacRepo.On("GetRoleByID", int64(99)).Return(nil, nil)

	err := s.service.AssignPermissionsToRole(99, req)

	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

func (s *RBACServiceTestSuite) TestRevokePermissionsFromRole_Success() {
	role := factories.MakeRole()
	req := &dto.AssignPermissionsRequest{PermissionIDs: []int64{1}}

	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)
	s.rbacRepo.On("RevokePermissionsFromRole", int64(1), []int64{1}).Return(nil)

	err := s.service.RevokePermissionsFromRole(1, req)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestSyncRolePermissions_Success() {
	role := factories.MakeRole()
	req := &dto.AssignPermissionsRequest{PermissionIDs: []int64{1, 2}}

	s.rbacRepo.On("GetRoleByID", int64(1)).Return(role, nil)
	s.rbacRepo.On("SyncRolePermissions", int64(1), []int64{1, 2}).Return(nil)

	err := s.service.SyncRolePermissions(1, req)

	s.NoError(err)
}

// ═══════════════════════════════════════════════════════════════════════════════
// USER ↔ ROLE TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func (s *RBACServiceTestSuite) TestGetUserRoles_Success() {
	user := factories.MakeUser()
	roles := []models.Role{*factories.MakeRole()}

	s.userRepo.On("GetByID", int64(1)).Return(user, nil)
	s.rbacRepo.On("GetUserRoles", int64(1)).Return(roles, nil)

	result, err := s.service.GetUserRoles(1)

	s.NoError(err)
	s.Len(result, 1)
	s.Equal("admin", result[0].Name)
}

func (s *RBACServiceTestSuite) TestGetUserRoles_UserNotFound() {
	s.userRepo.On("GetByID", int64(99)).Return(nil, nil)

	result, err := s.service.GetUserRoles(99)

	s.Nil(result)
	s.requireAppError(err, http.StatusNotFound)
}

func (s *RBACServiceTestSuite) TestGetUserRoles_EmptyRoles_Returns404() {
	user := factories.MakeUser()

	s.userRepo.On("GetByID", int64(1)).Return(user, nil)
	s.rbacRepo.On("GetUserRoles", int64(1)).Return([]models.Role{}, nil)

	result, err := s.service.GetUserRoles(1)

	s.Nil(result)
	s.requireAppError(err, http.StatusNotFound)
	s.Contains(err.Error(), "roles pada user")
}

func (s *RBACServiceTestSuite) TestGetUserRoles_DBError() {
	user := factories.MakeUser()

	s.userRepo.On("GetByID", int64(1)).Return(user, nil)
	s.rbacRepo.On("GetUserRoles", int64(1)).Return(nil, dbError())

	result, err := s.service.GetUserRoles(1)

	s.Nil(result)
	s.requireAppError(err, http.StatusInternalServerError)
}

func (s *RBACServiceTestSuite) TestAssignRolesToUser_Success() {
	req := &dto.AssignRolesRequest{RoleIDs: []int64{1, 2}}

	s.rbacRepo.On("AssignRolesToUser", int64(5), []int64{1, 2}, s.actorID).Return(nil)

	err := s.service.AssignRolesToUser(5, req, s.actorID)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestRevokeRolesFromUser_Success() {
	req := &dto.AssignRolesRequest{RoleIDs: []int64{1}}

	s.rbacRepo.On("RevokeRolesFromUser", int64(5), []int64{1}).Return(nil)

	err := s.service.RevokeRolesFromUser(5, req)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestSyncUserRoles_Success() {
	req := &dto.AssignRolesRequest{RoleIDs: []int64{2, 3}}

	s.rbacRepo.On("SyncUserRoles", int64(5), []int64{2, 3}, s.actorID).Return(nil)

	err := s.service.SyncUserRoles(5, req, s.actorID)

	s.NoError(err)
}

// ═══════════════════════════════════════════════════════════════════════════════
// USER ↔ PERMISSION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func (s *RBACServiceTestSuite) TestAssignDirectPermission_Grant() {
	req := &dto.AssignDirectPermissionRequest{
		PermissionID: 1,
		IsGranted:    true,
	}

	s.rbacRepo.On("AssignDirectPermission", int64(5), int64(1), true, s.actorID).Return(nil)

	err := s.service.AssignDirectPermission(5, req, s.actorID)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestAssignDirectPermission_Deny() {
	req := &dto.AssignDirectPermissionRequest{
		PermissionID: 2,
		IsGranted:    false, // explicit deny
	}

	s.rbacRepo.On("AssignDirectPermission", int64(5), int64(2), false, s.actorID).Return(nil)

	err := s.service.AssignDirectPermission(5, req, s.actorID)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestRevokeDirectPermission_Success() {
	s.rbacRepo.On("RevokeDirectPermission", int64(5), int64(1)).Return(nil)

	err := s.service.RevokeDirectPermission(5, 1)

	s.NoError(err)
}

func (s *RBACServiceTestSuite) TestGetUserDirectPermissions_Success() {
	userPerms := []models.UserPermission{
		factories.MakeUserPermission(5, 1, true),
		factories.MakeUserPermission(5, 2, false),
	}
	perm1 := factories.MakePermission()
	perm2 := factories.MakePermission(func(p *models.Permission) {
		p.ID = 2
		p.Name = "users:create"
	})

	s.rbacRepo.On("GetUserDirectPermissions", int64(5)).Return(userPerms, nil)
	s.rbacRepo.On("GetPermissionByID", int64(1)).Return(perm1, nil)
	s.rbacRepo.On("GetPermissionByID", int64(2)).Return(perm2, nil)

	result, err := s.service.GetUserDirectPermissions(5)

	s.NoError(err)
	s.Len(result, 2)
	s.True(result[0].IsGranted)
	s.False(result[1].IsGranted)
}

func (s *RBACServiceTestSuite) TestGetUserDirectPermissions_SkipMissingPermission() {
	// Jika permission sudah dihapus tapi user_permissions masih ada,
	// entry tersebut harus di-skip
	userPerms := []models.UserPermission{
		factories.MakeUserPermission(5, 999, true),
	}

	s.rbacRepo.On("GetUserDirectPermissions", int64(5)).Return(userPerms, nil)
	s.rbacRepo.On("GetPermissionByID", int64(999)).Return(nil, nil)

	result, err := s.service.GetUserDirectPermissions(5)

	s.NoError(err)
	s.Empty(result) // di-skip
}

// ═══════════════════════════════════════════════════════════════════════════════
// CHECK TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func (s *RBACServiceTestSuite) TestGetUserAllPermissions_Success() {
	perms := []string{"users:read", "users:create", "roles:read"}
	s.rbacRepo.On("GetUserAllPermissions", int64(1)).Return(perms, nil)

	result, err := s.service.GetUserAllPermissions(1)

	s.NoError(err)
	s.Len(result, 3)
	s.Contains(result, "users:read")
}

func (s *RBACServiceTestSuite) TestHasPermission_True() {
	s.rbacRepo.On("HasPermission", int64(1), "users:read").Return(true, nil)

	result, err := s.service.HasPermission(1, "users:read")

	s.NoError(err)
	s.True(result)
}

func (s *RBACServiceTestSuite) TestHasPermission_False() {
	s.rbacRepo.On("HasPermission", int64(1), "users:delete").Return(false, nil)

	result, err := s.service.HasPermission(1, "users:delete")

	s.NoError(err)
	s.False(result)
}

func (s *RBACServiceTestSuite) TestHasPermission_DBError() {
	s.rbacRepo.On("HasPermission", int64(1), "users:read").Return(false, dbError())

	result, err := s.service.HasPermission(1, "users:read")

	s.False(result)
	s.Error(err)
}

// ─── mock_any helper ──────────────────────────────────────────────────────────

// mock_any mengembalikan testify matcher yang cocok dengan argumen apapun
// dipakai saat membandingkan pointer struct yang tidak bisa direct-compare
func mock_any() interface{} {
	return mock_anything{}
}

type mock_anything struct{}

func (mock_anything) Matches(v interface{}) bool { return true }
func (mock_anything) String() string             { return "anything" }
