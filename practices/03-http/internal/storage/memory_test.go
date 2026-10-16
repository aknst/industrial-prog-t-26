package storage

import (
	"sync"
	"testing"
)

func TestConcurrentCreate(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); s.Create("task") }()
	}
	wg.Wait()
	items := s.List()
	if len(items) != 50 {
		t.Fatal(len(items))
	}
	for i, v := range items {
		if v.ID != int64(i+1) {
			t.Fatal(v)
		}
	}
}
