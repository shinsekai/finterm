.PHONY: build run test lint fmt vet clean

BIN_DIR := bin
BINARY := $(BIN_DIR)/finterm

# Build-time version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

build:
	@echo "Building finterm..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BINARY) ./cmd/finterm/

run: build
	@echo "Running finterm..."
	@./$(BINARY)

test:
	@echo "Running tests..."
	@go test ./...

lint:
	@echo "Running linter..."
	@golangci-lint run

fmt:
	@echo "Running gofmt..."
	@gofmt -l -w .
	@echo "Formatted Go files."

vet:
	@echo "Running go vet..."
	@go vet ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out
	@echo "Cleaned binaries and coverage files."
