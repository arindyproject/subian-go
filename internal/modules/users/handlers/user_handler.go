package handlers

import (
	"net/http"
	"strconv"

	rbacMiddlewares "subian_go/internal/modules/rbac/middlewares"
	userContracts "subian_go/internal/modules/users/contracts"
	"subian_go/internal/modules/users/dto"
	"subian_go/internal/shared/response"
	"subian_go/internal/shared/validator"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	service userContracts.Service
}

func NewHandler(service userContracts.Service) *Handler {
	return &Handler{service: service}
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

func parseID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// buildAuthContext membuat AuthContext dari JWT claims di context
func buildAuthContext(c *echo.Context) userContracts.AuthContext {
	userID, _ := rbacMiddlewares.GetUserIDFromContext(c)
	isSuperadmin := rbacMiddlewares.IsSuperadmin(c)
	return userContracts.AuthContext{
		UserID:       userID,
		IsSuperadmin: isSuperadmin,
	}
}

// ─── User CRUD ─────────────────────────────────────────────────────────────────────

// ─── CreateUserHandler ─────────────────────────────────────────────────────────────
// CreateUserHandler godoc
//
//	@Summary		Create user
//	@Description	Create New user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CreateUserRequest	true	"Login Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.UserSimpleResponse}
//	@Router			/users [post]
//
// CreateUserHandler handles POST /api/v1/users
func (h *Handler) CreateUserHandler(c *echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	actor := buildAuthContext(c)
	user, err := h.service.CreateUser(&req, actor)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusCreated, true, "User berhasil dibuat", user, nil)
} // ─── CreateUserHandler ───────────────────────────────────────────────────────────

// ─── GetUserHandler ────────────────────────────────────────────────────────────────
// GetUserHandler godoc
//
//	@Summary		Get user
//	@Description	Get user by :id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	response.MyGoResponse{data=dto.UserResponse}
//	@Router			/users/{id} [get]
//
// GetUserHandler handles GET /api/v1/users/:id
func (h *Handler) GetUserHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	actor := buildAuthContext(c)

	user, err := h.service.GetUserByID(id, actor)

	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusNotFound, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Berhasil mengambil data user", user, nil)
} // ─── GetUserHandler ──────────────────────────────────────────────────────────────

// ─── By Username ───────────────────────────────────────────────────────────────────
// GetByUsernameHandler godoc
//
//	@Summary		Get user
//	@Description	Get user by :username
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	response.MyGoResponse{data=dto.UserResponse}
//	@Router			/users/username/{username} [get]
//
// GetByUsernameHandler handles GET /api/v1/users/username/:username
func (h *Handler) GetByUsernameHandler(c *echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return response.Response(c, http.StatusBadRequest, false, "Username tidak boleh kosong", nil, nil)
	}

	actor := buildAuthContext(c)

	user, err := h.service.GetUserByUsername(username, actor)
	if err != nil {
		return response.Response(c, http.StatusNotFound, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Berhasil mengambil data user", user, nil)
} // ─── By Username ─────────────────────────────────────────────────────────────────

// ─── ListUsersHandler ──────────────────────────────────────────────────────────────
// ListUsersHandler godoc
//
//	@Summary		Get list of users
//	@Description	Get paginated list of users
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query		int		false	"Page number"
//	@Param			page_size		query		int		false	"Page size"
//	@Param			name			query		string	false	"Filter by name (partial match)"
//	@Param			username		query		string	false	"Filter by username (partial match)"
//	@Param			email			query		string	false	"Filter by email (partial match)"
//	@Param			is_superadmin	query		bool	false	"Filter by superadmin status"
//	@Param			is_active		query		bool	false	"Filter by active status"
//	@Param			is_staff		query		bool	false	"Filter by staff status"
//	@Success		200				{object}	response.MyGoResponse{data=[]dto.UserSimpleResponse}
//	@Router			/users [get]
//
// ListUsersHandler handles GET /api/v1/users
// Siapa yang bisa: semua yang login (data dirinya sendiri disaring di service)
func (h *Handler) ListUsersHandler(c *echo.Context) error {
	page, pageSize := 1, 10
	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	// Mengambil query parameter untuk filter
	filter := dto.UserFilter{
		Name:     c.QueryParam("name"),
		Username: c.QueryParam("username"),
		Email:    c.QueryParam("email"),
	}

	// Menggunakan pointer untuk boolean agar bisa membedakan antara "false" kiriman user vs default value Go (false)
	if isSuperadmin := c.QueryParam("is_superadmin"); isSuperadmin != "" {
		b, err := strconv.ParseBool(isSuperadmin)
		if err == nil {
			filter.IsSuperadmin = &b
		}
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		b, err := strconv.ParseBool(isActive)
		if err == nil {
			filter.IsActive = &b
		}
	}
	if isStaff := c.QueryParam("is_staff"); isStaff != "" {
		b, err := strconv.ParseBool(isStaff)
		if err == nil {
			filter.IsStaff = &b
		}
	}

	users, total, err := h.service.ListUsers(page, pageSize, &filter)
	if err != nil {
		// ─── CEK JENIS ERROR UNTUK RESPONS 404 ───────────────────────────
		if err.Error() == "user tidak ditemukan" {
			return response.Response(c, http.StatusNotFound, false, "User tidak ditemukan dengan filter tersebut", nil, nil)
		}
		// ──────────────────────────────────────────────────────────────────

		// Error internal sesungguhnya (misal: database down)
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil data user", nil, nil)
	}
	return response.Paginated(c, http.StatusOK, true, "Berhasil mengambil data user", users, total, page, pageSize)
} // ─── ListUsersHandler ────────────────────────────────────────────────────────────

// ─── UpdateUserHandler ─────────────────────────────────────────────────────────────
// UpdateUserHandler godoc
//
//	@Summary		Update user
//	@Description	Update user by :id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"User ID"
//	@Param			body	body		dto.UpdateUserRequest	true	"Update User Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.UserResponse}
//	@Router			/users/{id} [put]
//
// UpdateUserHandler handles PUT /api/v1/users/:id
//
// Authorization (dicek di service):
// - Superadmin → boleh
// - Diri sendiri → boleh (tapi tidak bisa ubah is_active, is_staff, is_superuser)
// - Punya permission "users:update" → boleh
// - Punya role "hrd" → boleh
func (h *Handler) UpdateUserHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	actor := buildAuthContext(c)
	user, err := h.service.UpdateUser(id, &req, actor)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "User berhasil diupdate", user, nil)
} // ─── UpdateUserHandler ───────────────────────────────────────────────────────────

