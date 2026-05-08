.PHONY: build test lint run clean fmt vet

# Build the binary
build:
	go build -o bin/vaultguard ./cmd/vaultguard

# Run all tests with verbose output
test:
	go test -v -race ./...

# Run tests with coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run the linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

# Run go vet
vet:
	go vet ./...

# Format all Go files
fmt:
	gofmt -w .

# Run the MCP server (reads from stdin, writes to stdout)
run: build
	./bin/vaultguard

# Run fuzz tests for 30 seconds
fuzz:
	go test -fuzz=. -fuzztime=30s ./pkg/scanner/

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
