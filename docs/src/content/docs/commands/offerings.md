---
title: offerings
description: Manage RevenueCat offerings
---

An offering is a presentation grouping of packages displayed on a
paywall. Each project has 0..N offerings; exactly one is "current" and is
returned by SDKs that ask for the current offering.

To set an offering current along with a paywall config in one shot, use
`revcat publish offering`.

## Subcommands

| Command | Description |
| --- | --- |
| `offerings archive <id>` | Archive an offering |
| `offerings create` | Create an offering |
| `offerings delete <id>` | Delete an offering |
| `offerings list` | List all offerings in the active project |
| `offerings preview` | Show what the SDK will receive from /v1/subscribers/{user}/offerings |
| `offerings set-current <id>` | Promote an offering to current |
| `offerings unarchive <id>` | Unarchive an offering |
| `offerings update <id>` | Update an offering |
| `offerings view <id>` | Show one offering by lookup_key |

Aliases: `offer`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat offerings list
revcat offerings view default
revcat offerings create --id pro --display-name "Pro"
revcat offerings set-current pro
```

Aliases: `offer`.
