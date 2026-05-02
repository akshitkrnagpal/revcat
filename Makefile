.PHONY: build install test typecheck docs-cli docs-groups docs drift-check verify help

# Build metadata. Captured at make-time so `make build` produces a
# binary stamped with the current git ref + UTC build timestamp - same
# shape goreleaser injects on tagged releases.
BIN         := revcat
PKG         := github.com/akshitkrnagpal/revcat/commands
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
               -X $(PKG).Version=$(VERSION) \
               -X $(PKG).CommitHash=$(COMMIT_HASH) \
               -X $(PKG).BuildTime=$(BUILD_TIME)

##@ Build

build: ## Build the CLI binary into ./$(BIN) with build metadata baked in
	go build -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/revcat

install: ## Build + install to $$GOPATH/bin (or $$GOBIN)
	go install -ldflags '$(LDFLAGS)' ./cmd/revcat

##@ Test + lint

test: ## go test ./...
	go test ./...

typecheck: ## go vet ./...
	go vet ./...

drift-check: ## Assert docs/skills don't reference missing commands
	go test ./commands/ -run TestDocsDontReferenceMissingCommands -v

verify: typecheck test drift-check ## Pre-commit: typecheck + tests + drift-check

##@ Docs

docs-cli: ## Regenerate the canonical CLI reference page
	go run ./scripts/gen-cli-reference

docs-groups: ## Regenerate per-group page headers (preserves prose below AUTOGEN_END)
	go run ./scripts/gen-group-docs

docs: docs-cli docs-groups ## Run docs-cli + docs-groups

##@ Help

# Awk-based help. Targets followed by `## doc` get listed under the
# section last marked with `##@ Section`.
help: ## Print this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Source-of-truth: cobra commands in commands/ are tier-1."
	@echo "After any command-tree change: make docs && make drift-check."
