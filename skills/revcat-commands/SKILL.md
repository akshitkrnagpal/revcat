---
name: revcat-commands
description: Use when constructing or verifying a revcat command, or when looking up which commands exist for a given resource. Reference for every subcommand with real syntax and examples for entitlements, offerings, packages, products, paywalls, subscribers, subscriptions, webhooks, metrics, charts, audit-logs, virtual-currencies, apps. Triggers on "revcat <command>", "how do I X with revcat", "which revcat command", "is there a revcat for", "what flags does revcat <subcommand> take", "list / view / create / delete / update with revcat".
---

# revcat - command reference

Commands below are the real surface as of the latest revcat. Verify with `revcat <group> --help` if anything looks off.

Conventions:

- `<id>` = literal id RC returns. Often a lookup_key like `premium`, sometimes an internal id like `pkg_xxx`.
- `--file <path>` = JSON body on disk (or `-` for stdin). Used wherever the v2 schema is broad.
- `-y / --confirm` = skip the destructive-action prompt.
- **Flags are scoped to the subcommand.** Don't assume a flag that works on one subcommand works on another. `revcat packages list` takes `--offering`, NOT `--app-id`. `revcat products list` has no per-app filter at all (filter by `app_id` in the JSON output instead). Always check `--help` on the exact subcommand if a guess doesn't take.

## auth

```sh
revcat auth login [--name my-app] [--client-id <id>]   # browser OAuth, only mode
revcat auth status [--validate]
revcat auth doctor
revcat auth use <name>
revcat auth list
revcat auth logout [<name>] [--all] [-y]
```

## init

```sh
revcat init                                    # interactive: pick project + apps
revcat init --project-id proj_xxx --no-apps    # scripted
revcat init --force                            # overwrite existing files
revcat init --no-local-creds                   # write only revcat.toml, skip .revcat/config.json
```

Writes `revcat.toml` (committed: project_id + apps) and `.revcat/config.json` (gitignored, mode 0600: credentials + project_id + apps). Auto-appends `.revcat/` to `.gitignore`.

## projects, apps

```sh
revcat projects list
revcat projects view [<id>]                   # default: resolved project
revcat projects create --name "My App"        # returns the new project_id

revcat apps list
revcat apps view <app_id>
revcat apps public-keys <app_id>
revcat apps storekit-config <app_id>          # raw JSON

# Create: shortcut flags for the common platforms
revcat apps create --type app_store --bundle com.acme.app --name "Acme iOS"
revcat apps create --type play_store --package com.acme.app --name "Acme Android"
# Anything else (stripe, rc_billing, paddle, roku, mac_app_store, optional fields)
revcat apps create --file ./app.json   # or --file - for stdin

# Update: --name shortcut, or --file for arbitrary fields. Send a nested
# field as null in the JSON to clear it.
revcat apps update <app_id> --name "renamed"
revcat apps update <app_id> --file ./patch.json

# Delete: hard delete. -y/--confirm to skip the prompt. Returns 409 if
# the app has dependent resources.
revcat apps delete <app_id> [-y]
```

## Catalog

```sh
revcat entitlements list
revcat entitlements view <id>
revcat entitlements create --id premium --display-name "Premium"
revcat entitlements create --file ./entitlement.json
revcat entitlements update <id> --display-name "New name" | --file ./patch.json
revcat entitlements delete <id> [-y]
revcat entitlements archive <id>
revcat entitlements unarchive <id>
revcat entitlements products <id>
revcat entitlements attach <id> <product_id> [<product_id> ...]
revcat entitlements detach <id> <product_id> [<product_id> ...]

revcat offerings list                           # current marked with *
revcat offerings view <id>
revcat offerings create --id pro --display-name "Pro" | --file ./offering.json
revcat offerings update <id> --display-name "..." | --file ./patch.json
revcat offerings delete <id> [-y]
revcat offerings archive <id>
revcat offerings unarchive <id>
revcat offerings set-current <id>

revcat packages list [--offering <id>]
revcat packages view <pkg_id>
revcat packages create --file ./package.json    # body: identifier + offering_id
revcat packages update <pkg_id> --file ./patch.json
revcat packages delete <pkg_id> [-y]
revcat packages products <pkg_id>
revcat packages attach <pkg_id> <product_id> ...
revcat packages detach <pkg_id> <product_id> ...

revcat products list
revcat products view <id>
revcat products create --file ./product.json    # body: store_identifier, type, display_name, app_id
revcat products update <id> --display-name "..." | --file ./patch.json
revcat products delete <id> [-y]
revcat products archive <id>
revcat products unarchive <id>
revcat products push-to-store <id>

revcat paywalls list
revcat paywalls view <id>                       # raw JSON
revcat paywalls create --file ./paywall.json
revcat paywalls delete <id> [-y]
```

