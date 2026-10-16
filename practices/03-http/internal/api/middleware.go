package api

import (
	"log"
	"net/http"
	"time"
)

type recorder struct {
	http.ResponseWriter
	status  int
	written bool
}

func (w *recorder) WriteHeader(code int) {
	if !w.written {
		w.status = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}
func (w *recorder) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(b)
}
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &recorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s status=%d elapsed=%s", r.Method, r.URL.RequestURI(), rec.status, time.Since(start))
	})
}
