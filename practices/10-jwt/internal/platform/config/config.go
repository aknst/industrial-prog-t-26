package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	JWTSecret     []byte
	AccessTTL     time.Duration // JWT_TTL, по умолчанию 15m
	RefreshTTL    time.Duration // REFRESH_TTL, по умолчанию 168h (7 дней)
	LoginAttempts int           // LOGIN_RATE_LIMIT: попыток входа с одного IP за окно
	LoginWindow   time.Duration // LOGIN_RATE_WINDOW, по умолчанию 5m
}

func Load() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	cfg := Config{Port: ":" + getenv("APP_PORT", "8080"), JWTSecret: []byte(secret)}

	var err error
	if cfg.AccessTTL, err = duration("JWT_TTL", "15m"); err != nil {
		return Config{}, err
	}
	if cfg.RefreshTTL, err = duration("REFRESH_TTL", "168h"); err != nil {
		return Config{}, err
	}
	if cfg.LoginWindow, err = duration("LOGIN_RATE_WINDOW", "5m"); err != nil {
		return Config{}, err
	}
	cfg.LoginAttempts, err = strconv.Atoi(getenv("LOGIN_RATE_LIMIT", "5"))
	if err != nil || cfg.LoginAttempts < 1 {
		return Config{}, fmt.Errorf("LOGIN_RATE_LIMIT must be a positive integer")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func duration(key, def string) (time.Duration, error) {
	d, err := time.ParseDuration(getenv(key, def))
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration like 15m or 2h", key)
	}
	return d, nil
}
