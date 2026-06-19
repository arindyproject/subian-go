package contracts

import (
	"io"
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/modules/users/models"
	he "subian_go/internal/shared/httputil"
)

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
	CreateUser(req *dto.CreateUserRequest, actor he.AuthContext) (*dto.UserSimpleResponse, error)
	GetUserByID(id int64, actor he.AuthContext) (*dto.UserResponse, error)
	GetUserByUsername(username string, actor he.AuthContext) (*dto.UserResponse, error)
	GetUserByEmail(email string, actor he.AuthContext) (*dto.UserResponse, error)
	ListUsers(page, pageSize int, filter *dto.UserFilter) ([]dto.UserSimpleResponse, int64, error)
	ListDeletedUsers(page, pageSize int, filter *dto.UserDeletedFilter, actor he.AuthContext) ([]dto.UserDeletedResponse, int64, error)
	UpdateUser(id int64, req *dto.UpdateUserRequest, actor he.AuthContext) (*dto.UserResponse, error)
	DeleteUser(id int64, reason string, actor he.AuthContext) error

	// Password
	ChangePassword(id int64, req *dto.ChangePasswordRequest, actor he.AuthContext) (*dto.UserResponse, error)
	ResetPassword(id int64, actor he.AuthContext) error
	UpdateLastLogin(id int64) error

	// Settings
	GetSettings(id int64, actor he.AuthContext) ([]models.UserSetting, error)
	UpdateSettings(id int64, req *dto.UpdateSettingsRequest, actor he.AuthContext) (*dto.UserResponse, error)

	//Photo
	UploadPhoto(id int64, filename string, reader io.Reader, actor he.AuthContext) (*dto.UserResponse, error)
	DeletePhoto(id int64, actor he.AuthContext) (*dto.UserResponse, error)
}
