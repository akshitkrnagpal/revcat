# revcat

[![Release](https://img.shields.io/github/v/release/akshitkrnagpal/revcat?label=release)](https://github.com/akshitkrnagpal/revcat/releases/latest)
[![License](https://img.shields.io/github/license/akshitkrnagpal/revcat)](./LICENSE)
[![Docs](https://img.shields.io/badge/docs-revcat.vercel.app-7c8aff)](https://revcat.vercel.app)
[![Go](https://img.shields.io/github/go-mod/go-version/akshitkrnagpal/revcat)](./go.mod)

The RevenueCat CLI. Run your RevenueCat project from the terminal instead of clicking through the dashboard.

![revcat demo](./demo/demo.gif)

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat subscribers info app_user_123
revcat metrics overview
revcat publish offering pro --paywall ./paywalls/pro.json
```

## Why

RevenueCat ships a dashboard, REST API, and (2025) an MCP server, but no first-party CLI. revcat covers the v2 REST API surface a per-project secret key can reach: full CRUD on entitlements, offerings, packages, products, paywalls, webhooks, virtual currencies; per-customer grants/refunds/transfers; metrics + charts; audit-logs.

Output is a colored table when you're at a terminal and JSON when you're piping into a script - no `--json` ceremony.

## Install

```sh
brew install akshitkrnagpal/tap/revcat

# or, from source (Go 1.23+)
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

Pre-built binaries for every platform are on the [Releases page](https://github.com/akshitkrnagpal/revcat/releases).

## Auth

revcat reads a RevenueCat v2 secret key (`sk_...`). One of:

1. `REVCAT_API_KEY` env (highest precedence, one-shot)
2. `--profile <name>` flag
3. `REVCAT_PROFILE` env
4. `~/.revcat/active` (set via `revcat auth use <name>`)
5. profile named `default`

Profiles live in your OS keychain by default. For containers/CI, pass `--bypass-keychain` (or `REVCAT_BYPASS_KEYCHAIN=1`) and the profile is written to `./.revcat/config.json` instead. A `.gitignore` is created on first write.

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat auth status --validate
revcat auth doctor                   # diagnose common breakage
revcat auth list
revcat auth use my-app
```

## Command surface

```
revcat projects           list | view
revcat apps               list | view | public-keys | storekit-config

revcat entitlements       list | view | create | update | delete | archive | unarchive
                          products | attach | detach
revcat offerings          list | view | create | update | delete | archive | unarchive
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

Audit-grant a refund:

```sh
revcat subscribers grant app_user_123 premium -d 7d -r "support ticket #2241"
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

A small slice of the v2 API is gated behind a project-management / partner-tier key (not the per-project secret keys revcat uses). Those are not implemented:

- `POST /projects` (project create)
- App CRUD (`POST /apps`, `POST /apps/{id}`, `DELETE /apps/{id}`)
- `GET /collaborators`

Manage these in the dashboard.

RC also has no REST events firehose; lifecycle events (purchases, renewals, cancellations) are delivered via webhooks. Use `revcat webhooks create` to subscribe your endpoint.

## Debug

```sh
REVCAT_DEBUG=api revcat metrics overview     # logs full request/response (key redacted)
revcat doctor                                # top-level health check
revcat auth doctor                           # auth-specific
```

## Documentation

Full docs at <https://revcat.vercel.app> - install, quickstart, every command, configuration, and guides. Source lives in [`docs/`](./docs/).

## AI agent support

revcat ships [Agent Skills](./skills/) (open standard, distributable via skills.sh) so Claude Code, Cursor, and Codex can compose revcat commands accurately:

- `revcat-getting-started` - install, auth, top-level command map
- `revcat-commands` - real syntax + examples for every subcommand
- `revcat-troubleshooting` - common errors and fixes

Install locally:

```sh
# Claude Code (project-scoped)
mkdir -p .claude/skills && cp -R skills/revcat-* .claude/skills/

# Claude Code (user-scoped, every project)
mkdir -p ~/.claude/skills && cp -R skills/revcat-* ~/.claude/skills/
```

See [`skills/README.md`](./skills/README.md) for Cursor and Codex install steps.

## License

MIT
