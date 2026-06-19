package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Helper functions untuk waktu yang lebih ekspresif
// ------------------------------------------------------------
func Days(n int) time.Duration {
	return time.Duration(n) * 24 * time.Hour
}

func Weeks(n int) time.Duration {
	return Days(n * 7)
}

func Months(n int) time.Duration {
	return Days(n * 30)
} //----------------------------------------------------------

// Manager mengelola operasi cache Redis
type Manager struct {
	client     *redis.Client
	enabled    bool
	defaultTTL time.Duration // <--- Default TTL dari config
}

// NewManager membuat instance cache manager baru
func NewManager(client *redis.Client, enabled bool, defaultTTLDays int) *Manager {
	return &Manager{
		client:     client,
		enabled:    enabled,
		defaultTTL: Days(defaultTTLDays), // <--- Konversi hari ke duration
	}
}

// GetDefaultTTL mengembalikan default TTL yang sudah dikonversi ke time.Duration
func (m *Manager) GetDefaultTTL() time.Duration {
	return m.defaultTTL
}

// SetDefault menyimpan data ke cache menggunakan default TTL dari config
func (m *Manager) SetDefault(ctx context.Context, key string, value interface{}) {
	m.Set(ctx, key, value, m.defaultTTL)
}

// Set menyimpan data ke cache dengan expiration time
func (m *Manager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	if !m.enabled || m.client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Cache marshal error key %s: %v", key, err)
		return
	}
	m.client.Set(ctx, key, data, expiration)
}

// Get mengambil data dari cache dan unmarshal ke destination
func (m *Manager) Get(ctx context.Context, key string, dest interface{}) bool {
	if !m.enabled || m.client == nil {
		return false
	}
	val, err := m.client.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false
	}
	return true
}

// InvalidateDetail menghapus cache berdasarkan key spesifik
func (m *Manager) InvalidateDetail(ctx context.Context, key string) {
	if !m.enabled || m.client == nil {
		return
	}
	m.client.Del(ctx, key)
}

// InvalidateList menghapus semua cache berdasarkan prefix (menggunakan SCAN)
func (m *Manager) InvalidateList(ctx context.Context, prefix string) {
	if !m.enabled || m.client == nil {
		return
	}
	var cursor uint64
	for {
		keys, nextCursor, err := m.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			m.client.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

// InvalidateAll menghapus semua cache berdasarkan multiple prefixes
func (m *Manager) InvalidateAll(ctx context.Context, prefixes ...string) {
	for _, prefix := range prefixes {
		m.InvalidateList(ctx, prefix)
	}
}

// IsEnabled mengecek apakah cache aktif
func (m *Manager) IsEnabled() bool {
	return m.enabled && m.client != nil
}
