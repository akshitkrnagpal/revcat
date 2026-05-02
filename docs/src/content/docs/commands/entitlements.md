---
title: entitlements
description: Manage RevenueCat entitlements
---

Entitlements are project-level access flags (e.g., "premium", "pro").
Customers gain entitlements via products attached on offerings, or via
promotional grants. Use `revcat subscribers info <user_id>` to see what
a specific customer has.

## Subcommands

| Command | Description |
| --- | --- |
| `entitlements archive <id>` | Archive an entitlement |
| `entitlements attach <id> <product_id> [<product_id> ...]` | Attach product(s) to an entitlement |
| `entitlements create` | Create an entitlement |
| `entitlements delete <id>` | Delete an entitlement |
| `entitlements detach <id> <product_id> [<product_id> ...]` | Detach product(s) from an entitlement |
| `entitlements list` | List all entitlements in the active project |
| `entitlements products <id>` | List products attached to an entitlement |
| `entitlements unarchive <id>` | Unarchive an entitlement |
| `entitlements update <id>` | Update an entitlement |
| `entitlements view <id>` | Show one entitlement by lookup_key |

Aliases: `ent`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat entitlements list
revcat entitlements create --id premium --display-name "Premium"
revcat entitlements attach premium prod_app_monthly prod_app_annual
revcat entitlements products premium
revcat entitlements archive premium -y
```

Aliases: `ent`.
