package core

import "time"

// Note — доменная сущность «Заметка».
type Note struct {
	ID        int64      `json:"id" example:"1"`
	Title     string     `json:"title" example:"Идея для лекции"`
	Content   string     `json:"content" example:"Добавить пример структуры API"`
	CreatedAt time.Time  `json:"createdAt" example:"2025-10-18T12:00:00Z"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" example:"2025-10-19T09:30:00Z"`
}

// NoteCreate — тело запроса POST /notes.
type NoteCreate struct {
	Title   string `json:"title" example:"Новая заметка"`
	Content string `json:"content" example:"Текст заметки"`
}

// NoteUpdate — тело запроса PATCH /notes/{id}; указанные поля заменяются, остальные не меняются.
type NoteUpdate struct {
	Title   *string `json:"title,omitempty" example:"Обновлено"`
	Content *string `json:"content,omitempty" example:"Новый текст"`
}

// ListParams — параметры списка: страница (с 1), размер страницы и подстрока заголовка.
type ListParams struct {
	Page  int
	Limit int
	Q     string
}
