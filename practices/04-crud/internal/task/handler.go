package task

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Handler struct{ repo *MemoryStore }

func NewHandler(repo *MemoryStore) *Handler { return &Handler{repo: repo} }
func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if code != 204 {
		json.NewEncoder(w).Encode(v)
	}
}
func Fail(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]string{"error": message})
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		Fail(w, 400, "invalid id")
		return 0, false
	}
	return id, true
}

type input struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func decode(w http.ResponseWriter, r *http.Request) (input, bool) {
	var in input
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		Fail(w, 400, "Content-Type must be application/json")
		return in, false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if dec.Decode(&in) != nil {
		Fail(w, 400, "invalid json")
		return in, false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		Fail(w, 400, "expected one JSON object")
		return in, false
	}
	in.Title = strings.TrimSpace(in.Title)
	if n := utf8.RuneCountInString(in.Title); n < 3 || n > 100 {
		Fail(w, 422, "title must contain 3 to 100 characters")
		return in, false
	}
	return in, true
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	in, ok := decode(w, r)
	if !ok {
		return
	}
	v := h.repo.Create(in.Title)
	w.Header().Set("Location", "/api/tasks/"+strconv.FormatInt(v.ID, 10))
	JSON(w, 201, v)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	v, err := h.repo.Get(id)
	if err != nil {
		Fail(w, 404, "task not found")
		return
	}
	JSON(w, 200, v)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := decode(w, r)
	if !ok {
		return
	}
	v, err := h.repo.Update(id, in.Title, in.Done)
	if err != nil {
		Fail(w, 404, "task not found")
		return
	}
	JSON(w, 200, v)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(id); err != nil {
		Fail(w, 404, "task not found")
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items := h.repo.List()
	if raw := r.URL.Query().Get("done"); raw != "" {
		done, err := strconv.ParseBool(raw)
		if err != nil {
			Fail(w, 400, "invalid done filter")
			return
		}
		filtered := make([]Task, 0)
		for _, v := range items {
			if v.Done == done {
				filtered = append(filtered, v)
			}
		}
		items = filtered
	}
	page, limit := 1, 10
	for key, dest := range map[string]*int{"page": &page, "limit": &limit} {
		if raw := r.URL.Query().Get(key); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 || v > 1000000 {
				Fail(w, 400, "invalid pagination")
				return
			}
			*dest = v
		}
	}
	if limit > 100 {
		Fail(w, 400, "limit must not exceed 100")
		return
	}
	start := (page - 1) * limit
	if start >= len(items) {
		items = []Task{}
	} else {
		end := start + limit
		if end > len(items) {
			end = len(items)
		}
		items = items[start:end]
	}
	JSON(w, 200, items)
}
