# Практическая работа 10

JWT-аутентификация: создание и проверка токенов, middleware авторизации.

Реализовано: HS256 JWT (sub, email, role, typ, iss, aud, iat, exp, jti), `/login`, `/refresh` с ротацией и blacklist отозванных refresh-токенов, `/logout`, middleware AuthN (Bearer) и AuthZ (RBAC по ролям), правило ABAC для `/users/{id}`, rate limit входа по IP, единый JSON-формат ошибок и логирование кодов и времени.

```bash
# Из корня репозитория
cd practices/10-jwt
export JWT_SECRET=dev-secret-change-me-please JWT_TTL=15m APP_PORT=8080
go run ./cmd/server
go build -o bin/app ./cmd/server
go vet ./...
go test ./...
```

Переменные окружения (образец в `.env.example`):

| Переменная | По умолчанию | Назначение |
| :--- | :--- | :--- |
| `JWT_SECRET` | — (обязательна) | ключ подписи HS256, в Git не коммитится |
| `JWT_TTL` | 15m | срок жизни access-токена |
| `REFRESH_TTL` | 168h | срок жизни refresh-токена (7 дней) |
| `APP_PORT` | 8080 | порт, сервер слушает 127.0.0.1 |
| `LOGIN_RATE_LIMIT` | 5 | попыток `/login` с одного IP за окно |
| `LOGIN_RATE_WINDOW` | 5m | размер окна rate limit |

Тестовые пользователи (пароль `secret123`): `admin@example.com` (admin, id 1) и `user@example.com` (user, id 2). Хранилище и список отозванных токенов находятся в памяти.

## Маршруты

| Маршрут | Доступ | Результат |
| :--- | :--- | :--- |
| `POST /api/v1/login` | публичный, rate limit | пара access + refresh |
| `POST /api/v1/refresh` | refresh-токен в теле | новая пара, старый refresh отозван |
| `POST /api/v1/logout` | refresh-токен в теле | refresh отозван, 204 |
| `GET /api/v1/me` | admin, user | профиль из клеймов |
| `GET /api/v1/admin/stats` | только admin | 403 для user |
| `GET /api/v1/users/{id}` | admin — любой id, user — только свой | ABAC: 403 для чужого id |

401 — токен отсутствует, неверен или просрочен; 403 — прав недостаточно; 429 — превышен лимит попыток входа.

## Проверка

```bash
curl -s -X POST localhost:8080/api/v1/login -H 'Content-Type: application/json' \
  -d '{"Email":"admin@example.com","Password":"secret123"}'
TOKEN=<access_token из ответа>
curl -s localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"
curl -s localhost:8080/api/v1/admin/stats -H "Authorization: Bearer $TOKEN"

TOKEN_USER=$(curl -s -X POST localhost:8080/api/v1/login -H 'Content-Type: application/json' \
  -d '{"Email":"user@example.com","Password":"secret123"}' | jq -r .access_token)
curl -i localhost:8080/api/v1/admin/stats -H "Authorization: Bearer $TOKEN_USER"   # 403
curl -i localhost:8080/api/v1/users/1 -H "Authorization: Bearer $TOKEN_USER"      # 403
curl -i localhost:8080/api/v1/users/2 -H "Authorization: Bearer $TOKEN_USER"      # 200

curl -s -X POST localhost:8080/api/v1/refresh -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'                                         # новая пара
```

```
cmd/server/main.go                    точка входа
internal/http/router.go               маршруты и цепочка middleware
internal/http/middleware/             authn.go, authz.go, ratelimit.go, logging.go
internal/core/                        user.go, service.go (login, refresh, обработчики)
internal/repo/                        user_mem.go (bcrypt), blacklist_mem.go
internal/platform/jwt/jwt.go          выпуск и проверка токенов
internal/platform/config/config.go    чтение переменных окружения
internal/platform/respond/respond.go  единый формат ответов и ошибок
```

[Отчёт по работе](report).
