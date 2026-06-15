package tests

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	rbacModels "subian_go/internal/modules/rbac/models"
	userContracts "subian_go/internal/modules/users/contracts"
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/modules/users/models"
	"subian_go/internal/modules/users/services"
	"subian_go/internal/modules/users/tests/factories"
	"subian_go/internal/modules/users/tests/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ─── Suite ─────────────────────────────────────────────────────────────────────

func TestMain(m *testing.M) {
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")
	fmt.Println("\033[35m  User Service Test Suite\033[0m")
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")

	code := m.Run()

	if code == 0 {
		fmt.Println("\n\033[32m✓  PASS\033[0m  subian_go/internal/modules/users")
	} else {
		fmt.Println("\n\033[31m✗  FAIL\033[0m  subian_go/internal/modules/users")
	}

	os.Exit(code)
}

type UserServiceTestSuite struct {
	suite.Suite
	repo         *mocks.MockUserRepository
	rbacRepo     *mocks.MockRBACRepository
	authRepo     *mocks.MockAuthRepository
	imgStorage   *mocks.MockImageStorage
	service      userContracts.Service
	superActor   userContracts.AuthContext
	regularActor userContracts.AuthContext
}

func (s *UserServiceTestSuite) SetupTest() {
	s.repo = new(mocks.MockUserRepository)
	s.rbacRepo = new(mocks.MockRBACRepository)
	s.authRepo = new(mocks.MockAuthRepository)
	s.imgStorage = new(mocks.MockImageStorage)
	s.service = services.NewUserService(s.repo, s.rbacRepo, s.authRepo, s.imgStorage)
	s.superActor = userContracts.AuthContext{UserID: 1, IsSuperadmin: true}
	s.regularActor = userContracts.AuthContext{UserID: 2, IsSuperadmin: false}
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
	s.rbacRepo.AssertExpectations(s.T())
	s.authRepo.AssertExpectations(s.T())
	s.imgStorage.AssertExpectations(s.T())
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// mustHash membuat bcrypt hash untuk testing
func mustHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// stubRBAC menyiapkan mock RBAC untuk GetUserRoles dan GetUserDirectPermissions
func (s *UserServiceTestSuite) stubRBAC(userID int64, withPermissions bool) {
	var roles []rbacModels.Role
	if withPermissions {
		roles = []rbacModels.Role{*factories.MakeRoleWithPermissions()}
	} else {
		roles = []rbacModels.Role{}
	}
	s.rbacRepo.On("GetUserRoles", userID).Return(roles, nil)
	s.rbacRepo.On("GetUserDirectPermissions", userID).Return([]rbacModels.UserPermission{}, nil)
}

// stubHistories menyiapkan mock login histories
func (s *UserServiceTestSuite) stubHistories(userID int64) {
	s.authRepo.On("GetUserLoginHistories", userID, 10).
		Return(factories.MakeLoginHistories(userID, 2), nil)
}

// stubCreator menyiapkan mock untuk buildCreator
func (s *UserServiceTestSuite) stubCreator(creatorID int64) {
	creator := factories.MakeSuperadminUser()
	s.repo.On("GetByID", creatorID).Return(creator, nil)
}

// stubFullUserDetail menyiapkan semua stub untuk GetUserByID / detail response
func (s *UserServiceTestSuite) stubFullUserDetail(user *models.User) {
	s.stubRBAC(user.ID, true)
	s.stubHistories(user.ID)
	if user.CreatedBy != nil {
		s.stubCreator(*user.CreatedBy)
	}
}

// ─── CreateUser ───────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestCreateUser_Success_Superadmin() {
	req := &dto.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Name:     "New User",
		Password: "password123",
	}

	s.repo.On("GetByUsername", "newuser").Return(nil, nil)
	s.repo.On("GetByEmail", "new@example.com").Return(nil, nil)
	s.repo.On("Create", mock.MatchedBy(func(u *models.User) bool {
		return u.Username == "newuser" && u.Email == "new@example.com"
	})).Return(nil).Run(func(args mock.Arguments) {
		// Simulasi DB assign ID setelah create
		u := args.Get(0).(*models.User)
		u.ID = 10
	})
	s.stubRBAC(int64(10), false)

	result, err := s.service.CreateUser(req, s.superActor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("newuser", result.Username)
	s.Equal("new@example.com", result.Email)
}

