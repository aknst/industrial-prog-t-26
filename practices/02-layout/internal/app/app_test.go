package app

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestPingRequestID(t *testing.T) {
	for _, id := range []string{"", "demo-123"} {
		req := httptest.NewRequest("GET", "/ping", nil)
		req.Header.Set("X-Request-Id", id)
		rr := httptest.NewRecorder()
		Routes().ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatal(rr.Code)
		}
		actual := rr.Header().Get("X-Request-Id")
		if actual == "" || (id != "" && actual != id) {
			t.Fatalf("request id %q", actual)
		}
		var body map[string]string
		if json.Unmarshal(rr.Body.Bytes(), &body) != nil || body["status"] != "ok" {
			t.Fatal(rr.Body.String())
		}
	}
}
