---
title: projects
description: Inspect RevenueCat projects.
---

A project is RevenueCat's top-level container - one per app or app family. revcat resolves the active project from `--project-id`, `REVCAT_PROJECT_ID`, the local `.revcat/config.json` written by `revcat init`, or `revcat.toml`.

Project create and collaborator CRUD aren't exposed by the v2 API revcat targets. Manage those in the dashboard.

## Subcommands

| Command | Description |
| --- | --- |
| `projects list` | List projects accessible to the active credential |
| `projects view [id]` | Show one project by id (defaults to the resolved project) |

## Examples

```sh
revcat projects list
revcat projects view              # the resolved project
revcat projects view proj_abc123
```

Aliases: `proj`.
