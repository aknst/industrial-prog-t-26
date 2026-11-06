SWAG ?= $(shell command -v swag 2>/dev/null || echo $(HOME)/go/bin/swag)

.PHONY: run swagger check-docs test

run:
	go run ./cmd/api

# Пересоздаёт docs/ из аннотаций и экспортирует спецификацию в api/swagger.yaml.
swagger:
	$(SWAG) init -g main.go -d cmd/api,internal/core,internal/http/handlers -o docs --parseInternal
	mkdir -p api
	cp docs/swagger.yaml api/swagger.yaml

# Падает, если docs/ не соответствует аннотациям (для CI).
check-docs: swagger
	git diff --exit-code -- docs api || (echo "docs are stale: run make swagger"; exit 1)

test:
	go vet ./...
	go test ./...
