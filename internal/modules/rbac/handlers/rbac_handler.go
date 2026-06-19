package handlers

import (
	"net/http"
	"strconv"

	"subian_go/internal/modules/rbac/dto"
	"subian_go/internal/shared/response"
	"subian_go/internal/shared/validator"

	he "subian_go/internal/shared/httputil"

	"github.com/labstack/echo/v5"
)

// ─── Permission Handlers ───────────────────────────────────────────────────────
// ListPermissions handles GET /permissions
//
// ListPermissions godoc
//
//	@Summary		List permissions
//	@Description	Get a paginated list of permissions
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number (default: 1)"
//	@Param			page_size	query		int	false	"Page size (default: 10, max: 100)"
//	@Success		200			{object}	response.MyGoResponse{data=[]dto.PermissionResponse}
//	@Router			/rbac/permissions [get]
//
// GetPermission handles GET /permissions/{id}
func (h *RBACHandler) ListPermissions(c *echo.Context) error {
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

	items, total, err := h.service.ListPermissions(page, pageSize)
	if err != nil {
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil data permission", nil, nil)
	}
	return response.Paginated(c, http.StatusOK, true, "Berhasil mengambil data permission", items, total, page, pageSize)
}

// GetPermission handles GET /permissions/{id}
// GetPermission godoc
//
//	@Summary		Get permission by ID
//	@Description	Get a permission by its ID
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Permission ID"
//	@Success		200	{object}	response.MyGoResponse{data=dto.PermissionResponse}
//	@Router			/rbac/permissions/{id} [get]
//
// UpdatePermission handles PUT /permissions/{id}
func (h *RBACHandler) GetPermission(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	item, err := h.service.GetPermissionByID(id)
	if err != nil {
		return response.Response(c, http.StatusNotFound, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Berhasil mengambil permission", item, nil)
}

// CreatePermission handles POST /permissions
//
// CreatePermission godoc
//
//	@Summary		Create a new permission
//	@Description	Create a new permission with name, display name, description, resource, and action
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.CreatePermissionRequest	true	"Permission data"
//	@Success		201		{object}	response.MyGoResponse{data=dto.PermissionResponse}
//	@Router			/rbac/permissions [post]
func (h *RBACHandler) CreatePermission(c *echo.Context) error {
	var req dto.CreatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	item, err := h.service.CreatePermission(&req, he.GetActorID(c))
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusCreated, true, "Permission berhasil dibuat", item, nil)
}

// UpdatePermission handles PUT /permissions/{id}
// UpdatePermission godoc
//
//	@Summary		Update a permission
//	@Description	Update a permission's display name, description, resource, or action
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Permission ID"
//	@Param			request	body	dto.UpdatePermissionRequest	true	"Updated permission data"
//	@Success		200		{object}	response.MyGoResponse{data=dto.PermissionResponse}
//	@Router			/rbac/permissions/{id} [put]
func (h *RBACHandler) UpdatePermission(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	var req dto.UpdatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	item, err := h.service.UpdatePermission(id, &req, he.GetActorID(c))
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Permission berhasil diupdate", item, nil)
}

// DeletePermission handles DELETE /permissions/{id}
// DeletePermission godoc
//
//	@Summary		Delete a permission
//	@Description	Delete a permission by its ID
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Permission ID"
//	@Success		200	{object}	response.MyGoResponse{data=string} "Permission berhasil dihapus"
//	@Router			/rbac/permissions/{id} [delete]
func (h *RBACHandler) DeletePermission(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	if err := h.service.DeletePermission(id); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Permission berhasil dihapus", nil, nil)
}

// ─── Role Handlers ─────────────────────────────────────────────────────────────

// ListRoles handles GET /roles
//
// ListRoles godoc
//
//	@Summary		List roles
//	@Description	Get a paginated list of roles
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number (default: 1)"
//	@Param			page_size	query		int	false	"Page size (default: 10, max: 100)"
//	@Success		200			{object}	response.MyGoResponse{data=[]dto.RoleResponse}
//	@Router			/rbac/roles [get]
func (h *RBACHandler) ListRoles(c *echo.Context) error {
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
	items, total, err := h.service.ListRoles(page, pageSize)
	if err != nil {
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil data role", nil, nil)
	}
	return response.Paginated(c, http.StatusOK, true, "Berhasil mengambil data role", items, total, page, pageSize)
}

// GetRole handles GET /roles/{id}
//
// GetRole godoc
//
//	@Summary		Get role by ID
//	@Description	Get a role by its ID
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Role ID"
//	@Success		200	{object}	response.MyGoResponse{data=dto.RoleResponse}
//	@Router			/rbac/roles/{id} [get]
func (h *RBACHandler) GetRole(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	item, err := h.service.GetRoleByID(id)
	if err != nil {
		return response.Response(c, http.StatusNotFound, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Berhasil mengambil role", item, nil)
}

// CreateRole handles POST /roles
//
// CreateRole godoc
//
//	@Summary		Create a new role
//	@Description	Create a new role with the provided details
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role	body	dto.CreateRoleRequest	true	"Role details"
//	@Success		201	{object}	response.MyGoResponse{data=dto.RoleResponse}
//	@Router			/rbac/roles [post]
func (h *RBACHandler) CreateRole(c *echo.Context) error {
	var req dto.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	item, err := h.service.CreateRole(&req, he.GetActorID(c))
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusCreated, true, "Role berhasil dibuat", item, nil)
}

// UpdateRole handles PUT /roles/{id}
//
// UpdateRole godoc
//
//	@Summary		Update a role
//	@Description	Update a role's display name or description
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Role ID"
func (h *RBACHandler) UpdateRole(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	var req dto.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	item, err := h.service.UpdateRole(id, &req, he.GetActorID(c))
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Role berhasil diupdate", item, nil)
}

// DeleteRole handles DELETE /roles/{id}
// DeleteRole godoc
//
//	@Summary		Delete a role
//	@Description	Delete a role by its ID
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Role ID"
//	@Success		200	{object}	response.MyGoResponse{data=string} "Role berhasil dihapus"
//	@Router			/rbac/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	if err := h.service.DeleteRole(id); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Role berhasil dihapus", nil, nil)
}

// ─── Role ↔ Permission Handlers ────────────────────────────────────────────────

// AssignPermissionsToRole handles POST /roles/{id}/permissions
//
// AssignPermissionsToRole godoc
//
//	@Summary		Assign permissions to a role
//	@Description	Assign one or more permissions to a role
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int								true	"Role ID"
//	@Param			request	body	dto.AssignPermissionsRequest	true	"Permission IDs to assign"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Permission berhasil ditambahkan ke role"
//	@Router			/rbac/roles/{id}/permissions [post]
func (h *RBACHandler) AssignPermissionsToRole(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	var req dto.AssignPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	if err := h.service.AssignPermissionsToRole(id, &req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Permission berhasil ditambahkan ke role", nil, nil)
}

// SyncRolePermissions handles PUT /roles/{id}/permissions
// SyncRolePermissions godoc
//
//	@Summary		Sync permissions of a role
//	@Description	Sync permissions of a role (revoke all existing and assign new ones)
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int								true	"Role ID"
//	@Param			request	body	dto.AssignPermissionsRequest	true	"Permission IDs to sync"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Permission role berhasil disinkronkan"
//	@Router			/rbac/roles/{id}/permissions [put]
func (h *RBACHandler) SyncRolePermissions(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	var req dto.AssignPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if err := h.service.SyncRolePermissions(id, &req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Permission role berhasil disinkronkan", nil, nil)
}

// RevokePermissionsFromRole handles DELETE /roles/{id}/permissions
// RevokePermissionsFromRole godoc
//
//	@Summary		Revoke permissions from a role
//	@Description	Revoke one or more permissions from a role
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int								true	"Role ID"
//	@Param			request	body	dto.AssignPermissionsRequest	true	"Permission IDs to revoke"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Permission berhasil dicabut dari role"
//	@Router			/rbac/roles/{id}/permissions [delete]
func (h *RBACHandler) RevokePermissionsFromRole(c *echo.Context) error {
	id, err := he.ParseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}
	var req dto.AssignPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if err := h.service.RevokePermissionsFromRole(id, &req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Permission berhasil dicabut dari role", nil, nil)
}

// ─── User ↔ Role Handlers ──────────────────────────────────────────────────────

// RequirePermission memastikan user memiliki permission tertentu
// RequireAnyPermission memastikan user memiliki minimal satu permission
// RequireRole memastikan user memiliki role tertentu
// GetUserRoles handles GET /users/{user_id}/roles
// GetUserRoles godoc
//
//	@Summary		Get user roles
//	@Description	Get all roles assigned to a user
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path	int	true	"User ID"
//	@Success		200	{object}	response.MyGoResponse{data=[]dto.RoleResponse}
//	@Router			/rbac/users/{user_id}/roles [get]
func (h *RBACHandler) GetUserRoles(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "User ID tidak valid", nil, nil)
	}

	roles, err := h.service.GetUserRoles(userID)
	if err != nil {
		if appErr, ok := err.(interface{ StatusCode() int }); ok {
			return response.Response(c, appErr.StatusCode(), false, err.Error(), nil, nil)
		}
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil role user", nil, nil)
	}

	return response.Response(c, http.StatusOK, true, "Berhasil mengambil role user", roles, nil)
}

// AssignRolesToUser handles POST /users/{user_id}/roles
//
// AssignRolesToUser godoc
//
//	@Summary		Assign roles to a user
//	@Description	Assign one or more roles to a user
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path	int							true	"User ID"
//	@Param			request	body	dto.AssignRolesRequest	true	"Role IDs to assign"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Role berhasil ditambahkan ke user"
//	@Router			/rbac/users/{user_id}/roles [post]
func (h *RBACHandler) AssignRolesToUser(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "User ID tidak valid", nil, nil)
	}
	var req dto.AssignRolesRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	if err := h.service.AssignRolesToUser(userID, &req, he.GetActorID(c)); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Role berhasil ditambahkan ke user", nil, nil)
}

