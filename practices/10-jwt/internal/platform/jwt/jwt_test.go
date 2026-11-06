package jwt

import (
	"strings"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var secret = []byte("test-secret-test-secret")

func TestSignAndParse(t *testing.T) {
	h := NewHS256(secret, time.Minute, time.Hour)
	tok, signed, err := h.SignAccess(42, "a@example.com", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(tok, ".") != 2 {
		t.Fatal("token must be header.payload.signature:", tok)
	}
	c, err := h.Parse(tok, TypeAccess)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := c.UserID()
	if id != 42 || c.Email != "a@example.com" || c.Role != "admin" || c.Subject != "42" || c.ID != signed.ID || c.ID == "" {
		t.Fatalf("claims = %+v", c)
	}
	if got := c.ExpiresAt.Sub(c.IssuedAt.Time); got != time.Minute {
		t.Fatalf("exp-iat = %v", got)
	}
}

func TestRejectsExpiredTamperedForeignAndWrongType(t *testing.T) {
	h := NewHS256(secret, time.Minute, time.Hour)

	expired, _, _ := NewHS256(secret, -time.Minute, time.Hour).SignAccess(1, "a@example.com", "user")
	if _, err := h.Parse(expired, TypeAccess); err == nil {
		t.Fatal("expired token accepted")
	}

	good, _, _ := h.SignAccess(1, "a@example.com", "user")
	parts := strings.Split(good, ".")
	tampered := parts[0] + "." + parts[1] + "x." + parts[2]
	if _, err := h.Parse(tampered, TypeAccess); err == nil {
		t.Fatal("tampered payload accepted")
	}

	foreign, _, _ := NewHS256([]byte("another-secret-another"), time.Minute, time.Hour).SignAccess(1, "a@example.com", "admin")
	if _, err := h.Parse(foreign, TypeAccess); err == nil {
		t.Fatal("token signed with another secret accepted")
	}

	refresh, _, _ := h.SignRefresh(1, "a@example.com", "user")
	if _, err := h.Parse(refresh, TypeAccess); err != ErrWrongType {
		t.Fatalf("want ErrWrongType, got %v", err)
	}
}

func TestRejectsAlgNone(t *testing.T) {
	h := NewHS256(secret, time.Minute, time.Hour)
	c := &Claims{Email: "a@example.com", Role: "admin", Type: TypeAccess, RegisteredClaims: jwtlib.RegisteredClaims{
		Subject: "1", Issuer: Issuer, Audience: jwtlib.ClaimStrings{Audience},
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	none, err := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, c).SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Parse(none, TypeAccess); err == nil {
		t.Fatal("alg=none accepted")
	}
}

func TestRejectsWrongIssuerAndAudience(t *testing.T) {
	h := NewHS256(secret, time.Minute, time.Hour)
	for name, c := range map[string]*Claims{
		"issuer":   {Type: TypeAccess, RegisteredClaims: jwtlib.RegisteredClaims{Subject: "1", Issuer: "evil", Audience: jwtlib.ClaimStrings{Audience}, ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}},
		"audience": {Type: TypeAccess, RegisteredClaims: jwtlib.RegisteredClaims{Subject: "1", Issuer: Issuer, Audience: jwtlib.ClaimStrings{"other"}, ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}},
		"no exp":   {Type: TypeAccess, RegisteredClaims: jwtlib.RegisteredClaims{Subject: "1", Issuer: Issuer, Audience: jwtlib.ClaimStrings{Audience}}},
	} {
		tok, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, c).SignedString(secret)
		if _, err := h.Parse(tok, TypeAccess); err == nil {
			t.Errorf("%s: token accepted", name)
		}
	}
}
