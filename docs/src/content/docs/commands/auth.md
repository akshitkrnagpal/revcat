---
title: auth
description: Manage RevenueCat authentication profiles.
---

revcat stores credentials in your OS keychain by default. Each set of credentials is a "profile" with a name, secret key, and project id.

For CI, pass `--bypass-keychain` (or set `REVCAT_BYPASS_KEYCHAIN=1`) to use a local file instead, or pass `REVCAT_API_KEY` directly.

## Subcommands

| Command | Description |
| --- | --- |
| `auth login` | Save a RevenueCat secret key as a named profile |
| `auth status` | Show the active auth profile (`--validate` hits the API) |
| `auth doctor` | Self-diagnose auth setup |
| `auth use <name>` | Set the default auth profile |
| `auth list` | List stored auth profiles |
| `auth logout [name]` | Remove a stored auth profile (`--all` wipes them all) |

## Examples

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat auth login --name my-app --secret-key sk_xxx --project-id proj_xxx --no-verify   # CI
revcat auth status --validate
revcat auth doctor
revcat auth use my-app
revcat auth list
revcat auth logout my-app
```

## Resolution order

Active profile is resolved in this order:

1. `REVCAT_API_KEY` env (synthesizes a one-shot profile)
2. `--profile <name>` flag
3. `REVCAT_PROFILE` env
4. `~/.revcat/active` (set by `revcat auth use`)
5. profile named `default`

See [Configuration](/reference/configuration/) for env vars and defaults.
