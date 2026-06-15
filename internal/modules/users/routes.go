package users

import (
	authMiddlewares "subian_go/internal/modules/auth/middlewares"
	rbacContracts "subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/modules/users/handlers"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan semua routes users dengan RBAC
// db dibutuhkan oleh JWTMiddleware untuk cek isSuperadmin secara realtime
func RegisterRoutes(e *echo.Echo, h *handlers.Handler, rbacRepo rbacContracts.RBACRepository, jwtManager *utils.JWTManager, db *gorm.DB) {
	jwt := authMiddlewares.JWTMiddleware(jwtManager, db)

	// ─── Protected ─────────────────────────────────────────────
	protected := e.Group("/api/v1/users", jwt)
	protected.POST("", h.CreateUserHandler)
	protected.GET("/:id", h.GetUserHandler)
	protected.GET("/username/:username", h.GetByUsernameHandler)
	protected.GET("", h.ListUsersHandler)
	protected.PUT("/:id", h.UpdateUserHandler)
	protected.DELETE("/:id", h.DeleteUserHandler)
	protected.GET("/deleted", h.ListDeletedUsersHandler)
	protected.PUT("/:id/change-password", h.ChangePasswordHandler)
	protected.POST("/:id/reset-password", h.ResetPasswordHandler)
	protected.GET("/:id/settings", h.GetSettingsHandler)
	protected.PUT("/:id/settings", h.UpdateSettingsHandler)
	protected.PUT("/:id/photo", h.UploadPhoto)
	protected.DELETE("/:id/photo", h.DeletePhoto)
	// ─── End Protected ─────────────────────────────────────────
}
