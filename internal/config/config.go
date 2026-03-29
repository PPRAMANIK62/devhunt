package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	JWTExpiryMinutes int
	ServerPort       string
	Env              string
	ResendAPIKey     string
	AppBaseURL       string
}

func Load() (*Config, error) {
	// Loads .env into os environment. Won't override vars already set,
	// so real environment variables always win (important in production).
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     os.Getenv("REDIS_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		ServerPort:   os.Getenv("SERVER_PORT"),
		Env:          os.Getenv("ENV"),
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		AppBaseURL:   os.Getenv("APP_BASE_URL"),
	}

	// Fail immediately if critical values are missing.
	// Better to crash at startup than silently misbehave later.
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Defaults for optional values
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	if cfg.AppBaseURL == "" {
		cfg.AppBaseURL = "http://localhost:5173"
	}

	mins, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_MINUTES"))
	if err != nil {
		cfg.JWTExpiryMinutes = 10
	} else {
		cfg.JWTExpiryMinutes = mins
	}

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
