package repo

import (
	"sync"
	"time"
)

// BlacklistMem — in-memory список отозванных refresh-токенов: jti -> время истечения токена.
// После exp запись больше не нужна (токен всё равно не пройдёт проверку срока), поэтому удаляется.
type BlacklistMem struct {
	mu      sync.Mutex
	revoked map[string]time.Time
	now     func() time.Time
}

func NewBlacklistMem() *BlacklistMem {
	return &BlacklistMem{revoked: map[string]time.Time{}, now: time.Now}
}

// Revoke добавляет токен в список; false, если он там уже есть.
func (b *BlacklistMem) Revoke(jti string, exp time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()
	for id, e := range b.revoked {
		if e.Before(now) {
			delete(b.revoked, id)
		}
	}
	if _, ok := b.revoked[jti]; ok {
		return false
	}
	b.revoked[jti] = exp
	return true
}

func (b *BlacklistMem) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.revoked)
}
