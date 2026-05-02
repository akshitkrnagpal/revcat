---
title: apps
description: Manage RevenueCat apps (per-platform inside a project).
---

Each project has one app per platform/storefront (one for iOS, one for Android, etc.). v2 exposes the full lifecycle: create, update, delete, plus reads.

## Subcommands

| Command | Description |
| --- | --- |
| `apps list` | List apps in the active project |
| `apps view <id>` | Show one app |
| `apps public-keys <app_id>` | List the public SDK keys for an app |
| `apps storekit-config <app_id>` | Print the StoreKit configuration for an iOS app (raw JSON) |
| `apps create` | Create an app (`--type` shortcut for the common platforms, `--file` for the rest) |
| `apps update <app_id>` | Update an app (POST at the same path as GET; `--name` shortcut, `--file` for arbitrary fields) |
| `apps delete <app_id>` | Delete an app (hard delete; can 409 if dependent resources exist) |

## Read examples

```sh
revcat apps list
revcat apps view app_abc123
revcat apps public-keys app_abc123
revcat apps storekit-config app_abc123 | jq .
```

## Create

The v2 body is discriminated by `type`. Shortcut flags cover the most common platforms; for everything else pass `--file`.

```sh
# iOS
revcat apps create --type app_store --bundle com.acme.app --name "Acme iOS"

# Android
revcat apps create --type play_store --package com.acme.app --name "Acme Android"

# Anything else (Stripe, rc_billing, paddle, roku, mac_app_store, or app_store
# with optional shared_secret / App Store Connect API key, etc.)
revcat apps create --file ./new-stripe.json

# stdin
echo '{"name":"Web","type":"stripe","stripe":{"stripe_account_id":"acct_..."}}' \
  | revcat apps create --file -
```

## Update

`v2` uses `POST` (not `PATCH`) at `/v2/projects/{project_id}/apps/{app_id}` for updates. revcat does the right thing under the hood.

```sh
# Rename
revcat apps update app_abc --name "Acme iOS (renamed)"

# Patch arbitrary fields - send any nested field as null in the JSON to clear it
revcat apps update app_abc --file ./patch.json
```

Sample `patch.json` to clear a stored shared_secret:

```json
{ "app_store": { "shared_secret": null } }
```

## Delete

Hard delete. Confirm prompt in interactive mode; pass `-y/--confirm` to skip.

```sh
revcat apps delete app_abc           # prompts
revcat apps delete app_abc --confirm # skips prompt (CI / scripts)
```

If the app has dependent resources (offerings, products, etc.) the API returns 409. Drain those first or back out the dependents in the dashboard.
