# Changelog

Notable changes per release. Dates are UTC.

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
