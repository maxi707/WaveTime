package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	AppSecret     string
	TokenTTL      time.Duration
	ReserveTTLMin int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AppSecret:     os.Getenv("APP_SECRET"),
		TokenTTL:      time.Hour,
		ReserveTTLMin: 15,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AppSecret == "" {
		return Config{}, fmt.Errorf("APP_SECRET is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
