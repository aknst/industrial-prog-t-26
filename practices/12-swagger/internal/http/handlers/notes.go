package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/core/service"
)

type Handlers struct {
	Svc *service.NoteService
}

// ErrorResponse — единый формат ошибки API.
type ErrorResponse struct {
	Error string `json:"error" example:"note not found"`
}

// ListNotes godoc
//
//	@Summary		Список заметок
//	@Description	Возвращает страницу заметок, отсортированных по id, с фильтром по подстроке заголовка
//	@Tags			notes
//	@Produce		json
//	@Param			page	query		int		false	"Номер страницы (с 1)"			default(1)	minimum(1)
//	@Param			limit	query		int		false	"Размер страницы (1-100)"		default(20)	minimum(1)	maximum(100)
//	@Param			q		query		string	false	"Поиск по подстроке title без учёта регистра"
//	@Success		200		{array}		core.Note
//	@Header			200		{integer}	X-Total-Count	"Общее количество заметок, подходящих под фильтр"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/notes [get]
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

// CreateNote godoc
//
//	@Summary		Создать заметку
//	@Description	Создаёт заметку; title обязателен (до 200 символов), content необязателен (до 10000)
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Param			input	body		core.NoteCreate	true	"Данные новой заметки"
//	@Success		201		{object}	core.Note
//	@Header			201		{string}	Location	"URL созданной заметки"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/notes [post]
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

// GetNote godoc
//
//	@Summary		Получить заметку
//	@Description	Возвращает заметку по идентификатору
//	@Tags			notes
//	@Produce		json
//	@Param			id	path		int	true	"Идентификатор заметки"	minimum(1)
//	@Success		200	{object}	core.Note
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/notes/{id} [get]
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

// PatchNote godoc
//
//	@Summary		Обновить заметку (частично)
//	@Description	Меняет только переданные поля; нужно передать хотя бы одно из title, content
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Идентификатор заметки"	minimum(1)
//	@Param			input	body		core.NoteUpdate	true	"Поля для обновления"
//	@Success		200		{object}	core.Note
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/notes/{id} [patch]
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

// DeleteNote godoc
//
//	@Summary		Удалить заметку
//	@Tags			notes
//	@Param			id	path	int	true	"Идентификатор заметки"	minimum(1)
//	@Success		204	"No Content"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/notes/{id} [delete]
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
