package repo

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/core"
)

type userRecord struct {
	core.User
	hash []byte
}

// UserMem — фейковая «база» пользователей с bcrypt-хэшами паролей.
type UserMem struct {
	byEmail   map[string]userRecord
	byID      map[int64]userRecord
	dummyHash []byte
}

// NewUserMem создаёт двух пользователей с паролем "secret123":
// admin@example.com (admin, id 1) и user@example.com (user, id 2).
func NewUserMem(cost int) *UserMem {
	hash := func(s string) []byte {
		h, err := bcrypt.GenerateFromPassword([]byte(s), cost)
		if err != nil {
			panic(err)
		}
		return h
	}
	m := &UserMem{byEmail: map[string]userRecord{}, byID: map[int64]userRecord{}, dummyHash: hash("dummy")}
	for _, u := range []userRecord{
		{User: core.User{ID: 1, Email: "admin@example.com", Role: core.RoleAdmin}, hash: hash("secret123")},
		{User: core.User{ID: 2, Email: "user@example.com", Role: core.RoleUser}, hash: hash("secret123")},
	} {
		m.byEmail[u.Email] = u
		m.byID[u.ID] = u
	}
	return m
}

func (r *UserMem) ByID(id int64) (core.User, error) {
	u, ok := r.byID[id]
	if !ok {
		return core.User{}, core.ErrNotFound
	}
	return u.User, nil
}

// CheckPassword проверяет пару email+пароль. Для неизвестного email всё равно
// выполняется сравнение с фиктивным хэшем, чтобы время ответа не выдавало существование пользователя.
func (r *UserMem) CheckPassword(email, pass string) (core.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		_ = bcrypt.CompareHashAndPassword(r.dummyHash, []byte(pass))
		return core.User{}, core.ErrNotFound
	}
	if bcrypt.CompareHashAndPassword(u.hash, []byte(pass)) != nil {
		return core.User{}, core.ErrBadCreds
	}
	return u.User, nil
}
