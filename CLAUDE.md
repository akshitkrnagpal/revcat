# revcat

The RevenueCat CLI. Go + cobra.

## Working with this repo

- `go run ./cmd/revcat <args>` runs the CLI from source
- `go build -o revcat ./cmd/revcat` produces a single static binary
- `go test ./...` runs tests
- `go vet ./...` lints

## Architecture

- `cmd/revcat` - main() entrypoint, calls commands.Execute()
- `commands/` - cobra commands organized by resource (auth, apps, ...)
- `internal/api` - hand-rolled RC v2 REST client (no SDK exists)
- `internal/auth` - keychain + local-file backends, profile resolver
- `internal/output` - TTY-aware renderer (table on TTY, JSON when piped)

## Auth

Profiles in OS keychain by default. Override:

- `--bypass-keychain` or `REVCAT_BYPASS_KEYCHAIN=1` writes to `./.revcat/config.json`
- `REVCAT_API_KEY` env synthesizes a one-shot profile (highest precedence)
- `REVCAT_PROFILE` selects active profile when `--profile` is not set
- `REVCAT_PROJECT_ID` overrides stored project id

## Style

- No em-dashes anywhere (commit hook would catch them; manual for now)
- `go vet` clean before every commit
- Errors bubble up - do not log + return; log OR return
