package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/core/service"
)

type Handlers struct {
	Svc *service.NoteService
}

func (h *Handlers) ListNotes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, ok := intParam(w, q.Get("page"), "page", 1)
	if !ok {
		return
	}
	limit, ok := intParam(w, q.Get("limit"), "limit", service.DefaultLimit)
	if !ok {
		return
	}
	items, total, err := h.Svc.List(core.ListParams{Page: page, Limit: limit, Q: q.Get("q")})
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	writeJSON(w, http.StatusOK, items)
}

func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	var in core.NoteCreate
	if !decode(w, r, &in) {
		return
	}
	n, err := h.Svc.Create(in)
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/notes/"+strconv.FormatInt(n.ID, 10))
	writeJSON(w, http.StatusCreated, n)
}

func (h *Handlers) GetNote(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	n, err := h.Svc.Get(id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handlers) PatchNote(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var in core.NoteUpdate
	if !decode(w, r, &in) {
		return
	}
	n, err := h.Svc.Update(id, in)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handlers) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	if err := h.Svc.Delete(id); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func idParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

// intParam разбирает необязательный целочисленный query-параметр (пустой -> def).
func intParam(w http.ResponseWriter, raw, name string, def int) (int, bool) {
	if raw == "" {
		return def, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		writeErr(w, http.StatusBadRequest, name+" must be an integer")
		return 0, false
	}
	return n, true
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid input")
		return false
	}
	return true
}

func fail(w http.ResponseWriter, err error) {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeErr(w, http.StatusBadRequest, ve.Error())
	case errors.Is(err, service.ErrNotFound):
		writeErr(w, http.StatusNotFound, "note not found")
	default:
		writeErr(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false) // «>=» в сообщениях об ошибках не должно превращаться в >
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
