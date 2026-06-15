package repositories

import (
	"errors"

	"subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/modules/rbac/models"
	userModel "subian_go/internal/modules/users/models"

	"gorm.io/gorm"
)

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) contracts.RBACRepository {
	return &rbacRepository{db: db}
}

// IsSuperadmin mengambil langsung status terbaru dari database
func (r *rbacRepository) IsSuperadmin(userID int64) (bool, error) {
	var isSuperadmin bool

	// Kita gunakan .Pluck untuk mengambil satu kolom saja secara efisien
	err := r.db.Model(&userModel.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Pluck("is_superadmin", &isSuperadmin).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return isSuperadmin, nil
}

// ─── Permission ────────────────────────────────────────────────────────────────

func (r *rbacRepository) CreatePermission(p *models.Permission) error {
	return r.db.Create(p).Error
}

func (r *rbacRepository) GetPermissionByID(id int64) (*models.Permission, error) {
	var p models.Permission
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&p)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, result.Error
}

func (r *rbacRepository) GetPermissionByName(name string) (*models.Permission, error) {
	var p models.Permission
	result := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&p)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, result.Error
}

func (r *rbacRepository) ListPermissions(page, pageSize int) ([]models.Permission, int64, error) {
	var items []models.Permission
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.Model(&models.Permission{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Where("deleted_at IS NULL").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *rbacRepository) UpdatePermission(p *models.Permission) error {
	return r.db.Save(p).Error
}

func (r *rbacRepository) DeletePermission(id int64) error {
	return r.db.Where("id = ?", id).Delete(&models.Permission{}).Error
}

// ─── Role ──────────────────────────────────────────────────────────────────────

func (r *rbacRepository) CreateRole(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *rbacRepository) GetRoleByID(id int64) (*models.Role, error) {
	var role models.Role
	result := r.db.Preload("Permissions").Where("id = ? AND deleted_at IS NULL", id).First(&role)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, result.Error
}

func (r *rbacRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	result := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&role)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, result.Error
}

func (r *rbacRepository) ListRoles(page, pageSize int) ([]models.Role, int64, error) {
	var items []models.Role
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.Model(&models.Role{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Preload("Permissions").Where("deleted_at IS NULL").
		Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *rbacRepository) UpdateRole(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *rbacRepository) DeleteRole(id int64) error {
	// Cek apakah system role
	var role models.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("system role tidak bisa dihapus")
	}
	return r.db.Where("id = ?", id).Delete(&models.Role{}).Error
}

func (r *rbacRepository) GetUsersRoles(userIDs []int64) (map[int64][]models.Role, error) {
	// 1. Definisikan struct flat agar GORM tidak mengiranya sebagai relasi/foreign key
	type dbResult struct {
		UserID      int64
		ID          int64
		Name        string
		DisplayName string
		Description *string
		IsSystem    bool
		CreatedBy   *int64
		UpdatedBy   *int64
		// Masukkan yang lain jika memang kolom ini dibutuhkan di aplikasi
		// CreatedAt   time.Time
		// UpdatedAt   time.Time
	}
	var results []dbResult

	// 2. Lakukan query dan scan ke struct flat tadi
	err := r.db.Table("user_roles ur").
		Select("ur.user_id, r.id, r.name, r.display_name, r.description, r.is_system, r.created_by, r.updated_by").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.user_id IN ? AND r.deleted_at IS NULL", userIDs).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 3. Petakan hasil flat dbResult ke objek map[int64][]models.Role
	userRolesMap := make(map[int64][]models.Role)
	for _, res := range results {
		role := models.Role{
			ID:          res.ID,
			Name:        res.Name,
			DisplayName: res.DisplayName,
			Description: res.Description,
			IsSystem:    res.IsSystem,
			CreatedBy:   res.CreatedBy,
			UpdatedBy:   res.UpdatedBy,
		}
		userRolesMap[res.UserID] = append(userRolesMap[res.UserID], role)
	}

	return userRolesMap, nil
}

// ─── Role ↔ Permission ─────────────────────────────────────────────────────────

func (r *rbacRepository) AssignPermissionsToRole(roleID int64, permissionIDs []int64) error {
	var pivots []models.RolePermission
	for _, pID := range permissionIDs {
		pivots = append(pivots, models.RolePermission{RoleID: roleID, PermissionID: pID})
	}
	return r.db.Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		FirstOrCreate(&pivots).Error
}

func (r *rbacRepository) RevokePermissionsFromRole(roleID int64, permissionIDs []int64) error {
	return r.db.Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&models.RolePermission{}).Error
}

func (r *rbacRepository) GetRolePermissions(roleID int64) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ? AND permissions.deleted_at IS NULL", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *rbacRepository) SyncRolePermissions(roleID int64, permissionIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Hapus semua permission lama
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		// Assign yang baru
		if len(permissionIDs) == 0 {
			return nil
		}
		var pivots []models.RolePermission
		for _, pID := range permissionIDs {
			pivots = append(pivots, models.RolePermission{RoleID: roleID, PermissionID: pID})
		}
		return tx.Create(&pivots).Error
	})
}

