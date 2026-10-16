# Технологии индустриального программирования

Практические работы 1–6. **Арефьев К. В., ЭФМО-01-26.**
Среда выполнения — Linux (WSL 2). Проверено на Go 1.27.2 и PostgreSQL 16.15.

| № | Работа и исходники | Реализовано | Отчёт |
| :-: | :--- | :--- | :--- |
| 1 | [HTTP-сервер и UUID](practices/01-hello) | /hello, /user, UUID, /health, APP_PORT | [Отчёт по работе](practices/01-hello/report) |
| 2 | [Структура Go-проекта](practices/02-layout) | cmd/internal/utils, /ping, Request ID, JSON-ошибки | [Отчёт по работе](practices/02-layout/report) |
| 3 | [HTTP API на net/http](practices/03-http) | GET/POST, фильтр q, валидация, логи, graceful shutdown | [Отчёт по работе](practices/03-http/report) |
| 4 | [CRUD-сервис на chi](practices/04-crud) | CRUD, CORS, пагинация, фильтр done | [Отчёт по работе](practices/04-crud/report) |
| 5 | [PostgreSQL и database/sql](practices/05-postgres) | INSERT/SELECT, поиск, фильтр, транзакции, пул соединений | [Отчёт по работе](practices/05-postgres/report) |
| 6 | [GORM, миграции и связи](practices/06-gorm) | User/Note/Tag, AutoMigrate, 1:N и M:N, Preload | [Отчёт по работе](practices/06-gorm/report) |

Для запуска: Go 1.25+; практики 5–6 дополнительно требуют PostgreSQL.
Команды, переменные окружения и проверочные запросы приведены в README каждой работы.
Все работы собираются отдельно. Зависимости зафиксированы в go.mod и go.sum.
