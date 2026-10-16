# Практическая работа 2

Структура Go-проекта.

Реализовано: cmd/internal/utils, /ping, Request ID, JSON-ошибки.

```bash
# Из корня репозитория
cd practices/02-layout
go run ./cmd/myapp
go build -o bin/app ./cmd/myapp
go vet ./...
go test ./...
```

По умолчанию сервер доступен на http://localhost:8080. Порт меняется через PORT.

[Отчёт по работе](report).