// ─── User ↔ Role ───────────────────────────────────────────────────────────────

func (r *rbacRepository) AssignRolesToUser(userID int64, roleIDs []int64, assignedBy *int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, rID := range roleIDs {
			pivot := models.UserRole{UserID: userID, RoleID: rID, CreatedBy: assignedBy}
			if err := tx.Where(models.UserRole{UserID: userID, RoleID: rID}).
				FirstOrCreate(&pivot).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *rbacRepository) RevokeRolesFromUser(userID int64, roleIDs []int64) error {
	return r.db.Where("user_id = ? AND role_id IN ?", userID, roleIDs).
		Delete(&models.UserRole{}).Error
}

func (r *rbacRepository) GetUserRoles(userID int64) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}

func (r *rbacRepository) SyncUserRoles(userID int64, roleIDs []int64, assignedBy *int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		var pivots []models.UserRole
		for _, rID := range roleIDs {
			pivots = append(pivots, models.UserRole{UserID: userID, RoleID: rID, CreatedBy: assignedBy})
		}
		return tx.Create(&pivots).Error
	})
}

// ─── User ↔ Permission (direct) ───────────────────────────────────────────────

func (r *rbacRepository) AssignDirectPermission(userID, permissionID int64, isGranted bool, assignedBy *int64) error {
	pivot := models.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
		IsGranted:    isGranted,
		CreatedBy:    assignedBy,
	}
	return r.db.Where(models.UserPermission{UserID: userID, PermissionID: permissionID}).
		Assign(map[string]interface{}{"is_granted": isGranted, "created_by": assignedBy}).
		FirstOrCreate(&pivot).Error
}

func (r *rbacRepository) RevokeDirectPermission(userID, permissionID int64) error {
	return r.db.Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Delete(&models.UserPermission{}).Error
}

func (r *rbacRepository) GetUserDirectPermissions(userID int64) ([]models.UserPermission, error) {
	var items []models.UserPermission
	err := r.db.Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

// ─── Check ─────────────────────────────────────────────────────────────────────

// GetUserAllPermissions mengambil semua permission user dari role + direct permission
func (r *rbacRepository) GetUserAllPermissions(userID int64) ([]string, error) {
	// 1. Ambil permission dari role
	var rolePermNames []string
	r.db.Raw(`
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ? AND p.deleted_at IS NULL
	`, userID).Scan(&rolePermNames)

	// 2. Ambil direct permission
	var directPerms []struct {
		Name      string
		IsGranted bool
	}
	r.db.Raw(`
		SELECT p.name, up.is_granted
		FROM permissions p
		JOIN user_permissions up ON up.permission_id = p.id
		WHERE up.user_id = ? AND p.deleted_at IS NULL
	`, userID).Scan(&directPerms)

	// 3. Merge: direct permission override role permission
	permSet := make(map[string]bool)
	for _, name := range rolePermNames {
		permSet[name] = true
	}
	for _, dp := range directPerms {
		permSet[dp.Name] = dp.IsGranted // override
	}

	var result []string
	for name, granted := range permSet {
		if granted {
			result = append(result, name)
		}
	}
	return result, nil
}

func (r *rbacRepository) HasPermission(userID int64, permission string) (bool, error) {
	perms, err := r.GetUserAllPermissions(userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p == permission {
			return true, nil
		}
	}
	return false, nil
}
