# Changelog

Notable changes per release. Dates are UTC.

## [v0.4.0-alpha.3](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.4.0-alpha.3) - 2026-05-01

UX polish on top of alpha.2: one passphrase prompt per invocation instead of three, and a self-healing fix for stale `auth use` markers that made every command fail with "no profile found" after `auth logout --all`.

### Fixed

- `revcat init` (and any command opening the global keyring more than once) now prompts for the file-backed keyring passphrase exactly once per invocation. Previously each open re-prompted, so a single init would ask 2-3 times. The cache is process-scoped, so subsequent invocations still ask once. The macOS Keychain / Linux Secret Service backends are unaffected (the OS handles the passphrase and never calls our prompt func).
- `revcat auth logout --all` and `revcat auth logout <name>` now clear `~/.revcat/active` when the deleted profile was the active one. Without this, subsequent commands resolved to a missing profile and errored with "no profile found".
- The credential resolver self-heals when `~/.revcat/active` points to a profile that no longer exists: falls through to the `default` profile (if it exists) and clears the stale marker. Saves one round-trip of confused error messages for users upgrading from earlier alpha.

### Internal

- New `cliutil.ClientFromResolved` lets long-running commands (init, doctor) build the API client from an already-resolved credential without re-running `Resolve` (each Resolve re-opens the global store).
- `internal/auth.cachedPassphrase` wraps `keyring.TerminalPrompt` in a `sync.Once`. First call prompts, subsequent calls within the process reuse the value.
- New `internal/auth.ClearActive` removes `~/.revcat/active`. Called from logout and from the resolver's self-heal path.

## [v0.4.0-alpha.2](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.4.0-alpha.2) - 2026-05-01

OAuth becomes the only auth flow. Per-directory credentials let agents and sandboxes operate inside a repo without touching the user's keychain. Breaking change: secret-key auth is removed.

### Changed (BREAKING)

- `revcat auth login` is OAuth-only. `--secret-key`, `--secret-key-stdin`, and the `REVCAT_API_KEY` env var are removed. Existing v0.3 secret-key profiles error on read with: "this profile was created under v0.3 secret-key auth, which was removed in v0.4. run `revcat auth login` to reauth via OAuth".
- `Profile` struct collapses to OAuth-only fields (`access_token`, `refresh_token`, `expires_at_ms`, `scope`, `client_id`). Profiles bound to a `project_id` no longer carry one — project context lives in revcat.toml / .revcat/config.json now.
- `--bypass-keychain` (and `REVCAT_BYPASS_KEYCHAIN=1`) now writes the global file backend to `~/.revcat/config.json` (HOME), not `./.revcat/config.json` (cwd). The cwd path is now exclusively for project-local credentials.
- `internal/api.Client` requires a `TokenSource`. `Options.SecretKey` removed. Programming-error panic if New is called without one.

### Added

- Project-local credential file at `./.revcat/config.json` (mode 0600, walked up from cwd, gitignored). Holds the OAuth credential blob plus `project_id` and optional apps. When present, every revcat command in the directory uses it instead of the global keychain. This is the "agents and sandboxes" story — drop the file in, no keychain needed.
- `revcat init` now writes both halves: `revcat.toml` (committed: project_id + apps) and `.revcat/config.json` (gitignored: credentials + project_id + apps). Auto-appends `.revcat/` to `.gitignore` (idempotent). New `--no-local-creds` flag to write only the toml.
- `REVCAT_REFRESH_TOKEN` env: CI / sandbox / agent escape hatch. When set, resolution short-circuits and synthesizes a virtual profile carrying just the refresh token. Pair with `REVCAT_PROJECT_ID` for full headless. Refreshed tokens stay in-memory for the duration of the invocation.
- `auth status` and `auth doctor` now show `source` (local / keychain / file / env) and `source_path` so you can tell at a glance which credential is winning.

### Internal

- Three storage roles in `internal/auth`:
  - `keychainStore` (default) — OS keychain via 99designs/keyring.
  - `globalFileStore` — `~/.revcat/config.json`, profiles map, mode 0600.
  - `LocalConfig` — `./.revcat/config.json`, single credential blob + project + apps.
