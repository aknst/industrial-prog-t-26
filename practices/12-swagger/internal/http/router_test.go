package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/core/service"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/http/handlers"
	"github.com/aknst/industrial-prog-t-26/practices/12-swagger/internal/repo"
)

func server() http.Handler {
	return NewRouter(&handlers.Handlers{Svc: service.NewNoteService(repo.NewNoteRepoMem())})
}

func do(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}

func create(t *testing.T, h http.Handler, title string) core.Note {
	t.Helper()
	rr := do(h, "POST", "/api/v1/notes", `{"title":"`+title+`","content":"c"}`)
	if rr.Code != 201 {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var n core.Note
	if err := json.Unmarshal(rr.Body.Bytes(), &n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCRUD(t *testing.T) {
	h := server()

	n := create(t, h, "Первая заметка")
	if n.ID != 1 || n.Title != "Первая заметка" || n.CreatedAt.IsZero() || n.UpdatedAt != nil {
		t.Fatalf("created = %+v", n)
	}
	if rr := do(h, "POST", "/api/v1/notes", `{"title":"x"}`); rr.Header().Get("Location") != "/api/v1/notes/2" {
		t.Fatal("Location:", rr.Header().Get("Location"))
	}

	rr := do(h, "GET", "/api/v1/notes/1", "")
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), "Первая заметка") {
		t.Fatal(rr.Code, rr.Body.String())
	}

	rr = do(h, "PATCH", "/api/v1/notes/1", `{"content":"Новый текст"}`)
	var patched core.Note
	_ = json.Unmarshal(rr.Body.Bytes(), &patched)
	if rr.Code != 200 || patched.Content != "Новый текст" || patched.Title != "Первая заметка" || patched.UpdatedAt == nil {
		t.Fatalf("patched = %d %+v", rr.Code, patched)
	}

	if rr := do(h, "DELETE", "/api/v1/notes/1", ""); rr.Code != 204 || rr.Body.Len() != 0 {
		t.Fatal(rr.Code)
	}
	if rr := do(h, "GET", "/api/v1/notes/1", ""); rr.Code != 404 {
		t.Fatal(rr.Code)
	}
	if rr := do(h, "DELETE", "/api/v1/notes/1", ""); rr.Code != 404 {
		t.Fatal(rr.Code)
	}
}

func TestListPaginationAndSearch(t *testing.T) {
	h := server()
	for _, title := range []string{"Alpha", "Beta", "Gamma", "alphabet", "Delta"} {
		create(t, h, title)
	}

	rr := do(h, "GET", "/api/v1/notes?page=2&limit=2", "")
	var page []core.Note
	_ = json.Unmarshal(rr.Body.Bytes(), &page)
	if rr.Code != 200 || len(page) != 2 || page[0].ID != 3 || rr.Header().Get("X-Total-Count") != "5" {
		t.Fatalf("page 2: %d %v total=%s", rr.Code, page, rr.Header().Get("X-Total-Count"))
	}

	rr = do(h, "GET", "/api/v1/notes?q=ALPHA", "")
	page = nil
	_ = json.Unmarshal(rr.Body.Bytes(), &page)
	if len(page) != 2 || rr.Header().Get("X-Total-Count") != "2" {
		t.Fatalf("search: %v", page)
	}

	rr = do(h, "GET", "/api/v1/notes?page=99", "")
	if rr.Code != 200 || strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("out-of-range page must be an empty array: %d %q", rr.Code, rr.Body.String())
	}
}

func TestDocsAreServed(t *testing.T) {
	h := server()

	if rr := do(h, "GET", "/docs", ""); rr.Code != http.StatusFound || rr.Header().Get("Location") != "/docs/index.html" {
		t.Fatal("/docs must redirect to Swagger UI:", rr.Code, rr.Header().Get("Location"))
	}
	if rr := do(h, "GET", "/docs/index.html", ""); rr.Code != 200 || !strings.Contains(rr.Body.String(), "swagger-ui") {
		t.Fatal("Swagger UI:", rr.Code)
	}
	if rr := do(h, "GET", "/redoc", ""); rr.Code != 200 || !strings.Contains(rr.Body.String(), "<redoc") {
		t.Fatal("ReDoc:", rr.Code)
	}

	rr := do(h, "GET", "/docs/doc.json", "")
	if rr.Code != 200 {
		t.Fatal("doc.json:", rr.Code)
	}
	var spec struct {
		BasePath string                                `json:"basePath"`
		Paths    map[string]map[string]json.RawMessage `json:"paths"`
		Defs     map[string]json.RawMessage            `json:"definitions"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.BasePath != "/api/v1" {
		t.Fatal("basePath =", spec.BasePath)
	}
	// Все пять методов /notes описаны в спецификации.
	for path, methods := range map[string][]string{"/notes": {"get", "post"}, "/notes/{id}": {"get", "patch", "delete"}} {
		for _, m := range methods {
			if _, ok := spec.Paths[path][m]; !ok {
				t.Errorf("%s %s is not documented", strings.ToUpper(m), path)
			}
		}
	}
	for _, def := range []string{"core.Note", "core.NoteCreate", "core.NoteUpdate", "handlers.ErrorResponse"} {
		if _, ok := spec.Defs[def]; !ok {
			t.Errorf("schema %s is missing", def)
		}
	}
}

func TestErrorsAndStatusCodes(t *testing.T) {
	h := server()
	create(t, h, "exists")
	cases := []struct {
		method, target, body string
		want                 int
	}{
		{"POST", "/api/v1/notes", `{`, 400},
		{"POST", "/api/v1/notes", `{"title":"   "}`, 400},
		{"POST", "/api/v1/notes", `{"title":"` + strings.Repeat("я", 201) + `"}`, 400},
		{"POST", "/api/v1/notes", `{"title":"x","id":5}`, 400},
		{"GET", "/api/v1/notes/abc", "", 400},
		{"GET", "/api/v1/notes/0", "", 400},
		{"GET", "/api/v1/notes/999", "", 404},
		{"GET", "/api/v1/notes?page=0", "", 400},
		{"GET", "/api/v1/notes?page=x", "", 400},
		{"GET", "/api/v1/notes?limit=101", "", 400},
		{"GET", "/api/v1/notes?limit=0", "", 400},
		{"PATCH", "/api/v1/notes/1", `{}`, 400},
		{"PATCH", "/api/v1/notes/1", `{"title":""}`, 400},
		{"PATCH", "/api/v1/notes/999", `{"title":"x"}`, 404},
		{"GET", "/health", "", 200},
	}
	for _, c := range cases {
		rr := do(h, c.method, c.target, c.body)
		if rr.Code != c.want {
			t.Errorf("%s %s %q: got %d, want %d (%s)", c.method, c.target, c.body, rr.Code, c.want, rr.Body.String())
		}
		if c.want >= 400 && !strings.Contains(rr.Body.String(), `"error"`) {
			t.Errorf("%s %s: error body is not JSON error: %s", c.method, c.target, rr.Body.String())
		}
	}
}
