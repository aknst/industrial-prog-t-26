package task

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type MemoryStore struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]Task
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{items: make(map[int64]Task)} }
func (s *MemoryStore) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	now := time.Now().UTC()
	v := Task{ID: s.seq, Title: title, CreatedAt: now, UpdatedAt: now}
	s.items[v.ID] = v
	return v
}
func (s *MemoryStore) Get(id int64) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return v, nil
}
func (s *MemoryStore) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (s *MemoryStore) Update(id int64, title string, done bool) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	v.Title = title
	v.Done = done
	v.UpdatedAt = time.Now().UTC()
	s.items[id] = v
	return v, nil
}
func (s *MemoryStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