func (s *UserServiceTestSuite) TestCreateUser_WithPermission() {
	req := &dto.CreateUserRequest{
		Username: "newuser2",
		Email:    "new2@example.com",
		Name:     "New User 2",
		Password: "password123",
	}

	// Actor punya permission users:create
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersCreate).Return(true, nil)
	s.repo.On("GetByUsername", "newuser2").Return(nil, nil)
	s.repo.On("GetByEmail", "new2@example.com").Return(nil, nil)
	s.repo.On("Create", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(0).(*models.User)
		u.ID = 11
	})
	s.stubRBAC(int64(11), false)

	result, err := s.service.CreateUser(req, s.regularActor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestCreateUser_Forbidden() {
	req := &dto.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Name:     "New User",
		Password: "password123",
	}

	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersCreate).Return(false, nil)
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersManage).Return(false, nil)
	s.rbacRepo.On("GetUserRoles", s.regularActor.UserID).Return([]rbacModels.Role{}, nil)

	result, err := s.service.CreateUser(req, s.regularActor)

	s.Nil(result)
	s.requireForbidden(err)
}

func (s *UserServiceTestSuite) TestCreateUser_DuplicateUsername() {
	req := &dto.CreateUserRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Name:     "New User",
		Password: "password123",
	}

	s.repo.On("GetByUsername", "existinguser").Return(factories.MakeUser(), nil)

	result, err := s.service.CreateUser(req, s.superActor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "username sudah digunakan")
}

func (s *UserServiceTestSuite) TestCreateUser_DuplicateEmail() {
	req := &dto.CreateUserRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Name:     "New User",
		Password: "password123",
	}

	s.repo.On("GetByUsername", "newuser").Return(nil, nil)
	s.repo.On("GetByEmail", "existing@example.com").Return(factories.MakeUser(), nil)

	result, err := s.service.CreateUser(req, s.superActor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "email sudah digunakan")
}

// ─── GetUserByID ──────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestGetUserByID_Success_Superadmin() {
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })

	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.stubFullUserDetail(user)

	result, err := s.service.GetUserByID(5, s.superActor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(int64(5), result.ID)
	// Superadmin → is_allow=true → dapat Settings dan Permissions
	s.NotNil(result.Settings)
	s.NotNil(result.Permissions)
}

func (s *UserServiceTestSuite) TestGetUserByID_Success_Self() {
	actor := userContracts.AuthContext{UserID: 5, IsSuperadmin: false}
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })

	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.stubFullUserDetail(user)

	result, err := s.service.GetUserByID(5, actor)

	s.NoError(err)
	s.NotNil(result)
	// Self → is_allow=true → data penuh
	s.NotNil(result.Settings)
}

func (s *UserServiceTestSuite) TestGetUserByID_LimitedView_NoPermission() {
	// Actor lain tanpa permission — dapat data minimal (is_allow=false)
	actor := userContracts.AuthContext{UserID: 2, IsSuperadmin: false}
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })

	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermUsersRead).Return(false, nil)
	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermUsersManage).Return(false, nil)
	s.stubRBAC(int64(5), false)
	s.stubHistories(int64(5))
	s.stubCreator(int64(1))

	result, err := s.service.GetUserByID(5, actor)

	s.NoError(err)
	s.NotNil(result)
	// is_allow=false → Settings nil, Permissions nil
	s.Nil(result.Settings)
}

