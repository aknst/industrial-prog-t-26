package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitWindowAndPerIP(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	h := rateLimit(2, time.Minute, clock)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	call := func(ip string) int {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", nil)
		req.RemoteAddr = ip + ":1111"
		h.ServeHTTP(rr, req)
		return rr.Code
	}

	if call("192.0.2.1") != 200 || call("192.0.2.1") != 200 {
		t.Fatal("first two requests must pass")
	}
	if call("192.0.2.1") != 429 {
		t.Fatal("third request must be limited")
	}
	if call("192.0.2.2") != 200 {
		t.Fatal("another IP has its own budget")
	}
	now = now.Add(61 * time.Second)
	if call("192.0.2.1") != 200 {
		t.Fatal("budget must be restored after the window")
	}
}
