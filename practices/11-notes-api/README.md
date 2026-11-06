# Практическая работа 11

Проектирование REST API: CRUD для заметок и структура проекта.

Реализовано: слоистая архитектура handler → service → repository, in-memory репозиторий с блокировками, CRUD `/api/v1/notes`, пагинация `page`/`limit` с заголовком `X-Total-Count`, поиск `q`, валидация, JSON-ошибки, спецификация [api/openapi.yaml](api/openapi.yaml).

```bash
# Из корня репозитория
cd practices/11-notes-api
go run ./cmd/api
go build -o bin/app ./cmd/api
go vet ./...
go test ./...
```

По умолчанию сервер доступен на http://localhost:8080. Порт меняется через PORT. Данные хранятся в памяти и пропадают при перезапуске.

```
cmd/api/main.go                      точка входа
internal/http/router.go              маршруты и middleware (chi)
internal/http/handlers/notes.go      HTTP-обработчики
internal/core/note.go                модель и DTO
internal/core/service/note_service.go  бизнес-логика, валидация
internal/repo/note_mem.go            in-memory репозиторий
api/openapi.yaml                     описание API
```

## Запросы

```bash
curl -i -X POST localhost:8080/api/v1/notes -H 'Content-Type: application/json' \
  -d '{"title":"Первая заметка","content":"Это тест"}'                      # 201
curl -i 'localhost:8080/api/v1/notes?page=1&limit=10&q=перв'                # 200 + X-Total-Count
curl -i localhost:8080/api/v1/notes/1                                       # 200
curl -i -X PATCH localhost:8080/api/v1/notes/1 -H 'Content-Type: application/json' \
  -d '{"content":"Обновлено"}'                                              # 200
curl -i -X DELETE localhost:8080/api/v1/notes/1                             # 204
curl -i localhost:8080/api/v1/notes/1                                       # 404
```

Коды: 200, 201 (+ `Location`), 204, 400 (неверный JSON, пустой или слишком длинный title, неверные `id`/`page`/`limit`), 404.

[Отчёт по работе](report).
