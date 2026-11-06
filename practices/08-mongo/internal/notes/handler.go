package notes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ store Store }

func NewHandler(s Store) *Handler { return &Handler{store: s} }

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/stats", h.stats)
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.patch)
	r.Delete("/{id}", h.del)
	return r
}

func reqCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// fail переводит ошибку репозитория в HTTP-ответ.
func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeErr(w, http.StatusNotFound, "not_found")
	case errors.Is(err, ErrDuplicate):
		writeErr(w, http.StatusConflict, "title_already_exists")
	case errors.Is(err, context.DeadlineExceeded):
		writeErr(w, http.StatusGatewayTimeout, "database_timeout")
	default:
		writeErr(w, http.StatusInternalServerError, "internal_error")
	}
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		writeErr(w, http.StatusBadRequest, "title_required")
		return
	}
	ctx, cancel := reqCtx(r)
	defer cancel()
	n, err := h.store.Create(ctx, in.Title, in.Content)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := reqCtx(r)
	defer cancel()
	n, err := h.store.ByID(ctx, chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.ParseInt(q.Get("limit"), 10, 64)
	skip, _ := strconv.ParseInt(q.Get("skip"), 10, 64)
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if skip < 0 {
		skip = 0
	}
	ctx, cancel := reqCtx(r)
	defer cancel()
	items, err := h.store.List(ctx, ListParams{Q: q.Get("q"), Limit: limit, Skip: skip, After: q.Get("after")})
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Title == nil && in.Content == nil {
		writeErr(w, http.StatusBadRequest, "nothing_to_update")
		return
	}
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			writeErr(w, http.StatusBadRequest, "title_required")
			return
		}
		in.Title = &t
	}
	ctx, cancel := reqCtx(r)
	defer cancel()
	n, err := h.store.Update(ctx, chi.URLParam(r, "id"), in.Title, in.Content)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) del(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := reqCtx(r)
	defer cancel()
	if err := h.store.Delete(ctx, chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := reqCtx(r)
	defer cancel()
	s, err := h.store.Stats(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}
