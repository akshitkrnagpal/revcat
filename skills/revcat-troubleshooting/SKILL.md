---
name: revcat-troubleshooting
description: Use when a revcat command fails or shows an error. Common errors: 401 unauthorized, no profile, no project_id, 404 customer/subscription, partner-tier endpoints, paywall PUT failures, ndjson parsing, Test Store quirks, v1-only endpoints, dashboard-only operations. Triggers on revcat error output, "revcat doesn't work", "command not supported", "atob error", "endpoint missing".
---

# revcat - troubleshooting

Run `revcat doctor` first - it walks the most common breakage points and tells you what's wrong. `revcat auth doctor` is the auth-specific variant.

For deep debugging:

```sh
REVCAT_DEBUG=api revcat <command>
```

Logs the full request and response (with the bearer token redacted).

## "the credentials were rejected (401 unauthorized)"

The login step or any later command got 401.

- The OAuth token may have been revoked (Account Security UI in the RC dashboard) or the refresh token expired. Re-run `revcat auth login`.
- If you're trying to call something the v2 REST API doesn't expose (project create, app CRUD, collaborators), you'll see 401/403/404. Those are out of scope for revcat.

## "this profile was created under v0.3 secret-key auth"

You're upgrading from v0.3 (or earlier). Secret-key auth was removed in v0.4.

```sh
revcat auth logout --all   # clear stale profiles
revcat auth login          # re-auth via OAuth
cd ~/your-repo && revcat init
```

## "no profile found - run `revcat auth login`"

There is no credential to use.

- First time: `revcat auth login` (browser OAuth).
- Already logged in elsewhere? Check what's set: `revcat auth list`.
- Switch global profiles with `revcat auth use <name>` or per-command `--profile <name>`.
- Inside a repo: `revcat init` writes a `.revcat/config.json` so subsequent commands in that directory pick it up automatically.
- In CI: set `REVCAT_REFRESH_TOKEN=rtk_...` and `REVCAT_PROJECT_ID=proj_...`, or pass `--bypass-keychain` and ship a populated `~/.revcat/config.json`.

## "no project_id resolved" / project context missing

The credential resolved fine but no project id is set.

- Inside a repo: run `revcat init` to bind it.
- One-off override: `--project-id proj_xxx` or `REVCAT_PROJECT_ID=proj_xxx`.

## "toml/local mismatch" in `revcat auth doctor`

`revcat.toml` (committed) and `.revcat/config.json` (gitignored) disagree about which project this directory is bound to.

- Common cause: someone hand-edited `revcat.toml` without rerunning `revcat init`.
- Fix: `revcat init --force` to realign, or edit `revcat.toml` to match the local config and commit.

## "no customer with id ... in this project"

`revcat subscribers info` returned 404 for the user.

- The id passed is probably an alias, not the canonical `app_user_id`. Search by store id:
  - `revcat subscriptions search <store_transaction_id>`
  - `revcat purchases search <store_purchase_id>`
- Or list a recent slice and grep: `revcat subscribers list --output json | jq '.[] | select(.email == "X")'` (only if attributes contain the email).

## "compile error" / build fails on `go install`

revcat targets Go 1.23+. Check with `go version`. Older Go won't build the cobra dependency tree.

## `revcat publish offering --paywall` keeps a no-op for me

revcat hashes the file with sorted-key canonicalization and compares it to the live paywall. Identical payloads short-circuit silently. If you expect a change:

- Inspect the live paywall: pipe `revcat publish offering <id> --paywall ./file.json --dry-run` and read the plan.
- Diff against the live paywall manually: pull it via `revcat paywalls list` then `revcat paywalls view <id>`.

## `revcat <resource> view <key>` returns 404

The v2 API requires a system id (`ofrng...`, `entl...`, `prod...`) for path lookups. revcat resolves human-friendly lookup_keys to ids by listing first - if listing returns the resource and the lookup_key matches, view will succeed. If you still see 404:

- Run the list command and confirm the resource is actually there: `revcat offerings list`, `revcat entitlements list`, etc.
- Try the system id directly (the `id` column in list output).

## "input too large" on `--file`

revcat caps JSON file input at 4 MB to defend against accidental binary blobs. If you genuinely need to push more (rare for paywall configs), open an issue.

## Tables look broken / mojibake

`--no-color` (or `NO_COLOR=1`) helps in terminals that don't speak ANSI. For dumb pipes, force JSON: `--output json`.

## Test Store quirks ("my offering returns 0 packages on the SDK")

The Test Store is RC's sandbox-style storefront for development. It has two gotchas:

1. **Prices are dashboard-only.** v2 has no endpoint for setting prices on Test Store products. revcat cannot help here — set the price in the RC dashboard UI under each product.
2. **A product without a price is invisible to the SDK.** `/v1/subscribers/{user_id}/offerings` returns `packages: []` for offerings whose products have no Test Store price. The dashboard will show the product attached to the package; the SDK will still see nothing.

If the user is debugging this end-to-end, route to `revcat-storefront-debug` — that skill walks the full diagnostic flow.

## v1-only endpoints (revcat does not wrap these)

revcat tracks v2. The v1 surface is intentionally out of scope. The few v1 endpoints that come up in real debugging:

- `GET /v1/subscribers/{user_id}/offerings` - what the SDK actually receives. Use this to diff "what the dashboard shows" vs "what `Purchases.getOfferings()` returns." See `revcat-storefront-debug` for the curl pattern.
- `POST /v1/subscribers/{user_id}/receipts` - validate a store receipt. Used by SDKs internally; not a debugging endpoint.

When you fall back to curl for v1, use the **public SDK key** (one of the per-platform public keys), not the OAuth bearer token. Pull it via `revcat apps public-keys <app_id>`.

## Dashboard-only operations

A small set of operations have no v2 endpoint at all — revcat cannot wrap what doesn't exist. Known cases as of revcat's current shipped version:

- Test Store product prices (covered above)
- StoreKit configuration generation for local Xcode testing (export from dashboard manually)
- Webhook signing-secret rotation (rotate via dashboard, then re-fetch with `revcat webhooks view`)

If the user reaches for the dashboard for one of these, that's expected. If they reach for the dashboard for anything else and a v2 endpoint exists for it, that's a revcat coverage gap worth filing.

## Where revcat does NOT work

- Project create, app CRUD, collaborators - not exposed by the v2 REST API.
- An events firehose - RC delivers lifecycle events via webhooks (`revcat webhooks`), not a REST stream.
- Anything not in `revcat <group> --help` - revcat tracks v2; v1-only endpoints are not wrapped (see above for the curl fallback).
