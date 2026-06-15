package factories

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	authModels "subian_go/internal/modules/auth/models"
	rbacModels "subian_go/internal/modules/rbac/models"
	"subian_go/internal/modules/users/models"

	"golang.org/x/crypto/bcrypt"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ─── User Factory Object ───────────────────────────────────────────────────────

// UserFactory membuat data user untuk keperluan testing/seeding dengan builder pattern
type UserFactory struct {
	overrides map[string]interface{}
}

// NewUserFactory membuat instance UserFactory baru
func NewUserFactory() *UserFactory {
	return &UserFactory{
		overrides: make(map[string]interface{}),
	}
}

// With menambahkan override field sebelum build
func (f *UserFactory) With(field string, value interface{}) *UserFactory {
	f.overrides[field] = value
	return f
}

// Make membuat satu User model tanpa menyimpan ke DB
func (f *UserFactory) Make() *models.User {
	idx := rng.Intn(999999)
	now := time.Now()
	createdBy := int64(1)

	// Default values
	id := int64(idx)
	username := fmt.Sprintf("user_%d", idx)
	email := fmt.Sprintf("user_%d@example.com", idx)
	name := fmt.Sprintf("User %d", idx)
	password := "password123"

	// Apply string/numeric overrides
	if v, ok := f.overrides["id"]; ok {
		id = v.(int64)
	}
	if v, ok := f.overrides["username"]; ok {
		username = v.(string)
	}
	if v, ok := f.overrides["email"]; ok {
		email = v.(string)
	}
	if v, ok := f.overrides["name"]; ok {
		name = v.(string)
	}
	if v, ok := f.overrides["password"]; ok {
		password = v.(string)
	}

	// Menggunakan bcrypt ringkas jika tidak di-override dengan hash spesifik
	var hashedPass string
	if v, ok := f.overrides["hashed_password"]; ok {
		hashedPass = v.(string)
	} else {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
		hashedPass = string(hashed)
	}

	user := &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		Name:         name,
		Password:     hashedPass,
		IsActive:     true,
		IsVerified:   true,
		IsSuperadmin: false,
		IsStaff:      false,
		Settings:     DefaultSettings(),
		CreatedBy:    &createdBy,
		UpdatedBy:    &createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Apply boolean overrides
	if v, ok := f.overrides["is_active"]; ok {
		user.IsActive = v.(bool)
	}
	if v, ok := f.overrides["is_superadmin"]; ok {
		user.IsSuperadmin = v.(bool)
	}
	if v, ok := f.overrides["is_staff"]; ok {
		user.IsStaff = v.(bool)
	}
	if v, ok := f.overrides["is_verified"]; ok {
		user.IsVerified = v.(bool)
	}

	// Apply pointer / special fields overrides if exists
	if v, ok := f.overrides["photo"]; ok {
		strVal := v.(string)
		user.Photo = &strVal
	}
	if v, ok := f.overrides["photo_thumbnail"]; ok {
		strVal := v.(string)
		user.PhotoThumbnail = &strVal
	}
	if v, ok := f.overrides["deleted_by"]; ok {
		val := v.(int64)
		user.DeletedBy = &val
	}
	if v, ok := f.overrides["delete_reason"]; ok {
		val := v.(string)
		user.DeleteReason = &val
	}
	if v, ok := f.overrides["deleted_at_time"]; ok {
		user.DeletedAt.Time = v.(time.Time)
		user.DeletedAt.Valid = true
	}

	return user
}

// MakeMany membuat banyak User model tanpa menyimpan ke DB
func (f *UserFactory) MakeMany(count int) []*models.User {
	users := make([]*models.User, count)
	for i := 0; i < count; i++ {
		// Melakukan clone factory agar state overrides tidak bercampur antar urutan item jika diinginkan,
		// atau membuat factory polosan per item.
		users[i] = NewUserFactory().Make()
	}
	return users
}

// ─── User Preset Factories ─────────────────────────────────────────────────────

// DefaultSettings menghasilkan format JSONB untuk konfigurasi awal user
func DefaultSettings() models.JSONB {
	settings := []models.UserSetting{
		{Key: "is_dark_mode", Type: "boolean", Value: false, Description: "Aktifkan tema gelap"},
		{Key: "language", Type: "string", Value: "id", Description: "Bahasa antarmuka"},
		{Key: "notification_email", Type: "boolean", Value: true, Description: "Notifikasi email"},
	}
	b, _ := json.Marshal(settings)
	return models.JSONB(b)
}

// MakeUser membuat user standar (kompatibilitas fungsi dari kode lama)
func MakeUser(overrides ...func(*models.User)) *models.User {
	user := NewUserFactory().With("hashed_password", "$2a$12$hashedpassword").Make()
	for _, fn := range overrides {
		fn(user)
	}
	return user
}

