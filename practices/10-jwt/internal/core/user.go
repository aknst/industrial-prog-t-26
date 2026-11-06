package core

import "errors"

// User — доменная модель. Хэш пароля хранится только в репозитории.
type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"` // "user" или "admin"
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

var (
	ErrNotFound = errors.New("user not found")
	ErrBadCreds = errors.New("bad credentials")
)
