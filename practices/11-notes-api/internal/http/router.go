package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/http/handlers"
)

func NewRouter(h *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Route("/api/v1/notes", func(r chi.Router) {
		r.Get("/", h.ListNotes)
		r.Post("/", h.CreateNote)
		r.Get("/{id}", h.GetNote)
		r.Patch("/{id}", h.PatchNote)
		r.Delete("/{id}", h.DeleteNote)
	})
	return r
}
