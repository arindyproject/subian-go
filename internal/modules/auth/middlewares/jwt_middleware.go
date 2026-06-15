package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"subian_go/internal/shared/response"
	"subian_go/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// ─── Context Keys ──────────────────────────────────────────────────────────────

const (
	CtxUserID       = "userID"
	CtxUsername     = "username"
	CtxIsStaff      = "isStaff"
	CtxIsSuperadmin = "isSuperadmin" // diisi realtime dari DB, bukan dari JWT
)

// ─── JWT Middleware ────────────────────────────────────────────────────────────

// JWTMiddleware memvalidasi access token dan mengambil isSuperadmin realtime dari DB
func JWTMiddleware(jwtManager *utils.JWTManager, db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Response(c, http.StatusUnauthorized, false, "Token tidak ditemukan.", nil, nil)
			}

			tokenStr := authHeader
			if strings.Contains(authHeader, " ") {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					return response.Response(c, http.StatusUnauthorized, false, "Format token tidak valid.", nil, nil)
				}
				tokenStr = parts[1]
			}

			claims, err := jwtManager.ParseToken(tokenStr)
			if err != nil {
				return response.Response(c, http.StatusUnauthorized, false, "Token tidak valid atau sudah kedaluwarsa.", nil, nil)
			}

			if claims.TokenType != "access" {
				return response.Response(c, http.StatusUnauthorized, false, "Tipe token tidak valid.", nil, nil)
			}

			// Set dari JWT claims dulu
			c.Set(CtxUserID, claims.UserID)
			c.Set(CtxUsername, claims.Username)
			c.Set(CtxIsStaff, claims.IsStaff)

			// ─── Realtime check isSuperadmin dari DB ───────────────
			// Query hanya kolom is_superadmin — ringan dan tidak load seluruh model
			var isSuperadmin bool
			db.Raw("SELECT is_superadmin FROM users WHERE id = ? AND deleted_at IS NULL", claims.UserID).
				Scan(&isSuperadmin)

			c.Set(CtxIsSuperadmin, isSuperadmin)

			return next(c)
		}
	}
}

// ─── Context Helpers ───────────────────────────────────────────────────────────

// IsSuperadmin mengambil status superadmin dari context (sudah realtime dari DB)
func IsSuperadmin(c *echo.Context) bool {
	v, _ := c.Get(CtxIsSuperadmin).(bool)
	return v
}

// GetUserID mengambil userID dari context
func GetUserID(c *echo.Context) (int64, bool) {
	id, ok := c.Get(CtxUserID).(int64)
	return id, ok
}

// GetUsername mengambil username dari context
func GetUsername(c *echo.Context) (string, bool) {
	username, ok := c.Get(CtxUsername).(string)
	return username, ok
}

// IsStaff mengambil status staff dari context
func IsStaff(c *echo.Context) bool {
	v, _ := c.Get(CtxIsStaff).(bool)
	return v
}

// ─── Alias untuk backward compatibility ────────────────────────────────────────

// GetUserIDFromContext alias dari GetUserID
func GetUserIDFromContext(c *echo.Context) (int64, bool) {
	return GetUserID(c)
}

// GetTargetUserID mengambil :id dari path param
func GetTargetUserID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// IsSelf mengembalikan true jika actor mengakses datanya sendiri
func IsSelf(c *echo.Context) bool {
	currentUserID, ok := GetUserID(c)
	if !ok {
		return false
	}
	targetUserID, err := GetTargetUserID(c)
	if err != nil {
		return false
	}
	return currentUserID == targetUserID
}

// ─── Authorization Middlewares ─────────────────────────────────────────────────

func RequireSuperuser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !IsSuperadmin(c) {
				return response.Response(c, http.StatusForbidden, false, "Akses ditolak. Hanya superadmin.", nil, nil)
			}
			return next(c)
		}
	}
}

func RequireStaff() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !IsStaff(c) && !IsSuperadmin(c) {
				return response.Response(c, http.StatusForbidden, false, "Akses ditolak. Hanya staff.", nil, nil)
			}
			return next(c)
		}
	}
}
