---
name: revcat-getting-started
description: Use when the user wants to start using revcat, the RevenueCat CLI. Covers install (Homebrew, go install), auth (browser OAuth + per-repo .revcat/config.json), and the top-level command map. Triggers on "revcat", "RevenueCat CLI", "set up revcat", "what can revcat do".
---

# revcat - getting started

`revcat` is a CLI for RevenueCat. It runs every v2 API operation reachable via OAuth, so you don't have to keep clicking through the dashboard. Single static Go binary, JSON-first when piped, table output on a TTY.

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

```sh
revcat auth login                # browser OAuth, saves tokens to OS keychain
cd ~/your-repo && revcat init    # bind this repo to a project
revcat auth doctor               # verify
```

`revcat auth login` opens the browser for the OAuth flow and stores the tokens in your OS keychain. `revcat init` walks the user through picking a project (and optionally apps), then writes:

- `revcat.toml` (committed): project_id + apps. Documents which RC project this repo belongs to.
- `.revcat/config.json` (gitignored, mode 0600): copies the credential into the directory. Walked up from cwd by every revcat command, so agents and sandboxes inside the directory inherit the credential without keychain access.

`.revcat/` is auto-appended to `.gitignore`.

### Linux / containers without secret-service

```sh
revcat auth login --bypass-keychain    # writes ~/.revcat/config.json instead of keychain
```

### Headless / CI

```sh
export REVCAT_REFRESH_TOKEN=rtk_...
export REVCAT_PROJECT_ID=proj_...
revcat offerings list
```

revcat synthesizes a virtual profile, refreshes tokens in-memory, no keychain or login flow. Pull the refresh token from your CI secret manager.

### Multi-account

```sh
revcat auth login --name work
revcat auth login --name personal
revcat auth use personal               # default for global commands
revcat --profile work auth status      # one-shot override
```

When you `revcat init` inside a repo, it copies whichever profile is active at that moment. To switch credentials for an existing repo, rerun `revcat init --force`.

## Top-level command map

Resources (CRUD + actions):

- `projects` (list/view/create), `apps` (list/view + public-keys + storekit-config + create/update/delete), `collaborators` (list)
- `entitlements`, `offerings`, `packages`, `products`, `paywalls`
- `subscribers` (customers), `subscriptions`, `purchases`, `invoices`
- `webhooks`, `virtual-currencies`

Activity:

- `metrics overview`, `charts get/options`, `audit-logs list`

Verb-orchestrators:

- `publish offering <id>` (set-current + paywall PUT in one shot)

Auth + housekeeping:

- `auth`, `init`, `doctor`, `completion`, `version`

## Global flags (every command)

- `--profile <name>` - active global profile (overridden by walked-up `.revcat/config.json`)
- `--project-id <id>` - override the resolved project id for this invocation
- `--bypass-keychain` - use `~/.revcat/config.json` (file backend) instead of OS keychain
- `--output table|json|csv|markdown` - force a format (auto: table on TTY, JSON when piped)
- `--pretty` - indent JSON
- `-v / -q / --no-color / --debug`

## What revcat does NOT cover

- Project update or delete - v2 has create + list only.
- An events firehose - RC delivers lifecycle events via webhooks, not a REST stream. Use `revcat webhooks create` to register your endpoint.

## First useful commands

```sh
revcat doctor                              # is everything OK?
revcat subscribers info app_user_123        # debug a customer
revcat metrics overview                     # headline numbers
revcat offerings list                       # catalog
revcat audit-logs list                      # who changed what
```
