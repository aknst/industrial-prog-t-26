package api

import (
	"encoding/json"
	"errors"
	"github.com/aknst/industrial-prog-t-26/practices/03-http/internal/storage"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Handlers struct{ Store *storage.MemoryStore }

func NewRouter() http.Handler {
	h := &Handlers{Store: storage.NewMemoryStore()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /tasks", h.List)
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("GET /tasks/{id}", h.Get)
	return Logging(mux)
}
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	items := h.Store.List()
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if q != "" {
		filtered := make([]storage.Task, 0)
		for _, v := range items {
			if strings.Contains(strings.ToLower(v.Title), q) {
				filtered = append(filtered, v)
			}
		}
		items = filtered
	}
	JSON(w, 200, items)
}
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		Error(w, 400, "Content-Type must be application/json")
		return
	}
	var in struct {
		Title string `json:"title"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		Error(w, 400, "invalid json")
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		Error(w, 400, "expected one JSON object")
		return
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		Error(w, 400, "title is required")
		return
	}
	n := utf8.RuneCountInString(title)
	if n < 3 || n > 140 {
		Error(w, 422, "title must contain 3 to 140 characters")
		return
	}
	v := h.Store.Create(title)
	w.Header().Set("Location", "/tasks/"+strconv.FormatInt(v.ID, 10))
	JSON(w, 201, v)
}
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		Error(w, 400, "invalid id")
		return
	}
	v, err := h.Store.Get(id)
	if errors.Is(err, storage.ErrNotFound) {
		Error(w, 404, "task not found")
		return
	}
	if err != nil {
		Error(w, 500, "storage error")
		return
	}
	JSON(w, 200, v)
}
