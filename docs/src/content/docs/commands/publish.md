---
title: publish
description: One-shot deploy verbs (offering, paywall, ...).
---

Publish-style verbs compose several API calls behind a single command. The intent is to mirror the dashboard's higher-level actions ("set as current", "deploy paywall") rather than mirror REST endpoints.

## Subcommands

| Command | Description |
| --- | --- |
| `publish offering <id>` | Set an offering as current and / or push a paywall config |

## `publish offering` flags

| Flag | Description |
| --- | --- |
| `--current` | Set the offering as current |
| `--no-current` | Do NOT set as current (overrides the default when only `--paywall` is used) |
| `--paywall <path>` | Path to a paywall config JSON file to PUT |
| `-y, --confirm` | Skip the confirmation prompt |
| `--dry-run` | Print the plan without making changes |

## Behavior

1. Verify the offering exists in the active project.
2. (Optional) Validate and PUT the paywall config from `--paywall`. If the canonicalized hash matches what's live, the step is skipped silently.
3. (Optional) Promote the offering to current.

The plan is printed before execution; pass `--confirm` to skip the prompt, or `--dry-run` to preview.

## Examples

```sh
revcat publish offering default --current --confirm
revcat publish offering pro --paywall ./paywalls/pro.json --current
revcat publish offering pro --paywall ./paywalls/pro.json --dry-run
```
