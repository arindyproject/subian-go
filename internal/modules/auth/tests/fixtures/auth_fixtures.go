package fixtures

import (
	"time"

	"subian_go/internal/modules/auth/dto"
	userDto "subian_go/internal/modules/users/dto"
)

// ─── Request Fixtures ──────────────────────────────────────────────────────────

func ValidLoginRequest() *dto.LoginRequest {
	return &dto.LoginRequest{
		Identifier: "testuser",
		Password:   "password123",
	}
}

func ValidRegisterRequest() *dto.RegisterRequest {
	name := "Test User"
	return &dto.RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     &name,
	}
}

func ValidRefreshTokenRequest() *dto.RefreshTokenRequest {
	return &dto.RefreshTokenRequest{
		RefreshToken: "valid.refresh.token",
	}
}

func ValidLogoutRequest() *dto.LogoutRequest {
	return &dto.LogoutRequest{
		RefreshToken: "valid.refresh.token",
	}
}

func ValidForgotPasswordRequest() *dto.ForgotPasswordRequest {
	return &dto.ForgotPasswordRequest{
		Identifier: "testuser@example.com",
	}
}

func ValidResetPasswordRequest() *dto.ResetPasswordRequest {
	return &dto.ResetPasswordRequest{
		Token:           "valid-reset-token",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}
}

// ─── Response Fixtures ─────────────────────────────────────────────────────────

func TokenResponse() *dto.TokenResponse {
	return &dto.TokenResponse{
		AccessToken:  "eyJ.access.token",
		RefreshToken: "eyJ.refresh.token",
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 menit
		User: userDto.UserResponse{
			ID:             1,
			Photo:          nil,
			PhotoThumbnail: nil,
			Username:       "testuser",
			Email:          "test@example.com",
			Name:           "Test User",
			IsSuperadmin:   false,
			IsStaff:        false,
			IsVerified:     true,
			Histories:      nil,
			Settings:       nil,
		},
	}
}

func RegisterResponse() *dto.RegisterResponse {
	return &dto.RegisterResponse{
		ID:         1,
		Username:   "newuser",
		Email:      "newuser@example.com",
		Name:       "Test User",
		IsActive:   true,
		IsVerified: true,
		CreatedAt:  time.Now(),
	}
}
