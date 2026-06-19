package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"subian_go/config"
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/modules/users/models"

	appErrors "subian_go/internal/shared/errors"

	"golang.org/x/crypto/bcrypt"

	he "subian_go/internal/shared/httputil"
)

// ─── CRUD ──────────────────────────────────────────────────────────────────────

// CreateUser --------------------------------------------------------------------
func (s *service) CreateUser(req *dto.CreateUserRequest, actor he.AuthContext) (*dto.UserSimpleResponse, error) {
	can, err := s.canCreateUser(actor)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak memiliki hak akses untuk membuat user baru.", nil)
	}

	if existing, _ := s.repo.GetByUsername(req.Username); existing != nil {
		return nil, appErrors.BadRequest("username sudah digunakan")
	}
	if existing, _ := s.repo.GetByEmail(req.Email); existing != nil {
		return nil, appErrors.BadRequest("email sudah digunakan")
	}

	hashed, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, appErrors.Internal("gagal memproses password")
	}

	defaultSettingsList := models.DefaultSettings()
	settingsBytes, err := json.Marshal(defaultSettingsList)
	if err != nil {
		return nil, appErrors.Internal("gagal memproses setting bawaan")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	isStaff := false
	if req.IsStaff != nil {
		isStaff = *req.IsStaff
	}
	isSuperadmin := false
	if req.IsSuperadmin != nil {
		isSuperadmin = *req.IsSuperadmin
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		Name:         req.Name,
		Password:     hashed,
		IsActive:     isActive,
		IsStaff:      isStaff,
		IsSuperadmin: isSuperadmin,
		Settings:     models.JSONB(settingsBytes),
		CreatedBy:    &actor.UserID,
		UpdatedBy:    &actor.UserID,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, appErrors.Internal("gagal membuat user")
	}

	roles, _ := s.buildUserRBAC(user.ID)

	return dto.ToUserSimpleResponse(dto.UserSimpleResponseParams{
		User:  user,
		Roles: roles,
	}), nil
} // CreateUser ------------------------------------------------------------------

// GetUserByID -------------------------------------------------------------------
func (s *service) GetUserByID(id int64, actor he.AuthContext) (*dto.UserResponse, error) {
	can, err := s.canReadUser(actor, id)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	// Ambil RBAC data
	roles, permissions := s.buildUserRBAC(user.ID)

	// Ambil creator
	creator := s.buildCreator(user.CreatedBy)

	// Ambil login histories
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, can), nil
} // GetUserByID -----------------------------------------------------------------

// GetUserByUsername -------------------------------------------------------------
func (s *service) GetUserByUsername(username string, actor he.AuthContext) (*dto.UserResponse, error) {

	user, err := s.repo.GetByUsername(username)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	can, err := s.canReadUser(actor, user.ID)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, can), nil
} // GetUserByUsername -----------------------------------------------------------

// GetUserByEmail ----------------------------------------------------------------
func (s *service) GetUserByEmail(email string, actor he.AuthContext) (*dto.UserResponse, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	can, err := s.canReadUser(actor, user.ID)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, can), nil
} // GetUserByEmail --------------------------------------------------------------

// ListUsers ---------------------------------------------------------------------
func (s *service) ListUsers(page, pageSize int, filter *dto.UserFilter) ([]dto.UserSimpleResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := s.repo.List(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}

	// ─── AMANKAN DI SINI: CEK JIKA DATA KOSONG ───────────────────────────
	if len(users) == 0 {
		// Mengembalikan error spesifik bahwa data tidak ditemukan
		return nil, 0, errors.New("user tidak ditemukan")
	}
	// ──────────────────────────────────────────────────────────────────────

	// Amankan pemanggilan index [0], sekarang sudah pasti aman karena len > 0
	// 1. Kumpulkan semua User ID untuk batching query
	userIDs := make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// 2. Ambil data roles secara batch (Hanya 1x query tambahan, bukan N kali)
	userRolesMap := s.buildUsersRBAC(userIDs)

	return dto.ToUserListResponse(users, userRolesMap), total, nil
} // ListUsers -------------------------------------------------------------------

