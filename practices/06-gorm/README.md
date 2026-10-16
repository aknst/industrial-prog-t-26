# Практическая работа 6

GORM, миграции и связи.

Реализовано: User/Note/Tag, AutoMigrate, 1:N и M:N, Preload.

Требуется PostgreSQL 14+ и отдельная база industrial_pz06. Задать DB_DSN (см. .env.example); таблицы создаются через AutoMigrate при запуске.

```bash
# Из корня репозитория
cd practices/06-gorm
go run ./cmd/server
go build -o bin/app ./cmd/server
go vet ./...
go test ./...
```

По умолчанию сервер доступен на http://localhost:8080. Порт меняется через PORT.

[Отчёт по работе](report).