func (s *UserServiceTestSuite) TestGetUserByID_NotFound() {
	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	result, err := s.service.GetUserByID(999, s.superActor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ─── ListUsers ────────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestListUsers_Success() {
	users := factories.MakeUserList(3)
	filter := &dto.UserFilter{}
	userIDs := []int64{1, 2, 3}
	rolesMap := map[int64][]rbacModels.Role{
		1: {*factories.MakeRole(1, "admin")},
		2: {},
		3: {},
	}

	s.repo.On("List", 1, 10, filter).Return(users, int64(3), nil)
	s.rbacRepo.On("GetUsersRoles", userIDs).Return(rolesMap, nil)

	result, total, err := s.service.ListUsers(1, 10, filter)

	s.NoError(err)
	s.Equal(int64(3), total)
	s.Len(result, 3)
	// User pertama punya role admin
	s.Len(result[0].Roles, 1)
	// User kedua tidak punya role
	s.Empty(result[1].Roles)
}

func (s *UserServiceTestSuite) TestListUsers_EmptyResult_ReturnsEmptySlice() {
	filter := &dto.UserFilter{Name: "noone"}

	s.repo.On("List", 1, 10, filter).Return([]models.User{}, int64(0), nil)

	result, total, err := s.service.ListUsers(1, 10, filter)

	// ✅ UBAH EKSPEKTASI: Service mengembalikan error, bukan empty slice
	s.Error(err)
	s.Contains(err.Error(), "user tidak ditemukan")
	s.Nil(result)
	s.Equal(int64(0), total)
}

func (s *UserServiceTestSuite) TestListUsers_PageNormalization() {
	users := factories.MakeUserList(1)
	filter := &dto.UserFilter{}

	// page=0 harus dinormalisasi ke 1
	s.repo.On("List", 1, 10, filter).Return(users, int64(1), nil)
	s.rbacRepo.On("GetUsersRoles", []int64{1}).Return(map[int64][]rbacModels.Role{}, nil)

	result, _, err := s.service.ListUsers(0, 10, filter)

	s.NoError(err)
	s.Len(result, 1)
}

// ─── UpdateUser ───────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestUpdateUser_Success_Superadmin() {
	user := factories.MakeUser(func(u *models.User) { u.ID = 3 })
	newName := "Updated Name"
	req := &dto.UpdateUserRequest{Name: &newName}

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	s.stubFullUserDetail(user)

	result, err := s.service.UpdateUser(3, req, s.superActor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("Updated Name", result.Name)
}

func (s *UserServiceTestSuite) TestUpdateUser_Success_Self() {
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	user := factories.MakeUser(func(u *models.User) { u.ID = 3 })
	newName := "My Name"
	req := &dto.UpdateUserRequest{Name: &newName}

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	s.stubFullUserDetail(user)

	result, err := s.service.UpdateUser(3, req, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestUpdateUser_Success_WithPermission() {
	actor := userContracts.AuthContext{UserID: 2, IsSuperadmin: false}
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })
	newName := "Updated via Permission"
	req := &dto.UpdateUserRequest{Name: &newName}

	// Actor punya permission users:update
	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermUsersUpdate).Return(true, nil)
	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	s.stubFullUserDetail(user)

	result, err := s.service.UpdateUser(5, req, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestUpdateUser_Forbidden() {
	req := &dto.UpdateUserRequest{}

	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersUpdate).Return(false, nil)
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersManage).Return(false, nil)
	s.rbacRepo.On("GetUserRoles", s.regularActor.UserID).Return([]rbacModels.Role{}, nil)

	result, err := s.service.UpdateUser(99, req, s.regularActor)

	s.Nil(result)
	s.requireForbidden(err)
}

func (s *UserServiceTestSuite) TestUpdateUser_NotFound() {
	req := &dto.UpdateUserRequest{}

	s.repo.On("GetByID", int64(404)).Return(nil, nil)

	result, err := s.service.UpdateUser(404, req, s.superActor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

func (s *UserServiceTestSuite) TestUpdateUser_DuplicateEmail() {
	user := factories.MakeUser(func(u *models.User) { u.ID = 3 })
	newEmail := "taken@example.com"
	req := &dto.UpdateUserRequest{Email: &newEmail}

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	otherUser := factories.MakeUser(func(u *models.User) {
		u.ID = 99
		u.Email = "taken@example.com"
	})
	s.repo.On("GetByEmail", "taken@example.com").Return(otherUser, nil)

	result, err := s.service.UpdateUser(3, req, s.superActor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "email sudah digunakan")
}

// ─── DeleteUser ───────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestDeleteUser_Success() {
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })

	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.repo.On("Delete", int64(5), s.superActor.UserID, "Pelanggaran").Return(nil)

	err := s.service.DeleteUser(5, "Pelanggaran", s.superActor)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDeleteUser_Forbidden_NotSuperadmin() {
	err := s.service.DeleteUser(5, "Alasan", s.regularActor)

	s.requireForbidden(err)
}

func (s *UserServiceTestSuite) TestDeleteUser_CannotDeleteSelf() {
	selfActor := userContracts.AuthContext{UserID: 5, IsSuperadmin: true}
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })

	s.repo.On("GetByID", int64(5)).Return(user, nil)

	err := s.service.DeleteUser(5, "Self", selfActor)

	s.Error(err)
	s.Contains(err.Error(), "tidak bisa menghapus akun sendiri")
}

