.PHONY: all check-arch generate build run test test-e2e mock docker-up docker-down clean
SHELL := /bin/bash

all: check-arch test

check-arch:
	@echo "==> Running Architecture Linter..."
	@bash scripts/check_arch.sh

generate:
	@echo "==> Bundling OpenAPI specification..."
	@mkdir -p api/dist
	@if command -v npx >/dev/null 2>&1; then \
		npx --yes @redocly/cli bundle api/openapi/openapi.yaml -o api/dist/openapi.json; \
	elif [ -f node_modules/.bin/redocly ]; then \
		./node_modules/.bin/redocly bundle api/openapi/openapi.yaml -o api/dist/openapi.json; \
	else \
		echo "Warning: redocly cli not found, skipping bundle"; \
	fi
	@echo "==> OpenAPI bundled."
	@echo "==> Generating Go Echo server & types from OpenAPI via oapi-codegen..."
	@cd backend && if command -v oapi-codegen >/dev/null 2>&1; then \
		oapi-codegen --config oapi-codegen.yaml ../api/dist/openapi.json; \
	elif [ -f $(HOME)/go/bin/oapi-codegen ]; then \
		$(HOME)/go/bin/oapi-codegen --config oapi-codegen.yaml ../api/dist/openapi.json; \
	else \
		echo "oapi-codegen not found. Run: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"; exit 1; \
	fi
	@echo "==> Echo server & types generated."
	@echo "==> Generating sqlc code..."
	@cd backend && if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	elif [ -f $(HOME)/go/bin/sqlc ]; then \
		$(HOME)/go/bin/sqlc generate; \
	else \
		echo "sqlc not found. Run: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; exit 1; \
	fi
	@echo "==> sqlc code generated."

mock:
	@echo "==> Generating mocks via mockery..."
	@cd backend && if command -v mockery >/dev/null 2>&1; then \
		mockery; \
	elif [ -f $(HOME)/go/bin/mockery ]; then \
		$(HOME)/go/bin/mockery; \
	else \
		echo "mockery not found. Run: go install github.com/vektra/mockery/v3@latest"; exit 1; \
	fi
	@echo "==> Mocks generated."

build:
	@echo "==> Building Backend..."
	@cd backend && go build -o bin/server ./cmd/server
	@echo "==> Building Frontend..."
	@cd frontend && npm run build

run:
	@echo "==> Running Backend..."
	@cd backend && go run ./cmd/server

test:
	@echo "==> Running Go unit tests..."
	@cd backend && go test -v -race ./...

test-e2e:
	@echo "==> Running Pytest E2E suite via uv..."
	@if command -v uv >/dev/null 2>&1; then \
		uv run pytest tests/ -v; \
	else \
		python3 -m pytest tests/ -v; \
	fi

docker-up:
	@docker compose up -d --build

docker-down:
	@docker compose down

clean:
	@rm -rf backend/bin frontend/dist api/dist .pytest_cache

