package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/jwt"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/respond"
)

// AuthZRoles (RBAC) пропускает только пользователей с одной из перечисленных ролей.
// Должен стоять после AuthN: без клеймов в контексте запрос получает 401.
func AuthZRoles(allowed ...string) func(http.Handler) http.Handler {
	set := toSet(allowed)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := jwt.ClaimsFrom(r.Context())
			if !ok {
				unauthorized(w, "authentication required")
				return
			}
			if _, ok := set[c.Role]; !ok {
				respond.Error(w, http.StatusForbidden, "forbidden", "insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthZSelfOrRoles (ABAC) разрешает доступ владельцу ресурса — когда параметр
// маршрута param равен sub из токена — либо пользователю с одной из ролей privileged.
func AuthZSelfOrRoles(param string, privileged ...string) func(http.Handler) http.Handler {
	set := toSet(privileged)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := jwt.ClaimsFrom(r.Context())
			if !ok {
				unauthorized(w, "authentication required")
				return
			}
			_, isPrivileged := set[c.Role]
			if !isPrivileged && chi.URLParam(r, param) != c.Subject {
				respond.Error(w, http.StatusForbidden, "forbidden", "only the owner or an admin may access this resource")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, it := range items {
		set[it] = struct{}{}
	}
	return set
}
