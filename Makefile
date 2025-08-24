APP?=streamforge
PKG=./...
BIN_DIR=bin
DC = docker compose -f compose/docker-compose.yml --env-file .env
MIGRATE_IMG = migrate/migrate:v4.18.3
MIGRATE_NETWORK = streamforge_default
DB_URL ?= $(DATABASE_URL)
ifeq ($(DB_URL),)
DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@timescale:5432/$(POSTGRES_DB)?sslmode=disable
endif
MIGRATIONS_DIR = $(PWD)/db/migrations


.PHONY: deps
deps:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(PKG)

.PHONY: race
race:
	go test -race $(PKG)

.PHONY: cover
cover:
	go test -coverprofile=coverage.out $(PKG) && go tool cover -func=coverage.out

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/...

.PHONY: vet
vet:
	go vet $(PKG)

.PHONY: tools
tools:
	@test -x "$$(command -v golangci-lint)" || (echo "Installing golangci-lint"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)

.PHONY: up
up:
	$(DC) up -d

.PHONY: down
down:
	$(DC) down -v

.PHONY: ps
ps:
	$(DC) ps

.PHONY: logs
logs:
	$(DC) logs -f --tail=200

.PHONY: restart
restart:
	$(DC) down
	$(DC) up -d


.PHONY: migrate-up
migrate-up:
	@echo "Running migrations UP against: $(DB_URL)"
	docker run --rm \
		-v "$(MIGRATIONS_DIR)":/migrations \
		--network $(MIGRATE_NETWORK) \
		$(MIGRATE_IMG) \
		-path=/migrations -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back ONE migration (down 1) against: $(DB_URL)"
	docker run --rm \
		-v "$(MIGRATIONS_DIR)":/migrations \
		--network $(MIGRATE_NETWORK) \
		$(MIGRATE_IMG) \
		-path=/migrations -database "$(DB_URL)" down 1

.PHONY: migrate-version
migrate-version:
	docker run --rm \
		-v "$(MIGRATIONS_DIR)":/migrations \
		--network $(MIGRATE_NETWORK) \
		$(MIGRATE_IMG) \
		-path=/migrations -database "$(DB_URL)" version

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(v)" ]; then echo "Usage: make migrate-force v=<version>"; exit 1; fi
	docker run --rm \
		-v "$(MIGRATIONS_DIR)":/migrations \
		--network $(MIGRATE_NETWORK) \
		$(MIGRATE_IMG) \
		-path=/migrations -database "$(DB_URL)" force $(v)


.PHONY: psql
psql:
	docker exec -it sf-timescale psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)