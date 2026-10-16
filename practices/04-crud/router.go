package main

import (
	"github.com/aknst/industrial-prog-t-26/practices/04-crud/internal/task"
	my "github.com/aknst/industrial-prog-t-26/practices/04-crud/pkg/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func routes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.Recoverer, my.Logger, my.SimpleCORS)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { task.JSON(w, 200, map[string]string{"status": "ok"}) })
	h := task.NewHandler(task.NewMemoryStore())
	r.Mount("/api/tasks", h.Routes())
	return r
}
