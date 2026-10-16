# Проверка API

```bash
curl -i http://localhost:8080/health
curl -i -H 'Content-Type: application/json' -d '{"title":"Купить молоко"}' http://localhost:8080/tasks
curl -i http://localhost:8080/tasks
curl -i 'http://localhost:8080/tasks?q=молоко'
curl -i http://localhost:8080/tasks/1
curl -i -H 'Content-Type: application/json' -d '{}' http://localhost:8080/tasks
curl -i http://localhost:8080/tasks/abc
curl -i http://localhost:8080/tasks/9999
```

Ожидаемые статусы: 200, 201, 200, 200, 200, 400, 400, 404.
