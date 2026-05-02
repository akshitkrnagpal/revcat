.PHONY: build test typecheck docs-cli docs-groups docs drift-check help

BIN := revcat

help:
	@echo "Available targets:"
	@echo "  make build         - go build the CLI binary"
	@echo "  make test          - go test ./..."
	@echo "  make typecheck     - go vet ./..."
	@echo "  make docs-cli      - regen the canonical CLI reference page"
	@echo "  make docs-groups   - regen per-group page headers (preserves prose"
	@echo "                       below <!-- AUTOGEN_END --> markers)"
	@echo "  make docs          - run docs-cli + docs-groups"
	@echo "  make drift-check   - assert docs/skills don't reference missing commands"
	@echo ""
	@echo "Source-of-truth: cobra commands in commands/ are tier-1."
	@echo "docs/, skills/, README.md are tier-3 and may drift. Run docs"
	@echo "after any command-tree change. Run drift-check before commit."

build:
	go build -o $(BIN) ./cmd/revcat

test:
	go test ./...

typecheck:
	go vet ./...

docs-cli:
	go run ./scripts/gen-cli-reference

docs-groups:
	go run ./scripts/gen-group-docs

docs: docs-cli docs-groups

drift-check:
	go test ./commands/ -run TestDocsDontReferenceMissingCommands -v
