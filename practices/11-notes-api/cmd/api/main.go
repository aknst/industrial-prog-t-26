package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/core/service"
	httpx "github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/http"
	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/http/handlers"
	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/repo"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	notes := repo.NewNoteRepoMem()
	h := &handlers.Handlers{Svc: service.NewNoteService(notes)}

	srv := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           httpx.NewRouter(h),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("Server started at", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
