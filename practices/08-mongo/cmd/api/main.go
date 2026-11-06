package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aknst/industrial-prog-t-26/practices/08-mongo/internal/db"
	"github.com/aknst/industrial-prog-t-26/practices/08-mongo/internal/notes"
)

func main() {
	uri := getenv("MONGO_URI", "mongodb://root:secret@localhost:27017/?authSource=admin")
	dbName := getenv("MONGO_DB", "pz8")
	addr := getenv("HTTP_ADDR", "127.0.0.1:8080")

	ctx := context.Background()
	deps, err := db.ConnectMongo(ctx, uri, dbName)
	if err != nil {
		log.Fatal("mongo connect: ", err)
	}
	defer deps.Client.Disconnect(ctx)

	idxCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	repo, err := notes.NewRepo(idxCtx, deps.Database)
	cancel()
	if err != nil {
		log.Fatal("notes repo: ", err)
	}

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Mount("/api/v1/notes", notes.NewHandler(repo).Routes())

	srv := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	log.Println("listening on", addr)
	log.Fatal(srv.ListenAndServe())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