// SyncUserRoles handles PUT /users/{user_id}/roles
// SyncUserRoles godoc
//
//	@Summary		Sync roles of a user
//	@Description	Sync roles of a user (revoke all existing and assign new ones)
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path	int							true	"User ID"
//	@Param			request	body	dto.AssignRolesRequest	true	"Role IDs to sync"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Role user berhasil disinkronkan"
//	@Router			/rbac/users/{user_id}/roles [put]
func (h *RBACHandler) SyncUserRoles(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "User ID tidak valid", nil, nil)
	}
	var req dto.AssignRolesRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if err := h.service.SyncUserRoles(userID, &req, he.GetActorID(c)); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Role user berhasil disinkronkan", nil, nil)
}

// RevokeRolesFromUser handles DELETE /users/{user_id}/roles
// RevokeRolesFromUser godoc
//
//	@Summary		Revoke roles from a user
//	@Description	Revoke one or more roles from a user
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path	int							true	"User ID"
//	@Param			request	body	dto.AssignRolesRequest	true	"Role IDs to revoke"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Role berhasil dicabut dari user"
//	@Router			/rbac/users/{user_id}/roles [delete]
func (h *RBACHandler) RevokeRolesFromUser(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "User ID tidak valid", nil, nil)
	}
	var req dto.AssignRolesRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if err := h.service.RevokeRolesFromUser(userID, &req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Role berhasil dicabut dari user", nil, nil)
}

