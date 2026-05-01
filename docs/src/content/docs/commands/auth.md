---
title: auth
description: Manage RevenueCat OAuth credentials.
---

revcat authenticates against RevenueCat via OAuth (PKCE). One browser login populates a global profile in your OS keychain; running `revcat init` inside a repo writes a per-directory `.revcat/config.json` (gitignored, mode 0600) so agents and sandboxes operating in the directory inherit the credential without keychain access.

For Linux containers without secret-service, pass `--bypass-keychain` (or set `REVCAT_BYPASS_KEYCHAIN=1`) to use `~/.revcat/config.json` instead. For headless CI / fresh sandboxes, set `REVCAT_REFRESH_TOKEN` to skip both keychain and login flow.

## Subcommands

| Command | Description |
| --- | --- |
| `auth login` | Run the browser OAuth flow and save tokens |
| `auth status` | Show the resolved credential and where it came from (`--validate` hits the API) |
| `auth doctor` | Self-diagnose auth setup |
| `auth use <name>` | Set the default global profile |
| `auth list` | List stored global profiles |
| `auth logout [name]` | Remove a stored profile (`--all` wipes them all) |

## Storage tiers

| Tier | Path | Used when |
| --- | --- | --- |
| keychain | OS keychain | default for `auth login` |
| global file | `~/.revcat/config.json` | `--bypass-keychain` or `REVCAT_BYPASS_KEYCHAIN=1` |
| local file | `./.revcat/config.json` (walked up from cwd) | written by `revcat init` |

Resolution order: `REVCAT_REFRESH_TOKEN` env > walked-up local file > global active profile.

## Examples

```sh
# First-time setup
revcat auth login                      # browser OAuth, default profile
cd ~/your-repo && revcat init          # bind this repo to a project

# Multi-account
revcat auth login --name work
revcat auth login --name personal
revcat auth use personal               # default for global commands

# Status / health
revcat auth status --validate
revcat auth doctor

# Cleanup
revcat auth logout work                # remove a profile
revcat auth logout --all               # wipe all profiles
```

## CI / headless

```sh
# In a fresh container with no browser
export REVCAT_REFRESH_TOKEN=rtk_...
export REVCAT_PROJECT_ID=proj_...
revcat offerings list
```

The refresh token is account-scoped; treat it like a long-lived credential and store it in your CI secret manager.

## Resolution order (full)

Credential:

1. `REVCAT_REFRESH_TOKEN` env (synthesizes a virtual profile)
2. Walked-up `./.revcat/config.json`
3. Global keychain or `~/.revcat/config.json` for the active profile

Active global profile name: `--profile <name>` flag > `REVCAT_PROFILE` env > `~/.revcat/active` (set by `auth use`) > `default`.

Project id: `--project-id` flag > `REVCAT_PROJECT_ID` env > resolved credential's bound project (local config or env hatch) > walked-up `revcat.toml`.

See [Configuration](/reference/configuration/) for the full env var list.
