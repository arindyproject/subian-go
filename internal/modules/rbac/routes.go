package rbac

import (
	authMiddlewares "subian_go/internal/modules/auth/middlewares"
	"subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/modules/rbac/handlers"
	rbacMiddlewares "subian_go/internal/modules/rbac/middlewares"
	rbacModels "subian_go/internal/modules/rbac/models"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan semua routes RBAC
// db dibutuhkan JWTMiddleware untuk cek isSuperadmin realtime
func RegisterRoutes(e *echo.Echo, h *handlers.RBACHandler, repo contracts.RBACRepository, jwtManager *utils.JWTManager, db *gorm.DB) {
	jwt := authMiddlewares.JWTMiddleware(jwtManager, db) // ← tambah db

	// ─── Permissions (need Login) ─────────────────────────
	perms := e.Group("/api/v1/rbac/permissions", jwt)
	perms.GET("", h.ListPermissions)
	perms.GET("/:id", h.GetPermission)
	perms.POST("", h.CreatePermission,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsCreate),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage), // alternatif: bisa punya perm manage untuk akses semua aksi terkait permissions
		rbacMiddlewares.RequireSuperadmin(),
	)
	perms.PUT("/:id", h.UpdatePermission,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsUpdate),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage), // alternatif: bisa punya perm manage untuk akses semua aksi terkait permissions
		rbacMiddlewares.RequireSuperadmin(),
	)
	perms.DELETE("/:id", h.DeletePermission,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsDelete),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage), // alternatif: bisa punya perm manage untuk akses semua aksi terkait permissions
		rbacMiddlewares.RequireSuperadmin(),
	)

	// ─── Roles ─────────────────────────────────────────────────
	roles := e.Group("/api/v1/rbac/roles", jwt)
	roles.GET("", h.ListRoles)
	roles.GET("/:id", h.GetRole)
	roles.POST("", h.CreateRole,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesCreate),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)
	roles.PUT("/:id", h.UpdateRole,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesUpdate),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)
	roles.DELETE("/:id", h.DeleteRole,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesDelete),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)

	// ─── Role ↔ Permission ─────────────────────────────────────
	roles.POST("/:id/permissions", h.AssignPermissionsToRole,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage), // cukup punya perm manage untuk akses semua aksi terkait permissions
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)
	roles.PUT("/:id/permissions", h.SyncRolePermissions,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)
	roles.DELETE("/:id/permissions", h.RevokePermissionsFromRole,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage),
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)

	// ─── Users with Role Permissions ───────────────────────────
	// ─── User ↔ Role ───────────────────────────────────────────
	userRoles := e.Group("/api/v1/rbac/users/:user_id/roles", jwt)
	userRoles.GET("", h.GetUserRoles,
		rbacMiddlewares.RequireSelfOrPermission(repo, rbacModels.PermRolesManage), // cukup punya perm view untuk akses semua aksi terkait users
		rbacMiddlewares.RequireSuperadmin(),
	)
	userRoles.POST("", h.AssignRolesToUser,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage), // cukup punya perm manage untuk akses semua aksi terkait roles
		rbacMiddlewares.RequireSuperadmin(),
	)
	userRoles.PUT("", h.SyncUserRoles,
		rbacMiddlewares.RequirePermission(repo, rbacModels.PermRolesManage),
		rbacMiddlewares.RequireSuperadmin(),
	)
	userRoles.DELETE("", h.RevokeRolesFromUser,

		rbacMiddlewares.RequireSuperadmin(),
	)

	// ─── User Permissions ──────────────────────────────────────
	userPerms := e.Group("/api/v1/rbac/users/:user_id/permissions", jwt)
	userPerms.GET("", h.GetUserAllPermissions, rbacMiddlewares.RequireSelfOrPermission(repo, rbacModels.PermPermissionsManage),
		rbacMiddlewares.RequireSuperadmin())
	userPerms.POST("", h.AssignDirectPermission, rbacMiddlewares.RequirePermission(repo, rbacModels.PermPermissionsManage),
		rbacMiddlewares.RequireSuperadmin())
}
