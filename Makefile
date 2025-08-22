APP?=streamforge
PKG=./...
BIN_DIR=bin

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
