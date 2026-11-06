package middleware

import (
	"net/http"
	"strings"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/jwt"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/respond"
)

// AuthN достаёт токен из Authorization: Bearer ..., проверяет подпись и срок действия
// и кладёт клеймы в контекст запроса. Принимаются только access-токены.
func AuthN(v *jwt.HS256) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			raw, ok := strings.CutPrefix(h, "Bearer ")
			if !ok || strings.TrimSpace(raw) == "" {
				unauthorized(w, "missing bearer token")
				return
			}
			claims, err := v.Parse(strings.TrimSpace(raw), jwt.TypeAccess)
			if err != nil {
				unauthorized(w, "invalid or expired token")
				return
			}
			next.ServeHTTP(w, r.WithContext(jwt.WithClaims(r.Context(), claims)))
		})
	}
}

func unauthorized(w http.ResponseWriter, details string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="pz10-auth"`)
	respond.Error(w, http.StatusUnauthorized, "unauthorized", details)
}
