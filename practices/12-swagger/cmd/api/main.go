// Package main Notes API server.
//
//	@title			Notes API
//	@version		1.0
//	@description	Учебный REST API для заметок (CRUD). Данные хранятся в памяти процесса.
//	@contact.name	Backend Course
//	@contact.email	example@university.ru
//	@host			localhost:8080
//	@BasePath		/api/v1
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/core/service"
	httpx "github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/http"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/http/handlers"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/repo"
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
	log.Printf("Server started at %s, Swagger UI: http://%s/docs/index.html", srv.Addr, srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
