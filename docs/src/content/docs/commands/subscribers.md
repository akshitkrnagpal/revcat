---
title: subscribers
description: Inspect and manage RevenueCat subscribers
---

Subscribers (a.k.a. customers, app users) are the end-users of your app.
revcat treats them as the unit of debugging - one command surfaces their
entitlements, subscriptions, purchases, and aliases in a single card.

## Subcommands

| Command | Description |
| --- | --- |
| `subscribers attributes <user_id>` | Get or set subscriber attributes |
| `subscribers create <user_id>` | Pre-create a customer (migration / import) |
| `subscribers delete <user_id>` | Permanently delete a customer (GDPR / test cleanup) |
| `subscribers grant <user_id> <entitlement>` | Grant a promotional entitlement to a subscriber |
| `subscribers info <user_id>` | Show a full debug card for a subscriber |
| `subscribers invoices <user_id>` | List invoices for a customer |
| `subscribers list` | List customers in the active project (paged) |
| `subscribers refund <subscription_id> <transaction_id>` | Refund a transaction on a subscription |
| `subscribers revoke <user_id> <entitlement>` | Revoke a promotional entitlement from a subscriber |
| `subscribers transfer <src_user_id> <dst_user_id>` | Transfer entitlements/subscriptions from one customer to another |

Aliases: `customers`, `subs`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat subscribers info app_user_123
revcat subscribers grant app_user_123 premium -d 7d
revcat subscribers attributes app_user_123 --set lang=en --set plan_intent=pro
revcat subscribers transfer app_user_old app_user_new -y
revcat subscribers refund sub_xxx txn_xxx
```

## Grant duration

`-d / --duration` accepts:

- `forever` / `lifetime` (~100 years out)
- `7d`, `30d`, `90d` (days)
- `1m`, `3m`, `6m` (calendar months)
- `1y`, `2y`, `5y` (calendar years)

revcat translates duration to an absolute `expires_at` (Unix ms) before sending - that's what RC v2 wants.

## Attributes

`subscribers attributes <user>` lists current attributes when called with no flags. Pass `--set key=value` (repeatable) to upsert. revcat normalizes input into the `[{name, value}]` shape v2 expects.

## What's not exposed (RC v2 limitations)

These actions don't have v2 endpoints. Manage in the dashboard:

- override an offering for one customer
- force a Play Store entitlement re-check
- per-customer virtual currency balances and transactions

Aliases: `customers`, `subs`.
