package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	// Пакет docs создаёт `swag init`; импорт регистрирует спецификацию для Swagger UI.
	_ "github.com/aknst/industrial-prog-t-26/practices/12-swagger/docs"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/http/handlers"
)

// redocPage — страница ReDoc поверх той же спецификации, что и Swagger UI.
const redocPage = `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <title>Notes API — ReDoc</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
  <redoc spec-url="/docs/doc.json"></redoc>
  <script src="https://cdn.jsdelivr.net/npm/redoc@2.1.5/bundles/redoc.standalone.js"></script>
</body>
</html>`

func NewRouter(h *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Документация: Swagger UI на /docs/index.html, ReDoc на /redoc.
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusFound)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)
	r.Get("/redoc", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(redocPage))
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
