.DEFAULT_GOAL := help

.PHONY: help setup fmt fmt-check lint test test-race build build-armv7 integration-test image security check clean

help: ## Show available targets
	@printf 'Available targets:\n'
	@awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z0-9_-]+:.*## / { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

setup: ## Install pinned tools with mise
	mise install

fmt: ## Format Go sources
	gofmt -w $$(find cmd internal -name '*.go' -type f)

fmt-check: ## Check Go formatting
	@test -z "$$(gofmt -l $$(find cmd internal -name '*.go' -type f))"

lint: ## Run Go vet
	go vet ./...

test: ## Run Go tests
	go test ./...

test-race: ## Run Go race tests
	go test -race ./...

build: ## Build the host executable
	mkdir -p build
	go build -o build/multitool ./cmd/multitool

build-armv7: ## Build the Linux ARMv7 executable
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o build/multitool-linux-armv7 ./cmd/multitool

integration-test: ## Run integration tests
	go test ./...

image: ## Build a recovery image (requires privileged Linux image tools)
	@echo 'Image build workflow is not implemented yet' >&2
	@exit 1

security: ## Run security checks
	@echo 'Security checks are not implemented yet' >&2
	@exit 1

check: fmt-check lint test build ## Run local quality checks

clean: ## Remove generated Go build output
	rm -rf build
