package repositories

import (
	"errors"

	"subian_go/internal/modules/users/contracts"
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/modules/users/models"

	"gorm.io/gorm"
)

// ─── Init ──────────────────────────────────────────────────────────────────────
// repository implements the contracts.Repository interface
type repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) contracts.Repository {
	return &repository{db: db}
}

// ─── End Init ──────────────────────────────────────────────────────────────────

// Create creates a new user
func (r *repository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *repository) GetByID(id int64) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil

}

// GetByUsername retrieves a user by username
func (r *repository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *repository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// List retrieves paginated list of users
func (r *repository) List(page, pageSize int, filter *dto.UserFilter) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 1. Inisialisasi basis query & pastikan record yang di-soft delete tidak ikut terbawa
	query := r.db.Model(&models.User{}).Where("deleted_at IS NULL")

	// 2. Filter Teks (Menggunakan ILIKE untuk case-insensitive)
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Username != "" {
		query = query.Where("username ILIKE ?", "%"+filter.Username+"%")
	}
	if filter.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filter.Email+"%")
	}

	// 3. Filter Boolean (Dicek nilainya lewat pointer agar nilai false tidak otomatis memfilter)
	if filter.IsSuperadmin != nil {
		query = query.Where("is_superadmin = ?", *filter.IsSuperadmin)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.IsStaff != nil {
		query = query.Where("is_staff = ?", *filter.IsStaff)
	}

	// 4. Hitung total data berdasarkan filter yang aktif
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. Ambil data dengan paginasi dan sorting
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates an existing user
func (r *repository) Update(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

// Delete soft deletes a user
// Delete soft deletes a user dengan mencatat pengedit dan alasan
func (r *repository) Delete(id int64, deletedBy int64, reason string) error {
	// Jalankan dalam Transaction agar kedua proses (update & delete) aman dan atomik
	return r.db.Transaction(func(tx *gorm.DB) error {

		// 1. Update kolom deleted_by dan delete_reason terlebih dahulu
		dataUpdate := map[string]interface{}{
			"deleted_by":    deletedBy,
			"delete_reason": reason,
		}

		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(dataUpdate).Error; err != nil {
			return err
		}

		// 2. Eksekusi soft delete (GORM akan mengisi deleted_at secara otomatis)
		if err := tx.Delete(&models.User{}, id).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

// DeletedList retrieves paginated list of soft-deleted users based on filter
func (r *repository) DeletedList(page, pageSize int, filter *dto.UserDeletedFilter) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 1. Menggunakan Unscoped() untuk menembus proteksi soft delete GORM
	//    dan filter hanya yang deleted_at TIDAK NULL
	query := r.db.Unscoped().Model(&models.User{}).Where("deleted_at IS NOT NULL")

	// 2. Filter Teks Dinamis
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Username != "" {
		query = query.Where("username ILIKE ?", "%"+filter.Username+"%")
	}
	if filter.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filter.Email+"%")
	}

	// 4. Hitung total data yang terhapus berdasarkan filter
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. Ambil data terhapus dengan urutan yang paling baru dihapus (deleted_at DESC)
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetSettings retrieves user settings
func (r *repository) GetSettings(id int64) ([]models.UserSetting, error) {
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return user.GetSettings()
}

// UpdateSettings updates user settings
func (r *repository) UpdateSettings(id int64, settings []models.UserSetting) error {
	user, err := r.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return gorm.ErrRecordNotFound
	}

	if err := user.SetSettings(settings); err != nil {
		return err
	}

	return r.Update(user)
}
