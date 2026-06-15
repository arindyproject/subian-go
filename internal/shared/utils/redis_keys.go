package utils

import (
	"crypto/sha256"
	"fmt"
)

// ─── Redis Key Helpers ─────────────────────────────────────────────────────────
// Semua key Redis untuk auth terpusat di sini

// hash membuat sha256 hash dari string
func hash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// KeyLoginAttempts key untuk counter percobaan login per identifier
// TTL: LOGIN_LOCK_DURATION_MINUTES
func KeyLoginAttempts(identifier string) string {
	return fmt.Sprintf("login_attempts:%s", hash(identifier))
}

// KeyLoginLock key untuk lock akun setelah max attempts
// TTL: LOGIN_LOCK_DURATION_MINUTES
func KeyLoginLock(identifier string) string {
	return fmt.Sprintf("login_lock:%s", hash(identifier))
}

// KeyLoginAttemptsIP key untuk counter percobaan login per IP
// TTL: 1 menit
func KeyLoginAttemptsIP(ip string) string {
	return fmt.Sprintf("login_attempts_ip:%s", hash(ip))
}

// KeyIPBlacklist key untuk blacklist IP sementara
// TTL: LOGIN_LOCK_DURATION_MINUTES
func KeyIPBlacklist(ip string) string {
	return fmt.Sprintf("ip_blacklist:%s", hash(ip))
}

// KeyResetPassword key untuk token reset password
// TTL: MAIL_RESET_TOKEN_EXP_MINUTES
func KeyResetPassword(token string) string {
	return fmt.Sprintf("reset_password:%s", token)
}