// UpdateUser --------------------------------------------------------------------
func (s *service) UpdateUser(id int64, req *dto.UpdateUserRequest, actor he.AuthContext) (*dto.UserResponse, error) {
	can, err := s.canUpdateUser(actor, id)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda Tidak bisa mengubah data ini.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		if existing, _ := s.repo.GetByEmail(*req.Email); existing != nil && existing.ID != id {
			return nil, appErrors.BadRequest("email sudah digunakan")
		}
		user.Email = *req.Email
	}

	if req.Username != nil {
		if existing, _ := s.repo.GetByUsername(*req.Username); existing != nil && existing.ID != id {
			return nil, appErrors.BadRequest("username sudah digunakan")
		}
		user.Username = *req.Username
	}

	user.UpdatedBy = &actor.UserID
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(user); err != nil {
		return nil, appErrors.Internal("gagal mengupdate user")
	}

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, true), nil
} // UpdateUser ------------------------------------------------------------------

// DeleteUser --------------------------------------------------------------------
func (s *service) DeleteUser(id int64, reason string, actor he.AuthContext) error {
	can, err := s.canDeleteUser(actor)
	if err != nil {
		return appErrors.Internal("gagal cek akses")
	}
	if !can {
		return appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak bisa menghapus user.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return appErrors.NotFound("user tidak ditemukan")
	}
	if user.ID == actor.UserID {
		return appErrors.BadRequest("tidak bisa menghapus akun sendiri")
	}

	// Teruskan ID, ID Penghapus (Actor), dan Alasan ke repository
	return s.repo.Delete(id, actor.UserID, reason)
} // DeleteUser ------------------------------------------------------------------

// ListDeletedUsers --------------------------------------------------------------
func (s *service) ListDeletedUsers(page, pageSize int, filter *dto.UserDeletedFilter, actor he.AuthContext) ([]dto.UserDeletedResponse, int64, error) {
	can, err := s.canDeleteUser(actor)
	if err != nil {
		return nil, 0, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, 0, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak bisa menghapus user.", nil)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Panggil repo khusus deleted list
	users, total, err := s.repo.DeletedList(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}

	// Antisipasi Error 500 jika data kosong
	if len(users) == 0 {
		return nil, 0, errors.New("data sampah user kosong")
	}

	// 1. Kumpulkan semua User ID untuk batching query
	userIDs := make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// 2. Ambil data roles secara batch (Hanya 1x query tambahan, bukan N kali)
	userRolesMap := s.buildUsersRBAC(userIDs)

	// 3. Ambil data creator dan deleter secara batch untuk semua user yang dihapus
	creatorsMap := make(map[int64]*models.UserCreator)
	deletersMap := make(map[int64]*models.UserCreator)
	for _, u := range users {
		creatorsMap[u.ID] = s.buildCreator(u.CreatedBy)
		deletersMap[u.ID] = s.buildCreator(u.DeletedBy)
	}

	// 4. Convert ke response DTO dengan data lengkap (roles, creator, deleter)

	// Mengonversi data models.User ke DTO response ringkas
	return dto.ToUserDeletedListResponse(users, userRolesMap, creatorsMap, deletersMap), total, nil
} // ListDeletedUsers ------------------------------------------------------------

// ─── Settings ──────────────────────────────────────────────────────────────────
func (s *service) GetSettings(id int64, actor he.AuthContext) ([]models.UserSetting, error) {
	can, err := s.canUpdateUser(actor, id)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda Tidak bisa mengubah data ini.", nil)
	}
	return s.repo.GetSettings(id)
}