- Unified `Resolve(ResolveOptions)` returning `(*Resolved, error)` walks the precedence chain in one place. Source enum (`SourceLocal | SourceKeychain | SourceGlobalFile | SourceEnv`) flows through to status/doctor for diagnostics.
- `OAuthTokenSource` writes refreshed tokens back to whichever tier they came from (local file vs global store; env hatch is in-memory only).
- Tests: local config walk-up + roundtrip, gitignore append idempotency, env-hatch precedence, legacy secret-key profile rejection, full Resolve precedence chain.

### Migration notes

- Old keychain entries from v0.3 secret-key auth: error on first use. Run `revcat auth login` to reauth via OAuth.
- Old `--bypass-keychain` users with `./.revcat/config.json` from v0.3 era: that path is now the project-local credentials file. The shape has changed (single blob, not profiles map). Old files won't load. Run `revcat auth login --bypass-keychain` to write fresh creds at `~/.revcat/config.json`.
- v0.4.0-alpha.1 OAuth profiles in keychain: keep working; their saved `project_id` is ignored (resolution is now flag/env/file > error, no profile fallback). To get the per-directory model: run `revcat init` in your repo.

## [v0.4.0-alpha.1](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.4.0-alpha.1) - 2026-05-01

OAuth (PKCE) login as an alternative to v2 secret keys, plus a per-repo `revcat.toml` that pins project context Terraform-style. Alpha for early feedback before v0.4.0 final.

### Added

- `revcat auth login --oauth` runs the PKCE flow against the public RevenueCat OAuth client (`UmV2Q2F0`). Opens the browser, captures the redirect on a loopback port, exchanges the code, stores tokens on a profile. Override the client with `REVCAT_OAUTH_CLIENT_ID` or `-ldflags '-X .../auth.EmbeddedClientID=<id>'`. Auto-refreshes via a `TokenSource` on the API client; mutex-guarded so concurrent commands don't double-refresh, with a 60s skew before expiry.
- New `revcat init` writes `revcat.toml` at the current directory pinning the active `project_id` (and optional apps[]). Walked up from cwd like `.git` / `go.mod`, so any command run inside the repo inherits the project automatically. Interactive (lists projects, optional app multi-select) or scripted (`--project-id`, `--app-id`, `--no-apps`, `--force`).
- New global `--project-id` flag and `REVCAT_PROJECT_ID` env. Resolution order: `--project-id` > env > `revcat.toml` > profile binding. The legacy fallback keeps existing secret-key profiles working unchanged.
- `revcat auth status` now shows `auth_type`, `project_id`, and a new `project_source` row pointing at the `revcat.toml` path / env / flag / profile so you can debug "why did it hit the wrong project?".
- `revcat auth list` shows a new `auth_type` column and redacts OAuth access tokens.
- `revcat auth doctor` and `revcat doctor` rewrite their project-binding check in terms of the resolved project context.

### Changed

- OAuth profiles save credentials only (no `project_id` binding). Switching projects no longer requires re-logging in. Existing secret-key profiles still bind a project at login time for backward compatibility.
- `internal/api.Client` accepts a `TokenSource` so OAuth and secret-key auth share one transport.
- `internal/cliutil.Client` now resolves project_id via the new precedence chain. Every command that builds clients through cliutil picks up the new behavior automatically.

### Internal

- New `internal/oauth` flow split across `internal/api/oauth.go` (PKCE pair, AuthorizeURL builder, ExchangeCode, RefreshToken, loopback server, browser opener) and `internal/auth/oauth.go` (refreshing TokenSource backed by the credential store).
- New `internal/project` package: TOML loader (BurntSushi/toml), walk-up lookup, atomic Save.
- New `internal/cliutil.ResolveProjectID` and `cliutil.ClientForProject` helpers.
- Tests: PKCE S256 challenge correctness, AuthorizeURL shape, ExchangeCode/RefreshToken form encoding, error-body surfacing, loopback callback capture, project-config walk-up + roundtrip, full ResolveProjectID precedence chain.

### Migration notes

- Existing secret-key profiles deserialize unchanged; `auth_type` defaults to `secret_key` when missing.
- After upgrading: existing OAuth profiles created during the alpha may carry a stale `project_id` from earlier iterations. Re-login via `revcat auth login --oauth --name <profile>` to drop it, then `cd` into your repo and run `revcat init`.

