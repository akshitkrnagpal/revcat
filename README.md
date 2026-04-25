# revcat

The RevenueCat CLI. Run your RevenueCat project from the terminal instead of clicking through the dashboard.

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat subscribers info app_user_123
revcat events tail --type INITIAL_PURCHASE,CANCELLATION
revcat publish offering pro --paywall ./paywalls/pro.json
revcat metrics overview
```

## Why

RevenueCat ships a dashboard, REST API, and (2025) an MCP server, but no first-party CLI. revcat covers the full v2 REST API surface that a per-project secret key can reach: full CRUD on entitlements, offerings, packages, products, paywalls, webhooks, virtual currencies; per-customer grants/refunds/transfers; live event tail; metrics + charts.

Output is a colored table when you're at a terminal and JSON when you're piping into a script - no `--json` ceremony.

## Install

From source (Go 1.23+):

```sh
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

Homebrew + GitHub release binaries land with v0.1.

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
                          grant | revoke | transfer | override-offering
                          attributes | invoices | restore-google-play
                          vc-balance | vc-tx | vc-set-balance
                          refund (delegates to subscriptions)
revcat subscriptions      view | transactions | entitlements
                          cancel | refund | management-url | search
revcat purchases          view | entitlements | refund | search
revcat invoices           view

revcat publish offering   <id> [--paywall ./paywall.json] [--current] [-y] [--dry-run]

revcat events             list | tail
revcat metrics            overview
revcat charts             get <name> | options <name>

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

`revcat events tail` emits one JSON object per line in JSON mode (ndjson), so you can pipe into `jq` mid-stream.

## Examples

Debug a customer report from support:

```sh
revcat subscribers info app_user_123
```

Find a subscription from a store id (App Store transaction id, Play purchase token, Stripe sub_):

```sh
revcat subscriptions search ABC123XYZ
```

Tail purchases and cancellations during a launch:

```sh
revcat events tail --type INITIAL_PURCHASE,CANCELLATION
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

Compute a daily revenue summary in CI:

```sh
revcat events list --type INITIAL_PURCHASE,RENEWAL --since 24h --output json | \
    jq '[.[] | .price] | add'
```

Show the dashboard headline numbers:

```sh
revcat metrics overview
```

## Out of scope

A small slice of the v2 API is gated behind a project-management / partner-tier key (not the per-project secret keys revcat uses). Those are not implemented:

- `POST /projects` (project create)
- App CRUD (`POST /apps`, `POST /apps/{id}`, `DELETE /apps/{id}`)
- `GET /audit_logs`
- `GET /collaborators`

Manage these in the dashboard.

## Debug

```sh
REVCAT_DEBUG=api revcat events list      # logs full request/response (key redacted)
revcat doctor                            # top-level health check
revcat auth doctor                       # auth-specific
```

## License

MIT
