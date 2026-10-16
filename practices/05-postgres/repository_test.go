package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"
)

func TestRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for isolated integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := openDB(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var database string
	if err := db.QueryRowContext(ctx, "SELECT current_database()").Scan(&database); err != nil {
		t.Fatal(err)
	}
	if database != "industrial_pz05_test" {
		t.Fatal("integration test requires industrial_pz05_test")
	}
	// A dedicated test database is required; demo runner provides it.
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS tasks(id SERIAL PRIMARY KEY,title TEXT NOT NULL CHECK(length(trim(title))>0),done BOOLEAN NOT NULL DEFAULT false,created_at TIMESTAMPTZ NOT NULL DEFAULT now()); TRUNCATE tasks RESTART IDENTITY`); err != nil {
		t.Fatal(err)
	}
	repo := &Repo{DB: db}
	id, err := repo.CreateTask(ctx, "O'Reilly; DROP TABLE tasks;")
	if err != nil {
		t.Fatal(err)
	}
	task, err := repo.FindByID(ctx, id)
	if err != nil || task.Title != "O'Reilly; DROP TABLE tasks;" {
		t.Fatalf("%+v %v", task, err)
	}
	if err := repo.CreateMany(ctx, []string{"must rollback", ""}); err == nil {
		t.Fatal("expected constraint error")
	}
	list, err := repo.ListTasks(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("transaction partially committed: %v %v", list, err)
	}
	if _, err := repo.FindByID(ctx, 99999); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	done, err := repo.ListDone(ctx, true)
	if err != nil || len(done) != 0 {
		t.Fatal(done, err)
	}
}
