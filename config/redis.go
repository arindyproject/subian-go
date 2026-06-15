package config

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis membuat koneksi ke Redis menggunakan config
func (c *Config) ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort),
		Password: c.RedisPassword,
		DB:       c.RedisDB,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("gagal koneksi ke Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully")
	return client, nil
}

// CloseRedis menutup koneksi Redis
func CloseRedis(client *redis.Client) {
	if client != nil {
		if err := client.Close(); err != nil {
			log.Printf("Warning: gagal menutup koneksi Redis: %v", err)
		}
	}
}
