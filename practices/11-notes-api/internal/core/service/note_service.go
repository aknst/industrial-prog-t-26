package service

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/core"
	"github.com/aknst/industrial-prog-t-26/practices/11-notes-api/internal/repo"
)

const (
	MaxTitleLen   = 200
	MaxContentLen = 10000
	DefaultLimit  = 20
	MaxLimit      = 100
)

var (
	ErrNotFound   = errors.New("note not found")
	ErrValidation = errors.New("validation failed")
)

// ValidationError описывает, какое поле и почему не прошло проверку.
type ValidationError struct{ Field, Msg string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }
func (e *ValidationError) Unwrap() error { return ErrValidation }

type NoteRepo interface {
	Create(core.Note) (core.Note, error)
	Get(id int64) (core.Note, error)
	List(core.ListParams) ([]core.Note, int, error)
	Update(id int64, apply func(*core.Note)) (core.Note, error)
	Delete(id int64) error
}

// NoteService содержит бизнес-логику: валидацию и правила пагинации.
type NoteService struct{ repo NoteRepo }

func NewNoteService(r NoteRepo) *NoteService { return &NoteService{repo: r} }

func (s *NoteService) Create(in core.NoteCreate) (core.Note, error) {
	title, err := validateTitle(in.Title)
	if err != nil {
		return core.Note{}, err
	}
	if err := validateContent(in.Content); err != nil {
		return core.Note{}, err
	}
	return s.repo.Create(core.Note{Title: title, Content: in.Content, CreatedAt: time.Now().UTC()})
}

func (s *NoteService) Get(id int64) (core.Note, error) {
	n, err := s.repo.Get(id)
	return n, mapErr(err)
}

// List проверяет параметры (page >= 1, 1 <= limit <= MaxLimit) и возвращает страницу и общее число.
// Значения по умолчанию (page=1, limit=DefaultLimit) подставляет вызывающий код.
func (s *NoteService) List(p core.ListParams) ([]core.Note, int, error) {
	if p.Page < 1 {
		return nil, 0, &ValidationError{"page", "must be >= 1"}
	}
	if p.Limit < 1 || p.Limit > MaxLimit {
		return nil, 0, &ValidationError{"limit", "must be from 1 to 100"}
	}
	items, total, err := s.repo.List(p)
	return items, total, mapErr(err)
}

func (s *NoteService) Update(id int64, in core.NoteUpdate) (core.Note, error) {
	if in.Title == nil && in.Content == nil {
		return core.Note{}, &ValidationError{"body", "at least one of title, content is required"}
	}
	var title string
	if in.Title != nil {
		t, err := validateTitle(*in.Title)
		if err != nil {
			return core.Note{}, err
		}
		title = t
	}
	if in.Content != nil {
		if err := validateContent(*in.Content); err != nil {
			return core.Note{}, err
		}
	}
	n, err := s.repo.Update(id, func(n *core.Note) {
		if in.Title != nil {
			n.Title = title
		}
		if in.Content != nil {
			n.Content = *in.Content
		}
	})
	return n, mapErr(err)
}

func (s *NoteService) Delete(id int64) error { return mapErr(s.repo.Delete(id)) }

func validateTitle(t string) (string, error) {
	t = strings.TrimSpace(t)
	if t == "" {
		return "", &ValidationError{"title", "is required"}
	}
	if utf8.RuneCountInString(t) > MaxTitleLen {
		return "", &ValidationError{"title", "must be at most 200 characters"}
	}
	return t, nil
}

func validateContent(c string) error {
	if utf8.RuneCountInString(c) > MaxContentLen {
		return &ValidationError{"content", "must be at most 10000 characters"}
	}
	return nil
}

func mapErr(err error) error {
	if errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