## Customers (subscribers)

```sh
revcat subscribers info <user_id>               # the headline debug card
revcat subscribers list                          # paged
revcat subscribers create <user_id> [--file ./body.json]
revcat subscribers delete <user_id> [-y]
revcat subscribers transfer <src> <dst> [-y]
revcat subscribers grant <user> <ent> -d <dur> [-y]
revcat subscribers revoke <user> <ent> [-y]
revcat subscribers refund <subscription_id> <transaction_id> [-y]

revcat subscribers attributes <user>             # get
revcat subscribers attributes <user> --set k=v --set k2=v2 [--file ./attrs.json]

revcat subscribers invoices <user>
```

`grant -d / --duration` accepts: `forever` / `lifetime` (~100 years), `7d` / `30d` / `90d`, `1m` / `3m` / `6m` (calendar months), `1y` / `2y`. revcat translates to an absolute `expires_at` (Unix ms) before sending.

`subscribers revoke` is implemented as "grant with a near-future expires_at" because v2 has no first-class revoke endpoint. The entitlement appears expired on the next read.

`subscribers attributes --set k=v` is normalized to v2's `[{name, value}]` shape automatically.

## Subscriptions, purchases, invoices

```sh
revcat subscriptions search <store_id>           # find by App Store / Play / Stripe id
revcat subscriptions view <sub_id>
revcat subscriptions transactions <sub_id>
revcat subscriptions entitlements <sub_id>
revcat subscriptions management-url <sub_id>
revcat subscriptions cancel <sub_id> [-y]        # Web Billing
revcat subscriptions refund <sub_id> [-y]        # Web Billing

revcat purchases search <store_id>
revcat purchases view <purchase_id>
revcat purchases entitlements <purchase_id>
revcat purchases refund <purchase_id> [-y]       # Web Billing

revcat invoices view <invoice_id>                # raw JSON
```

## Activity

```sh
revcat metrics overview

revcat charts options <chart_name>
revcat charts get <chart_name> [--start YYYY-MM-DD] [--end YYYY-MM-DD] \
    [--period day|week|month] [--filter k=v ...]

revcat audit-logs list

revcat collaborators list                # alias: members
```

Valid chart names: `actives`, `actives_movement`, `actives_new`, `arr`, `churn`, `cohort_explorer`, `conversion_to_paying`, `customers_new`, `ltv_per_customer`, `ltv_per_paying_customer`, `mrr`, `mrr_movement`, `refund_rate`, `revenue`, `subscription_retention`, `subscription_status`, `trials`, `trials_movement`.

Lifecycle events (purchases, renewals, cancellations, refunds, ...) are NOT exposed via REST. Use `revcat webhooks create` to subscribe your endpoint to them. Common event types delivered via webhook: `INITIAL_PURCHASE`, `RENEWAL`, `CANCELLATION`, `EXPIRATION`, `TRIAL_STARTED`, `TRIAL_CONVERTED`, `REFUND`, `BILLING_ISSUE`, `PRODUCT_CHANGE`.

## Verb-orchestrator

```sh
revcat publish offering <id> [--current] [--no-current] [--paywall ./paywall.json] [-y] [--dry-run]
```

Composes set-current + paywall PUT behind one plan. Existing paywall is fetched lazily; if its hash matches the file, the PUT is skipped.

## Integrations

```sh
revcat webhooks list
revcat webhooks view <id>
revcat webhooks create --name "..." --url https://... --events initial_purchase,renewal,cancellation
revcat webhooks create --file ./webhook.json
revcat webhooks update <id> --url ... | --events ... | --name ... | --file ./patch.json
revcat webhooks delete <id> [-y]

# webhooks notes:
# - event_types are LOWERCASE (initial_purchase, not INITIAL_PURCHASE).
#   --events accepts either form; revcat lowercases.
# - URL must pass RC's reachability check; example.com / localhost fail.
# - body shape on --file: {name, url, event_types:[...]}

revcat virtual-currencies list
revcat virtual-currencies view <CODE>                  # uppercase, e.g. COIN
revcat virtual-currencies create --name Coins --code COIN [--description "..."]
revcat virtual-currencies create --file ./vc.json      # body: {name, code, description?}
revcat virtual-currencies update <CODE> --name "..." [--description "..."] | --file ./patch.json
revcat virtual-currencies delete <CODE> [-y]
revcat virtual-currencies archive <CODE>
revcat virtual-currencies unarchive <CODE>
```

## Output controls (every command)

```sh
... --output json --pretty               # force JSON pretty
... --output table                       # force table even when piped
REVCAT_DEFAULT_OUTPUT=json revcat ...   # session default
NO_COLOR=1 revcat ...                    # disable color
REVCAT_DEBUG=api revcat ...              # log full request/response (key redacted)
```
