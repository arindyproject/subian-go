package mocks

import (
	authModels "subian_go/internal/modules/auth/models"

	"github.com/stretchr/testify/mock"
)

// MockAuthRepository adalah mock untuk authContracts.AuthRepository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) SaveToken(token *authModels.AuthToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAuthRepository) GetTokenByJTI(jti string) (*authModels.AuthToken, error) {
	args := m.Called(jti)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authModels.AuthToken), args.Error(1)
}

func (m *MockAuthRepository) BlacklistToken(jti string) error {
	args := m.Called(jti)
	return args.Error(0)
}

func (m *MockAuthRepository) BlacklistAllUserTokens(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthRepository) CountActiveTokens(userID int64) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuthRepository) SaveLoginHistory(history *authModels.LoginHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *MockAuthRepository) GetUserLoginHistories(userID int64, limit int) ([]authModels.LoginHistory, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]authModels.LoginHistory), args.Error(1)
}

func (m *MockAuthRepository) SavePasswordHistory(history *authModels.PasswordHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *MockAuthRepository) GetPasswordHistories(userID int64, limit int) ([]authModels.PasswordHistory, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]authModels.PasswordHistory), args.Error(1)
}
