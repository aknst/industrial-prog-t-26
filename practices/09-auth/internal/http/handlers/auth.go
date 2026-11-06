package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/09-auth/internal/repo"
)

const (
	minPasswordLen = 8
	// bcrypt учитывает только первые 72 байта пароля.
	maxPasswordLen = 72
)

// UserStore — то, что нужно обработчикам от хранилища пользователей.
type UserStore interface {
	Create(ctx context.Context, u *core.User) error
	ByEmail(ctx context.Context, email string) (core.User, error)
}

type AuthHandler struct {
	Users      UserStore
	BcryptCost int

	// dummyHash сравнивается с паролем, когда email не найден, чтобы время ответа
	// не выдавало существование пользователя.
	dummyHash []byte
}

func NewAuthHandler(users UserStore, cost int) *AuthHandler {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("not-a-real-password"), cost)
	return &AuthHandler{Users: users, BcryptCost: cost, dummyHash: dummy}
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResp struct {
	Status string `json:"status"`
	User   any    `json:"user,omitempty"`
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func normalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decode(w, r, &in) {
		return
	}
	in.Email = normalizeEmail(in.Email)
	if _, err := mail.ParseAddress(in.Email); err != nil || len(in.Email) > 255 {
		writeErr(w, http.StatusBadRequest, "invalid_email")
		return
	}
	if len(in.Password) < minPasswordLen || len(in.Password) > maxPasswordLen {
		writeErr(w, http.StatusBadRequest, "password_length_8_to_72")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), h.BcryptCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash_failed")
		return
	}
	u := core.User{Email: in.Email, PasswordHash: string(hash)}
	if err := h.Users.Create(r.Context(), &u); err != nil {
		if errors.Is(err, repo.ErrEmailTaken) {
			writeErr(w, http.StatusConflict, "email_taken")
			return
		}
		writeErr(w, http.StatusInternalServerError, "db_error")
		return
	}
	writeJSON(w, http.StatusCreated, authResp{Status: "ok", User: map[string]any{"id": u.ID, "email": u.Email}})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decode(w, r, &in) {
		return
	}
	in.Email = normalizeEmail(in.Email)
	if in.Email == "" || in.Password == "" {
		writeErr(w, http.StatusBadRequest, "email_and_password_required")
		return
	}

	u, err := h.Users.ByEmail(r.Context(), in.Email)
	if err != nil {
		if !errors.Is(err, repo.ErrUserNotFound) {
			writeErr(w, http.StatusInternalServerError, "db_error")
			return
		}
		_ = bcrypt.CompareHashAndPassword(h.dummyHash, []byte(in.Password))
		writeErr(w, http.StatusUnauthorized, "invalid_credentials")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "invalid_credentials")
		return
	}
	// В ПЗ 10 здесь выдаётся JWT; пока возвращаем только статус.
	writeJSON(w, http.StatusOK, authResp{Status: "ok", User: map[string]any{"id": u.ID, "email": u.Email}})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
