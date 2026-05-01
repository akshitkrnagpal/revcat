---
title: apps
description: Inspect RevenueCat apps (per-platform inside a project).
---

Each project has one app per platform/storefront (one for iOS, one for Android, etc.). App create / update / delete isn't exposed by the v2 API revcat targets; this group is read-only.

## Subcommands

| Command | Description |
| --- | --- |
| `apps list` | List apps in the active project |
| `apps view <id>` | Show one app |
| `apps public-keys <app_id>` | List the public SDK keys for an app |
| `apps storekit-config <app_id>` | Print the StoreKit configuration for an iOS app (raw JSON) |

## Examples

```sh
revcat apps list
revcat apps view app_abc123
revcat apps public-keys app_abc123
revcat apps storekit-config app_abc123 | jq .
```
