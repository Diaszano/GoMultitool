.DEFAULT_GOAL := help

.PHONY: help setup fmt fmt-check lint test test-race build build-armv7 integration-test image security check clean lint-shell lint-docs release-snapshot

help: ## Show available targets
	@printf 'Available targets:\n'
	@awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z0-9_-]+:.*## / { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

setup: ## Install the pinned toolchain from .mise.toml
	mise install

fmt: ## Format Go sources
	bash scripts/check-go-format.sh --write $$(find cmd internal tests -name '*.go' -type f)

fmt-check: ## Check Go formatting
	bash scripts/check-go-format.sh --check $$(find cmd internal tests -name '*.go' -type f)

lint: ## Run Go static analysis
	golangci-lint run --config .golangci.yml ./...

lint-shell: ## Check maintained shell scripts
	bash scripts/check-shell.sh --check $$(find scripts tests/policy tests/integration tools/policy -maxdepth 1 -type f -name '*.sh' | sort)

lint-docs: ## Lint Markdown, YAML, and whitespace
	bash scripts/check-docs.sh --check $$(find README.md CONTRIBUTING.md docs -type f -name '*.md' ! -path 'docs/superpowers/plans/*' | sort) $$(find . -type f \( -name '*.yaml' -o -name '*.yml' \) ! -path './.git/*' ! -path './.tools/*' ! -path './dist/*' ! -path './build/*' | sort)
	git diff --check

GO_TEST_PACKAGES ?= ./...

test: ## Run Go tests
	bash scripts/test-go.sh $(GO_TEST_PACKAGES)

test-race: ## Run Go race tests
	go test -race $(GO_TEST_PACKAGES)

build: ## Build the host executable
	mkdir -p build
	go build -o build/multitool ./cmd/multitool

build-armv7: ## Build the Linux ARMv7 executable
	bash scripts/build-armv7.sh build/multitool-linux-armv7

integration-test: ## Run integration tests
	go test ./...
	bash tests/policy/build_image_safety_test.sh
	bash tests/integration/rootfs_test.sh
	bash tests/integration/image_build_test.sh

IMAGE_TARGET ?= rk322x

image: ## Build a recovery image (requires root and privileged Linux image tools)
	bash scripts/build-image.sh "$(IMAGE_TARGET)"

security: ## Run vulnerability and secret checks
	govulncheck ./...
	gitleaks dir . --redact --no-banner
	actionlint .github/workflows/*.yml
	zizmor --offline .github/workflows

release-snapshot: ## Build local release archives without publishing
	goreleaser check
	goreleaser release --snapshot --clean

check: fmt-check lint lint-shell lint-docs test test-race integration-test build-armv7 ## Run all local quality checks

clean: ## Remove generated Go build output
	rm -rf build
