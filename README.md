# revcat

The RevenueCat CLI. Run your RevenueCat project from the terminal instead of clicking through the dashboard.

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat subscribers info app_user_123
revcat events tail --type INITIAL_PURCHASE,CANCELLATION
revcat publish offering pro --paywall ./paywalls/pro.json
```

## Why

RevenueCat ships a dashboard, REST API, and (2025) an MCP server, but no first-party CLI. Common workflows like "debug a customer's entitlements", "tail recent events", and "set an offering as current with a fresh paywall" still mean clicking through the dashboard.

revcat composes those workflows behind one command each. Output is a colored table when you're at a terminal and JSON when you're piping into a script - no `--json` ceremony.

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

## Commands

```
revcat subscribers info <user_id>                 # full debug card
revcat subscribers grant <user> <ent> -d 7d -r "ticket #2241"
revcat subscribers revoke <user> <ent>
revcat subscribers refund <user> -t <txn_id>

revcat entitlements list | view <id>
revcat offerings    list | view <id>
revcat packages     list [-o <offering>]

revcat publish offering <id> [--paywall ./paywall.json] [--current] [-y]

revcat events list [--type ...] [--since 1h]
revcat events tail [--type ...] [--since 1h] [--interval 5s]

revcat doctor
revcat auth login | status | doctor | use | list | logout
revcat completion bash | zsh | fish
revcat version
```

## Output

By default, output is TTY-aware:

- **Interactive terminal**: tables (lipgloss) with color
- **Piped or in CI**: JSON, one object per row

Override with `--output table|json|csv|markdown` or env `REVCAT_DEFAULT_OUTPUT`. Use `--pretty` for indented JSON.

`revcat events tail` emits one JSON object per line in JSON mode (ndjson), so you can pipe into `jq` mid-stream.

## Examples

Debug a paywall report from support:

```sh
revcat subscribers info app_user_123
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

Compute a daily revenue summary in CI:

```sh
revcat events list --type INITIAL_PURCHASE,RENEWAL --since 24h --output json | \
    jq '[.[] | .price] | add'
```

## Debug

```sh
REVCAT_DEBUG=api revcat events list      # logs full request/response (key redacted)
revcat doctor                            # top-level health check
revcat auth doctor                       # auth-specific
```

## License

MIT
