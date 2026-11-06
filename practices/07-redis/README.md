# Практическая работа 7

Подключение и работа с Redis (SET/GET, TTL). Простой кэш.

Реализовано: пакет `internal/cache` на go-redis v9 (Set, Get, TTL, Del, cache-aside `GetOrLoad`), эндпоинты /set, /get, /ttl, /del, /product, /health, таймауты на все операции Redis.

Требуется Redis 6+ (локально или в Docker):

```bash
docker run -d --name redis -p 127.0.0.1:6379:6379 redis:7-alpine
```

```bash
# Из корня репозитория
cd practices/07-redis
go run ./cmd/server
go build -o bin/app ./cmd/server
go vet ./...
go test ./...
```

Тесты используют встроенный miniredis и не требуют запущенного Redis.

Переменные окружения (образец в `.env.example`, приложение читает их из окружения процесса): `REDIS_ADDR` (localhost:6379), `REDIS_PASSWORD`, `REDIS_DB` (0), `PORT` (8080).

```bash
curl "http://localhost:8080/set?key=test&value=hello"        # OK: test=hello (TTL 10s)
curl "http://localhost:8080/set?key=test&value=hello&ttl=60" # свой TTL в секундах
curl "http://localhost:8080/get?key=test"                    # VALUE: test=hello
curl "http://localhost:8080/ttl?key=test"                    # TTL for test: 9.9s
curl "http://localhost:8080/product?id=1"                    # source: db (≈300 мс)
curl "http://localhost:8080/product?id=1"                    # source: cache (≈1 мс)
```

После истечения TTL `/get` и `/ttl` возвращают 404.

По умолчанию сервер доступен на http://localhost:8080. Порт меняется через PORT.

[Отчёт по работе](report).