## [v0.3.0](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.3.0) - 2026-04-29

Security audit pass. Five findings ranging from a transitive CVE down to inconsistent URL escaping, all addressed before the OAuth public-client work lands. `govulncheck ./...` now reports clean.

### Added

- `revcat auth login --secret-key-stdin` reads the v2 secret key from stdin instead of a flag value. The previous `--secret-key sk_xxx` form leaks the key into shell history; documentation now leads with the stdin pattern. The flag-value form still works but the help text flags the risk.

### Fixed

- **Security** — `github.com/dvsekhvalnov/jose2go` pinned to v1.7.0 via a `go.mod` replace directive, closing two known DoS CVEs (GO-2025-4123, GO-2023-2409) that reached the keychain backend transitively via `99designs/keyring`. `govulncheck ./...` reports `No vulnerabilities found.`
- **Security** — `internal/api/projects.go` now URL-escapes `appID` consistently with the rest of `internal/api/`. The other resource files were already escaping correctly; projects.go was the lone holdout.
- `internal/auth/local.go` writes to `~/.revcat/config.json` are now atomic (tempfile + chmod 0600 + sync + rename). Ctrl-C, kernel panic, or two concurrent revcat invocations can no longer leave the config in a partially-written state. Same fix applied to `internal/auth/active.go` `SetActive`.
- `--file <path>` now enforces the same 4 MiB cap as stdin. Previously revcat would happily try to read a multi-GB file into memory. The paywall loader (`commands/publish/offering.loadPaywall`) inherits the same cap. Error message: `"input too large: file is N bytes, max is 4 MiB. Pipe via stdin if you really need more."`

### Internal

- New `cliutil.MaxJSONSize` constant + `cliutil.ReadCappedFile` helper used by both stdin and file-path branches.
- New `internal/auth/atomic.go` with layered helpers (`atomicWriteJSON` → `atomicWriteFile` → `atomicWriteWith`); the lowest takes a `func(io.Writer) error` so tests can inject mid-write failures.
- Added regression tests across all five fixes including a concurrent-writes test for the atomic config helper and an at-the-cap-boundary test for the file size limit.

## [v0.2.0](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.2.0) - 2026-04-29

Diagnostic + JSON-completeness pass driven by a real-world session where the SDK was seeing 0 packages from a Test Store offering. Adds the missing piece for that workflow, fixes two latent unmarshaling bugs, and stops `--output json` from silently dropping fields the v2 API returns.

### Added

- `revcat offerings preview <id>` - hits the v1 `/subscribers/{user_id}/offerings` endpoint (the one the SDK actually calls) and renders what `Purchases.getOfferings()` will receive. Auto-fetches the public SDK key, defaults `--as` to a synthetic user id, and auto-detects the project's app on `--platform`. Closes the most common "why is the SDK seeing nothing" loop in one command instead of a curl fan-out.
- New skill `revcat-storefront-debug` walks the 7-step diagnostic flow (offering current → packages → product binding → store binding → push-to-store → Test Store price → v1 verify). Includes the dashboard-only Test Store price gotcha that has no v2 endpoint.

### Fixed

- `<resource> view --output json` now passes through the full v2 response. Previously revcat decoded into a typed struct and re-serialized the curated subset, dropping fields like `app_id`, `created_at`, `state`, `subscription`, `store_identifier`, etc. Affected `products`, `packages`, `offerings`, `entitlements`, `paywalls`, `webhooks`, `virtual-currencies`, `subscriptions`, `purchases`, `apps`, `projects`, and `subscribers info`. New regression test injects a `future_field_revcat_doesnt_model` and asserts it survives the round trip.
- `revcat packages products` was returning rows with every field empty. v2's `GET /packages/{id}/products` returns binding objects (`{eligibility_criteria, product:{...}}`), not bare products. revcat was unmarshaling into `[]Product` directly, so every nested field - id, display_name, store_identifier, app_id - dropped silently. Now uses a `PackageProductBinding` type matching the actual shape; the table view shows the full product + eligibility, JSON returns the raw v2 binding objects.
- `revcat entitlements products --output json` was still curating to 4 fields. Same passthrough fix as the `view` commands; full v2 product shape now reaches users.

