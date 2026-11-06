package main

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	router "github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/http"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if len(cfg.JWTSecret) < 16 {
		log.Print("warning: JWT_SECRET is short, use at least 32 random characters outside of training")
	}
	srv := &http.Server{
		Addr:              "127.0.0.1" + cfg.Port,
		Handler:           router.Build(cfg, bcrypt.DefaultCost),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s (access %s, refresh %s)", srv.Addr, cfg.AccessTTL, cfg.RefreshTTL)
	log.Fatal(srv.ListenAndServe())
}
