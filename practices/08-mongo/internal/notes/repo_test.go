package notes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/08-mongo/internal/db"
)

// newTestRepo подключается к MongoDB из MONGO_URI и создаёт изолированную базу,
// которая удаляется по завершении теста. Без MONGO_URI тест пропускается.
func newTestRepo(t *testing.T) *Repo {
	t.Helper()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI is not set")
	}
	ctx := context.Background()
	name := fmt.Sprintf("pz8_test_%d", time.Now().UnixNano())
	deps, err := db.ConnectMongo(ctx, uri, name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = deps.Database.Drop(ctx)
		_ = deps.Client.Disconnect(ctx)
	})
	r, err := NewRepo(ctx, deps.Database)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestCreateAndGet(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	created, err := r.Create(ctx, "T1", "C1")
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.ByID(ctx, created.ID.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "T1" || got.Content != "C1" {
		t.Fatalf("got %+v", got)
	}
}

func TestUniqueTitle(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	if _, err := r.Create(ctx, "same", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(ctx, "same", ""); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}

func TestPartialUpdateKeepsOtherFields(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	n, _ := r.Create(ctx, "title", "old")
	content := "new"
	got, err := r.Update(ctx, n.ID.Hex(), nil, &content)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "title" || got.Content != "new" || !got.UpdatedAt.After(n.CreatedAt) && !got.UpdatedAt.Equal(n.CreatedAt) {
		t.Fatalf("got %+v", got)
	}
}

func TestListSearchCursorAndStats(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		if _, err := r.Create(ctx, fmt.Sprintf("Note %d", i), "12345678"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := r.Create(ctx, "Other (x)", "1234"); err != nil {
		t.Fatal(err)
	}

	found, err := r.List(ctx, ListParams{Q: "note", Limit: 20})
	if err != nil || len(found) != 5 {
		t.Fatalf("search: %d %v", len(found), err)
	}
	// «(x)» — не регулярное выражение, а обычная подстрока.
	if found, _ = r.List(ctx, ListParams{Q: "(x)", Limit: 20}); len(found) != 1 {
		t.Fatalf("literal search: %d", len(found))
	}

	page1, _ := r.List(ctx, ListParams{Q: "Note", Limit: 2, After: "ffffffffffffffffffffffff"})
	if len(page1) != 2 {
		t.Fatalf("page1: %d", len(page1))
	}
	page2, _ := r.List(ctx, ListParams{Q: "Note", Limit: 2, After: page1[1].ID.Hex()})
	if len(page2) != 2 || page2[0].ID == page1[1].ID {
		t.Fatalf("page2: %+v", page2)
	}

	s, err := r.Stats(ctx)
	if err != nil || s.Count != 6 || s.AvgContentLength < 7.3 || s.AvgContentLength > 7.4 { // (5*8 + 4) / 6
		t.Fatalf("stats: %+v %v", s, err)
	}
}

func TestDeleteAndInvalidID(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	n, _ := r.Create(ctx, "x", "")
	if err := r.Delete(ctx, n.ID.Hex()); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(ctx, n.ID.Hex()); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := r.ByID(ctx, "bad-id"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}
