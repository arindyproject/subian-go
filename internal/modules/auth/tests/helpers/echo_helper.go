package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v5"
)

// NewTestEcho membuat instance Echo untuk testing
func NewTestEcho() *echo.Echo {
	e := echo.New()
	return e
}

// NewJSONRequest membuat HTTP request dengan body JSON
func NewJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(b)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req, nil
}

// NewContext membuat echo.Context untuk testing
func NewContext(e *echo.Echo, method, path string, body interface{}) (*echo.Context, *httptest.ResponseRecorder, error) {
	req, err := NewJSONRequest(method, path, body)
	if err != nil {
		return nil, nil, err
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec, nil
}

// NewContextWithJWT membuat echo.Context dengan JWT claims di context
func NewContextWithJWT(e *echo.Echo, method, path string, body interface{}, userID int64, isSuperuser bool) (*echo.Context, *httptest.ResponseRecorder, error) {
	c, rec, err := NewContext(e, method, path, body)
	if err != nil {
		return nil, nil, err
	}

	// Set JWT claims seperti yang dilakukan JWTMiddleware
	c.Set("userID", userID)
	c.Set("isSuperuser", isSuperuser)
	c.Set("username", "testuser")

	return c, rec, nil
}

// ParseResponse mem-parse response JSON ke map
func ParseResponse(rec *httptest.ResponseRecorder) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ParseResponseBody mem-parse response JSON ke struct tertentu
func ParseResponseBody(rec *httptest.ResponseRecorder, v interface{}) error {
	return json.Unmarshal(rec.Body.Bytes(), v)
}
