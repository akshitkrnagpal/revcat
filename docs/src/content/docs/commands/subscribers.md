---
title: subscribers
description: Inspect and manage RevenueCat subscribers (a.k.a. customers, app users).
---

Subscribers (a.k.a. customers, app users) are the end-users of your app. revcat treats them as the unit of debugging - one command surfaces their entitlements, subscriptions, purchases, and aliases in a single card.

## Subcommands

### Inspect

| Command | Description |
| --- | --- |
| `subscribers info <user_id>` | Full debug card: entitlements, subs, purchases, aliases, attribution |
| `subscribers list` | List customers in the active project (paged) |
| `subscribers attributes <user_id>` | Get / set subscriber attributes |
| `subscribers invoices <user_id>` | List invoices for a customer |
| `subscribers vc-balance <user_id>` | Show a customer's virtual currency balances |

### Manage

| Command | Description |
| --- | --- |
| `subscribers create <user_id>` | Pre-create a customer (migration / import) |
| `subscribers delete <user_id>` | Permanently delete (GDPR / test cleanup) |
| `subscribers transfer <src> <dst>` | Move entitlements / subscriptions between customers |
| `subscribers grant <user> <ent>` | Grant a promotional entitlement (`-d 7d`, `-r "ticket #2241"`) |
| `subscribers revoke <user> <ent>` | Revoke a promotional entitlement |
| `subscribers refund <sub_id> <txn_id>` | Refund a transaction on a subscription |
| `subscribers override-offering <user> [offering]` | Force a specific offering (`--clear` to reset) |
| `subscribers restore-google-play <user>` | Force a Play Store entitlement re-check |
| `subscribers vc-tx <user>` | Post a virtual currency transaction (credit or debit) |
| `subscribers vc-set-balance <user>` | Directly set a virtual currency balance |

## Examples

```sh
revcat subscribers info app_user_123
revcat subscribers grant app_user_123 premium -d 7d -r "support ticket #2241"
revcat subscribers attributes app_user_123 --set lang=en --set plan_intent=pro
revcat subscribers transfer app_user_old app_user_new -y
revcat subscribers refund sub_xxx txn_xxx
```

## Grant duration

`-d / --duration` accepts:

- `forever` / `lifetime`
- `7d`, `30d`, `90d` (days)
- `1m`, `3m`, `6m`, `1y`, `2y`
- RC built-ins: `daily`, `weekly`, `monthly`, `yearly`

Aliases: `customers`, `subs`.