// ─── DeleteUserHandler ─────────────────────────────────────────────────────────────
// DeleteUserHandler godoc
//
//	@Summary		Delete user
//	@Description	Delete user by :id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"User ID"
//	@Param			body	body		dto.DeleteUserRequest	true	"Delete User Request"
//	@Success		200		{object}	response.MyGoResponse
//	@Router			/users/{id} [delete]
//
// DeleteUserHandler handles DELETE /api/v1/users/:id
// Siapa yang bisa: superadmin
func (h *Handler) DeleteUserHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	// Bind body request untuk mengambil alasan dihapus
	var req dto.DeleteUserRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Format data tidak valid", nil, nil)
	}

	// Opsional: Jika Anda menggunakan validator/struct tag di echo
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	actor := buildAuthContext(c)

	// Kirim req.Reason ke service
	if err := h.service.DeleteUser(id, req.Reason, actor); err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "User berhasil dihapus", nil, nil)
} // ─── DeleteUserHandler ───────────────────────────────────────────────────────────

// ─── ListDeletedUsersHandler ───────────────────────────────────────────────────────
// ListDeletedUsersHandler godoc
//
//	@Summary		Get list of deleted users
//	@Description	Get paginated list of soft-deleted users
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"
//	@Param			page_size	query		int		false	"Page size"
//	@Param			name		query		string	false	"Filter by name (partial match)"
//	@Param			username	query		string	false	"Filter by username (partial match)"
//	@Param			email		query		string	false	"Filter by email (partial match)"
//	@Success		200			{object}	response.MyGoResponse{data=[]dto.UserSimpleResponse}
//	@Router			/users/deleted [get]
//
// ListDeletedUsersHandler handles GET /api/v1/users/deleted
// Siapa yang bisa: superadmin
func (h *Handler) ListDeletedUsersHandler(c *echo.Context) error {
	actor := buildAuthContext(c)
	page, pageSize := 1, 10
	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	// Tangkap kriteria filter dari Query Param URL
	filter := dto.UserDeletedFilter{
		Name:     c.QueryParam("name"),
		Username: c.QueryParam("username"),
		Email:    c.QueryParam("email"),
	}

	// Eksekusi ke Service
	users, total, err := h.service.ListDeletedUsers(page, pageSize, &filter, actor)
	if err != nil {
		// Jika data memang tidak ada/tidak cocok dengan filter
		if err.Error() == "data sampah user kosong" {
			return response.Response(c, http.StatusNotFound, false, "Tidak ada data user yang telah dihapus", nil, nil)
		}
		// Jika ada kendala internal server database
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil data sampah user", nil, nil)
	}

	return response.Paginated(c, http.StatusOK, true, "Berhasil mengambil data sampah user", users, total, page, pageSize)
} // ─── ListDeletedUsersHandler ─────────────────────────────────────────────────────

