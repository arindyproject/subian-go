package logger

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	appErrors "subian_go/internal/shared/errors"

	"github.com/labstack/echo/v5"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// InitLogger menginisialisasi logger terstruktur (JSON) ke Console dan File
func InitLogger(logLevel string) *slog.Logger {
	// 1. Pastikan folder logs ada
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic("gagal membuat folder logs: " + err.Error())
	}

	// 2. Konfigurasi rotasi file berdasarkan WAKTU (Harian)
	// Menggunakan file-rotatelogs karena lumberjack hanya mendukung rotasi berdasarkan ukuran.
	writer, err := rotatelogs.New(
		"logs/app_error_%Y-%m-%d.log",                 // Pattern nama file: app_error_2026-06-03.log
		rotatelogs.WithLinkName("logs/app_error.log"), // Membuat symlink 'app_error.log' yang selalu menunjuk ke file hari ini
		rotatelogs.WithMaxAge(30*24*time.Hour),        // Hapus file log yang lebih tua dari 30 hari
		rotatelogs.WithRotationTime(24*time.Hour),     // Lakukan rotasi (ganti file baru) setiap 24 jam
	)
	if err != nil {
		panic("gagal menginisialisasi rotatelogs: " + err.Error())
	}

	// 3. Tulis ke Console (stdout) DAN File secara bersamaan
	multiWriter := io.MultiWriter(os.Stdout, writer)

	// 4. Parse string level dari config menjadi slog.Level
	level := parseLogLevel(logLevel)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	return slog.New(slog.NewJSONHandler(multiWriter, opts))
}

// parseLogLevel mengubah string level (dari .env) menjadi slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// RequestLoggerMiddleware mencatat SEMUA request (sukses/gagal) + durasi
func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			// Proses request
			err := next(c)

			req := c.Request()
			duration := time.Since(start).Milliseconds()
			reqID := req.Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = "no-request-id"
			}

			// ─── PERBAIKAN ECHO V5 ───────────────────────────────────────────
			status := 0
			if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
				status = resp.Status
			}

			if status == 0 && err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else if ae, ok := err.(*appErrors.AppError); ok {
					status = ae.Code
				} else {
					status = http.StatusInternalServerError
				}
			}

			if status == 0 {
				status = http.StatusOK
			}
			// ─────────────────────────────────────────────────────────────────

			ctx := req.Context()

			logAttrs := []any{
				slog.String("request_id", reqID),
				slog.String("method", req.Method),
				slog.String("url", req.URL.Path),
				slog.Int("status", status),
				slog.Int64("duration_ms", duration),
				slog.String("ip", c.RealIP()),
			}

			if err != nil {
				logAttrs = append(logAttrs, slog.String("error", err.Error()))
			}

			if status >= 500 {
				c.Logger().ErrorContext(ctx, "REQUEST_FAILED", logAttrs...)
			} else if status >= 400 {
				c.Logger().WarnContext(ctx, "CLIENT_ERROR", logAttrs...)
			} else {
				c.Logger().InfoContext(ctx, "REQUEST_SUCCESS", logAttrs...)
			}

			return err
		}
	}
}
