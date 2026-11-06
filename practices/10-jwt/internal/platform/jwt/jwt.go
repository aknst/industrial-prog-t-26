package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	Issuer   = "pz10-auth"
	Audience = "pz10-clients"

	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

var ErrWrongType = errors.New("wrong token type")

// Claims — registered claims (sub, exp, iat, iss, aud, jti) плюс private claims.
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"typ"`
	jwtlib.RegisteredClaims
}

// UserID возвращает числовой идентификатор из sub.
func (c *Claims) UserID() (int64, error) { return strconv.ParseInt(c.Subject, 10, 64) }

type HS256 struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewHS256(secret []byte, accessTTL, refreshTTL time.Duration) *HS256 {
	return &HS256{secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (h *HS256) AccessTTL() time.Duration  { return h.accessTTL }
func (h *HS256) RefreshTTL() time.Duration { return h.refreshTTL }

func (h *HS256) SignAccess(userID int64, email, role string) (string, *Claims, error) {
	return h.sign(TypeAccess, h.accessTTL, userID, email, role)
}

func (h *HS256) SignRefresh(userID int64, email, role string) (string, *Claims, error) {
	return h.sign(TypeRefresh, h.refreshTTL, userID, email, role)
}

func (h *HS256) sign(typ string, ttl time.Duration, userID int64, email, role string) (string, *Claims, error) {
	jti, err := randomID()
	if err != nil {
		return "", nil, err
	}
	now := time.Now()
	claims := &Claims{
		Email: email,
		Role:  role,
		Type:  typ,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    Issuer,
			Audience:  jwtlib.ClaimStrings{Audience},
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
			ID:        jti,
		},
	}
	signed, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(h.secret)
	return signed, claims, err
}

// Parse проверяет подпись (только HS256), срок действия, iss, aud и тип токена.
func (h *HS256) Parse(tokenStr, wantType string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwtlib.ParseWithClaims(tokenStr, claims,
		func(*jwtlib.Token) (any, error) { return h.secret, nil },
		jwtlib.WithValidMethods([]string{"HS256"}),
		jwtlib.WithAudience(Audience),
		jwtlib.WithIssuer(Issuer),
		jwtlib.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != wantType {
		return nil, ErrWrongType
	}
	return claims, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type ctxKey struct{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func ClaimsFrom(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok && c != nil
}
