package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	RedisAddr             string
	IdempotencyHashSecret string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	idempotencyHashSecret := os.Getenv("IDEMPOTENCY_HASH_SECRET")
	if idempotencyHashSecret == "" {
		return nil, fmt.Errorf("IDEMPOTENCY_HASH_SECRET is required")
	}

	return &Config{
		Port:                  getEnvOrDefault("PORT", "3000"),
		DatabaseURL:           dbURL,
		RedisAddr:             getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		IdempotencyHashSecret: idempotencyHashSecret,
	}, nil
}

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	return fallback
}
