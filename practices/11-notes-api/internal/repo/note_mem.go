package repo

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/core"
)

var ErrNotFound = errors.New("note not found")

// NoteRepoMem — потокобезопасное хранилище заметок в памяти.
type NoteRepoMem struct {
	mu    sync.RWMutex
	notes map[int64]*core.Note
	next  int64
}

func NewNoteRepoMem() *NoteRepoMem {
	return &NoteRepoMem{notes: make(map[int64]*core.Note)}
}

func (r *NoteRepoMem) Create(n core.Note) (core.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next++
	n.ID = r.next
	n.UpdatedAt = nil
	r.notes[n.ID] = &n
	return n, nil
}

func (r *NoteRepoMem) Get(id int64) (core.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.notes[id]
	if !ok {
		return core.Note{}, ErrNotFound
	}
	return *n, nil
}

// List возвращает страницу заметок, отсортированных по id, и общее число подходящих заметок.
func (r *NoteRepoMem) List(p core.ListParams) ([]core.Note, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := strings.ToLower(p.Q)
	all := make([]core.Note, 0, len(r.notes))
	for _, n := range r.notes {
		if q == "" || strings.Contains(strings.ToLower(n.Title), q) {
			all = append(all, *n)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	total := len(all)
	start := (p.Page - 1) * p.Limit
	if start >= total {
		return []core.Note{}, total, nil
	}
	end := min(start+p.Limit, total)
	return all[start:end], total, nil
}

// Update применяет apply к заметке под блокировкой и проставляет updatedAt.
func (r *NoteRepoMem) Update(id int64, apply func(*core.Note)) (core.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notes[id]
	if !ok {
		return core.Note{}, ErrNotFound
	}
	apply(n)
	now := time.Now().UTC()
	n.UpdatedAt = &now
	return *n, nil
}

func (r *NoteRepoMem) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.notes[id]; !ok {
		return ErrNotFound
	}
	delete(r.notes, id)
	return nil
}
