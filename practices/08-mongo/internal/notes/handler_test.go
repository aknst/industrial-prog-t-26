package notes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// fakeStore — минимальная in-memory реализация Store для тестов HTTP-слоя.
type fakeStore struct {
	items map[string]Note
}

func newFake() *fakeStore { return &fakeStore{items: map[string]Note{}} }

func (f *fakeStore) Create(_ context.Context, title, content string) (Note, error) {
	for _, n := range f.items {
		if n.Title == title {
			return Note{}, ErrDuplicate
		}
	}
	n := Note{ID: primitive.NewObjectID(), Title: title, Content: content, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	f.items[n.ID.Hex()] = n
	return n, nil
}

func (f *fakeStore) ByID(_ context.Context, id string) (Note, error) {
	n, ok := f.items[id]
	if !ok {
		return Note{}, ErrNotFound
	}
	return n, nil
}

func (f *fakeStore) List(_ context.Context, p ListParams) ([]Note, error) {
	out := []Note{}
	for _, n := range f.items {
		if strings.Contains(strings.ToLower(n.Title), strings.ToLower(p.Q)) {
			out = append(out, n)
		}
	}
	return out, nil
}

func (f *fakeStore) Update(_ context.Context, id string, title, content *string) (Note, error) {
	n, ok := f.items[id]
	if !ok {
		return Note{}, ErrNotFound
	}
	if title != nil {
		n.Title = *title
	}
	if content != nil {
		n.Content = *content
	}
	f.items[id] = n
	return n, nil
}

func (f *fakeStore) Delete(_ context.Context, id string) error {
	if _, ok := f.items[id]; !ok {
		return ErrNotFound
	}
	delete(f.items, id)
	return nil
}

func (f *fakeStore) Stats(context.Context) (Stats, error) { return Stats{Count: len(f.items)}, nil }

func call(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}

func TestCRUDFlow(t *testing.T) {
	h := NewHandler(newFake()).Routes()

	rr := call(h, "POST", "/", `{"title":"First note","content":"Hello Mongo!"}`)
	if rr.Code != 201 || !strings.Contains(rr.Body.String(), `"title":"First note"`) {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var created Note
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil || created.ID.IsZero() {
		t.Fatal(err, rr.Body.String())
	}
	id := created.ID.Hex()

	if rr := call(h, "GET", "/"+id, ""); rr.Code != 200 {
		t.Fatal(rr.Code)
	}
	if rr := call(h, "GET", "/?q=first", ""); rr.Code != 200 || !strings.Contains(rr.Body.String(), id) {
		t.Fatal(rr.Code, rr.Body.String())
	}
	rr = call(h, "PATCH", "/"+id, `{"content":"Updated content"}`)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), `"content":"Updated content"`) || !strings.Contains(rr.Body.String(), `"title":"First note"`) {
		t.Fatal(rr.Code, rr.Body.String())
	}
	if rr := call(h, "DELETE", "/"+id, ""); rr.Code != 204 {
		t.Fatal(rr.Code)
	}
	if rr := call(h, "GET", "/"+id, ""); rr.Code != 404 {
		t.Fatal(rr.Code)
	}
}

func TestStatusCodes(t *testing.T) {
	h := NewHandler(newFake()).Routes()
	call(h, "POST", "/", `{"title":"dup"}`)
	cases := []struct {
		method, target, body string
		want                 int
	}{
		{"POST", "/", `{"title":"dup"}`, 409},
		{"POST", "/", `{"title":"  "}`, 400},
		{"POST", "/", `{`, 400},
		{"POST", "/", `{"title":"x","extra":1}`, 400},
		{"GET", "/not-an-object-id", "", 404},
		{"PATCH", "/not-an-object-id", `{"title":"x"}`, 404},
		{"PATCH", "/507f1f77bcf86cd799439011", `{}`, 400},
		{"DELETE", "/507f1f77bcf86cd799439011", "", 404},
		{"GET", "/stats", "", 200},
	}
	for _, c := range cases {
		if rr := call(h, c.method, c.target, c.body); rr.Code != c.want {
			t.Errorf("%s %s %s: got %d, want %d (%s)", c.method, c.target, c.body, rr.Code, c.want, rr.Body.String())
		}
	}
}
