package tests

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"subian_go/internal/modules/auth/handlers"
	"subian_go/internal/modules/auth/services"
	"subian_go/internal/modules/auth/tests/fixtures"
	"subian_go/internal/modules/auth/tests/helpers"
	"subian_go/internal/modules/auth/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ─── Test Suite ────────────────────────────────────────────────────────────────

// AuthHandlerTestSuite mengelompokkan semua test handler auth
type AuthHandlerTestSuite struct {
	suite.Suite
	mockService *mocks.MockAuthService
	handler     *handlers.AuthHandler
}

// SetupTest dipanggil sebelum setiap test — reset mock
func (s *AuthHandlerTestSuite) SetupTest() {
	s.mockService = new(mocks.MockAuthService)
	s.handler = handlers.NewAuthHandler(s.mockService)
}

// TearDownTest dipanggil setelah setiap test — verifikasi semua expectation terpenuhi
func (s *AuthHandlerTestSuite) TearDownTest() {
	s.mockService.AssertExpectations(s.T())
}

// Entrypoint suite
func TestAuthHandler(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func TestMain(m *testing.M) {
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")
	fmt.Println("\033[35m  Auth Handler Test Suite\033[0m")
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")

	code := m.Run()

	if code == 0 {
		fmt.Println("\n\033[32m✓  PASS\033[0m  subian_go/internal/modules/auth")
	} else {
		fmt.Println("\n\033[31m✗  FAIL\033[0m  subian_go/internal/modules/auth")
	}

	os.Exit(code)
}

// ─── Login Tests ───────────────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestLogin_Success() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidLoginRequest()
	tokenResp := fixtures.TokenResponse()

	// Setup mock expectation
	s.mockService.On("Login", req, "192.0.2.1", "").
		Return(tokenResp, nil)

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/login", req)
	s.Require().NoError(err)

	// Simulasi RealIP
	c.Request().RemoteAddr = "192.0.2.1:1234"

	s.Require().NoError(s.handler.Login(c))

	// Assert response
	s.Equal(http.StatusOK, rec.Code)

	body, err := helpers.ParseResponse(rec)
	s.Require().NoError(err)
	s.True(body["status"].(bool))
	s.Equal("Login berhasil", body["message"])
	s.NotNil(body["data"])
}

