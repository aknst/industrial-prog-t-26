package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/aknst/industrial-prog-t-26/practices/07-redis/internal/cache"
)

func newServer(t *testing.T) (http.Handler, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c := cache.New(mr.Addr(), "", 0)
	t.Cleanup(func() { c.Close() })
	return Router(&Handlers{Cache: c}), mr
}

func do(h http.Handler, target string) (int, string) {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
	return rr.Code, strings.TrimSpace(rr.Body.String())
}

func TestSetGetTTL(t *testing.T) {
	h, _ := newServer(t)
	if code, body := do(h, "/set?key=test&value=hello"); code != 200 || body != "OK: test=hello (TTL 10s)" {
		t.Fatal(code, body)
	}
	if code, body := do(h, "/get?key=test"); code != 200 || body != "VALUE: test=hello" {
		t.Fatal(code, body)
	}
	if code, body := do(h, "/ttl?key=test"); code != 200 || body != "TTL for test: 10s" {
		t.Fatal(code, body)
	}
}

func TestKeyExpiresAfterTTL(t *testing.T) {
	h, mr := newServer(t)
	do(h, "/set?key=k&value=v&ttl=5")
	mr.FastForward(6 * time.Second)
	if code, _ := do(h, "/get?key=k"); code != http.StatusNotFound {
		t.Fatal("expired key must be 404, got", code)
	}
	if code, _ := do(h, "/ttl?key=k"); code != http.StatusNotFound {
		t.Fatal("ttl of expired key must be 404, got", code)
	}
}

func TestValidation(t *testing.T) {
	h, _ := newServer(t)
	for _, target := range []string{"/set?key=a", "/set?value=b", "/set?key=a&value=b&ttl=0", "/set?key=a&value=b&ttl=x", "/get", "/ttl", "/del", "/product"} {
		if code, _ := do(h, target); code != http.StatusBadRequest {
			t.Fatal(target, code)
		}
	}
	if code, _ := do(h, "/get?key=missing"); code != http.StatusNotFound {
		t.Fatal(code)
	}
}

func TestDel(t *testing.T) {
	h, _ := newServer(t)
	do(h, "/set?key=k&value=v")
	if _, body := do(h, "/del?key=k"); body != "DEL k: true" {
		t.Fatal(body)
	}
	if _, body := do(h, "/del?key=k"); body != "DEL k: false" {
		t.Fatal(body)
	}
}

func TestProductIsCachedAfterFirstRequest(t *testing.T) {
	h, mr := newServer(t)
	if _, body := do(h, "/product?id=7"); !strings.Contains(body, `"source":"db"`) {
		t.Fatal(body)
	}
	if _, body := do(h, "/product?id=7"); !strings.Contains(body, `"source":"cache"`) {
		t.Fatal(body)
	}
	mr.FastForward(31 * time.Second)
	if _, body := do(h, "/product?id=7"); !strings.Contains(body, `"source":"db"`) {
		t.Fatal("cache must be refreshed after TTL:", body)
	}
}

func TestHealthFailsWhenRedisIsDown(t *testing.T) {
	h, mr := newServer(t)
	if code, _ := do(h, "/health"); code != 200 {
		t.Fatal(code)
	}
	mr.Close()
	if code, _ := do(h, "/health"); code != http.StatusServiceUnavailable {
		t.Fatal(code)
	}
}
