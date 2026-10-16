# Практическая работа 5

PostgreSQL и database/sql.

Реализовано: INSERT/SELECT, поиск, фильтр, транзакции, пул соединений.

Требуется PostgreSQL 14+ и база industrial_pz05. Выполнить схему из sql/001_schema.sql и задать DATABASE_URL (см. .env.example).

```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f sql/001_schema.sql
```

```bash
# Из корня репозитория
cd practices/05-postgres
go run .
go build -o bin/app .
go vet ./...
go test ./...
```

Интеграционный тест запускается при TEST_DATABASE_URL, указывающем строго на industrial_pz05_test. Эта тестовая база очищается в начале теста.

[Отчёт по работе](report).
