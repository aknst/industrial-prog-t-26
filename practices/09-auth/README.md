# Практическая работа 9

Регистрация и вход пользователей. Хэширование паролей с bcrypt.

Реализовано: POST /auth/register и POST /auth/login на chi + GORM, bcrypt (cost из BCRYPT_COST), уникальный email, валидация ввода, одинаковый ответ `invalid_credentials` для неверного пароля и несуществующего email.

Требуется PostgreSQL 14+ и отдельная база industrial_pz09. Задать DB_DSN (см. .env.example); таблица users создаётся через AutoMigrate при запуске.

```bash
# Из корня репозитория
cd practices/09-auth
export DB_DSN='postgres://USER:PASSWORD@127.0.0.1:5432/industrial_pz09?sslmode=disable'
go run ./cmd/api
go build -o bin/app ./cmd/api
go vet ./...
go test ./...
```

Переменные окружения: `DB_DSN` (обязательно), `BCRYPT_COST` (4–31, по умолчанию 12), `APP_ADDR` (127.0.0.1:8080).

Тесты обработчиков используют in-memory хранилище. Тест репозитория выполняется, если задан `TEST_DB_DSN` (отдельная база, таблица users пересоздаётся), иначе пропускается.

```bash
# регистрация -> 201
curl -i -X POST localhost:8080/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Secret123!"}'
# повторная регистрация -> 409
curl -i -X POST localhost:8080/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"AnotherPass"}'
# вход -> 200
curl -i -X POST localhost:8080/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Secret123!"}'
# неверный пароль -> 401
curl -i -X POST localhost:8080/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"wrong"}'
```

Проверка хэша в базе:

```bash
psql "$DB_DSN" -c 'select id, email, left(password_hash, 7) as hash_prefix, length(password_hash) from users'
```

`GET /health` возвращает `{"status":"ok"}` и служит проверкой готовности сервера.

Коды: 201 создан, 400 неверный JSON/email/длина пароля (8–72 байта), 401 неверная пара email+пароль, 409 email занят.

По умолчанию сервер доступен на http://localhost:8080. Адрес меняется через APP_ADDR.

[Отчёт по работе](report).
