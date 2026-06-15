package mocks

import (
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/modules/users/models"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository adalah mock untuk contracts.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int64) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) List(page, pageSize int, filter *dto.UserFilter) ([]models.User, int64, error) {
	args := m.Called(page, pageSize, filter)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int64, deletedBy int64, reason string) error {
	args := m.Called(id, deletedBy, reason)
	return args.Error(0)
}

func (m *MockUserRepository) DeletedList(page, pageSize int, filter *dto.UserDeletedFilter) ([]models.User, int64, error) {
	args := m.Called(page, pageSize, filter)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) GetSettings(id int64) ([]models.UserSetting, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserSetting), args.Error(1)
}

func (m *MockUserRepository) UpdateSettings(id int64, settings []models.UserSetting) error {
	args := m.Called(id, settings)
	return args.Error(0)
}
