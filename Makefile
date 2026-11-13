DATABASE_URL ?= postgres://postgres:testtest@localhost:5432/postgres?sslmode=disable

COMPOSE := docker compose -f backend/compose.yaml

# Pinned so `verify-gen` cannot fail on a generator version bump alone: the
# version string is baked into every generated file's header comment.
SQLC_VERSION ?= v1.31.1
OAPI_VERSION ?= v2.8.0
# Keep in sync with GOLANGCI_VERSION in .github/workflows/ci.yml.
GOLANGCI_VERSION ?= v2.15.0

# Only the `migration` target shells out to goose; the stack migrates itself.
export GOOSE_DRIVER := postgres
export GOOSE_DBSTRING := $(DATABASE_URL)
export GOOSE_MIGRATION_DIR := backend/db/migrations

.PHONY: help api rebuild-api watch-api stop-api logs psql seed install tools generate build check test lint verify-gen migration

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

seed: ## Re-apply backend/db/seed/seed.sql (idempotent; `make api` seeds only an empty DB)
	$(COMPOSE) exec -T db psql -U postgres -v ON_ERROR_STOP=1 --single-transaction < backend/db/seed/seed.sql

install: ## Install the coaster CLI to $$GOPATH/bin
	cd cli && go install ./cmd/coaster

tools: ## Install the pinned code generators and linter into $$GOPATH/bin
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

generate: ## Regenerate sqlc queries and the OpenAPI server/client
	@for t in sqlc oapi-codegen; do \
		command -v $$t >/dev/null || { echo "$$t not found on PATH: run 'make tools'"; exit 1; }; \
	done
	cd backend && sqlc generate
	oapi-codegen --config=oapi-codegen.server.yaml ./openapi.yaml
	oapi-codegen --config=oapi-codegen.client.yaml ./openapi.yaml

build: ## Build and vet both modules
	cd backend && go build ./... && go vet ./...
	cd cli && go build ./... && go vet ./...

check: build ## Build, vet, test, lint, and verify generated code (mirrors CI)
	$(MAKE) test
	$(MAKE) lint
	$(MAKE) verify-gen

test: ## Run the unit tests for both modules
	cd backend && go test ./...
	cd cli && go test ./...

lint: ## Run golangci-lint on both modules
	cd backend && golangci-lint run
	cd cli && golangci-lint run

# Checksums the working tree before and after regenerating, so it fails only
# when `generate` would change something, not merely because correct output is
# uncommitted. In CI the checkout is clean, so this still catches generated code
# that was never committed (including brand-new untracked files).
# Does not build or lint: if this passes, the bytes are what `build` already ran on.
GEN_DIRS := backend/internal/gen cli/internal/gen
GEN_SUMS = find $(GEN_DIRS) -type f | LC_ALL=C sort | xargs shasum
verify-gen: ## Fail if generated code is stale relative to its sources
	@$(GEN_SUMS) > .gen-before.sum
	@$(MAKE) --no-print-directory generate
	@$(GEN_SUMS) > .gen-after.sum
	@diff .gen-before.sum .gen-after.sum >/dev/null; status=$$?; \
	if [ $$status -ne 0 ]; then \
		diff .gen-before.sum .gen-after.sum | sed -n 's/^> [0-9a-f]*  /  changed: /p'; \
		echo "generated code was stale and has been regenerated: review and commit the result"; \
	fi; \
	rm -f .gen-before.sum .gen-after.sum; exit $$status

migration: ## Create a migration: make migration name=add_something (needs goose)
ifndef name
	$(error usage: make migration name=add_something)
endif
	goose create $(name) sql
