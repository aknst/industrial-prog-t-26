package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/07-redis/internal/cache"
)

const (
	defaultTTL  = 10 * time.Second
	productTTL  = 30 * time.Second
	maxTTLInSec = 86400
)

type Handlers struct {
	Cache *cache.Cache
	// LoadDelay имитирует медленный запрос к основной БД в /product.
	LoadDelay time.Duration
}

func Router(h *Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/set", h.set)
	mux.HandleFunc("/get", h.get)
	mux.HandleFunc("/ttl", h.ttl)
	mux.HandleFunc("/del", h.del)
	mux.HandleFunc("/product", h.product)
	return mux
}

func (h *Handlers) health(w http.ResponseWriter, r *http.Request) {
	if err := h.Cache.Ping(r.Context()); err != nil {
		http.Error(w, "redis unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (h *Handlers) set(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	if key == "" || value == "" {
		http.Error(w, "key and value required", http.StatusBadRequest)
		return
	}
	ttl := defaultTTL
	if raw := r.URL.Query().Get("ttl"); raw != "" {
		sec, err := strconv.Atoi(raw)
		if err != nil || sec < 1 || sec > maxTTLInSec {
			http.Error(w, "ttl must be a number of seconds from 1 to 86400", http.StatusBadRequest)
			return
		}
		ttl = time.Duration(sec) * time.Second
	}
	if err := h.Cache.Set(r.Context(), key, value, ttl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "OK: %s=%s (TTL %s)", key, value, ttl)
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	val, err := h.Cache.Get(r.Context(), key)
	switch {
	case errors.Is(err, cache.ErrMiss):
		http.Error(w, "key not found or expired", http.StatusNotFound)
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	default:
		fmt.Fprintf(w, "VALUE: %s=%s", key, val)
	}
}

func (h *Handlers) ttl(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	ttl, err := h.Cache.TTL(r.Context(), key)
	switch {
	case errors.Is(err, cache.ErrMiss):
		http.Error(w, "key not found or expired", http.StatusNotFound)
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	case ttl < 0:
		fmt.Fprintf(w, "TTL for %s: no expiration", key)
	default:
		fmt.Fprintf(w, "TTL for %s: %v", key, ttl)
	}
}

func (h *Handlers) del(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	deleted, err := h.Cache.Del(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "DEL %s: %t", key, deleted)
}

// product показывает кэширование: первый запрос «ходит в БД» (LoadDelay),
// повторные в течение productTTL отдаются из Redis.
func (h *Handlers) product(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	started := time.Now()
	val, cached, err := h.Cache.GetOrLoad(r.Context(), "product:"+id, productTTL, func(ctx context.Context) (string, error) {
		select {
		case <-time.After(h.LoadDelay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
		return fmt.Sprintf("Product #%s", id), nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	source := "db"
	if cached {
		source = "cache"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": id, "name": val, "source": source, "tookMs": time.Since(started).Milliseconds(),
	})
}
