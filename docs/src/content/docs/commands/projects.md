---
title: projects
description: Inspect RevenueCat projects.
---

A project is RevenueCat's top-level container - one per app or app family. revcat is bound to a single project per profile (the one the secret key has access to). Use these commands to inspect the project and switch between profiles for different ones.

Project create, app CRUD, audit logs, and collaborators all require a higher key tier than the per-project secret key revcat uses, so they are not exposed here. Manage those in the dashboard.

## Subcommands

| Command | Description |
| --- | --- |
| `projects list` | List projects accessible to the active secret key |
| `projects view [id]` | Show one project by id (defaults to the active profile's project) |

## Examples

```sh
revcat projects list
revcat projects view              # the active profile's project
revcat projects view proj_abc123
```

Aliases: `proj`.
