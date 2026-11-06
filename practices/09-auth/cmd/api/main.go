package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/http/handlers"
	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/platform/config"
	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/repo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := repo.Open(cfg.DBDSN)
	if err != nil {
		log.Fatal("db connect: ", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	users := repo.NewUserRepo(db)
	if err := users.AutoMigrate(); err != nil {
		log.Fatal("migrate: ", err)
	}
	log.Print("AutoMigrate: users ready")

	auth := handlers.NewAuthHandler(users, cfg.BcryptCost)
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Post("/auth/register", auth.Register)
	r.Post("/auth/login", auth.Login)

	srv := &http.Server{Addr: cfg.Addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s (bcrypt cost %d)", cfg.Addr, cfg.BcryptCost)
	log.Fatal(srv.ListenAndServe())
}
