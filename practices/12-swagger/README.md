# Практическая работа 12

Swagger/OpenAPI: автоматическая генерация документации для API заметок из ПЗ 11.

Подход code-first: аннотации swaggo/swag над обработчиками и в `cmd/api/main.go`, генерация `docs/`, Swagger UI на `/docs`, ReDoc на `/redoc`, экспорт спецификации в `api/swagger.yaml`.

Требуется Go 1.25+ и утилита swag:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

```bash
# Из корня репозитория
cd practices/12-swagger
make swagger      # swag init + копия спецификации в api/swagger.yaml
make run          # go run ./cmd/api
make test         # go vet ./... && go test ./...
```

Без make: `swag init -g main.go -d cmd/api,internal/core,internal/http/handlers -o docs --parseInternal` (каталоги поиска перечислены явно, чтобы swag не предупреждал о каталогах без Go-файлов).

После запуска:

- Swagger UI — http://localhost:8080/docs/index.html (`/docs` перенаправляет сюда)
- спецификация JSON — http://localhost:8080/docs/doc.json
- ReDoc — http://localhost:8080/redoc (скрипт ReDoc подгружается с cdn.jsdelivr.net)

Порт меняется через PORT. Каталог `docs/` создаётся генератором и вручную не редактируется; после изменения аннотаций выполните `make swagger`. Тест `TestDocsAreServed` проверяет, что в спецификации описаны все пять методов `/notes` и схемы Note, NoteCreate, NoteUpdate, ErrorResponse, поэтому забытая регенерация или удалённая аннотация заметны при `go test`.

Генератор swag v1 выпускает спецификацию Swagger 2.0; для OpenAPI 3 используется ручной `api/openapi.yaml` из ПЗ 11 либо swag v2.

```
cmd/api/main.go                   общие аннотации (@title, @version, @BasePath)
internal/http/handlers/notes.go   аннотации методов (@Summary, @Param, @Success, @Failure, @Router)
internal/core/note.go             модели с example-тегами
docs/                             docs.go, swagger.json, swagger.yaml (генерируется)
api/swagger.yaml                  экспорт спецификации
Makefile                          run, swagger, test
```

[Отчёт по работе](report).
