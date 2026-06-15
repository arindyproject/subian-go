package errors

import (
	"errors"
	"log/slog"
	"net/http"

	"subian_go/internal/shared/response"

	"github.com/labstack/echo/v5"
)

// ─── App Error ─────────────────────────────────────────────────────────────────

type AppError struct {
	Code    int
	Message string
	Err     error // original error (opsional, untuk logging)
}

func (e *AppError) Error() string   { return e.Message }
func (e *AppError) Unwrap() error   { return e.Err }
func (e *AppError) StatusCode() int { return e.Code }

// ─── Constructor Shortcuts ─────────────────────────────────────────────────────

func BadRequest(message string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: message}
}

func NotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: message}
}

func UnprocessableEntity(message string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: message}
}

func TooManyRequests(message string) *AppError {
	return &AppError{Code: http.StatusTooManyRequests, Message: message}
}

func Internal(message string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: message}
}

func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// ─── Global Error Handler ──────────────────────────────────────────────────────

func Handler(c *echo.Context, err error) {
	// Cek apakah response sudah dikirim
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	// Ambil HTTP status code dari error chain
	code := http.StatusInternalServerError
	var sc echo.HTTPStatusCoder
	if errors.As(err, &sc) {
		if tmp := sc.StatusCode(); tmp != 0 {
			code = tmp
		}
	}

	logger := c.Logger()
	req := c.Request()
	reqID := req.Header.Get(echo.HeaderXRequestID)
	if reqID == "" {
		reqID = "no-request-id"
	}

	// PERBAIKAN 1: Ambil context dari request.
	// JANGAN pernah melewatkan nil ke fungsi yang butuh context.Context
	ctx := req.Context()

	// Fokus logging pada Server Error (5xx)
	if code >= 500 {
		var appErr *AppError
		if errors.As(err, &appErr) && appErr.Err != nil {
			// PERBAIKAN 2: Gunakan ErrorContext dan lewati ctx.
			// Langsung gunakan variadic args (slog.String) tanpa membungkusnya
			// di dalam []any{...} untuk menghindari warning redundant type.
			logger.ErrorContext(ctx, "INTERNAL_SERVER_ERROR",
				slog.String("request_id", reqID),
				slog.String("url", req.URL.Path),
				slog.String("method", req.Method),
				slog.String("public_message", appErr.Message),
				slog.String("internal_error", appErr.Err.Error()),
			)
		} else {
			logger.ErrorContext(ctx, "UNHANDLED_ERROR",
				slog.String("request_id", reqID),
				slog.String("url", req.URL.Path),
				slog.String("method", req.Method),
				slog.String("error", err.Error()),
			)
		}
	}

	message := resolveMessage(err, code)

	if req.Method == http.MethodHead {
		if cErr := c.NoContent(code); cErr != nil {
			// Gunakan slog.Any, pastikan tidak menulis any(errors.Join(...))
			logger.ErrorContext(ctx, "failed to send no content",
				slog.Any("error", errors.Join(err, cErr)),
			)
		}
		return
	}

	// Kirim JSON response
	// Catatan: Jika response.Response Anda memiliki parameter slice/map di belakang,
	// pastikan Anda tidak menulis tipe eksplisit di dalamnya.
	// Contoh SALAH: response.Response(c, code, false, message, nil, []string{string("err")})
	// Contoh BENAR: response.Response(c, code, false, message, nil, []string{"err"})
	if cErr := response.Response(c, code, false, message, nil, nil); cErr != nil {
		logger.ErrorContext(ctx, "failed to send error response",
			slog.Any("error", errors.Join(err, cErr)),
		)
	}
}

// ─── Helper ────────────────────────────────────────────────────────────────────

func resolveMessage(err error, code int) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.Message != "" {
			return httpErr.Message
		}
	}

	return defaultMessage(code)
}

func defaultMessage(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "Request tidak valid"
	case http.StatusUnauthorized:
		return "Autentikasi diperlukan"
	case http.StatusForbidden:
		return "Akses ditolak"
	case http.StatusNotFound:
		return "Endpoint tidak ditemukan (404)"
	case http.StatusMethodNotAllowed:
		return "Method tidak diizinkan"
	case http.StatusRequestEntityTooLarge:
		return "Ukuran request terlalu besar"
	case http.StatusUnsupportedMediaType:
		return "Tipe konten tidak didukung"
	case http.StatusUnprocessableEntity:
		return "Validasi gagal"
	case http.StatusTooManyRequests:
		return "Terlalu banyak permintaan, coba lagi nanti"
	case http.StatusInternalServerError:
		return "Terjadi kesalahan sistem"
	case http.StatusServiceUnavailable:
		return "Layanan tidak tersedia"
	default:
		return "Terjadi kesalahan"
	}
}
