package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/respond"
)

// RateLimit ограничивает число запросов с одного IP: не более limit за окно window.
// IP берётся из RemoteAddr; заголовкам X-Forwarded-For доверять нельзя без доверенного прокси.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return rateLimit(limit, window, time.Now)
}

func rateLimit(limit int, window time.Duration, now func() time.Time) func(http.Handler) http.Handler {
	var mu sync.Mutex
	hits := map[string][]time.Time{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			t := now()

			mu.Lock()
			for k, list := range hits { // очистка устаревших записей, чтобы карта не росла
				if len(list) == 0 || t.Sub(list[len(list)-1]) >= window {
					delete(hits, k)
				}
			}
			recent := hits[ip][:0:0]
			for _, h := range hits[ip] {
				if t.Sub(h) < window {
					recent = append(recent, h)
				}
			}
			if len(recent) >= limit {
				retry := int(window-t.Sub(recent[0])) / int(time.Second)
				hits[ip] = recent
				mu.Unlock()
				w.Header().Set("Retry-After", strconv.Itoa(retry+1))
				respond.Error(w, http.StatusTooManyRequests, "too_many_requests", "login attempts limit exceeded")
				return
			}
			hits[ip] = append(recent, t)
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
