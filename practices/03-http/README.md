# Практическая работа 3

HTTP API на net/http.

Реализовано: GET/POST, фильтр q, валидация, логи, graceful shutdown.

```bash
# Из корня репозитория
cd practices/03-http
go run ./cmd/server
go build -o bin/app ./cmd/server
go vet ./...
go test ./...
```

По умолчанию сервер доступен на http://localhost:8080. Порт меняется через PORT.

Примеры проверочных запросов: [requests.md](requests.md).

[Отчёт по работе](report).
