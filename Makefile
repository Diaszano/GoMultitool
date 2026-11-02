.DEFAULT_GOAL := help

.PHONY: help setup fmt fmt-check lint test test-race build build-armv7 integration-test image security check clean

help: ## Show available targets
	@printf 'Available targets:\n'
	@awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z0-9_-]+:.*## / { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

setup: ## Install pinned Go tools locally
	GOBIN=$(shell go env GOPATH)/bin go install golang.org/x/tools/cmd/goimports@$$(sed -n 's/^goimports=//p' .tools-versions)
	GOBIN=$(shell go env GOPATH)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$$(sed -n 's/^golangci-lint=//p' .tools-versions)

fmt: ## Format Go sources
	PATH=$(shell go env GOPATH)/bin:$$PATH bash scripts/check-go-format.sh --write $$(find cmd internal tests -name '*.go' -type f)

fmt-check: ## Check Go formatting
	PATH=$(shell go env GOPATH)/bin:$$PATH bash scripts/check-go-format.sh --check $$(find cmd internal tests -name '*.go' -type f)

lint: ## Run Go static analysis
	PATH=$(shell go env GOPATH)/bin:$$PATH golangci-lint run --config .golangci.yml ./...

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
