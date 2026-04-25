---
name: revcat-getting-started
description: Use when the user wants to start using revcat, the RevenueCat CLI. Covers install (Homebrew, go install), auth (sk_ keys + keychain, profiles), and the top-level command map. Triggers on "revcat", "RevenueCat CLI", "set up revcat", "what can revcat do".
---

# revcat - getting started

`revcat` is a CLI for RevenueCat. It runs every API operation a per-project secret key can reach, so you don't have to keep clicking through the dashboard. Single static Go binary, JSON-first when piped, table output on a TTY.

GitHub: <https://github.com/akshitkrnagpal/revcat>
Docs: <https://revcat.vercel.app>

## Install

```sh
brew install akshitkrnagpal/tap/revcat

# or, from source (Go 1.23+)
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

Pre-built binaries for every platform are on the [GitHub Releases page](https://github.com/akshitkrnagpal/revcat/releases).

## Auth (one-time)

revcat reads a v2 secret key (`sk_...`) and stores it in the OS keychain.

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat auth doctor             # verify
```

For CI / Docker (no keychain), use one of:

```sh
# A: env var only (one-shot)
REVCAT_API_KEY=sk_xxx revcat metrics overview

# B: write a profile to ./.revcat/config.json (auto-gitignored)
revcat auth login --bypass-keychain --name ci --secret-key sk_xxx --no-verify
```

Switch profiles with `revcat auth use <name>` or `--profile <name>` per-command.

## Top-level command map

Resources (CRUD + actions):

- `projects` (list/view), `apps` (list/view + public-keys + storekit-config)
- `entitlements`, `offerings`, `packages`, `products`, `paywalls`
- `subscribers` (customers), `subscriptions`, `purchases`, `invoices`
- `webhooks`, `virtual-currencies`

Activity:

- `metrics overview`, `charts get/options`, `audit-logs list`

Verb-orchestrators:

- `publish offering <id>` (set-current + paywall PUT in one shot)

Auth + housekeeping:

- `auth`, `doctor`, `completion`, `version`

## Global flags (every command)

- `--profile <name>` - active auth profile
- `--bypass-keychain` - read/write profile from `./.revcat/config.json`
- `--output table|json|csv|markdown` - force a format (auto: table on TTY, JSON when piped)
- `--pretty` - indent JSON
- `-v / -q / --no-color / --debug`

## What revcat does NOT cover (out of scope)

- `POST /projects` (project create) - partner-tier key
- App CRUD (`POST /apps`, etc.) - partner-tier key
- `GET /collaborators` - partner-tier key
- An events firehose - RC delivers lifecycle events via webhooks, not a REST stream. Use `revcat webhooks create` to register your endpoint.

Manage the partner-tier items in the dashboard. Everything else a project secret key can do is wrapped.

## First useful commands

```sh
revcat doctor                              # is everything OK?
revcat subscribers info app_user_123        # debug a customer
revcat metrics overview                     # headline numbers
revcat offerings list                       # catalog
revcat audit-logs list                      # who changed what
```
