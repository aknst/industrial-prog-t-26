package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http/httptest"
	"testing"
)

func TestUserUUID(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		routes().ServeHTTP(rr, httptest.NewRequest("GET", "/user", nil))
		var u struct {
			ID   string
			Name string
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &u); err != nil {
			t.Fatal(err)
		}
		if _, err := uuid.Parse(u.ID); err != nil {
			t.Fatal(err)
		}
		if seen[u.ID] || u.Name != "Gopher" {
			t.Fatal(u)
		}
		seen[u.ID] = true
	}
}
