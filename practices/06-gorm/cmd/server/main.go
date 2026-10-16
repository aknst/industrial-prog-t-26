package main

import (
	"github.com/aknst/industrial-prog-t-26/practices/06-gorm/internal/db"
	"github.com/aknst/industrial-prog-t-26/practices/06-gorm/internal/httpapi"
	"github.com/aknst/industrial-prog-t-26/practices/06-gorm/internal/models"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	d, err := db.Connect(os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	pool, _ := d.DB()
	defer pool.Close()
	if err := d.AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{}); err != nil {
		log.Fatal(err)
	}
	log.Print("AutoMigrate: users, notes, tags, note_tags ready")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: "127.0.0.1:" + port, Handler: httpapi.BuildRouter(d), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