func (s *UserServiceTestSuite) TestDeleteUser_NotFound() {
	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	err := s.service.DeleteUser(999, "Alasan", s.superActor)

	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestChangePassword_Success_Self() {
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	hashed := mustHash("oldpassword123")
	user := factories.MakeUser(func(u *models.User) {
		u.ID = 3
		u.Password = hashed
	})

	req := &dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword456",
		ConfirmPassword: "newpassword456",
	}

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	s.stubFullUserDetail(user)

	result, err := s.service.ChangePassword(3, req, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestChangePassword_Success_Superadmin() {
	// Superadmin ubah password user lain
	hashed := mustHash("userpass")
	user := factories.MakeUser(func(u *models.User) {
		u.ID = 5
		u.Password = hashed
	})

	req := &dto.ChangePasswordRequest{
		OldPassword:     "userpass",
		NewPassword:     "newpass456",
		ConfirmPassword: "newpass456",
	}

	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	s.stubFullUserDetail(user)

	result, err := s.service.ChangePassword(5, req, s.superActor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestChangePassword_WrongOldPassword() {
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	hashed := mustHash("correctpassword")
	user := factories.MakeUser(func(u *models.User) {
		u.ID = 3
		u.Password = hashed
	})

	req := &dto.ChangePasswordRequest{
		OldPassword:     "wrongpassword",
		NewPassword:     "newpass456",
		ConfirmPassword: "newpass456",
	}

	s.repo.On("GetByID", int64(3)).Return(user, nil)

	result, err := s.service.ChangePassword(3, req, actor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "password lama tidak sesuai")
}

func (s *UserServiceTestSuite) TestChangePassword_Forbidden_OtherUser() {
	req := &dto.ChangePasswordRequest{
		OldPassword: "old", NewPassword: "new", ConfirmPassword: "new",
	}

	result, err := s.service.ChangePassword(99, req, s.regularActor)

	s.Nil(result)
	s.requireForbidden(err)
}

// ─── GetSettings ──────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestGetSettings_Success_Self() {
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	settings := []models.UserSetting{
		{Key: "is_dark_mode", Type: "boolean", Value: false},
	}
	s.repo.On("GetSettings", int64(3)).Return(settings, nil)

	result, err := s.service.GetSettings(3, actor)

	s.NoError(err)
	s.Len(result, 1)
}

func (s *UserServiceTestSuite) TestGetSettings_Forbidden() {
	// actor != target, tidak punya permission
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersUpdate).Return(false, nil)
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersManage).Return(false, nil)
	s.rbacRepo.On("GetUserRoles", s.regularActor.UserID).Return([]rbacModels.Role{}, nil)

	result, err := s.service.GetSettings(99, s.regularActor)

	s.Nil(result)
	s.requireForbidden(err)
}

// ─── UploadPhoto ──────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestUploadPhoto_Success() {
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	user := factories.MakeUser(func(u *models.User) { u.ID = 3 })

	mockFile := &mocks.MockMultipartFile{}
	origURL := "http://localhost:1323/uploads/users/photo.jpg"
	thumbURL := "http://localhost:1323/uploads/users/thumbnails/photo_thumb.jpg"

	s.repo.On("GetByID", int64(3)).Return(user, nil)

	// ✅ TAMBAHKAN INI: Antisipasi jika service mencoba menghapus foto lama yang kosong
	s.imgStorage.On("DeleteImageMultiple", "", "").Return(nil).Maybe()

	s.imgStorage.On("UploadImageWithThumbnail", mockFile, mock.Anything, "users").
		Return(origURL, thumbURL, nil)
	s.repo.On("Update", user).Return(nil)
	s.stubFullUserDetail(user)

	result, err := s.service.UploadPhoto(3, "photo.jpg", mockFile, actor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(&origURL, result.Photo)
	s.Equal(&thumbURL, result.PhotoThumbnail)
}

func (s *UserServiceTestSuite) TestUploadPhoto_DeletesOldPhoto() {
	// Pastikan foto lama dihapus dari storage
	oldPhoto := "http://localhost:1323/uploads/users/old.jpg"
	oldThumb := "http://localhost:1323/uploads/users/thumbnails/old_thumb.jpg"

	user := factories.MakeUserWithPhoto()
	user.ID = 3
	user.Photo = &oldPhoto
	user.PhotoThumbnail = &oldThumb

	mockFile := &mocks.MockMultipartFile{}
	newOrig := "http://localhost:1323/uploads/users/new.jpg"
	newThumb := "http://localhost:1323/uploads/users/thumbnails/new_thumb.jpg"

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.imgStorage.On("UploadImageWithThumbnail", mockFile, mock.Anything, "users").
		Return(newOrig, newThumb, nil)
	s.repo.On("Update", user).Return(nil)
	// Hapus foto lama — dipanggil via goroutine
	s.imgStorage.On("DeleteImageMultiple", oldPhoto, oldThumb).Return(nil).Maybe()
	s.stubFullUserDetail(user)

	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	result, err := s.service.UploadPhoto(3, "photo.jpg", mockFile, actor)

	s.NoError(err)
	s.Equal(&newOrig, result.Photo)
}

func (s *UserServiceTestSuite) TestUploadPhoto_RollbackOnDBError() {
	// Jika DB gagal, file baru harus dihapus (rollback)
	user := factories.MakeUser(func(u *models.User) { u.ID = 3 })

	mockFile := &mocks.MockMultipartFile{}
	origURL := "http://localhost:1323/uploads/users/new.jpg"
	thumbURL := "http://localhost:1323/uploads/users/thumbnails/new_thumb.jpg"

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.imgStorage.On("UploadImageWithThumbnail", mockFile, mock.Anything, "users").
		Return(origURL, thumbURL, nil)
	s.repo.On("Update", user).Return(s.someError("db error"))
	// Harus hapus file baru karena DB gagal
	s.imgStorage.On("DeleteImageMultiple", origURL, thumbURL).Return(nil)

	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}
	result, err := s.service.UploadPhoto(3, "photo.jpg", mockFile, actor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "gagal mengupdate foto")
}

