# Технологии индустриального программирования

Практические работы 1–12. **Арефьев К. В., ЭФМО-01-26.**
Среда выполнения — Linux (WSL 2). Проверено на Go 1.27.2, PostgreSQL 16.15, Redis 7 и MongoDB 7.

| № | Работа и исходники | Реализовано | Отчёт |
| :-: | :--- | :--- | :--- |
| 1 | [HTTP-сервер и UUID](practices/01-hello) | /hello, /user, UUID, /health, APP_PORT | [Отчёт по работе](practices/01-hello/report) |
| 2 | [Структура Go-проекта](practices/02-layout) | cmd/internal/utils, /ping, Request ID, JSON-ошибки | [Отчёт по работе](practices/02-layout/report) |
| 3 | [HTTP API на net/http](practices/03-http) | GET/POST, фильтр q, валидация, логи, graceful shutdown | [Отчёт по работе](practices/03-http/report) |
| 4 | [CRUD-сервис на chi](practices/04-crud) | CRUD, CORS, пагинация, фильтр done | [Отчёт по работе](practices/04-crud/report) |
| 5 | [PostgreSQL и database/sql](practices/05-postgres) | INSERT/SELECT, поиск, фильтр, транзакции, пул соединений | [Отчёт по работе](practices/05-postgres/report) |
| 6 | [GORM, миграции и связи](practices/06-gorm) | User/Note/Tag, AutoMigrate, 1:N и M:N, Preload | [Отчёт по работе](practices/06-gorm/report) |
| 7 | [Redis: SET/GET, TTL и кэш](practices/07-redis) | /set, /get, /ttl, /del, cache-aside /product, таймауты | [Отчёт по работе](practices/07-redis/report) |
| 8 | [MongoDB: подключение и CRUD](practices/08-mongo) | CRUD notes, уникальный индекс, поиск, курсорная пагинация, aggregation | [Отчёт по работе](practices/08-mongo/report) |
| 9 | [Регистрация и вход, bcrypt](practices/09-auth) | /auth/register, /auth/login, bcrypt, GORM, одинаковый ответ при ошибках входа | [Отчёт по работе](practices/09-auth/report) |
| 10 | [JWT и middleware](practices/10-jwt) | access/refresh, AuthN/AuthZ, RBAC, ABAC, blacklist, rate limit | [Отчёт по работе](practices/10-jwt/report) |
| 11 | [REST API: CRUD заметок](practices/11-notes-api) | handler → service → repository, пагинация, OpenAPI-файл | [Отчёт по работе](practices/11-notes-api/report) |
| 12 | [Swagger/OpenAPI](practices/12-swagger) | swag init, Swagger UI на /docs, ReDoc, Makefile | [Отчёт по работе](practices/12-swagger/report) |

Для запуска: Go 1.25+; практики 5–6 и 9 дополнительно требуют PostgreSQL, практика 7 — Redis, практика 8 — MongoDB (в README этих работ есть команды Docker).
Команды, переменные окружения и проверочные запросы приведены в README каждой работы.
Все работы собираются отдельно. Зависимости зафиксированы в go.mod и go.sum.
