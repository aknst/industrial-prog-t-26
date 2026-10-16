package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/aknst/industrial-prog-t-26/practices/06-gorm/internal/models"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Handlers struct{ db *gorm.DB }

func write(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, code int, message string) {
	write(w, code, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		fail(w, 400, "Content-Type must be application/json")
		return false
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		fail(w, 400, "invalid json")
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		fail(w, 400, "expected one JSON object")
		return false
	}
	return true
}
func dbError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gorm.ErrDuplicatedKey):
		fail(w, 409, "unique value already exists")
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		fail(w, 400, "user does not exist")
	case errors.Is(err, gorm.ErrRecordNotFound):
		fail(w, 404, "note not found")
	default:
		fail(w, 500, "database operation failed")
	}
}
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	pool, err := h.db.DB()
	if err != nil || pool.PingContext(r.Context()) != nil {
		fail(w, 503, "database unavailable")
		return
	}
	write(w, 200, map[string]string{"status": "ok"})
}

type userDTO struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type tagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
type noteDTO struct {
	ID      uint     `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	UserID  uint     `json:"userId"`
	User    userDTO  `json:"user"`
	Tags    []tagDTO `json:"tags"`
}

func toNote(n models.Note) noteDTO {
	tags := make([]tagDTO, 0, len(n.Tags))
	for _, t := range n.Tags {
		tags = append(tags, tagDTO{t.ID, t.Name})
	}
	return noteDTO{n.ID, n.Title, n.Content, n.UserID, userDTO{n.User.ID, n.User.Name, n.User.Email}, tags}
}
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Name == "" || !strings.Contains(in.Email, "@") || utf8.RuneCountInString(in.Name) > 100 || len(in.Email) > 200 {
		fail(w, 400, "valid name and email are required")
		return
	}
	u := models.User{Name: in.Name, Email: in.Email}
	if err := h.db.WithContext(r.Context()).Create(&u).Error; err != nil {
		dbError(w, err)
		return
	}
	write(w, 201, userDTO{u.ID, u.Name, u.Email})
}
func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		UserID  uint     `json:"userId"`
		Tags    []string `json:"tags"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" || utf8.RuneCountInString(in.Title) > 200 || in.UserID == 0 || len(in.Tags) > 20 {
		fail(w, 400, "title and userId are required; at most 20 tags")
		return
	}
	names := make([]string, 0)
	seen := map[string]bool{}
	for _, name := range in.Tags {
		name = strings.TrimSpace(name)
		if name == "" || utf8.RuneCountInString(name) > 50 {
			fail(w, 400, "tag must contain 1 to 50 characters")
			return
		}
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	note := models.Note{Title: in.Title, Content: in.Content, UserID: in.UserID}
	err := h.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, in.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrForeignKeyViolated
			}
			return err
		}
		for _, name := range names {
			tag := models.Tag{Name: name}
			if err := tx.Where("name = ?", name).FirstOrCreate(&tag).Error; err != nil {
				return err
			}
			note.Tags = append(note.Tags, tag)
		}
		if err := tx.Create(&note).Error; err != nil {
			return err
		}
		return tx.Preload("User").Preload("Tags").First(&note, note.ID).Error
	})
	if err != nil {
		dbError(w, err)
		return
	}
	w.Header().Set("Location", "/notes/"+strconv.FormatUint(uint64(note.ID), 10))
	write(w, 201, toNote(note))
}
func (h *Handlers) GetNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil || id == 0 {
		fail(w, 400, "invalid id")
		return
	}
	var note models.Note
	if err := h.db.WithContext(r.Context()).Preload("User").Preload("Tags").First(&note, uint(id)).Error; err != nil {
		dbError(w, err)
		return
	}
	write(w, 200, toNote(note))
}
