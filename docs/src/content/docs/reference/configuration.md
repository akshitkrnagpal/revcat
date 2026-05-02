---
title: Configuration
description: Environment variables, credential storage, and output controls.
---

revcat reads its configuration from environment variables, `~/.revcat/config.json`, and (when present) `./.revcat/config.json` and `./revcat.toml` walked up from cwd.

## Credential resolution order

1. `REVCAT_REFRESH_TOKEN` env var (synthesizes a virtual profile; highest precedence, in-memory only)
2. Walked-up `./.revcat/config.json` (written by `revcat init`)
3. Active global profile from `~/.revcat/config.json` (mode 0600, written by `revcat auth login`)

Within step 3, the active global profile name is: `--profile <name>` flag > `REVCAT_PROFILE` env > `~/.revcat/active` (set by `revcat auth use`) > `default`.

## Project id resolution order

1. `--project-id` flag
2. `REVCAT_PROJECT_ID` env var
3. Resolved credential's bound project (from local config or env hatch)
4. Walked-up `./revcat.toml`

## Environment variables

| Variable | Purpose |
| --- | --- |
| `REVCAT_REFRESH_TOKEN` | OAuth refresh token. Synthesizes a virtual profile, refreshes in-memory each invocation. The headless / CI / sandbox hatch. |
| `REVCAT_PROJECT_ID` | Override the project id used by project-scoped commands. |
| `REVCAT_PROFILE` | Active global profile name when `--profile` is not set. |
| `REVCAT_OAUTH_CLIENT_ID` | Override the embedded public OAuth client id. |
| `REVCAT_DEFAULT_OUTPUT` | Default output format (`table`, `json`, `csv`, `markdown`) when `--output` is not set. |
| `REVCAT_DEBUG` | Set to `api` to log full request / response (token redacted). |
| `NO_COLOR` | Standard env that disables color when set. |

## Storage tiers

| Tier | Path | Used when |
| --- | --- | --- |
| global file | `~/.revcat/config.json` (mode 0600) | written by `revcat auth login` |
| local file | `./.revcat/config.json` (walked up from cwd, mode 0600) | written by `revcat init`, gitignored |

The active-profile pointer (`~/.revcat/active`) is a plain text file, not a secret.

## Output

revcat is TTY-aware:

| Stdout | Default format |
| --- | --- |
| Interactive terminal | Tables (lipgloss + color) |
| Piped or redirected | JSON |

Force a format with `--output table|json|csv|markdown`. `--pretty` indents JSON.

## Debugging

```sh
REVCAT_DEBUG=api revcat metrics overview   # full request/response, token redacted
revcat doctor                              # top-level health check
revcat auth doctor                         # auth-specific (incl. toml/local mismatch)
```
