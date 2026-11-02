.DEFAULT_GOAL := help

.PHONY: help setup fmt fmt-check lint test test-race build build-armv7 integration-test image security check clean lint-shell lint-docs release-snapshot

help: ## Show available targets
	@printf 'Available targets:\n'
	@awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z0-9_-]+:.*## / { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

setup: ## Install pinned Go tools locally
	GOBIN=$(shell go env GOPATH)/bin go install golang.org/x/tools/cmd/goimports@$$(sed -n 's/^goimports=//p' .tools-versions)
	GOBIN=$(shell go env GOPATH)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$$(sed -n 's/^golangci-lint=//p' .tools-versions)
	GOBIN=$(shell go env GOPATH)/bin go install mvdan.cc/sh/v3/cmd/shfmt@$$(sed -n 's/^shfmt=//p' .tools-versions)
	npm install --prefix $(shell go env GOPATH) markdownlint-cli2@$$(sed -n 's/^markdownlint-cli2=//p' .tools-versions)
	python3 -m pip install --target $(shell go env GOPATH)/python-tools yamllint==$$(sed -n 's/^yamllint=//p' .tools-versions)
	GOBIN=$(shell go env GOPATH)/bin go install github.com/rhysd/actionlint/cmd/actionlint@$$(sed -n 's/^actionlint=//p' .tools-versions)
	bash scripts/install-shellcheck.sh
	GOBIN=$(shell go env GOPATH)/bin go install golang.org/x/vuln/cmd/govulncheck@$$(sed -n 's/^govulncheck=//p' .tools-versions)
	GOBIN=$(shell go env GOPATH)/bin go install github.com/zricethezav/gitleaks/v8@$$(sed -n 's/^gitleaks=//p' .tools-versions)
	bash scripts/install-goreleaser.sh
	bash scripts/install-zizmor.sh

fmt: ## Format Go sources
	PATH=$(shell go env GOPATH)/bin:$$PATH bash scripts/check-go-format.sh --write $$(find cmd internal tests -name '*.go' -type f)

fmt-check: ## Check Go formatting
	PATH=$(shell go env GOPATH)/bin:$$PATH bash scripts/check-go-format.sh --check $$(find cmd internal tests -name '*.go' -type f)

lint: ## Run Go static analysis
	PATH=$(shell go env GOPATH)/bin:$$PATH golangci-lint run --config .golangci.yml ./...

lint-shell: ## Check maintained shell scripts
	PATH=$(shell go env GOPATH)/bin:$$PATH bash scripts/check-shell.sh --check $$(find scripts tests/policy tools/policy -maxdepth 1 -type f -name '*.sh' | sort)

lint-docs: ## Lint Markdown, YAML, and whitespace
	PATH=$(shell go env GOPATH)/bin:$$PATH PYTHONPATH=$(shell go env GOPATH)/python-tools bash scripts/check-docs.sh --check $$(find README.md CONTRIBUTING.md docs -type f -name '*.md' ! -path 'docs/superpowers/plans/*' | sort) $$(find . -type f \( -name '*.yaml' -o -name '*.yml' \) ! -path './.git/*' ! -path './.tools/*' ! -path './dist/*' ! -path './build/*' | sort)
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

image: ## Build a recovery image (requires privileged Linux image tools)
	@echo 'Image build workflow is not implemented yet' >&2
	@exit 1

security: ## Run vulnerability and secret checks
	PATH=$(shell go env GOPATH)/bin:$$PATH govulncheck ./...
	PATH=$(shell go env GOPATH)/bin:$$PATH gitleaks dir . --redact --no-banner
	PATH=$(shell go env GOPATH)/bin:$$PATH actionlint .github/workflows/*.yml
	PATH=$(shell go env GOPATH)/bin:$$PATH zizmor --offline .github/workflows

release-snapshot: ## Build local release archives without publishing
	PATH=$(shell go env GOPATH)/bin:$$PATH goreleaser check
	PATH=$(shell go env GOPATH)/bin:$$PATH goreleaser release --snapshot --clean

check: fmt-check lint lint-shell lint-docs test test-race integration-test build-armv7 ## Run all local quality checks

clean: ## Remove generated Go build output
	rm -rf build
