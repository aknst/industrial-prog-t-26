package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}
type Repo struct{ DB *sql.DB }

func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx, `INSERT INTO tasks(title) VALUES($1) RETURNING id`, title).Scan(&id)
	return id, err
}
func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	return t, err
}
func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
	t, err := scanTask(r.DB.QueryRowContext(ctx, `SELECT id,title,done,created_at FROM tasks WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	return &t, nil
}
func collect(rows *sql.Rows) ([]Task, error) {
	defer rows.Close()
	out := make([]Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,title,done,created_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return collect(rows)
}
func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,title,done,created_at FROM tasks WHERE done=$1 ORDER BY id`, done)
	if err != nil {
		return nil, err
	}
	return collect(rows)
}
func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, title := range titles {
		if _, err := tx.ExecContext(ctx, `INSERT INTO tasks(title) VALUES($1)`, title); err != nil {
			return fmt.Errorf("batch insert: %w", err)
		}
	}
	return tx.Commit()
}
