package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCRUD(t *testing.T) {
	h := routes()
	for _, c := range []struct {
		method, path, body string
		code               int
	}{{"POST", "/api/tasks", `{"title":"Выучить chi"}`, 201}, {"GET", "/api/tasks/1", "", 200}, {"PUT", "/api/tasks/1", `{"title":"Готовая задача","done":true}`, 200}, {"GET", "/api/tasks?done=true&page=1&limit=5", "", 200}, {"OPTIONS", "/api/tasks", "", 204}, {"DELETE", "/api/tasks/1", "", 204}, {"GET", "/api/tasks/1", "", 404}, {"GET", "/api/tasks/abc", "", 400}, {"POST", "/api/tasks", `{"title":"ab"}`, 422}, {"GET", "/api/tasks?limit=0", "", 400}} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(rr, req)
		if rr.Code != c.code {
			t.Fatalf("%s %s: %d want %d %s", c.method, c.path, rr.Code, c.code, rr.Body.String())
		}
		if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatal("missing CORS")
		}
		if c.code == 204 && rr.Body.Len() != 0 {
			t.Fatal("204 must be empty")
		}
	}
}
