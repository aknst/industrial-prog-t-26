package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/07-redis/internal/api"
	"github.com/aknst/industrial-prog-t-26/practices/07-redis/internal/cache"
)

func main() {
	addr := getenv("REDIS_ADDR", "localhost:6379")
	db, _ := strconv.Atoi(getenv("REDIS_DB", "0"))
	c := cache.New(addr, os.Getenv("REDIS_PASSWORD"), db)
	defer c.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err := c.Ping(pingCtx)
	cancel()
	if err != nil {
		log.Fatalf("redis %s: %v", addr, err)
	}

	srv := &http.Server{
		Addr:              "127.0.0.1:" + getenv("PORT", "8080"),
		Handler:           api.Router(&api.Handlers{Cache: c, LoadDelay: 300 * time.Millisecond}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("redis %s, listening on %s", addr, srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