### Internal

- `Client.DoRaw` returns the verbatim response body alongside the typed decode. `Get*Raw` and `paginateBoth` helpers cover every read path so JSON output is field-complete by default.
- Skills updated: `revcat-troubleshooting` adds Test Store quirks + v1-fallback sections; `revcat-commands` description broadened so the skill loads on flag/command-discovery questions.

## [v0.1.2](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.1.2) - 2026-04-25

API shape fixes from a full smoke test against a live RC v2 project. Most things worked but a pile of bodies and paths didn't match what v2 actually accepts.

### Fixed

- `packages create`: now POSTs to `/offerings/{oid}/packages`, not `/packages`. Takes `<offering>` as a positional arg with `--id` / `--display-name` / `--position` shortcut flags.
- `packages attach`: body is `{products:[{product_id, eligibility_criteria}]}`. New `--eligibility` flag (default `all`; also `google_sdk_lt_6`, `google_sdk_ge_6`).
- `packages list`: returns `[]` not `null` when there are no packages.
- `packages` struct: `lookup_key` field (was `Identifier` - was returning empty).
- `subscribers grant`: body uses `{entitlement_id, expires_at:<unix_ms>}`. revcat translates `--duration` to `expires_at` internally.
- `subscribers revoke`: implemented as "grant with `expires_at = now+1s`" because v2 has no revoke endpoint and rejects past `expires_at`.
- `subscribers attributes set`: body is `{attributes:[{name,value},...]}`. `--set k=v` shortcut now normalizes.
- `webhooks`: requires `name`; `event_types` field (not `events`); values must be lowercase. revcat lowercases `--events` automatically.
- `virtual-currencies`: keyed by uppercase `code` (e.g., `COIN`). Struct uses `{name, code, description, state}`.
- `paywalls create`: only `{offering_id}` is accepted. New `--offering` shortcut.
- `projects view`: implemented via list-and-filter (v2 has no `GET /projects/{id}`).
- `subscriptions/purchases search`: 404 normalized to `[]`.

### Removed

- `subscribers override-offering`, `subscribers restore-google-play`, `subscribers vc-balance`, `subscribers vc-tx`, `subscribers vc-set-balance`. None of these endpoints exist on the v2 customer surface.

## [v0.1.1](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.1.1) - 2026-04-25

### Fixed

- 4 v1-style API paths corrected to match v2 docs (grant_promotional, revoke_promotional, transactions:refund, _set_current).

### Added

- `revcat audit-logs list`. The `/audit_logs` endpoint is reachable with project secret keys; revcat now wraps it.
- `subscribers info` lookup_key resolver: pass either the system id or the lookup_key to view/update/delete commands across resources.

### Removed

- `revcat events list/tail`. RC v2 has no REST events firehose; lifecycle events are webhook-delivered. Use `revcat webhooks create` to subscribe an endpoint.

## [v0.1.0](https://github.com/akshitkrnagpal/revcat/releases/tag/v0.1.0) - 2026-04-25

First public release.

### Added

- `auth` (login, status, doctor, use, logout, list) with OS keychain backend and `--bypass-keychain` for CI.
- Catalog: full CRUD + archive on `entitlements`, `offerings`, `packages`, `products`. Top-level `paywalls`. Attach/detach for entitlements ↔ products and packages ↔ products.
- Subscribers: `info`, `list`, `create`, `delete`, `transfer`, `grant`, `revoke`, `refund`, `attributes`, `invoices`.
- Activity: `metrics overview`, `charts get`, `charts options`.
- Integrations: `webhooks` (CRUD), `virtual-currencies` (CRUD + archive).
- Verb-orchestrator: `revcat publish offering` (set-current + paywall PUT in one shot).
- Subscriptions / purchases / invoices read commands and `search`.
- `doctor`, `version`, shell completion via cobra.
- Homebrew tap publication via goreleaser. macOS/Linux/Windows on amd64/arm64.
- Astro Starlight docs at <https://revcat.vercel.app>.
- Three Agent Skills (`revcat-getting-started`, `revcat-commands`, `revcat-troubleshooting`).
