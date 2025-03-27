.PHONY: run build test lint clean

# Default Go build flags
GOFLAGS := -v

# Build the application
build:
	go build $(GOFLAGS) -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test ./... -v

# Run linter
lint:
	go vet ./...
	# Add golangci-lint when configured

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

.DEFAULT_GOAL := build
