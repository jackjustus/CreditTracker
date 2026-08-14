DATABASE_URL ?= postgres://postgres:testtest@localhost:5432/postgres?sslmode=disable

COMPOSE := docker compose -f backend/compose.yaml

# Pinned so `verify-gen` cannot fail on a generator version bump alone: the
# version string is baked into every generated file's header comment.
SQLC_VERSION ?= v1.31.1
OAPI_VERSION ?= v2.8.0

# Only the `migration` target shells out to goose; the stack migrates itself.
export GOOSE_DRIVER := postgres
export GOOSE_DBSTRING := $(DATABASE_URL)
export GOOSE_MIGRATION_DIR := backend/db/migrations

.PHONY: help api rebuild-api watch-api stop-api logs psql install tools generate check test lint verify-gen migration

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  \033[36m%-11s\033[0m %s\n", $$1, $$2}'

api: ## Start Postgres, migrate, seed, and serve the API on :8080
	$(COMPOSE) up -d --build

rebuild-api: ## Rebuild and restart just the API (leaves Postgres running)
	$(COMPOSE) up -d --build api-service

watch-api: ## Rebuild the API automatically whenever Go source changes
	$(COMPOSE) watch api-service

stop-api: ## Stop the stack (data preserved)
	$(COMPOSE) down

logs: ## Follow the API logs
	$(COMPOSE) logs -f api-service

psql: ## Open a psql shell
	$(COMPOSE) exec db psql -U postgres

install: ## Install the coaster CLI to $$GOPATH/bin
	cd cli && go install ./cmd/coaster

tools: ## Install the pinned code generators into $$GOPATH/bin
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_VERSION)

generate: ## Regenerate sqlc queries and the OpenAPI server/client
	cd backend && sqlc generate
	oapi-codegen --config=oapi-codegen.server.yaml ./openapi.yaml
	oapi-codegen --config=oapi-codegen.client.yaml ./openapi.yaml

# Builds and vets first: golangci-lint reports a package that does not compile
# as an inscrutable typecheck error, so let the compiler say it plainly.
check: ## Build, vet, and lint both modules
	cd backend && go build ./... && go vet ./...
	cd cli && go build ./... && go vet ./...
	$(MAKE) lint

test: ## Run the unit tests for both modules
	cd backend && go test ./...
	cd cli && go test ./...

lint: ## Run golangci-lint on both modules
	cd backend && golangci-lint run
	cd cli && golangci-lint run

# Uses `git status --porcelain`, not `git diff`: the latter ignores untracked
# files, so a newly generated .sql.go that was never committed slips past it.
verify-gen: generate check ## CI: fail if generated code is not committed
	@test -z "$$(git status --porcelain -- backend/internal/gen cli/internal/gen)" || ( \
		git status --short -- backend/internal/gen cli/internal/gen; \
		echo "generated code is out of date: run 'make generate' and commit the result"; \
		exit 1)

migration: ## Create a migration: make migration name=add_something (needs goose)
ifndef name
	$(error usage: make migration name=add_something)
endif
	goose create $(name) sql
