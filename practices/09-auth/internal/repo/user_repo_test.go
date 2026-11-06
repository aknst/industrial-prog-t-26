package repo

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/core"
)

// Тест работает с реальным PostgreSQL: нужна отдельная тестовая база в TEST_DB_DSN.
func TestUserRepoAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN is not set")
	}
	db, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrator().DropTable(&core.User{}); err != nil {
		t.Fatal(err)
	}
	r := NewUserRepo(db)
	if err := r.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	u := core.User{Email: "a@example.com", PasswordHash: "hash"}
	if err := r.Create(ctx, &u); err != nil || u.ID == 0 {
		t.Fatal(err, u.ID)
	}
	if err := r.Create(ctx, &core.User{Email: "a@example.com", PasswordHash: "h2"}); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("want ErrEmailTaken, got %v", err)
	}
	got, err := r.ByEmail(ctx, "a@example.com")
	if err != nil || got.ID != u.ID {
		t.Fatal(got, err)
	}
	if _, err := r.ByEmail(ctx, "missing@example.com"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}
