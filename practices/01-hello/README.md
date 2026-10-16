# Практическая работа 1

HTTP-сервер и UUID.

Реализовано: /hello, /user, UUID, /health, APP_PORT.

```bash
# Из корня репозитория
cd practices/01-hello
go run ./cmd/server
go build -o bin/app ./cmd/server
go vet ./...
go test ./...
```

По умолчанию сервер доступен на http://localhost:8080. Порт меняется через APP_PORT.

[Отчёт по работе](report).