func (s *service) UpdateSettings(id int64, req *dto.UpdateSettingsRequest, actor he.AuthContext) (*dto.UserResponse, error) {
	can, err := s.canUpdateUser(actor, id)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda Tidak bisa mengubah data ini.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	if err := s.repo.UpdateSettings(id, req.Settings); err != nil {
		return nil, appErrors.Internal("gagal mengupdate settings user")
	}

	user.UpdatedBy = &actor.UserID
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		return nil, appErrors.Internal("gagal mengupdate user setelah update settings")
	}

	// Refresh data user setelah update settings
	user, err = s.repo.GetByID(id)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan setelah update settings")
	}

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, true), nil
} // ─── Settings ────────────────────────────────────────────────────────────────

// ─── Password ──────────────────────────────────────────────────────────────────
func (s *service) ChangePassword(id int64, req *dto.ChangePasswordRequest, actor he.AuthContext) (*dto.UserResponse, error) {
	if !actor.IsSuperadmin && actor.UserID != id {
		return nil, appErrors.Wrap(http.StatusForbidden, "Akses ditolak. Hanya bisa mengubah password sendiri.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}
	if !s.verifyPassword(req.OldPassword, user.Password) {
		return nil, appErrors.BadRequest("password lama tidak sesuai")
	}

	hashed, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return nil, appErrors.Internal("gagal memproses password")
	}

	now := time.Now()
	user.Password = hashed
	user.PasswordChangedAt = &now
	user.UpdatedBy = &actor.UserID
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(user); err != nil {
		return nil, appErrors.Internal("gagal mengupdate password")
	}

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, true), nil
} // ─── Password ────────────────────────────────────────────────────────────────

// ─── Reset Password ────────────────────────────────────────────────────────────
func (s *service) ResetPassword(id int64, actor he.AuthContext) error {
	cfg := config.LoadConfig()

	can, err := s.canDeleteUser(actor) // Cek akses superadmin
	if err != nil {
		return appErrors.Internal("gagal cek akses")
	}
	if !can {
		return appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak bisa mereset password.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return appErrors.NotFound("user tidak ditemukan")
	}

	hashed, err := s.hashPassword(cfg.DefaultPassword)
	if err != nil {
		return appErrors.Internal("gagal memproses password")
	}

	user.Password = hashed
	user.PasswordChangedAt = func() *time.Time { t := time.Now(); return &t }()

	return s.repo.Update(user)
}

// ─── Upload Foto ───────────────────────────────────────────────────────────────
func (s *service) UploadPhoto(id int64, filename string, reader io.Reader, actor he.AuthContext) (*dto.UserResponse, error) {
	can, err := s.canUpdateUser(actor, id)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda Tidak bisa mengubah data ini.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	//lakukan update foto di sini, misal:
	//-----------------------------------------------------------------------------
	// ─── UPDATE FOTO DI SINI ───────────────────────────────────────────────────

	// 1. Simpan URL foto lama untuk dihapus nanti jika upload sukses
	var oldPhoto, oldThumbnail string
	if user.Photo != nil {
		oldPhoto = *user.Photo
	}
	if user.PhotoThumbnail != nil {
		oldThumbnail = *user.PhotoThumbnail
	}

	// 2. Konversi io.Reader ke multipart.File (aman untuk di-assert)
	multipartFile, ok := reader.(multipart.File)
	if !ok {
		multipartFile = nopCloser{reader}
	}

	// 3. Buat FileHeader buatan (Gunakan header.Size asli jika dikirim dari handler)
	fileHeader := &multipart.FileHeader{
		Filename: filename,
		Size:     0,
	}

	// 4. Jalankan upload foto original + thumbnail otomatis
	origURL, thumbURL, err := s.storageService.UploadImageWithThumbnail(multipartFile, fileHeader, "users")
	if err != nil {
		return nil, appErrors.BadRequest(fmt.Sprintf("gagal mengunggah foto: %v", err))
	}

	// 5. Pasang URL baru ke struct user
	user.Photo = &origURL
	user.PhotoThumbnail = &thumbURL

	// NYALAKAN NIL SAFETY CHECK UNTUK ACTOR
	// Menghindari panic jika context actor dikirim kosong/nil dari handler
	if actor != (he.AuthContext{}) {
		userID := actor.UserID // Menggunakan getter jika interface, atau cek nil pointer jika struct
		user.UpdatedBy = &userID
	}
	user.UpdatedAt = time.Now()

	// ─── END UPDATE FOTO ───────────────────────────────────────────────────────

	// 6. Simpan Perubahan ke Database
	if err := s.repo.Update(user); err != nil {
		// ROLLBACK FISIK: Hapus file baru di storage jika DB gagal menyimpan data
		_ = s.storageService.DeleteImageMultiple(origURL, thumbURL)
		return nil, appErrors.Internal("gagal mengupdate foto user")
	}

	// 7. HAPUS FOTO LAMA DARI STORAGE (Asynchronous / Background Process)
	go func(urls ...string) {
		if len(urls) > 0 {
			_ = s.storageService.DeleteImageMultiple(urls...)
		}
	}(oldPhoto, oldThumbnail)
	//-----------------------------------------------------------------------------

	// CATATAN: BLOK KODE DI BAWAH INI YANG SEBELUMNYA DUPLIKAT SUDAH DIHAPUS
	// (s.repo.Update(user) yang kedua dibuang)

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, true), nil
}

