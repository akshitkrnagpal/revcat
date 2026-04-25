---
title: entitlements
description: Manage RevenueCat entitlements.
---

Entitlements are project-level access flags (e.g., `premium`, `pro`). Customers gain entitlements via products attached on offerings, or via promotional grants. Use [`revcat subscribers info <user_id>`](/commands/subscribers/) to see what a specific customer has.

## Subcommands

| Command | Description |
| --- | --- |
| `entitlements list` | List all entitlements in the active project |
| `entitlements view <id>` | Show one entitlement by lookup_key |
| `entitlements create` | Create an entitlement (`--id <key>` + `--display-name`, or `--file <path>`) |
| `entitlements update <id>` | Update an entitlement (`--display-name`, or `--file <path>`) |
| `entitlements delete <id>` | Delete an entitlement |
| `entitlements archive <id>` | Archive an entitlement |
| `entitlements unarchive <id>` | Unarchive an entitlement |
| `entitlements products <id>` | List products attached to an entitlement |
| `entitlements attach <id> <product_id> ...` | Attach product(s) to an entitlement |
| `entitlements detach <id> <product_id> ...` | Detach product(s) from an entitlement |

## Examples

```sh
revcat entitlements list
revcat entitlements create --id premium --display-name "Premium"
revcat entitlements attach premium prod_app_monthly prod_app_annual
revcat entitlements products premium
revcat entitlements archive premium -y
```

Aliases: `ent`.