// MakeSuperadminUser membuat user superadmin dengan data fixed/preset
func MakeSuperadminUser() *models.User {
	return NewUserFactory().
		With("id", int64(1)).
		With("username", "superadmin").
		With("email", "superadmin@example.com").
		With("name", "Super Admin").
		With("is_superadmin", true).
		With("is_staff", true).
		Make()
}

// MakeStaffUser membuat user dengan flag staff aktif
func MakeStaffUser() *models.User {
	return NewUserFactory().
		With("id", int64(2)).
		With("username", "staffuser").
		With("email", "staff@example.com").
		With("name", "Staff User").
		With("is_staff", true).
		Make()
}

// MakeStaff membuat user dengan role staff

func MakeStaffsUser(idx int) *models.User {

	return NewUserFactory().
		With("username", fmt.Sprintf("staff_%d", idx)).
		With("email", fmt.Sprintf("staff_%d@example.com", idx)).
		With("name", fmt.Sprintf("Staff %d", idx)).
		With("is_staff", true).
		With("is_verified", true).
		Make()

}

// MakeInactiveUser membuat user berstatus tidak aktif
func MakeInactiveUser() *models.User {
	return NewUserFactory().
		With("id", int64(3)).
		With("is_active", false).
		Make()
}

// MakeUserWithPhoto membuat user yang memiliki properti foto profil
func MakeUserWithPhoto() *models.User {
	return NewUserFactory().
		With("photo", "http://localhost:1323/uploads/users/photo.jpg").
		With("photo_thumbnail", "http://localhost:1323/uploads/users/thumbnails/photo_thumb.jpg").
		Make()
}

// MakeDeletedUser membuat data user yang berstatus soft-deleted
func MakeDeletedUser() *models.User {
	return NewUserFactory().
		With("id", int64(99)).
		With("deleted_by", int64(1)).
		With("delete_reason", "Pelanggaran kebijakan").
		With("deleted_at_time", time.Now()).
		Make()
}

// MakeUserList membuat slice berisi sekumpulan data user berurutan
func MakeUserList(count int) []models.User {
	users := make([]models.User, count)
	for i := 0; i < count; i++ {
		idx := i + 1
		u := NewUserFactory().
			With("id", int64(idx)).
			With("username", fmt.Sprintf("user%d", idx)).
			With("email", fmt.Sprintf("user%d@example.com", idx)).
			Make()
		users[i] = *u
	}
	return users
}

// ─── RBAC Factory ──────────────────────────────────────────────────────────────

// MakeRole membuat role untuk testing
func MakeRole(id int64, name string) *rbacModels.Role {
	desc := "Role " + name
	return &rbacModels.Role{
		ID:          id,
		Name:        name,
		DisplayName: "Role " + name,
		Description: &desc,
		IsSystem:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// MakeRoleWithPermissions membuat role lengkap dengan daftar permission di dalamnya
func MakeRoleWithPermissions() *rbacModels.Role {
	role := MakeRole(1, "admin")
	role.Permissions = []rbacModels.Permission{
		{ID: 1, Name: "users:read", DisplayName: "Read Users", Resource: "users", Action: "read"},
		{ID: 2, Name: "users:update", DisplayName: "Update Users", Resource: "users", Action: "update"},
	}
	return role
}

// MakePermission membuat item permission individu untuk testing
func MakePermission(id int64, name string) *rbacModels.Permission {
	desc := "Permission " + name
	resource, action := parsePermission(name)
	return &rbacModels.Permission{
		ID:          id,
		Name:        name,
		DisplayName: "Permission " + name,
		Description: &desc,
		Resource:    resource,
		Action:      action,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func parsePermission(name string) (string, string) {
	for i, c := range name {
		if c == ':' {
			return name[:i], name[i+1:]
		}
	}
	return name, ""
}

// ─── Auth Factory ──────────────────────────────────────────────────────────────

// MakeLoginHistory membuat riwayat login untuk testing
func MakeLoginHistory(userID int64, status string) authModels.LoginHistory {
	return authModels.LoginHistory{
		ID:         1,
		UserID:     &userID,
		Identifier: "testuser",
		IPAddress:  "127.0.0.1",
		Status:     status,
		CreatedAt:  time.Now(),
	}
}

// MakeLoginHistories membuat slice kumpulan riwayat login
func MakeLoginHistories(userID int64, count int) []authModels.LoginHistory {
	histories := make([]authModels.LoginHistory, count)
	for i := 0; i < count; i++ {
		h := MakeLoginHistory(userID, "success")
		h.ID = int64(i + 1)
		histories[i] = h
	}
	return histories
}
