BINARY  := rssify
IMAGE   := rssify
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build test vet fmt lint clean run docker-build docker-run help

all: vet test build ## Run vet, test, and build

build: ## Build the binary
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test: ## Run all tests
	go test ./...

test-verbose: ## Run all tests with verbose output
	go test -v ./...

test-cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofmt -w .

lint: vet ## Run linters (vet + staticcheck if available)
	@which staticcheck >/dev/null 2>&1 && staticcheck ./... || true

clean: ## Remove build artifacts
	rm -f $(BINARY)

run: build ## Build and run the server
	./$(BINARY)

docker-build: ## Build Docker image
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-run: docker-build ## Build and run Docker container
	docker run --rm -p 8080:8080 $(IMAGE):latest

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'
