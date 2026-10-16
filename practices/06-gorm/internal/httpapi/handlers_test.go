package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRejectsTrailingAndUnknownFields(t *testing.T) {
	for _, body := range []string{`{"title":"ok"} {}`, `{"unknown":"value"}`, `{`} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/notes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		var in struct {
			Title string `json:"title"`
		}
		if decode(rr, req, &in) {
			t.Fatal("accepted", body)
		}
		if rr.Code != 400 {
			t.Fatal(rr.Code)
		}
	}
}
