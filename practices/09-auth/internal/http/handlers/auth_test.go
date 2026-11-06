package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/repo"
)

type memStore struct {
	mu    sync.Mutex
	users map[string]core.User
}

func (m *memStore) Create(_ context.Context, u *core.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[u.Email]; ok {
		return repo.ErrEmailTaken
	}
	u.ID = int64(len(m.users) + 1)
	m.users[u.Email] = *u
	return nil
}

func (m *memStore) ByEmail(_ context.Context, email string) (core.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[email]
	if !ok {
		return core.User{}, repo.ErrUserNotFound
	}
	return u, nil
}

func setup() (*AuthHandler, *memStore) {
	st := &memStore{users: map[string]core.User{}}
	return NewAuthHandler(st, bcrypt.MinCost), st
}

func post(fn http.HandlerFunc, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	fn(rr, req)
	return rr
}

func TestRegisterStoresBcryptHashNotPassword(t *testing.T) {
	h, st := setup()
	rr := post(h.Register, `{"email":" User@Example.com ","password":"Secret123!"}`)
	if rr.Code != http.StatusCreated {
		t.Fatal(rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "Secret123") || strings.Contains(rr.Body.String(), "$2") {
		t.Fatal("response leaks password or hash:", rr.Body.String())
	}
	u := st.users["user@example.com"]
	if u.PasswordHash == "" || u.PasswordHash == "Secret123!" || !strings.HasPrefix(u.PasswordHash, "$2") {
		t.Fatalf("hash = %q", u.PasswordHash)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("Secret123!")) != nil {
		t.Fatal("stored hash does not match the password")
	}
}

func TestSamePasswordGivesDifferentHashes(t *testing.T) {
	h, st := setup()
	post(h.Register, `{"email":"a@example.com","password":"Secret123!"}`)
	post(h.Register, `{"email":"b@example.com","password":"Secret123!"}`)
	if st.users["a@example.com"].PasswordHash == st.users["b@example.com"].PasswordHash {
		t.Fatal("salt is not applied")
	}
}

func TestRegisterValidationAndConflict(t *testing.T) {
	h, _ := setup()
	cases := []struct {
		body string
		want int
	}{
		{`{"email":"u@example.com","password":"Secret123!"}`, 201},
		{`{"email":"U@EXAMPLE.com","password":"AnotherPass"}`, 409},
		{`{"email":"not-an-email","password":"Secret123!"}`, 400},
		{`{"email":"","password":"Secret123!"}`, 400},
		{`{"email":"v@example.com","password":"short"}`, 400},
		{`{"email":"v@example.com","password":"` + strings.Repeat("a", 73) + `"}`, 400},
		{`{`, 400},
		{`{"email":"v@example.com","password":"Secret123!","role":"admin"}`, 400},
	}
	for _, c := range cases {
		if rr := post(h.Register, c.body); rr.Code != c.want {
			t.Errorf("%s: got %d, want %d (%s)", c.body, rr.Code, c.want, rr.Body.String())
		}
	}
}

func TestLogin(t *testing.T) {
	h, _ := setup()
	post(h.Register, `{"email":"user@example.com","password":"Secret123!"}`)

	if rr := post(h.Login, `{"email":"USER@example.com","password":"Secret123!"}`); rr.Code != 200 {
		t.Fatal(rr.Code, rr.Body.String())
	}
	wrongPass := post(h.Login, `{"email":"user@example.com","password":"wrong"}`)
	noUser := post(h.Login, `{"email":"nobody@example.com","password":"Secret123!"}`)
	if wrongPass.Code != 401 || noUser.Code != 401 {
		t.Fatal(wrongPass.Code, noUser.Code)
	}
	if wrongPass.Body.String() != noUser.Body.String() {
		t.Fatalf("responses differ and reveal whether email exists: %q vs %q", wrongPass.Body.String(), noUser.Body.String())
	}
	if rr := post(h.Login, `{"email":"user@example.com"}`); rr.Code != 400 {
		t.Fatal(rr.Code)
	}
}
