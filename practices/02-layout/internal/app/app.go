package app

import (
	"github.com/aknst/industrial-prog-t-26/practices/02-layout/internal/app/handlers"
	"github.com/aknst/industrial-prog-t-26/practices/02-layout/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Hello, Go project structure!")
	})
	mux.HandleFunc("GET /ping", handlers.Ping)
	mux.HandleFunc("GET /fail", func(w http.ResponseWriter, r *http.Request) { utils.WriteErr(w, 400, "bad_request_example") })
	return utils.Logging(mux)
}
func Run() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: "127.0.0.1:" + port, Handler: Routes(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
