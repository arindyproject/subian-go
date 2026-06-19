package httputil

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

func ParseID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func ParsePagination(c *echo.Context) (page, pageSize int) {
	page, pageSize = 1, 10
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
	return
}

func ParseOptionalInt64Query(c *echo.Context, key string) *int64 {
	raw := c.QueryParam(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}
