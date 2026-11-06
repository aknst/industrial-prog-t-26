package core

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/jwt"
	"github.com/aknst/industrial-prog-t-26/practices/10-jwt/internal/platform/respond"
)

type UserStore interface {
	CheckPassword(email, pass string) (User, error)
	ByID(id int64) (User, error)
}

// Blacklist хранит отозванные refresh-токены (по jti) до их естественного истечения.
type Blacklist interface {
	// Revoke помечает токен отозванным; false, если он уже был отозван.
	Revoke(jti string, exp time.Time) bool
}

type Service struct {
	users   UserStore
	tokens  *jwt.HS256
	revoked Blacklist
}

func NewService(u UserStore, t *jwt.HS256, b Blacklist) *Service {
	return &Service{users: u, tokens: t, revoked: b}
}

type tokenPair struct {
	Token        string `json:"token"` // совместимость с ПЗ: то же, что access_token
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *Service) issue(u User) (tokenPair, error) {
	access, _, err := s.tokens.SignAccess(u.ID, u.Email, u.Role)
	if err != nil {
		return tokenPair{}, err
	}
	refresh, _, err := s.tokens.SignRefresh(u.ID, u.Email, u.Role)
	if err != nil {
		return tokenPair{}, err
	}
	return tokenPair{Token: access, AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: int(s.tokens.AccessTTL().Seconds())}, nil
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(v); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

// LoginHandler: POST /api/v1/login. Поля JSON регистронезависимы (Email и email).
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if !decode(w, r, &in) {
		return
	}
	if in.Email == "" || in.Password == "" {
		respond.Error(w, http.StatusBadRequest, "invalid_credentials", "email and password are required")
		return
	}
	u, err := s.users.CheckPassword(in.Email, in.Password)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	pair, err := s.issue(u)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "token_error")
		return
	}
	respond.JSON(w, http.StatusOK, pair)
}

// RefreshHandler: POST /api/v1/refresh. Старый refresh-токен отзывается (ротация),
// повторное использование того же токена даёт 401.
func (s *Service) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &in) {
		return
	}
	claims, err := s.tokens.Parse(in.RefreshToken, jwt.TypeRefresh)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid_refresh_token")
		return
	}
	id, err := claims.UserID()
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid_refresh_token")
		return
	}
	if !s.revoked.Revoke(claims.ID, claims.ExpiresAt.Time) {
		respond.Error(w, http.StatusUnauthorized, "refresh_token_revoked")
		return
	}
	// Роль берём из хранилища, а не из старого токена: изменение прав вступает в силу при обновлении.
	u, err := s.users.ByID(id)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid_refresh_token")
		return
	}
	pair, err := s.issue(u)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "token_error")
		return
	}
	respond.JSON(w, http.StatusOK, pair)
}

// LogoutHandler: POST /api/v1/logout отзывает переданный refresh-токен.
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &in) {
		return
	}
	claims, err := s.tokens.Parse(in.RefreshToken, jwt.TypeRefresh)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid_refresh_token")
		return
	}
	s.revoked.Revoke(claims.ID, claims.ExpiresAt.Time)
	w.WriteHeader(http.StatusNoContent)
}

// MeHandler возвращает профиль из клеймов токена (AuthN уже положил их в контекст).
func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	c, ok := jwt.ClaimsFrom(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, _ := c.UserID()
	respond.JSON(w, http.StatusOK, map[string]any{"id": id, "email": c.Email, "role": c.Role})
}

func (s *Service) AdminStats(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]any{"users": 2, "version": "1.0"})
}

// UserHandler: GET /api/v1/users/{id}. Правило ABAC (владелец или admin) проверяет
// middleware AuthZSelfOrRoles; здесь только поиск пользователя.
func (s *Service) UserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid_id")
		return
	}
	u, err := s.users.ByID(id)
	if errors.Is(err, ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, u)
}
