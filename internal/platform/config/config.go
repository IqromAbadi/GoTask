package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	AppEnv  string
	AppPort int

	DatabaseURL string

	JwtAccessSecret  string
	JwtRefreshSecret string
	JwtAccessTTL     time.Duration
	JwtRefreshTTL    time.Duration

	CorsAllowedOrigins string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	port, err := getEnvInt("APP_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: port,

		DatabaseURL: getEnv("DATABASE_URL", "postgres://gotask:gotask@localhost:5432/gotask?sslmode=disable"),

		JwtAccessSecret:  requireEnv("JWT_ACCESS_SECRET"),
		JwtRefreshSecret: requireEnv("JWT_REFRESH_SECRET"),
		JwtAccessTTL:     accessTTL,
		JwtRefreshTTL:    refreshTTL,

		CorsAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return val
}

func getEnvInt(key string, fallback int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return n, nil
}