func (s *AuthHandlerTestSuite) TestLogin_InvalidCredentials() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidLoginRequest()

	s.mockService.On("Login", req, "192.0.2.1", "").
		Return(nil, services.NewAuthError(401, "Username atau email tidak terdaftar."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/login", req)
	s.Require().NoError(err)
	c.Request().RemoteAddr = "192.0.2.1:1234"

	s.Require().NoError(s.handler.Login(c))

	s.Equal(http.StatusUnauthorized, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.False(body["status"].(bool))
	s.Equal("Username atau email tidak terdaftar.", body["message"])
}

func (s *AuthHandlerTestSuite) TestLogin_AccountInactive() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidLoginRequest()

	s.mockService.On("Login", req, "192.0.2.1", "").
		Return(nil, services.NewAuthError(403, "Akun Anda tidak aktif. Hubungi administrator."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/login", req)
	s.Require().NoError(err)
	c.Request().RemoteAddr = "192.0.2.1:1234"

	s.Require().NoError(s.handler.Login(c))

	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *AuthHandlerTestSuite) TestLogin_TooManyAttempts() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidLoginRequest()

	s.mockService.On("Login", req, "192.0.2.1", "").
		Return(nil, services.NewAuthError(429, "Terlalu banyak percobaan login. Coba lagi dalam 15 menit."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/login", req)
	s.Require().NoError(err)
	c.Request().RemoteAddr = "192.0.2.1:1234"

	s.Require().NoError(s.handler.Login(c))

	s.Equal(http.StatusTooManyRequests, rec.Code)
}

func (s *AuthHandlerTestSuite) TestLogin_EmptyIdentifier_ValidationFail() {
	e := helpers.NewTestEcho()

	// Request tanpa identifier
	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"identifier": "",
		"password":   "password123",
	})
	s.Require().NoError(err)

	// Tidak ada mock call karena validasi gagal sebelum service dipanggil
	s.Require().NoError(s.handler.Login(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.False(body["status"].(bool))
	s.Equal("Validasi gagal", body["message"])
	s.NotNil(body["errors"])
}

// ─── Register Tests ────────────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestRegister_Success() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidRegisterRequest()
	registerResp := fixtures.RegisterResponse()

	s.mockService.On("Register", req).Return(registerResp, nil)

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/register", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Register(c))

	s.Equal(http.StatusCreated, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.True(body["status"].(bool))
	s.Equal("Registrasi berhasil", body["message"])
}

func (s *AuthHandlerTestSuite) TestRegister_RegistrationDisabled() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidRegisterRequest()

	s.mockService.On("Register", req).
		Return(nil, services.NewAuthError(403, "Registrasi sedang tidak tersedia."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/register", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Register(c))

	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *AuthHandlerTestSuite) TestRegister_DuplicateUsername() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidRegisterRequest()

	s.mockService.On("Register", req).
		Return(nil, services.NewAuthError(422, "username sudah digunakan"))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/register", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Register(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *AuthHandlerTestSuite) TestRegister_InvalidEmail_ValidationFail() {
	e := helpers.NewTestEcho()

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "user",
		"email":    "invalid-email", // ← email tidak valid
		"password": "password123",
	})
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Register(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

// ─── RefreshToken Tests ────────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestRefreshToken_Success() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidRefreshTokenRequest()
	tokenResp := fixtures.TokenResponse()

	s.mockService.On("RefreshToken", req).Return(tokenResp, nil)

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/refresh", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.RefreshToken(c))

	s.Equal(http.StatusOK, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.True(body["status"].(bool))
	s.Equal("Token berhasil diperbarui", body["message"])
}

func (s *AuthHandlerTestSuite) TestRefreshToken_InvalidToken() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidRefreshTokenRequest()

	s.mockService.On("RefreshToken", req).
		Return(nil, services.NewAuthError(401, "Refresh token tidak valid."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/refresh", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.RefreshToken(c))

	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *AuthHandlerTestSuite) TestRefreshToken_BlacklistedToken() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidRefreshTokenRequest()

	s.mockService.On("RefreshToken", req).
		Return(nil, services.NewAuthError(401, "Token sudah tidak berlaku."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/refresh", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.RefreshToken(c))

	s.Equal(http.StatusUnauthorized, rec.Code)
}

// ─── Logout Tests ──────────────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestLogout_Success() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidLogoutRequest()

	s.mockService.On("Logout", req).Return(nil)

	c, rec, err := helpers.NewContextWithJWT(e, http.MethodPost, "/api/v1/auth/logout", req, 1, false)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Logout(c))

	s.Equal(http.StatusOK, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.True(body["status"].(bool))
	s.Equal("Logout berhasil", body["message"])
}

func (s *AuthHandlerTestSuite) TestLogout_ExpiredToken_StillSuccess() {
	// Logout dengan token expired harus tetap return 200
	e := helpers.NewTestEcho()
	req := fixtures.ValidLogoutRequest()

	// Service tetap return nil meski token expired
	s.mockService.On("Logout", req).Return(nil)

	c, rec, err := helpers.NewContextWithJWT(e, http.MethodPost, "/api/v1/auth/logout", req, 1, false)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Logout(c))

	s.Equal(http.StatusOK, rec.Code)
}

func (s *AuthHandlerTestSuite) TestLogout_MissingRefreshToken_ValidationFail() {
	e := helpers.NewTestEcho()

	c, rec, err := helpers.NewContextWithJWT(e, http.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": "", // ← kosong
	}, 1, false)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.Logout(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

// ─── LogoutAll Tests ───────────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestLogoutAll_Success() {
	e := helpers.NewTestEcho()

	var userID int64 = 1
	s.mockService.On("LogoutAll", userID).Return(nil)

	c, rec, err := helpers.NewContextWithJWT(e, http.MethodPost, "/api/v1/auth/logout-all", nil, userID, false)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.LogoutAll(c))

	s.Equal(http.StatusOK, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.True(body["status"].(bool))
	s.Equal("Berhasil logout dari semua perangkat", body["message"])
}

func (s *AuthHandlerTestSuite) TestLogoutAll_NoJWT_Unauthorized() {
	e := helpers.NewTestEcho()

	// Context tanpa JWT claims
	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/logout-all", nil)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.LogoutAll(c))

	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *AuthHandlerTestSuite) TestLogoutAll_SuperuserCanLogoutAll() {
	e := helpers.NewTestEcho()

	var userID int64 = 1
	s.mockService.On("LogoutAll", userID).Return(nil)

	c, rec, err := helpers.NewContextWithJWT(e, http.MethodPost, "/api/v1/auth/logout-all", nil, userID, true)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.LogoutAll(c))

	s.Equal(http.StatusOK, rec.Code)
}

// ─── ForgotPassword Tests ──────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestForgotPassword_AlwaysSuccess() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidForgotPasswordRequest()

	// Selalu return nil meski email tidak ditemukan (security)
	s.mockService.On("ForgotPassword", req).Return(nil)

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/forgot-password", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.ForgotPassword(c))

	// Selalu 200 meski email tidak ada
	s.Equal(http.StatusOK, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.True(body["status"].(bool))
	s.Contains(body["message"].(string), "Jika email terdaftar")
}

func (s *AuthHandlerTestSuite) TestForgotPassword_ValidationFail() {
	e := helpers.NewTestEcho()

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"identifier": "", // ← kosong
	})
	s.Require().NoError(err)

	s.Require().NoError(s.handler.ForgotPassword(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

// ─── ResetPassword Tests ───────────────────────────────────────────────────────

func (s *AuthHandlerTestSuite) TestResetPassword_Success() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidResetPasswordRequest()

	s.mockService.On("ResetPassword", req).Return(nil)

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/reset-password", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.ResetPassword(c))

	s.Equal(http.StatusOK, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	s.True(body["status"].(bool))
	s.Contains(body["message"].(string), "Password berhasil direset")
}

func (s *AuthHandlerTestSuite) TestResetPassword_InvalidToken() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidResetPasswordRequest()

	s.mockService.On("ResetPassword", req).
		Return(services.NewAuthError(400, "Token tidak valid atau sudah kedaluwarsa."))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/reset-password", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.ResetPassword(c))

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *AuthHandlerTestSuite) TestResetPassword_PasswordMismatch_ValidationFail() {
	e := helpers.NewTestEcho()

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":            "valid-token",
		"new_password":     "password123",
		"confirm_password": "", // ← kosong
	})
	s.Require().NoError(err)

	s.Require().NoError(s.handler.ResetPassword(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *AuthHandlerTestSuite) TestResetPassword_WeakPassword() {
	e := helpers.NewTestEcho()
	req := fixtures.ValidResetPasswordRequest()

	s.mockService.On("ResetPassword", req).
		Return(services.NewAuthError(422, "Password minimal 6 karakter"))

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/reset-password", req)
	s.Require().NoError(err)

	s.Require().NoError(s.handler.ResetPassword(c))

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

// ─── Standalone Tests (tanpa suite) ───────────────────────────────────────────
// Gunakan ini untuk test sederhana yang tidak perlu setup/teardown

func TestLogin_InvalidJSON(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	handler := handlers.NewAuthHandler(mockSvc)
	e := helpers.NewTestEcho()

	// Body bukan JSON valid
	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/login", "invalid-json")
	assert.NoError(t, err)

	assert.NoError(t, handler.Login(c))
	// Bind error → 400
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Tidak ada mock call karena bind gagal duluan
	mockSvc.AssertNotCalled(t, "Login")
}

func TestRegister_PasswordTooShort(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	handler := handlers.NewAuthHandler(mockSvc)
	e := helpers.NewTestEcho()

	c, rec, err := helpers.NewContext(e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "user123",
		"email":    "user@example.com",
		"password": "123", // ← terlalu pendek (min 8)
	})
	assert.NoError(t, err)
	assert.NoError(t, handler.Register(c))

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	body, _ := helpers.ParseResponse(rec)
	assert.False(t, body["status"].(bool))
	assert.NotNil(t, body["errors"])

	mockSvc.AssertNotCalled(t, "Register")
}
