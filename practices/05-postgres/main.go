package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func execute() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	db, err := openDB(context.Background(), dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	log.Print("Connected to PostgreSQL")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo := &Repo{DB: db}
	first := 0
	for _, title := range []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"} {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			return err
		}
		if first == 0 {
			first = id
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}
	if err := repo.CreateMany(ctx, []string{"Транзакция: задача A", "Транзакция: задача B"}); err != nil {
		return err
	}
	tasks, err := repo.ListTasks(ctx)
	if err != nil {
		return err
	}
	fmt.Println("=== Tasks ===")
	for _, v := range tasks {
		fmt.Printf("#%d | %s | done=%v | %s\n", v.ID, v.Title, v.Done, v.CreatedAt.Format(time.RFC3339))
	}
	t, err := repo.FindByID(ctx, first)
	if err != nil {
		return err
	}
	fmt.Printf("FindByID(%d): %s\n", first, t.Title)
	pending, err := repo.ListDone(ctx, false)
	if err != nil {
		return err
	}
	fmt.Printf("ListDone(false): %d tasks\n", len(pending))
	start := time.Now()
	for i := 0; i < 20; i++ {
		if err := db.PingContext(ctx); err != nil {
			return err
		}
	}
	stats := db.Stats()
	fmt.Printf("Pool: max=%d open=%d idle=%d waits=%d; 20 sequential pings: %s\n", stats.MaxOpenConnections, stats.OpenConnections, stats.Idle, stats.WaitCount, time.Since(start))
	return nil
}
func main() {
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}
