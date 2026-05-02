---
title: auth
description: Manage RevenueCat authentication
---

revcat authenticates against RevenueCat via OAuth. One browser login
writes the credential to ~/.revcat/config.json (mode 0600). A per-repo
.revcat/config.json (written by `revcat init`) carries that credential
into the directory so agents and sandboxes inside the directory inherit
it without touching the global file.

Most users only need:

    revcat auth login                # browser OAuth
    cd ~/your/repo && revcat init    # bind this repo to a project
    revcat auth status

For CI / fresh sandboxes with no browser: set REVCAT_REFRESH_TOKEN
(and REVCAT_PROJECT_ID) to skip both file and login flow.

## Subcommands

| Command | Description |
| --- | --- |
| `auth doctor` | Self-diagnose auth setup |
| `auth list` | List stored auth profiles |
| `auth login` | Authenticate revcat against RevenueCat via OAuth |
| `auth logout` | Remove a stored auth profile |
| `auth status` | Show the active auth profile and resolved project |
| `auth use <profile>` | Set the default auth profile |

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Storage tiers

| Tier | Path | Used when |
| --- | --- | --- |
| global file | `~/.revcat/config.json` (mode 0600) | written by `revcat auth login` |
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
3. `~/.revcat/config.json` for the active profile

Active global profile name: `--profile <name>` flag > `REVCAT_PROFILE` env > `~/.revcat/active` (set by `auth use`) > `default`.

Project id: `--project-id` flag > `REVCAT_PROJECT_ID` env > resolved credential's bound project (local config or env hatch) > walked-up `revcat.toml`.

See [Configuration](/reference/configuration/) for the full env var list.
