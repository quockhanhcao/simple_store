.PHONY: run-store run-admin build test vet fmt check clean

ENV_FILE ?= .env

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
