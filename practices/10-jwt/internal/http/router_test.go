package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/config"
)

func newHandler(attempts int) http.Handler {
	return Build(config.Config{
		JWTSecret:     []byte("test-secret-test-secret"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		LoginAttempts: attempts,
		LoginWindow:   5 * time.Minute,
	}, bcrypt.MinCost)
}

func do(h http.Handler, method, target, body, bearer string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.RemoteAddr = "192.0.2.1:1234"
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	h.ServeHTTP(rr, req)
	return rr
}

type pair struct {
	Token        string `json:"token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func login(t *testing.T, h http.Handler, email string) pair {
	t.Helper()
	rr := do(h, "POST", "/api/v1/login", `{"Email":"`+email+`","Password":"secret123"}`, "")
	if rr.Code != 200 {
		t.Fatalf("login %s: %d %s", email, rr.Code, rr.Body.String())
	}
	var p pair
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil || p.AccessToken == "" || p.RefreshToken == "" || p.Token != p.AccessToken {
		t.Fatalf("bad token pair: %s (%v)", rr.Body.String(), err)
	}
	return p
}

func TestLoginErrors(t *testing.T) {
	h := newHandler(100)
	cases := []struct {
		body string
		want int
	}{
		{`{"email":"admin@example.com","password":"wrong"}`, 401},
		{`{"email":"nobody@example.com","password":"secret123"}`, 401},
		{`{"email":"","password":""}`, 400},
		{`{`, 400},
	}
	for _, c := range cases {
		rr := do(h, "POST", "/api/v1/login", c.body, "")
		if rr.Code != c.want || !strings.Contains(rr.Body.String(), `"error"`) {
			t.Errorf("%s: %d %s", c.body, rr.Code, rr.Body.String())
		}
	}
}

func TestProtectedRoutesRequireToken(t *testing.T) {
	h := newHandler(100)
	for _, target := range []string{"/api/v1/me", "/api/v1/admin/stats", "/api/v1/users/1"} {
		rr := do(h, "GET", target, "", "")
		if rr.Code != 401 || rr.Header().Get("WWW-Authenticate") == "" {
			t.Errorf("%s without token: %d", target, rr.Code)
		}
		if rr := do(h, "GET", target, "", "garbage.token.value"); rr.Code != 401 {
			t.Errorf("%s with garbage token: %d", target, rr.Code)
		}
	}
}

func TestAdminFlowAndRBAC(t *testing.T) {
	h := newHandler(100)
	admin := login(t, h, "admin@example.com")
	user := login(t, h, "user@example.com")

	rr := do(h, "GET", "/api/v1/me", "", admin.AccessToken)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), `"role":"admin"`) || !strings.Contains(rr.Body.String(), `"email":"admin@example.com"`) {
		t.Fatal(rr.Code, rr.Body.String())
	}
	if rr := do(h, "GET", "/api/v1/admin/stats", "", admin.AccessToken); rr.Code != 200 {
		t.Fatal("admin stats:", rr.Code)
	}
	if rr := do(h, "GET", "/api/v1/me", "", user.AccessToken); rr.Code != 200 {
		t.Fatal("user /me:", rr.Code)
	}
	if rr := do(h, "GET", "/api/v1/admin/stats", "", user.AccessToken); rr.Code != 403 {
		t.Fatal("user must get 403 on admin stats, got", rr.Code)
	}
}

func TestRefreshTokenCannotBeUsedAsAccessAndViceVersa(t *testing.T) {
	h := newHandler(100)
	p := login(t, h, "user@example.com")
	if rr := do(h, "GET", "/api/v1/me", "", p.RefreshToken); rr.Code != 401 {
		t.Fatal("refresh token accepted as access token:", rr.Code)
	}
	if rr := do(h, "POST", "/api/v1/refresh", `{"refresh_token":"`+p.AccessToken+`"}`, ""); rr.Code != 401 {
		t.Fatal("access token accepted as refresh token:", rr.Code)
	}
}

func TestRefreshRotationAndReuseDetection(t *testing.T) {
	h := newHandler(100)
	first := login(t, h, "user@example.com")

	rr := do(h, "POST", "/api/v1/refresh", `{"refresh_token":"`+first.RefreshToken+`"}`, "")
	if rr.Code != 200 {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var second pair
	_ = json.Unmarshal(rr.Body.Bytes(), &second)
	if second.AccessToken == "" || second.AccessToken == first.AccessToken || second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh must issue a new pair")
	}
	if rr := do(h, "GET", "/api/v1/me", "", second.AccessToken); rr.Code != 200 {
		t.Fatal("new access token rejected:", rr.Code)
	}
	// Старый refresh отозван: повторное использование -> 401.
	if rr := do(h, "POST", "/api/v1/refresh", `{"refresh_token":"`+first.RefreshToken+`"}`, ""); rr.Code != 401 {
		t.Fatal("old refresh token reused:", rr.Code)
	}
	// Новый refresh продолжает работать.
	if rr := do(h, "POST", "/api/v1/refresh", `{"refresh_token":"`+second.RefreshToken+`"}`, ""); rr.Code != 200 {
		t.Fatal("new refresh token rejected:", rr.Code)
	}
}

func TestLogoutRevokesRefreshToken(t *testing.T) {
	h := newHandler(100)
	p := login(t, h, "user@example.com")
	if rr := do(h, "POST", "/api/v1/logout", `{"refresh_token":"`+p.RefreshToken+`"}`, ""); rr.Code != 204 {
		t.Fatal(rr.Code)
	}
	if rr := do(h, "POST", "/api/v1/refresh", `{"refresh_token":"`+p.RefreshToken+`"}`, ""); rr.Code != 401 {
		t.Fatal("refresh after logout:", rr.Code)
	}
}

func TestABACUserCanReadOnlyOwnProfile(t *testing.T) {
	h := newHandler(100)
	user := login(t, h, "user@example.com") // id 2
	admin := login(t, h, "admin@example.com")

	if rr := do(h, "GET", "/api/v1/users/2", "", user.AccessToken); rr.Code != 200 || !strings.Contains(rr.Body.String(), "user@example.com") {
		t.Fatal(rr.Code, rr.Body.String())
	}
	if rr := do(h, "GET", "/api/v1/users/1", "", user.AccessToken); rr.Code != 403 {
		t.Fatal("user reads another profile:", rr.Code)
	}
	// 403, а не 404: user не узнаёт, существует ли чужой id.
	if rr := do(h, "GET", "/api/v1/users/999", "", user.AccessToken); rr.Code != 403 {
		t.Fatal(rr.Code)
	}
	if rr := do(h, "GET", "/api/v1/users/2", "", admin.AccessToken); rr.Code != 200 {
		t.Fatal("admin must read any profile:", rr.Code)
	}
	if rr := do(h, "GET", "/api/v1/users/999", "", admin.AccessToken); rr.Code != 404 {
		t.Fatal(rr.Code)
	}
	if rr := do(h, "GET", "/api/v1/users/abc", "", admin.AccessToken); rr.Code != 400 {
		t.Fatal(rr.Code)
	}
}

func TestLoginRateLimit(t *testing.T) {
	h := newHandler(3)
	for i := 0; i < 3; i++ {
		if rr := do(h, "POST", "/api/v1/login", `{"email":"user@example.com","password":"bad"}`, ""); rr.Code != 401 {
			t.Fatalf("attempt %d: %d", i+1, rr.Code)
		}
	}
	rr := do(h, "POST", "/api/v1/login", `{"email":"user@example.com","password":"secret123"}`, "")
	if rr.Code != 429 || rr.Header().Get("Retry-After") == "" {
		t.Fatalf("4th attempt must be limited: %d", rr.Code)
	}
}

func TestUnifiedErrorFormat(t *testing.T) {
	h := newHandler(100)
	for _, c := range []struct{ method, target string }{{"GET", "/nope"}, {"DELETE", "/api/v1/me"}} {
		rr := do(h, c.method, c.target, "", "")
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || body["error"] == nil {
			t.Errorf("%s %s: not a JSON error: %s", c.method, c.target, rr.Body.String())
		}
		if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json") {
			t.Errorf("%s %s: content-type %q", c.method, c.target, rr.Header().Get("Content-Type"))
		}
	}
}
