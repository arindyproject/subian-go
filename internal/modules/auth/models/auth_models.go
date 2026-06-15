package models

import (
	"time"

	"gorm.io/gorm"
)

// ─── AuthToken (outstanding_tokens) ───────────────────────────────────────────

// AuthToken menyimpan refresh token yang aktif
type AuthToken struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID      int64          `gorm:"column:user_id;not null;index" json:"user_id"`
	JTI         string         `gorm:"column:jti;type:varchar(255);uniqueIndex;not null" json:"jti"`
	TokenType   string         `gorm:"column:token_type;type:varchar(50);not null;default:'refresh'" json:"token_type"`
	DeviceInfo  *string        `gorm:"column:device_info;type:text" json:"device_info"`
	IPAddress   *string        `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	IsBlacklist bool           `gorm:"column:is_blacklist;not null;default:false" json:"is_blacklist"`
	ExpiresAt   time.Time      `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:NOW()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"`
}

func (AuthToken) TableName() string {
	return "auth_tokens"
}

// IsExpired mengecek apakah token sudah expired
func (t *AuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid mengecek apakah token masih valid (tidak blacklist dan belum expired)
func (t *AuthToken) IsValid() bool {
	return !t.IsBlacklist && !t.IsExpired()
}

// ─── LoginHistory (users_login_histories) ─────────────────────────────────────

// LoginHistory menyimpan riwayat login user
type LoginHistory struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID        *int64    `gorm:"column:user_id;index" json:"user_id"`
	Identifier    string    `gorm:"column:identifier;type:varchar(255);not null" json:"identifier"`
	IPAddress     string    `gorm:"column:ip_address;type:varchar(45);not null" json:"ip_address"`
	UserAgent     *string   `gorm:"column:user_agent;type:text" json:"user_agent"`
	Status        string    `gorm:"column:status;type:varchar(20);not null" json:"status"` // success | failed
	FailureReason *string   `gorm:"column:failure_reason;type:varchar(255)" json:"failure_reason"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
}

func (LoginHistory) TableName() string {
	return "users_login_histories"
}

// Status constants
const (
	LoginStatusSuccess = "success"
	LoginStatusFailed  = "failed"
)

// ─── PasswordHistory (users_password_histories) ───────────────────────────────

// PasswordHistory menyimpan riwayat password user
type PasswordHistory struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID       int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
}

func (PasswordHistory) TableName() string {
	return "users_password_histories"
}