// ─── Settings ──────────────────────────────────────────────────────────────────────
// GetSettingsHandler godoc
//
//	@Summary		Get user settings
//	@Description	Get settings of a user by :id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	response.MyGoResponse{data=[]models.UserSetting}
//	@Router			/users/{id}/settings [get]
//
// GetSettingsHandler handles GET /api/v1/users/:id/settings
// Siapa yang bisa: diri sendiri atau superadmin
func (h *Handler) GetSettingsHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	actor := buildAuthContext(c)
	settings, err := h.service.GetSettings(id, actor)
	if err != nil {
		return response.Response(c, http.StatusInternalServerError, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Berhasil mengambil settings", settings, nil)
}

// UpdateSettingsHandler godoc
//
//	@Summary		Update user settings
//	@Description	Update settings of a user by :id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"User ID"
//	@Param			body	body		dto.UpdateSettingsRequest	true	"Update Settings Request"
//	@Success		200		{object}	response.MyGoResponse
//	@Router			/users/{id}/settings [put]
//
// UpdateSettingsHandler handles PUT /api/v1/users/:id/settings
// Siapa yang bisa: diri sendiri atau superadmin
func (h *Handler) UpdateSettingsHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	var req dto.UpdateSettingsRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	actor := buildAuthContext(c)
	user, err := h.service.UpdateSettings(id, &req, actor)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Settings berhasil diupdate", user, nil)
} // ─── Settings ────────────────────────────────────────────────────────────────────

// ─── Password ──────────────────────────────────────────────────────────────────────
// ChangePasswordHandler godoc
//
//	@Summary		Change password
//	@Description	Change password of a user by :id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"User ID"
//	@Param			body	body		dto.ChangePasswordRequest	true	"Change Password Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.UserResponse}
//	@Router			/users/{id}/change-password [put]
//
// GetChangePasswordHandler handles GET /api/v1/users/:id/change-password
// Siapa yang bisa: diri sendiri atau superadmin
func (h *Handler) ChangePasswordHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	if errs := dto.ValidatePasswordPolicy(req.NewPassword); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Password tidak memenuhi kebijakan keamanan", nil, errs)
	}

	actor := buildAuthContext(c)
	user, err := h.service.ChangePassword(id, &req, actor)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Password berhasil diubah", user, nil)
}

// ResetPasswordHandler godoc
//
//	@Summary		Reset password
//	@Description	Reset password of a user by :id to default password
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	response.MyGoResponse
//	@Router			/users/{id}/reset-password [post]
//
// ResetPasswordHandler handles POST /api/v1/users/:id/reset-password
// Siapa yang bisa: superadmin
func (h *Handler) ResetPasswordHandler(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	actor := buildAuthContext(c)
	if err := h.service.ResetPassword(id, actor); err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Password berhasil direset ke default", nil, nil)
} // ─── Password ──────────────────────────────────────────────────────────────────

// ─── Upload Photo	 ───────────────────────────────────────────────────────────────
// UploadPhotoHandler godoc
//
//	@Summary		Upload user photo
//	@Description	Upload or update user photo by :id
//	@Tags			Users
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"User ID"
//	@Param			photo	formData	file	true	"Photo file (jpg, jpeg, png, webp, heic)"
//	@Success		200		{object}	response.MyGoResponse{data=dto.UserResponse}
//	@Router			/users/{id}/photo [put]
//
// UploadPhotoHandler handles PUT /api/v1/users/:id/photo
// Siapa yang bisa: diri sendiri atau superadmin
// PUT /users/:id/photo
func (h *Handler) UploadPhoto(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	file, err := c.FormFile("photo")
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "file photo wajib diisi", nil, nil)
	}

	src, err := file.Open()
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "gagal membaca file", nil, nil)
	}
	defer src.Close()

	actor := buildAuthContext(c)
	result, err := h.service.UploadPhoto(id, file.Filename, src, actor)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}

	return response.Response(c, http.StatusOK, true, "Foto berhasil diperbarui", result, nil)
} // ─── Upload Photo	 ───────────────────────────────────────────────────────────────

// ─── Delete Photo	 ───────────────────────────────────────────────────────────────────
// DeletePhotoHandler godoc
//
//	@Summary		Delete user photo
//	@Description	Delete user photo by :id (set to default avatar)
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	response.MyGoResponse{data=dto.UserResponse}
//	@Router			/users/{id}/photo [delete]
//
// DeletePhotoHandler handles DELETE /api/v1/users/:id/photo
// Siapa yang bisa: diri sendiri atau superadmin
func (h *Handler) DeletePhoto(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	actor := buildAuthContext(c)
	result, err := h.service.DeletePhoto(id, actor)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}

	return response.Response(c, http.StatusOK, true, "Foto berhasil dihapus", result, nil)
} // ─── Delete Photo	 ─────────────────────────────────────────────────────────────────
