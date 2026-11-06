package config

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	DBDSN      string
	BcryptCost int
	Addr       string
}

// Load читает DB_DSN, BCRYPT_COST (по умолчанию 12) и APP_ADDR (по умолчанию 127.0.0.1:8080).
func Load() (Config, error) {
	cost := 12
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < bcrypt.MinCost || n > bcrypt.MaxCost {
			return Config{}, fmt.Errorf("BCRYPT_COST must be an integer from %d to %d, got %q", bcrypt.MinCost, bcrypt.MaxCost, v)
		}
		cost = n
	}
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}
	return Config{DBDSN: dsn, BcryptCost: cost, Addr: addr}, nil
}
