package repo

import (
	"sync"
	"testing"

	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/core"
)

func TestConcurrentCreateGivesUniqueIDs(t *testing.T) {
	r := NewNoteRepoMem()
	const n = 100
	ids := make(chan int64, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			note, _ := r.Create(core.Note{Title: "t"})
			ids <- note.ID
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[int64]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Fatal(len(seen))
	}
}

func TestStoredCopyIsIsolatedFromCaller(t *testing.T) {
	r := NewNoteRepoMem()
	n, _ := r.Create(core.Note{Title: "orig"})
	n.Title = "changed outside"
	got, _ := r.Get(n.ID)
	if got.Title != "orig" {
		t.Fatal("repository state was modified through a returned value")
	}
}
