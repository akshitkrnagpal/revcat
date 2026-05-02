.PHONY: build test typecheck docs-cli help

BIN := revcat

help:
	@echo "Available targets:"
	@echo "  make build      - go build the CLI binary"
	@echo "  make test       - go test ./..."
	@echo "  make typecheck  - go vet ./..."
	@echo "  make docs-cli   - regenerate the auto-generated CLI reference page"
	@echo ""
	@echo "Source-of-truth note: cobra commands in commands/ are tier-1."
	@echo "docs/, skills/, README.md are tier-3 and may drift. Run docs-cli"
	@echo "after any command-tree change."

build:
	go build -o $(BIN) ./cmd/revcat

test:
	go test ./...

typecheck:
	go vet ./...

docs-cli:
	go run ./scripts/gen-cli-reference
