---
title: projects
description: Inspect and create RevenueCat projects.
---

A project is RevenueCat's top-level container - one per app or app family. revcat resolves the active project from `--project-id`, `REVCAT_PROJECT_ID`, the local `.revcat/config.json` written by `revcat init`, or `revcat.toml`.

v2 exposes project create + list. There is no v2 update or delete by id; manage those in the dashboard.

## Subcommands

| Command | Description |
| --- | --- |
| `projects list` | List projects accessible to the active credential |
| `projects view [id]` | Show one project by id (defaults to the resolved project) |
| `projects create --name <name>` | Create a new project at the account level |

## Examples

```sh
revcat projects list
revcat projects view              # the resolved project
revcat projects view proj_abc123

revcat projects create --name "My App"
revcat projects create --name staging --output json | jq -r .id
```

After `projects create`, run `revcat init --project-id <new_id>` inside your repo to bind it.

Aliases: `proj`.
