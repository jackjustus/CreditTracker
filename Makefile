DATABASE_URL ?= postgres://postgres:testtest@localhost:5432/postgres?sslmode=disable

COMPOSE := docker compose -f backend/compose.yaml

# Only the `migration` target shells out to goose; the stack migrates itself.
export GOOSE_DRIVER := postgres
export GOOSE_DBSTRING := $(DATABASE_URL)
export GOOSE_MIGRATION_DIR := backend/db/migrations

.PHONY: help api stop-api logs psql install generate check migration

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  \033[36m%-11s\033[0m %s\n", $$1, $$2}'

api: ## Start Postgres, migrate, seed, and serve the API on :8080
	$(COMPOSE) up -d --build

stop-api: ## Stop the stack (data preserved)
	$(COMPOSE) down

logs: ## Follow the API logs
	$(COMPOSE) logs -f api-service

psql: ## Open a psql shell
	$(COMPOSE) exec db psql -U postgres

install: ## Install the coaster CLI to $$GOPATH/bin
	cd cli && go install ./cmd/coaster

generate: ## Regenerate sqlc queries and the OpenAPI server/client
	cd backend && sqlc generate
	oapi-codegen --config=oapi-codegen.server.yaml ./openapi.yaml
	oapi-codegen --config=oapi-codegen.client.yaml ./openapi.yaml

check: generate ## Build, vet, and verify generated code is current
	cd backend && go build ./... && go vet ./...
	cd cli && go build ./... && go vet ./...
	git diff --exit-code -- backend/internal/gen cli/internal/gen || (echo "generated code is stale: commit it" && exit 1)

migration: ## Create a migration: make migration name=add_something (needs goose)
ifndef name
	$(error usage: make migration name=add_something)
endif
	goose create $(name) sql
