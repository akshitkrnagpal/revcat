---
name: revcat-troubleshooting
description: Use when a revcat command fails or shows an error. Common errors: 401 unauthorized, no profile, no project_id, 404 customer/subscription, partner-tier endpoints, paywall PUT failures, ndjson parsing. Triggers on revcat error output, "revcat doesn't work".
---

# revcat - troubleshooting

Run `revcat doctor` first - it walks the most common breakage points and tells you what's wrong. `revcat auth doctor` is the auth-specific variant.

For deep debugging:

```sh
REVCAT_DEBUG=api revcat <command>
```

Logs the full request and response (with the secret key redacted).

## "the key was rejected (401 unauthorized)"

The login step or any later command got 401.

- Confirm you're using a **v2 secret key** (starts with `sk_`), not a public SDK key.
- The key may have been rotated. Generate a new one in the dashboard and re-run `revcat auth login --name <name> --secret-key sk_xxx`.
- If you're hitting a partner-tier endpoint (project create, app CRUD, audit logs, collaborators) with a project secret key, you'll see 401/403. Those are out of scope for revcat.

## "no profile found - run `revcat auth login`"

There is no active profile to use.

- First time: `revcat auth login --name my-app --secret-key sk_xxx`.
- Already logged in elsewhere? Check what's set: `revcat auth list`.
- Switch with `revcat auth use <name>` or per-command `--profile <name>`.
- In CI: pass `REVCAT_API_KEY=sk_xxx` directly, or use `--bypass-keychain` so revcat reads `./.revcat/config.json`.

## "no project_id on profile"

The active profile knows the secret key but not which project to call.

- Re-run `revcat auth login` and pick the project at the prompt.
- Or set `REVCAT_PROJECT_ID=proj_xxx` for a one-off override.

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

## Where revcat does NOT work

- Project create, app CRUD, collaborators - need a partner-tier API key.
- An events firehose - RC delivers lifecycle events via webhooks (`revcat webhooks`), not a REST stream.
- Anything not in `revcat <group> --help` - revcat tracks v2; v1-only endpoints are not wrapped.
