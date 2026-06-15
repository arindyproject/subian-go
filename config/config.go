package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5/middleware"
)

// Config holds all configuration for the application
type Config struct {
	EnvCode string
	// Server
	BaseURL    string
	ServerPort string
	LogLevel   string
	Env        string
	TimeZone   string

	// Default
	DefaultPageSize int
	DefaultPassword string

	// Database
	DatabaseURL  string
	DatabaseHost string
	DatabasePort string
	DatabaseUser string
	DatabasePass string
	DatabaseName string
	DatabaseSSL  string

	// CORS
	CORS middleware.CORSConfig

	// JWT
	JWTSecret                string
	JWTIssuer                string
	JWTAccessTokenExpMinutes int
	JWTRefreshTokenExpDays   int

	// Login Security
	LoginMaxAttempts             int
	LoginLockDurationMinutes     int
	MaxConcurrentSessions        int
	RateLimitLoginPerIPPerMinute int

	// Password Policy
	PasswordMinLength        int
	PasswordRequireUppercase bool
	PasswordRequireNumber    bool
	PasswordRequireSymbol    bool
	PasswordHistoryCount     int

	// Registration
	IsRegistrationActive bool
	AutoActiveUser       bool

	// Reset Password & Frontend
	MailResetTokenExpMinutes int
	AppFrontendURL           string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	var env string = "DEV" // default to DEV
	var envFile string
	switch env {
	case "DEV":
		envFile = "config/.env.dev"
	case "PROD":
		envFile = "config/.env.prod"
	default:
		log.Printf("Warning: Invalid env '%s', defaulting to DEV", env)
		envFile = "config/.env.dev"
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Error loading %s file: %v", envFile, err)
	}

	return &Config{
		EnvCode: env,
		// ─── Server ────────────────────────────────────────────
		BaseURL:    getEnv("BASE_URL", "http://localhost:1323"),
		ServerPort: getEnv("SERVER_PORT", "1323"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		Env:        getEnv("ENV", "development"),
		TimeZone:   getEnv("TimeZone", "Asia/Jakarta"),

		// ─── Default ─────────────────────────────────────────────
		DefaultPageSize: getEnvAsInt("DEFAULT_PAGE_SIZE", 10),
		DefaultPassword: getEnv("DEFAULT_PASSWORD", "password123"),

		// ─── Database ──────────────────────────────────────────
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		DatabaseHost: getEnv("DB_HOST", "localhost"),
		DatabasePort: getEnv("DB_PORT", "5432"),
		DatabaseUser: getEnv("DB_USER", "postgres"),
		DatabasePass: getEnv("DB_PASS", ""),
		DatabaseName: getEnv("DB_NAME", "subian"),
		DatabaseSSL:  getEnv("DB_SSL_MODE", "disable"),

		// ─── CORS ──────────────────────────────────────────────
		CORS: middleware.CORSConfig{
			AllowOrigins:     strings.Split(getEnv("ALLOW_ORIGINS", "*"), ","),
			AllowMethods:     strings.Split(getEnv("ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
			AllowHeaders:     strings.Split(getEnv("ALLOW_HEADERS", "Content-Type,Authorization"), ","),
			AllowCredentials: getEnvAsBool("ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 0),
		},

		// ─── JWT ───────────────────────────────────────────────
		JWTSecret:                getEnv("JWT_SECRET", ""),
		JWTIssuer:                getEnv("JWT_ISSUER", "subian"),
		JWTAccessTokenExpMinutes: getEnvAsInt("JWT_ACCESS_TOKEN_EXP_MINUTES", 15),
		JWTRefreshTokenExpDays:   getEnvAsInt("JWT_REFRESH_TOKEN_EXP_DAYS", 7),

		// ─── Login Security ────────────────────────────────────
		LoginMaxAttempts:             getEnvAsInt("LOGIN_MAX_ATTEMPTS", 5),
		LoginLockDurationMinutes:     getEnvAsInt("LOGIN_LOCK_DURATION_MINUTES", 15),
		MaxConcurrentSessions:        getEnvAsInt("MAX_CONCURRENT_SESSIONS", 5),
		RateLimitLoginPerIPPerMinute: getEnvAsInt("RATE_LIMIT_LOGIN_PER_IP_PER_MINUTE", 20),

		// ─── Password Policy ───────────────────────────────────
		PasswordMinLength:        getEnvAsInt("PASSWORD_MIN_LENGTH", 6),
		PasswordRequireUppercase: getEnvAsBool("PASSWORD_REQUIRE_UPPERCASE", false),
		PasswordRequireNumber:    getEnvAsBool("PASSWORD_REQUIRE_NUMBER", false),
		PasswordRequireSymbol:    getEnvAsBool("PASSWORD_REQUIRE_SYMBOL", false),
		PasswordHistoryCount:     getEnvAsInt("PASSWORD_HISTORY_COUNT", 3),

		// ─── Registration ──────────────────────────────────────
		IsRegistrationActive: getEnvAsBool("IS_REGISTRATION_ACTIVE", true),
		AutoActiveUser:       getEnvAsBool("AUTO_ACTIVE_USER", true),

		// ─── Reset Password ────────────────────────────────────
		MailResetTokenExpMinutes: getEnvAsInt("MAIL_RESET_TOKEN_EXP_MINUTES", 30),
		AppFrontendURL:           getEnv("APP_FRONTEND_URL", "http://localhost:3000"),

		// ─── Redis ─────────────────────────────────────────────
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// ─── SMTP ──────────────────────────────────────────────
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "subian"),
	}
}

// ─── Env Helpers ───────────────────────────────────────────────────────────────

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
