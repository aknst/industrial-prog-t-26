package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateAndErrors(t *testing.T) {
	h := NewRouter()
	cases := []struct {
		method, path, body string
		code               int
	}{{"POST", "/tasks", `{"title":"Купить молоко"}`, 201}, {"GET", "/tasks/1", "", 200}, {"GET", "/tasks?q=молоко", "", 200}, {"GET", "/tasks/999", "", 404}, {"GET", "/tasks/abc", "", 400}, {"POST", "/tasks", `{}`, 400}, {"POST", "/tasks", `{"title":"ab"}`, 422}, {"POST", "/tasks", `{"title":"valid"} {}`, 400}}
	for _, c := range cases {
		t.Run(c.method+c.path+c.body, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != c.code {
				t.Fatalf("got %d want %d: %s", rr.Code, c.code, rr.Body.String())
			}
			if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json") {
				t.Fatal(rr.Header())
			}
		})
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/tasks", nil))
	if !strings.Contains(rr.Body.String(), "Купить молоко") {
		t.Fatal(rr.Body.String())
	}
}
