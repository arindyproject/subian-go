package auth

import (
	"subian_go/internal/modules/auth/handlers"
	authMiddlewares "subian_go/internal/modules/auth/middlewares"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan semua routes untuk module auth
// db dibutuhkan oleh JWTMiddleware untuk cek isSuperadmin secara realtime
func RegisterRoutes(e *echo.Echo, h *handlers.AuthHandler, jwtManager *utils.JWTManager, db *gorm.DB) {
	// ─── Public routes ─────────────────────────────────────────
	public := e.Group("/api/v1/auth")
	public.POST("/login", h.Login)
	public.POST("/register", h.Register)
	public.POST("/refresh", h.RefreshToken)
	public.POST("/forgot-password", h.ForgotPassword)
	public.POST("/reset-password", h.ResetPassword)

	// ─── Protected routes ──────────────────────────────────────
	// JWTMiddleware sekarang terima db untuk query isSuperadmin realtime
	protected := e.Group("/api/v1/auth", authMiddlewares.JWTMiddleware(jwtManager, db))
	protected.POST("/logout", h.Logout)
	protected.POST("/logout-all", h.LogoutAll)
}
