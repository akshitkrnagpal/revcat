# revcat

[![Release](https://img.shields.io/github/v/release/akshitkrnagpal/revcat?label=release)](https://github.com/akshitkrnagpal/revcat/releases/latest)
[![License](https://img.shields.io/github/license/akshitkrnagpal/revcat)](./LICENSE)
[![Docs](https://img.shields.io/badge/docs-revcat.vercel.app-7c8aff)](https://revcat.vercel.app)
[![Go](https://img.shields.io/github/go-mod/go-version/akshitkrnagpal/revcat)](./go.mod)

The RevenueCat CLI. Run your RevenueCat project from the terminal instead of clicking through the dashboard.

![revcat demo](./demo/demo.gif)

```sh
revcat auth login                        # browser OAuth, saves to keychain
cd ~/your-repo && revcat init            # bind this repo to a project
revcat subscribers info app_user_123
revcat metrics overview
revcat publish offering pro --paywall ./paywalls/pro.json
```

## Why

RevenueCat ships a dashboard, a REST API, and an MCP server, but no first-party CLI. revcat covers the v2 REST API surface accessible via OAuth: full CRUD on entitlements, offerings, packages, products, paywalls, webhooks, virtual currencies; per-customer grants/refunds/transfers; metrics + charts; audit-logs.

Output is a colored table when you're at a terminal and JSON when you're piping into a script - no `--json` ceremony.

## Install

```sh
brew install akshitkrnagpal/tap/revcat

# or, from source (Go 1.26+)
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

Pre-built binaries for every platform are on the [Releases page](https://github.com/akshitkrnagpal/revcat/releases).

## Auth

revcat authenticates against RevenueCat via OAuth (PKCE). One browser login populates a global profile in your OS keychain; running `revcat init` inside a repo writes a per-directory `.revcat/config.json` (gitignored, mode 0600) so agents and sandboxes operating in that directory inherit the credential without keychain access.

```sh
revcat auth login                        # browser OAuth
cd ~/your-repo && revcat init            # binds this repo to a project
revcat auth status --validate            # confirm
revcat auth doctor                       # diagnose
```

### Storage tiers

| Tier | Path | Used when |
| --- | --- | --- |
| keychain | OS keychain | default for `auth login` |
| global file | `~/.revcat/config.json` | `--bypass-keychain` or `REVCAT_BYPASS_KEYCHAIN=1` |
| local file | `./.revcat/config.json` (walked up) | written by `revcat init` |

Resolution: `REVCAT_REFRESH_TOKEN` env > walked-up local file > global active profile.

### Multi-account

```sh
revcat auth login --name work
revcat auth login --name personal
revcat auth use personal                 # default for global commands
revcat --profile work auth status        # one-shot override
```

### Headless / CI

```sh
export REVCAT_REFRESH_TOKEN=rtk_...
export REVCAT_PROJECT_ID=proj_...
revcat offerings list
```

revcat synthesizes a virtual profile, refreshes tokens in-memory, no keychain or login flow. Pull the refresh token from your CI secret manager.

## Command surface

```
revcat projects           list | view
revcat apps               list | view | public-keys | storekit-config

revcat entitlements       list | view | create | update | delete | archive | unarchive
                          products | attach | detach
revcat offerings          list | view | preview | create | update | delete | archive | unarchive
                          set-current
revcat packages           list | view | create | update | delete | products | attach | detach
revcat products           list | view | create | update | delete | archive | unarchive
                          push-to-store
revcat paywalls           list | view | create | delete

revcat subscribers        info | list | create | delete
                          grant | revoke | transfer
                          attributes | invoices
                          refund (delegates to subscriptions)
revcat subscriptions      view | transactions | entitlements
                          cancel | refund | management-url | search
revcat purchases          view | entitlements | refund | search
revcat invoices           view

revcat publish offering   <id> [--paywall ./paywall.json] [--current] [-y] [--dry-run]

revcat metrics            overview
revcat charts             get <name> | options <name>
revcat audit-logs         list

revcat webhooks           list | view | create | update | delete
revcat virtual-currencies list | view | create | update | delete | archive | unarchive

revcat init               (run inside a repo to bind project context)
revcat doctor
revcat auth               login | status | doctor | use | list | logout
revcat completion         bash | zsh | fish
revcat version
```

## Output

By default, output is TTY-aware:

- **Interactive terminal**: tables (lipgloss) with color
- **Piped or in CI**: JSON, one object per row

Override with `--output table|json|csv|markdown` or env `REVCAT_DEFAULT_OUTPUT`. Use `--pretty` for indented JSON.

## Examples

Debug a customer report from support:

```sh
revcat subscribers info app_user_123
```

Find a subscription from a store id (App Store transaction id, Play purchase token, Stripe sub_):

```sh
revcat subscriptions search ABC123XYZ
```

Promote a new paywall to current:

```sh
revcat publish offering pro --paywall ./paywalls/pro.json
```

Grant a complimentary entitlement:

```sh
revcat subscribers grant app_user_123 premium -d 7d
```

Wire entitlement <-> product:

```sh
revcat entitlements attach premium prod_app_monthly prod_app_annual
```

Show the dashboard headline numbers:

```sh
revcat metrics overview
```

Audit who changed what:

```sh
revcat audit-logs list
```

## Out of scope

A small slice of the v2 API isn't exposed by REST at all. Those are not implemented:

- `POST /projects` (project create)
- App CRUD (`POST /apps`, `POST /apps/{id}`, `DELETE /apps/{id}`)
- `GET /collaborators`

Manage these in the dashboard.

RC also has no REST events firehose; lifecycle events (purchases, renewals, cancellations) are delivered via webhooks. Use `revcat webhooks create` to subscribe your endpoint.

## Debug

```sh
REVCAT_DEBUG=api revcat metrics overview     # logs full request/response (token redacted)
revcat doctor                                # top-level health check
revcat auth doctor                           # auth-specific
```

## Documentation

Full docs at <https://revcat.vercel.app> - install, quickstart, every command, configuration, and guides. Source lives in [`docs/`](./docs/).

## AI agent support

revcat ships four [Agent Skills](./skills/) (open standard) so Claude Code, Cursor, and Codex can compose revcat commands accurately:

- `revcat-getting-started` — install, auth, top-level command map
- `revcat-commands` — real syntax + examples for every subcommand
- `revcat-troubleshooting` — common errors and fixes
- `revcat-storefront-debug` — 7-step diagnostic for "the SDK sees 0 packages from my offering"

Install via [skills.sh](https://skills.sh) (auto-detects your agent):

```sh
npx skills add akshitkrnagpal/revcat
```

Or manually — see [`skills/README.md`](./skills/README.md) for the Claude Code / Cursor / Codex paths.

## License

MIT
