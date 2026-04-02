.PHONY: build run test lint fmt vet clean

BIN_DIR := bin
BINARY := $(BIN_DIR)/finterm

build:
	@echo "Building finterm..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BINARY) ./cmd/finterm/

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
