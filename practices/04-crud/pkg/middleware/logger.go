package middleware

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
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &recorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s status=%d elapsed=%s", r.Method, r.URL.RequestURI(), rw.status, time.Since(start))
	})
}
