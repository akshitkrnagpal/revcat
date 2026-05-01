# revcat

The RevenueCat CLI. Go + cobra.

## Working with this repo

- `go run ./cmd/revcat <args>` runs the CLI from source
- `go build -o revcat ./cmd/revcat` produces a single static binary
- `go test ./...` runs tests
- `go vet ./...` lints

## Architecture

- `cmd/revcat` - main() entrypoint, calls commands.Execute()
- `commands/` - cobra commands organized by resource (auth, init, apps, ...)
- `internal/api` - hand-rolled RC v2 REST client (no SDK exists) + OAuth flow
- `internal/auth` - keychain / global file / project-local credential stores + resolver
- `internal/project` - revcat.toml loader (committed half of project context)
- `internal/cliutil` - shared command helpers (Client, ResolveProjectID)
- `internal/output` - TTY-aware renderer (table on TTY, JSON when piped)

## Auth

OAuth-only since v0.4. Three credential storage tiers (resolution order):

1. `REVCAT_REFRESH_TOKEN` env (synthesizes a virtual profile, in-memory only)
2. Walked-up `./.revcat/config.json` (written by `revcat init`, gitignored, mode 0600)
3. Global keychain (default) or `~/.revcat/config.json` with `--bypass-keychain` / `REVCAT_BYPASS_KEYCHAIN=1`

Within tier 3, the active profile name is `--profile` flag > `REVCAT_PROFILE` env > `~/.revcat/active` > `default`.

Project id resolution: `--project-id` flag > `REVCAT_PROJECT_ID` env > resolved credential's bound project (local config or env hatch) > walked-up `revcat.toml`.

## Style

- No em-dashes anywhere (commit hook would catch them; manual for now)
- `go vet` clean before every commit
- Errors bubble up - do not log + return; log OR return
