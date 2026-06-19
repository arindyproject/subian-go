package handlers

import (
	"net/http"

	"subian_go/internal/modules/auth/dto"
	authMiddlewares "subian_go/internal/modules/auth/middlewares"
	"subian_go/internal/modules/auth/services"
	"subian_go/internal/shared/response"
	"subian_go/internal/shared/validator"

	"github.com/labstack/echo/v5"
)

// ─── End Init ──────────────────────────────────────────────────────────────────

// ─── Login ─────────────────────────────────────────────────────────────────────

// Login godoc
//
//	@Summary		Login user
//	@Description	Login menggunakan username atau email dan password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest	true	"Login Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.TokenResponse}
//	@Router			/auth/login [post]
//
// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	result, err := h.service.Login(&req, ip, userAgent)
	if err != nil {
		return handleAuthError(c, err)
	}

	return response.Response(c, http.StatusOK, true, "Login berhasil", result, nil)
}

// ─── Register ──────────────────────────────────────────────────────────────────
// Register godoc
//
//	@Summary		Register user baru
//	@Description	Mendaftarkan akun baru
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest	true	"Register Request"
//	@Success		201		{object}	response.MyGoResponse{data=dto.RegisterResponse}
//	@Router			/auth/register [post]
//
// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	result, err := h.service.Register(&req)
	if err != nil {
		return handleAuthError(c, err)
	}

	return response.Response(c, http.StatusCreated, true, "Registrasi berhasil", result, nil)
}

// ─── Refresh Token ─────────────────────────────────────────────────────────────
// RefreshToken godoc
//
//	@Summary		Refresh access token
//	@Description	Memperbarui access token menggunakan refresh token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RefreshTokenRequest	true	"Refresh Token Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.TokenResponse}
//	@Router			/auth/refresh [post]
//
// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	result, err := h.service.RefreshToken(&req)
	if err != nil {
		return handleAuthError(c, err)
	}

	return response.Response(c, http.StatusOK, true, "Token berhasil diperbarui", result, nil)
}

// ─── Forgot Password ───────────────────────────────────────────────────────────
// ForgotPassword godoc
//
//	@Summary		Lupa password
//	@Description	Mengirim email reset password. Selalu response 200 untuk keamanan.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.ForgotPasswordRequest	true	"Forgot Password Request"
//	@Success		200		{object}	response.MyGoResponse
//	@Router			/auth/forgot-password [post]
//
// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *echo.Context) error {
	var req dto.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	// Selalu return 200 — jangan bocorkan apakah identifier terdaftar
	h.service.ForgotPassword(&req)

	return response.Response(c, http.StatusOK, true,
		"Jika email terdaftar, kami telah mengirimkan instruksi reset password.", nil, nil)
}

// ─── Reset Password ────────────────────────────────────────────────────────────
// ResetPassword godoc
//
//	@Summary		Reset password
//	@Description	Mereset password menggunakan token dari email
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.ResetPasswordRequest	true	"Reset Password Request"
//	@Success		200		{object}	response.MyGoResponse
//	@Router			/auth/reset-password [post]
//
// ResetPassword handles POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *echo.Context) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	if err := h.service.ResetPassword(&req); err != nil {
		return handleAuthError(c, err)
	}

	return response.Response(c, http.StatusOK, true,
		"Password berhasil direset. Silakan login dengan password baru.", nil, nil)
}

// Logout godoc
//
//	@Summary		Logout dari device saat ini
//	@Description	Memblacklist refresh token sehingga tidak bisa dipakai lagi. Access token tetap valid hingga expired (simpan di client dan hapus saat logout).
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.LogoutRequest	true	"Logout Request"
//	@Success		200		{object}	response.MyGoResponse
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *echo.Context) error {
	var req dto.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	// Logout tidak return error meski token sudah expired/tidak valid
	// untuk mencegah informasi bocor
	h.service.Logout(&req)

	return response.Response(c, http.StatusOK, true, "Logout berhasil", nil, nil)
}

// LogoutAll godoc
//
//	@Summary		Logout dari semua device
//	@Description	Memblacklist semua refresh token user sehingga semua sesi aktif dihentikan
//	@Tags			Auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.MyGoResponse
//	@Router			/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *echo.Context) error {
	// Ambil userID dari JWT claims yang sudah di-set oleh JWTMiddleware
	userID, ok := authMiddlewares.GetUserID(c)
	if !ok {
		return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
	}

	if err := h.service.LogoutAll(userID); err != nil {
		return handleAuthError(c, err)
	}

	return response.Response(c, http.StatusOK, true, "Berhasil logout dari semua perangkat", nil, nil)
}

// ─── Helper ────────────────────────────────────────────────────────────────────

// handleAuthError mengubah AuthError menjadi response yang sesuai
func handleAuthError(c *echo.Context, err error) error {
	if authErr, ok := err.(*services.AuthError); ok {
		return response.Response(c, authErr.Code, false, authErr.Message, nil, nil)
	}
	return response.Response(c, http.StatusInternalServerError, false, "Terjadi kesalahan sistem.", nil, nil)
}