// ─── Delete Photo ──────────────────────────────────────────────────────────────
func (s *service) DeletePhoto(id int64, actor he.AuthContext) (*dto.UserResponse, error) {
	can, err := s.canUpdateUser(actor, id)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda Tidak bisa mengubah data ini.", nil)
	}

	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return nil, appErrors.NotFound("user tidak ditemukan")
	}

	// Simpan URL foto lama untuk dihapus nanti jika update sukses
	var oldPhoto, oldThumbnail string
	if user.Photo != nil {
		oldPhoto = *user.Photo
	}
	if user.PhotoThumbnail != nil {
		oldThumbnail = *user.PhotoThumbnail
	}

	user.Photo = nil
	user.PhotoThumbnail = nil

	if actor != (he.AuthContext{}) {
		userID := actor.UserID
		user.UpdatedBy = &userID
	}
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(user); err != nil {
		return nil, appErrors.Internal("gagal menghapus foto user")
	}

	// Hapus file dari storage secara asynchronous
	go func(urls ...string) {
		if len(urls) > 0 {
			_ = s.storageService.DeleteImageMultiple(urls...)
		}
	}(oldPhoto, oldThumbnail)

	roles, permissions := s.buildUserRBAC(user.ID)
	creator := s.buildCreator(user.CreatedBy)
	histories, _ := s.authRepo.GetUserLoginHistories(user.ID, 10)

	return dto.ToUserResponse(dto.UserResponseParams{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
		Histories:   histories,
		Creator:     creator,
	}, true), nil
}

// ─── Private Helpers ───────────────────────────────────────────────────────────
func (s *service) UpdateLastLogin(id int64) error {
	user, err := s.repo.GetByID(id)
	if err != nil || user == nil {
		return errors.New("user tidak ditemukan")
	}
	now := time.Now()
	user.LastLoginAt = &now
	return s.repo.Update(user)
}

func (s *service) verifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Helper struct ditaruh di luar fungsi (di bawah/atas file) jika reader bukan multipart.File
type nopCloser struct {
	io.Reader
}

// Read implements [multipart.File].
// Subtle: this method shadows the method (Reader).Read of nopCloser.Reader.
func (nopCloser) Read(p []byte) (n int, err error) {
	panic("unimplemented")
}

// Seek implements [multipart.File].
func (n nopCloser) Seek(offset int64, whence int) (int64, error) {
	panic("unimplemented")
}

func (nopCloser) Close() error                                  { return nil }
func (nopCloser) ReadAt(p []byte, off int64) (n int, err error) { return 0, io.EOF }
