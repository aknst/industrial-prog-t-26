package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/http/middleware"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/config"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/jwt"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/respond"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/repo"
)

// Build собирает зависимости и цепочку middleware. bcryptCost выбирается снаружи:
// сервер использует bcrypt.DefaultCost, тесты — bcrypt.MinCost.
func Build(cfg config.Config, bcryptCost int) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(middleware.Logger)

	users := repo.NewUserMem(bcryptCost)
	jwtv := jwt.NewHS256(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	svc := core.NewService(users, jwtv, repo.NewBlacklistMem())

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) { respond.Error(w, http.StatusNotFound, "not_found") })
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		respond.Error(w, http.StatusMethodNotAllowed, "method_not_allowed")
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Публичные маршруты. Вход ограничен по числу попыток с одного IP.
	r.With(middleware.RateLimit(cfg.LoginAttempts, cfg.LoginWindow)).Post("/api/v1/login", svc.LoginHandler)
	r.Post("/api/v1/refresh", svc.RefreshHandler)
	r.Post("/api/v1/logout", svc.LogoutHandler)

	// Защищённые маршруты: сначала AuthN (кто ты), затем AuthZ (что тебе можно).
	r.Group(func(priv chi.Router) {
		priv.Use(middleware.AuthN(jwtv))

		priv.With(middleware.AuthZRoles(core.RoleAdmin, core.RoleUser)).Get("/api/v1/me", svc.MeHandler)
		priv.With(middleware.AuthZRoles(core.RoleAdmin)).Get("/api/v1/admin/stats", svc.AdminStats)
		// ABAC: user видит только себя, admin — любого.
		priv.With(middleware.AuthZSelfOrRoles("id", core.RoleAdmin)).Get("/api/v1/users/{id}", svc.UserHandler)
	})
	return r
}
