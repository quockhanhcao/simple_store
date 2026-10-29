.PHONY: run-store run-admin build test vet fmt check clean migrate-create db-up db-down db-reset db-psql migrate

ENV_FILE ?= .env
MIGRATIONS_DIR ?= ./internal/database/migrations

# Export every variable from $(ENV_FILE) into the environment of the command that follows.
LOAD_ENV = @test -f $(ENV_FILE) || { echo "$(ENV_FILE) not found: run 'cp .env.example .env'"; exit 1; }; \
	set -a; . ./$(ENV_FILE); set +a;

run-store:
	$(LOAD_ENV) go run ./cmd/store

run-admin:
	$(LOAD_ENV) go run ./cmd/admin

build:
	go build -o bin/store ./cmd/store
	go build -o bin/admin ./cmd/admin

test:
	go test -race ./...

vet:
	go vet ./...

# Rewrites files in place.
fmt:
	gofmt -l -w .

# Read-only gate to run before committing: fails if anything is unformatted, vet complains, or a test fails.
check:
	@unformatted=$$(gofmt -l .); if [ -n "$$unformatted" ]; then echo "needs gofmt:"; echo "$$unformatted"; exit 1; fi
	go vet ./...
	go test -race ./...

clean:
	rm -rf bin

# Usage: make migrate-create name=add_products
migrate-create:
	@test -n "$(name)" || { echo "usage: make migrate-create name=<migration_name>"; exit 1; }
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

# Postgres (:5432) + pgweb (http://localhost:8081). --wait blocks until postgres is healthy.
db-up:
	docker compose up -d --wait

# Stops containers; data is kept in the pgdata volume.
db-down:
	docker compose down

# Wipes the database (deletes the volume) and starts fresh. Migrations rerun on make migrate.
db-reset:
	docker compose down -v
	docker compose up -d --wait

db-psql:
	docker compose exec postgres psql -U postgres -d storedb

migrate:
	$(LOAD_ENV) go run ./cmd/migrate
