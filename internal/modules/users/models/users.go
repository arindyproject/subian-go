package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ─── User Model ────────────────────────────────────────────────────────────────

// User represents the users table in database
type User struct {
	ID                int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Photo             *string        `gorm:"column:photo;type:varchar(500)" json:"photo"`
	PhotoThumbnail    *string        `gorm:"column:photo_thumbnail;type:varchar(500)" json:"photo_thumbnail"`
	Username          string         `gorm:"column:username;type:varchar(150);uniqueIndex;not null" json:"username"`
	Email             string         `gorm:"column:email;type:varchar(254);uniqueIndex;not null" json:"email"`
	Name              string         `gorm:"column:name;type:varchar(255);not null;default:''" json:"name"`
	IsSuperadmin      bool           `gorm:"column:is_superadmin;not null;default:false" json:"is_superadmin"`
	IsActive          bool           `gorm:"column:is_active;not null;default:true" json:"is_active"`
	IsStaff           bool           `gorm:"column:is_staff;not null;default:false" json:"is_staff"`
	IsVerified        bool           `gorm:"column:is_verified;not null;default:false" json:"is_verified"`
	Password          string         `gorm:"column:password;type:varchar(255);not null" json:"-"`
	PasswordChangedAt *time.Time     `gorm:"column:password_changed_at;type:timestamptz" json:"password_changed_at"`
	LastLoginAt       *time.Time     `gorm:"column:last_login_at;type:timestamptz" json:"last_login_at"`
	Settings          JSONB          `gorm:"column:settings;type:jsonb;not null;default:'[]'::jsonb" json:"settings"`
	CreatedBy         *int64         `gorm:"column:created_by" json:"created_by"`
	UpdatedBy         *int64         `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt         time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:NOW()" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"`
	DeletedBy         *int64         `gorm:"column:deleted_by" json:"deleted_by"`
	DeleteReason      *string        `gorm:"column:delete_reason;type:varchar(500)" json:"delete_reason"`
}

func (User) TableName() string {
	return "users"
}

// ─── Settings Helper ───────────────────────────────────────────────────────────

// UserSetting adalah struct untuk settings yang disimpan sebagai JSONB
// Bukan tabel — disimpan di kolom users.settings
type UserSetting struct {
	Key         string      `json:"key"`
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
}

func (u *User) GetSettings() ([]UserSetting, error) {
	var settings []UserSetting
	if err := json.Unmarshal(u.Settings, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (u *User) SetSettings(settings []UserSetting) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	u.Settings = data
	return nil
}

func DefaultSettings() []UserSetting {
	return []UserSetting{
		{Key: "is_dark_mode", Type: "boolean", Value: false, Description: "Aktifkan tema gelap"},
	}
}

// ─── User Creator ──────────────────────────────────────────────────────────────
type UserCreator struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// ─── JSONB Type ────────────────────────────────────────────────────────────────

type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = make(JSONB, len(v))
		copy(*j, v)
	case string:
		*j = JSONB(v)
	default:
		return json.Unmarshal([]byte(fmt.Sprintf("%v", value)), j)
	}
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return json.Unmarshal(data, j)
	}
	*j = make(JSONB, len(data))
	copy(*j, data)
	return nil
}
