# Changelog

Notable changes per release. Dates are UTC.

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