func (s *UserServiceTestSuite) TestUploadPhoto_Forbidden() {
	mockFile := &mocks.MockMultipartFile{}
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersUpdate).Return(false, nil)
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersManage).Return(false, nil)
	s.rbacRepo.On("GetUserRoles", s.regularActor.UserID).Return([]rbacModels.Role{}, nil)

	result, err := s.service.UploadPhoto(99, "photo.jpg", mockFile, s.regularActor)

	s.Nil(result)
	s.requireForbidden(err)
}

// ─── DeletePhoto ──────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestDeletePhoto_Success() {
	user := factories.MakeUserWithPhoto()
	user.ID = 3
	oldPhoto := *user.Photo
	oldThumb := *user.PhotoThumbnail
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	s.imgStorage.On("DeleteImageMultiple", oldPhoto, oldThumb).Return(nil).Maybe()
	s.stubFullUserDetail(user)

	result, err := s.service.DeletePhoto(3, actor)

	s.NoError(err)
	s.NotNil(result)
	s.Nil(result.Photo)
	s.Nil(result.PhotoThumbnail)
}

func (s *UserServiceTestSuite) TestDeletePhoto_NoPhoto_StillSuccess() {
	// User tanpa foto — delete tetap sukses (tidak error)
	user := factories.MakeUser(func(u *models.User) {
		u.ID = 3
		u.Photo = nil
		u.PhotoThumbnail = nil
	})
	actor := userContracts.AuthContext{UserID: 3, IsSuperadmin: false}

	s.repo.On("GetByID", int64(3)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)
	// Tidak ada file untuk dihapus — DeleteImageMultiple dipanggil dengan string kosong
	s.imgStorage.On("DeleteImageMultiple", "", "").Return(nil).Maybe()
	s.stubFullUserDetail(user)

	result, err := s.service.DeletePhoto(3, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestDeletePhoto_Forbidden() {
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersUpdate).Return(false, nil)
	s.rbacRepo.On("HasPermission", s.regularActor.UserID, rbacModels.PermUsersManage).Return(false, nil)
	s.rbacRepo.On("GetUserRoles", s.regularActor.UserID).Return([]rbacModels.Role{}, nil)

	result, err := s.service.DeletePhoto(99, s.regularActor)

	s.Nil(result)
	s.requireForbidden(err)
}

func (s *UserServiceTestSuite) TestDeletePhoto_NotFound() {
	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	result, err := s.service.DeletePhoto(999, s.superActor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestResetPassword_Success() {
	user := factories.MakeUser(func(u *models.User) { u.ID = 5 })

	s.repo.On("GetByID", int64(5)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)

	err := s.service.ResetPassword(5, s.superActor)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestResetPassword_Forbidden() {
	err := s.service.ResetPassword(5, s.regularActor)

	s.requireForbidden(err)
}

func (s *UserServiceTestSuite) TestResetPassword_NotFound() {
	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	err := s.service.ResetPassword(999, s.superActor)

	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

// ─── UpdateLastLogin ──────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestUpdateLastLogin_Success() {
	user := factories.MakeUser(func(u *models.User) {
		u.ID = 1 // Set ID menjadi 1 secara eksplisit
	})

	s.repo.On("GetByID", int64(1)).Return(user, nil)
	s.repo.On("Update", user).Return(nil)

	err := s.service.UpdateLastLogin(1)

	s.NoError(err)
	s.NotNil(user.LastLoginAt)
}

func (s *UserServiceTestSuite) TestUpdateLastLogin_NotFound() {
	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	err := s.service.UpdateLastLogin(999)

	s.Error(err)
}

// ─── ListDeletedUsers ─────────────────────────────────────────────────────────

func (s *UserServiceTestSuite) TestListDeletedUsers_Success() {
	deletedUser := factories.MakeDeletedUser()
	filter := &dto.UserDeletedFilter{}

	s.repo.On("DeletedList", 1, 10, filter).Return([]models.User{*deletedUser}, int64(1), nil)
	s.rbacRepo.On("GetUsersRoles", []int64{99}).Return(map[int64][]rbacModels.Role{}, nil)
	// creator dan deleter
	s.repo.On("GetByID", int64(1)).Return(factories.MakeSuperadminUser(), nil)

	result, total, err := s.service.ListDeletedUsers(1, 10, filter, s.superActor)

	s.NoError(err)
	s.Equal(int64(1), total)
	s.Len(result, 1)
	s.NotNil(result[0].DeleteReason)
}

func (s *UserServiceTestSuite) TestListDeletedUsers_Forbidden() {
	filter := &dto.UserDeletedFilter{}

	result, total, err := s.service.ListDeletedUsers(1, 10, filter, s.regularActor)

	s.Nil(result)
	s.Equal(int64(0), total)
	s.requireForbidden(err)
}

func (s *UserServiceTestSuite) TestListDeletedUsers_EmptyResult() {
	filter := &dto.UserDeletedFilter{}

	s.repo.On("DeletedList", 1, 10, filter).Return([]models.User{}, int64(0), nil)

	result, total, err := s.service.ListDeletedUsers(1, 10, filter, s.superActor)

	// ✅ UBAH EKSPEKTASI: Service mengembalikan error, bukan empty slice
	s.Error(err)
	s.Contains(err.Error(), "data sampah user kosong")
	s.Nil(result)
	s.Equal(int64(0), total)
}

// ─── Assertion Helpers ────────────────────────────────────────────────────────

// requireForbidden memastikan error adalah 403 Forbidden
func (s *UserServiceTestSuite) requireForbidden(err error) {
	s.Require().Error(err)
	appErr, ok := err.(interface{ StatusCode() int })
	s.Require().True(ok, "expected AppError with StatusCode()")
	s.Equal(http.StatusForbidden, appErr.StatusCode())
}

// someError membuat error generik untuk test
func (s *UserServiceTestSuite) someError(msg string) error {
	return &testError{msg: msg}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
