package repo

import (
	"testing"
	"time"
)

func TestBlacklistRevokeOnceAndCleanup(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	b := NewBlacklistMem()
	b.now = func() time.Time { return now }

	if !b.Revoke("a", now.Add(time.Hour)) {
		t.Fatal("first revoke must succeed")
	}
	if b.Revoke("a", now.Add(time.Hour)) {
		t.Fatal("second revoke of the same jti must report false")
	}
	b.Revoke("b", now.Add(time.Minute))
	if b.Len() != 2 {
		t.Fatal(b.Len())
	}
	now = now.Add(2 * time.Minute) // "b" истёк сам по себе
	b.Revoke("c", now.Add(time.Hour))
	if b.Len() != 2 {
		t.Fatalf("expired entries must be dropped, len = %d", b.Len())
	}
}
