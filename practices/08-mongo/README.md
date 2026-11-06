# Практическая работа 8

MongoDB: подключение, коллекция, CRUD-операции.

Реализовано: коллекция notes, уникальный индекс по title, CRUD (POST/GET/PATCH/DELETE), поиск по заголовку, пагинация limit/skip и по курсору (`after`), aggregation `/stats`, таймаут 5 с на каждый запрос к БД.

Требуются Go 1.25+, Docker (или локальная MongoDB 7), curl.

```bash
# Из корня репозитория
cd practices/08-mongo
docker compose up -d
cp .env.example .env
set -a; source .env; set +a
go run ./cmd/api
go build -o bin/app ./cmd/api
go vet ./...
go test ./...
```

Тесты HTTP-слоя работают без базы. Тесты репозитория выполняются, если задан `MONGO_URI` (для каждого теста создаётся и удаляется отдельная база `pz8_test_<время>`), иначе пропускаются:

```bash
MONGO_URI='mongodb://root:secret@localhost:27017/?authSource=admin' go test ./...
```

Переменные окружения: `MONGO_URI`, `MONGO_DB` (pz8), `HTTP_ADDR` (127.0.0.1:8080).

## Запросы

```bash
# создать (201)
curl -s -X POST localhost:8080/api/v1/notes -H 'Content-Type: application/json' \
  -d '{"title":"First note","content":"Hello Mongo!"}'
# список: поиск, лимит, смещение
curl -s 'localhost:8080/api/v1/notes?limit=5&skip=0&q=first'
# пагинация по курсору: заметки старше указанного id
curl -s 'localhost:8080/api/v1/notes?limit=5&after=<id>'
# получить по id
curl -s localhost:8080/api/v1/notes/<id>
# частично обновить
curl -s -X PATCH localhost:8080/api/v1/notes/<id> -H 'Content-Type: application/json' -d '{"content":"Updated content"}'
# удалить (204)
curl -i -X DELETE localhost:8080/api/v1/notes/<id>
# статистика (aggregation)
curl -s localhost:8080/api/v1/notes/stats
```

Коды ответов: 201 создано, 400 неверный JSON или пустой title, 404 заметка не найдена или `id` не ObjectID, 409 заголовок уже существует (уникальный индекс), 504 истёк таймаут запроса к БД.

## Типовые проблемы

- `connection refused` / `server selection timeout` — контейнер не запущен или порт 27017 не проброшен: `docker compose ps`.
- `Authentication failed` — используйте `?authSource=admin` и пару root/secret из docker-compose.yml.
- `409 title_already_exists` — сработал уникальный индекс по title.
- Неверный ObjectID — API возвращает 404.

[Отчёт по работе](report).
