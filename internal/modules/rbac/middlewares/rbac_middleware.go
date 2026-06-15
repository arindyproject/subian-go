package middlewares

import (
	"net/http"
	"strconv"

	"subian_go/internal/modules/rbac/contracts"
	"subian_go/internal/shared/response"

	"github.com/labstack/echo/v5"
)

// ─── Middleware ────────────────────────────────────────────────────────────────

// RequirePermission memastikan user memiliki permission tertentu
func RequirePermission(repo contracts.RBACRepository, permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if IsSuperadmin(c) {
				return next(c)
			}
			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}
			has, err := repo.HasPermission(userID, permission)
			if err != nil {
				return response.Response(c, http.StatusInternalServerError, false, "Gagal cek permission", nil, nil)
			}
			if !has {
				return response.Response(c, http.StatusForbidden, false,
					"Akses ditolak. Permission '"+permission+"' diperlukan.", nil, nil)
			}
			return next(c)
		}
	}
}

// RequireAnyPermission memastikan user memiliki minimal satu permission
func RequireAnyPermission(repo contracts.RBACRepository, permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if IsSuperadmin(c) {
				return next(c)
			}
			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}
			userPerms, err := repo.GetUserAllPermissions(userID)
			if err != nil {
				return response.Response(c, http.StatusInternalServerError, false, "Gagal cek permission", nil, nil)
			}
			permSet := toSet(userPerms)
			for _, p := range permissions {
				if permSet[p] {
					return next(c)
				}
			}
			return response.Response(c, http.StatusForbidden, false, "Akses ditolak. Permission tidak mencukupi.", nil, nil)
		}
	}
}

// RequireRole memastikan user memiliki role tertentu
func RequireRole(repo contracts.RBACRepository, roleName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if IsSuperadmin(c) {
				return next(c)
			}
			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}
			has, err := HasRole(repo, userID, roleName)
			if err != nil {
				return response.Response(c, http.StatusInternalServerError, false, "Gagal cek role", nil, nil)
			}
			if !has {
				return response.Response(c, http.StatusForbidden, false,
					"Akses ditolak. Role '"+roleName+"' diperlukan.", nil, nil)
			}
			return next(c)
		}
	}
}

// RequireSuperadmin memastikan user adalah superadmin
func RequireSuperadmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// 1. Langsung izinkan jika user adalah superadmin
			if IsSuperadmin(c) {
				return next(c)
			}

			// 2. Cek apakah user sudah terautentikasi (punya userID di context)
			// Ini berguna untuk membedakan response antara yang belum login (401)
			// dan yang sudah login tapi bukan superadmin (403).
			_, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}

			// 3. Jika sudah login tapi bukan superadmin, tolak akses
			return response.Response(c, http.StatusForbidden, false,
				"Akses ditolak. Hak akses superadmin diperlukan.", nil, nil)
		}
	}
}

// RequireSelf memastikan user yang sedang login hanya bisa mengakses data miliknya sendiri.
func RequireSelf() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			_, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}

			if IsSelf(c) {
				return next(c)
			}

			return response.Response(c, http.StatusForbidden, false,
				"Akses ditolak. Anda hanya dapat mengakses data Anda sendiri.", nil, nil)
		}
	}
}

// RequireSelfOrPermission mengizinkan akses jika user adalah pemilik data (self)
// ATAU user memiliki permission tertentu (misal: admin yang mengedit user lain).
func RequireSelfOrPermission(repo contracts.RBACRepository, permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if IsSuperadmin(c) {
				return next(c)
			}

			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}

			// Izinkan jika ini adalah data milik user itu sendiri
			if IsSelf(c) {
				return next(c)
			}

			// Jika bukan data sendiri, cek apakah punya permission
			has, err := repo.HasPermission(userID, permission)
			if err != nil {
				return response.Response(c, http.StatusInternalServerError, false, "Gagal cek permission", nil, nil)
			}
			if has {
				return next(c)
			}

			return response.Response(c, http.StatusForbidden, false,
				"Akses ditolak. Anda hanya dapat mengakses data sendiri atau memerlukan permission '"+permission+"'.", nil, nil)
		}
	}
}

// RequireSelfOrRole mengizinkan akses jika user adalah pemilik data (self)
// ATAU user memiliki role tertentu.
func RequireSelfOrRole(repo contracts.RBACRepository, roleName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if IsSuperadmin(c) {
				return next(c)
			}

			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return response.Response(c, http.StatusUnauthorized, false, "Autentikasi diperlukan", nil, nil)
			}

			// Izinkan jika ini adalah data milik user itu sendiri
			if IsSelf(c) {
				return next(c)
			}

			// Jika bukan data sendiri, cek apakah punya role
			has, err := HasRole(repo, userID, roleName)
			if err != nil {
				return response.Response(c, http.StatusInternalServerError, false, "Gagal cek role", nil, nil)
			}
			if has {
				return next(c)
			}

			return response.Response(c, http.StatusForbidden, false,
				"Akses ditolak. Anda hanya dapat mengakses data sendiri atau memerlukan role '"+roleName+"'.", nil, nil)
		}
	}
}

// ─── Context Helpers ───────────────────────────────────────────────────────────

func IsSuperadmin(c *echo.Context) bool {
	v, _ := c.Get("isSuperadmin").(bool)
	return v
}

func GetUserIDFromContext(c *echo.Context) (int64, bool) {
	id, ok := c.Get("userID").(int64)
	return id, ok
}

func GetTargetUserID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// IsSelf mengembalikan true jika user yang sedang login mengakses datanya sendiri.
// Berguna jika Anda ingin mengizinkan user non-admin mengedit profil mereka sendiri tanpa perlu permission admin.
func IsSelf(c *echo.Context) bool {
	currentUserID, ok := GetUserIDFromContext(c)
	if !ok {
		return false
	}
	targetUserID, err := GetTargetUserID(c)
	if err != nil {
		return false
	}
	return currentUserID == targetUserID
}

// ─── Programmatic Helpers (untuk service/handler) ─────────────────────────────

func HasPermission(repo contracts.RBACRepository, userID int64, permission string) (bool, error) {
	return repo.HasPermission(userID, permission)
}

func HasRole(repo contracts.RBACRepository, userID int64, roleName string) (bool, error) {
	roles, err := repo.GetUserRoles(userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r.Name == roleName {
			return true, nil
		}
	}
	return false, nil
}

func HasAnyRole(repo contracts.RBACRepository, userID int64, roleNames ...string) (bool, error) {
	roles, err := repo.GetUserRoles(userID)
	if err != nil {
		return false, err
	}
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r.Name] = true
	}
	for _, name := range roleNames {
		if roleSet[name] {
			return true, nil
		}
	}
	return false, nil
}

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, v := range items {
		set[v] = true
	}
	return set
}
