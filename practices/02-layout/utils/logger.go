package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

func NewID16() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = NewID16()
		}
		w.Header().Set("X-Request-Id", id)
		log.Printf("request_id=%s remote=%s method=%s path=%s", id, r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("request_id=%s elapsed=%s", id, time.Since(start))
	})
}
