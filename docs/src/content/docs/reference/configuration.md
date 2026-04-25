---
title: Configuration
description: Environment variables, credential storage, and output controls.
---

revcat reads its configuration from environment variables and the OS keychain. There is no global config file (auth profiles in the keychain are the only persisted state).

## Auth resolution order

When you run any command that needs the API, revcat resolves the active profile in this order:

1. `REVCAT_API_KEY` env var - synthesizes a one-shot profile (highest precedence)
2. `--profile <name>` flag
3. `REVCAT_PROFILE` env var
4. `~/.revcat/active` (set by `revcat auth use <name>`)
5. profile named `default`

## Environment variables

| Variable | Purpose |
| --- | --- |
| `REVCAT_API_KEY` | One-shot RC v2 secret key. Synthesizes a profile, takes precedence over everything else. |
| `REVCAT_PROFILE` | Active profile name when `--profile` is not set. |
| `REVCAT_PROJECT_ID` | Override the stored project id on the active profile. |
| `REVCAT_BYPASS_KEYCHAIN` | Set to `1` to read/write profiles from `./.revcat/config.json` instead of the OS keychain. |
| `REVCAT_DEFAULT_OUTPUT` | Default output format (`table`, `json`, `csv`, `markdown`) when `--output` is not set. |
| `REVCAT_DEBUG` | Set to `api` to log full request / response (key redacted). |
| `NO_COLOR` | Standard env that disables color when set. |

## Credential storage

By default, profiles live in the OS keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager) under service `revcat`, account = profile name.

For CI / Docker, pass `--bypass-keychain` (or set `REVCAT_BYPASS_KEYCHAIN=1`). Profiles are written to `./.revcat/config.json`. revcat creates a `.gitignore` in `./.revcat/` on first write so the file isn't committed by accident.

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
REVCAT_DEBUG=api revcat metrics overview   # full request/response, key redacted
revcat doctor                         # top-level health check
revcat auth doctor                    # auth-specific
```
