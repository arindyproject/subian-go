package contracts

import (
	"io"
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/modules/users/models"
)

// AuthContext berisi informasi user yang sedang login untuk authorization
type AuthContext struct {
	UserID       int64
	IsSuperadmin bool
}

// ─── Repository ────────────────────────────────────────────────────────────────

type Repository interface {
	Create(user *models.User) error
	GetByID(id int64) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	List(page, pageSize int, filter *dto.UserFilter) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id int64, deletedBy int64, reason string) error
	DeletedList(page, pageSize int, filter *dto.UserDeletedFilter) ([]models.User, int64, error)
	GetSettings(id int64) ([]models.UserSetting, error)
	UpdateSettings(id int64, settings []models.UserSetting) error
}

// ─── Service ───────────────────────────────────────────────────────────────────

type Service interface {
	// CRUD — operasi yang butuh auth context
	CreateUser(req *dto.CreateUserRequest, actor AuthContext) (*dto.UserSimpleResponse, error)
	GetUserByID(id int64, actor AuthContext) (*dto.UserResponse, error)
	GetUserByUsername(username string, actor AuthContext) (*dto.UserResponse, error)
	GetUserByEmail(email string, actor AuthContext) (*dto.UserResponse, error)
	ListUsers(page, pageSize int, filter *dto.UserFilter) ([]dto.UserSimpleResponse, int64, error)
	ListDeletedUsers(page, pageSize int, filter *dto.UserDeletedFilter, actor AuthContext) ([]dto.UserDeletedResponse, int64, error)
	UpdateUser(id int64, req *dto.UpdateUserRequest, actor AuthContext) (*dto.UserResponse, error)
	DeleteUser(id int64, reason string, actor AuthContext) error

	// Password
	ChangePassword(id int64, req *dto.ChangePasswordRequest, actor AuthContext) (*dto.UserResponse, error)
	ResetPassword(id int64, actor AuthContext) error
	UpdateLastLogin(id int64) error

	// Settings
	GetSettings(id int64, actor AuthContext) ([]models.UserSetting, error)
	UpdateSettings(id int64, req *dto.UpdateSettingsRequest, actor AuthContext) (*dto.UserResponse, error)

	//Photo
	UploadPhoto(id int64, filename string, reader io.Reader, actor AuthContext) (*dto.UserResponse, error)
	DeletePhoto(id int64, actor AuthContext) (*dto.UserResponse, error)
}