// ─── User Permissions ──────────────────────────────────────────────────────────

// GetUserAllPermissions handles GET /users/{user_id}/permissions
//
// GetUserAllPermissions godoc
//
//	@Summary		Get all permissions of a user
//	@Description	Get all permissions of a user, including those inherited from roles and directly assigned
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path	int	true	"User ID"
//	@Success		200	{object}	response.MyGoResponse{data=[]dto.PermissionResponse}
//	@Router			/rbac/users/{user_id}/permissions [get]
func (h *RBACHandler) GetUserAllPermissions(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "User ID tidak valid", nil, nil)
	}
	perms, err := h.service.GetUserAllPermissions(userID)
	if err != nil {
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil permission", nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Berhasil mengambil semua permission user", perms, nil)
}

// AssignDirectPermission handles POST /users/{user_id}/permissions
// AssignDirectPermission godoc
//
//	@Summary		Assign direct permission to a user
//	@Description	Assign a direct permission to a user
//	@Tags			RBAC
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path	int							true	"User ID"
//	@Param			request	body	dto.AssignDirectPermissionRequest	true	"Permission to assign"
//	@Success		200		{object}	response.MyGoResponse{data=string} "Direct permission berhasil ditetapkan"
//	@Router			/rbac/users/{user_id}/permissions [post]
func (h *RBACHandler) AssignDirectPermission(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "User ID tidak valid", nil, nil)
	}
	var req dto.AssignDirectPermissionRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}
	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}
	if err := h.service.AssignDirectPermission(userID, &req, he.GetActorID(c)); err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}
	return response.Response(c, http.StatusOK, true, "Direct permission berhasil ditetapkan", nil, nil)
}
