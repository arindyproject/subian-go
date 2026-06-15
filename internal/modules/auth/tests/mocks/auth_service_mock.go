package mocks

import (
	"subian_go/internal/modules/auth/dto"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(
	req *dto.LoginRequest,
	ip,
	userAgent string,
) (*dto.TokenResponse, error) {

	args := m.Called(req, ip, userAgent)

	var resp *dto.TokenResponse
	if v := args.Get(0); v != nil {
		resp = v.(*dto.TokenResponse)
	}

	return resp, args.Error(1)
}

func (m *MockAuthService) Register(
	req *dto.RegisterRequest,
) (*dto.RegisterResponse, error) {

	args := m.Called(req)

	var resp *dto.RegisterResponse
	if v := args.Get(0); v != nil {
		resp = v.(*dto.RegisterResponse)
	}

	return resp, args.Error(1)
}

func (m *MockAuthService) RefreshToken(
	req *dto.RefreshTokenRequest,
) (*dto.TokenResponse, error) {

	args := m.Called(req)

	var resp *dto.TokenResponse
	if v := args.Get(0); v != nil {
		resp = v.(*dto.TokenResponse)
	}

	return resp, args.Error(1)
}

func (m *MockAuthService) Logout(
	req *dto.LogoutRequest,
) error {

	args := m.Called(req)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAll(
	userID int64,
) error {

	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) ForgotPassword(
	req *dto.ForgotPasswordRequest,
) error {

	args := m.Called(req)
	return args.Error(0)
}

func (m *MockAuthService) ResetPassword(
	req *dto.ResetPasswordRequest,
) error {

	args := m.Called(req)
	return args.Error(0)
}
