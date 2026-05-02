---
title: projects
description: Inspect RevenueCat projects
---

A project is RevenueCat's top-level container - one per app or app
family. revcat resolves the active project from --project-id,
REVCAT_PROJECT_ID, the local .revcat/config.json written by
`revcat init`, or revcat.toml.

v2 exposes project create + list. There is no v2 update or delete by
id; manage those in the dashboard.

## Subcommands

| Command | Description |
| --- | --- |
| `projects create` | Create a new project |
| `projects list` | List projects accessible to this credential |
| `projects view` | Show one project by id (defaults to the resolved project) |

Aliases: `proj`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

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
