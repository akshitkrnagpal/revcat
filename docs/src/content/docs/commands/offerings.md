---
title: offerings
description: Manage RevenueCat offerings.
---

An offering is a presentation grouping of packages displayed on a paywall. Each project has 0..N offerings; exactly one is "current" and is returned by SDKs that ask for the current offering.

To set an offering current along with a paywall config in one shot, use [`revcat publish offering`](/commands/publish/).

## Subcommands

| Command | Description |
| --- | --- |
| `offerings list` | List all offerings in the active project (current marked with `*`) |
| `offerings view <id>` | Show one offering by lookup_key (includes packages on TTY) |
| `offerings create` | Create an offering (`--id` + `--display-name`, or `--file`) |
| `offerings update <id>` | Update an offering (`--display-name`, or `--file`) |
| `offerings delete <id>` | Delete an offering |
| `offerings archive <id>` | Archive |
| `offerings unarchive <id>` | Unarchive |
| `offerings set-current <id>` | Promote an offering to current |

## Examples

```sh
revcat offerings list
revcat offerings view default
revcat offerings create --id pro --display-name "Pro"
revcat offerings set-current pro
```

Aliases: `offer`.
