---
title: '"User paid but doesn''t have premium"'
description: Walk through a real support ticket end-to-end with revcat.
---

The single most common RC support flow: a customer contacts you saying they paid but the app doesn't unlock the paid features. Here's how to debug it without leaving the terminal.

![revcat customer debug demo](https://raw.githubusercontent.com/akshitkrnagpal/revcat/main/demo/customer-debug.gif)

## 1. Find the customer

If you have the **store transaction id** from their receipt or App Store / Play / Stripe email:

```sh
revcat subscriptions search 1000000123456789
```

Returns one or more subscription rows. Note the `customer` column - that's the `app_user_id` to use next.

If the customer told you their **app user id** directly, skip to step 2.

## 2. Pull the full debug card

```sh
revcat subscribers info app_user_123
```

Sample output:

```
subscriber
╭───────────────────────────────╮
│  id           app_user_123    │
│  project      proj_xxx        │
│  first seen   2024-08-12      │
│  last seen    today (3h ago)  │
╰───────────────────────────────╯

active entitlements (0)
  (none)

subscriptions (1)
┌──────────────┬────────┬─────────────────────┬────────────┐
│ product      │ status │ current period ends │ store      │
├──────────────┼────────┼─────────────────────┼────────────┤
│ app.monthly  │ active │ 2026-05-25 (in 29d) │ app_store  │
└──────────────┴────────┴─────────────────────┴────────────┘
```

The mismatch is right there: there's an active App Store subscription, but no active entitlement. The product is selling but it isn't wired to grant `premium`.

## 3. Confirm the catalog wiring

```sh
revcat entitlements list
revcat entitlements products premium
```

If the second command returns an empty list (or doesn't include `app.monthly`), that's the bug: the product isn't attached to the entitlement.

## 4. Fix the wiring

```sh
revcat products list                                # find the product id
revcat entitlements attach premium prod_xxx         # attach
```

The customer's next purchase / renewal will grant the entitlement. To unblock them right now, grant it manually as a one-off:

```sh
revcat subscribers grant app_user_123 premium -d 30d
```

That gives them 30 days while the catalog fix propagates. Tag the audit reason if you want it recorded:

```sh
revcat subscribers attributes app_user_123 --set support_grant=ticket_2241
```

## 5. Verify

```sh
revcat subscribers info app_user_123
```

`active entitlements` should now show `premium` with an expiry. Reply to the ticket.

## What this saved you

Without revcat, the same flow needs you to:

1. Open the dashboard, search for the user.
2. Click into their profile.
3. Click "subscriptions" tab.
4. Open another tab to entitlements > inspect the product attachments.
5. Click "Grant promotional entitlement" modal, fill in dates, submit.
6. Click back to the user to verify.

That's six page loads vs three commands.
