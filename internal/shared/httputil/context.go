package httputil

import (
	rbacMiddlewares "subian_go/internal/modules/rbac/middlewares"

	"github.com/labstack/echo/v5"
)

type AuthContext struct {
	UserID       int64
	IsSuperadmin bool
}

func BuildAuthContext(c *echo.Context) AuthContext {
	userID, _ := rbacMiddlewares.GetUserIDFromContext(c)
	isSuperadmin := rbacMiddlewares.IsSuperadmin(c)
	return AuthContext{
		UserID:       userID,
		IsSuperadmin: isSuperadmin,
	}
}

func GetActorID(c *echo.Context) *int64 {
	if userID, ok := c.Get("userID").(int64); ok {
		return &userID
	}
	return nil
}
