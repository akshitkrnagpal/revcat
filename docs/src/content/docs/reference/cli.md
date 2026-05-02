---
title: CLI reference
description: Auto-generated reference for every revcat command. Regenerate with `go run ./scripts/gen-cli-reference`.
---

This page is generated from the cobra command tree (`go run ./scripts/gen-cli-reference`). It's the canonical surface — when prose elsewhere disagrees, this page wins.

## `revcat apps`

Manage RevenueCat apps (per-platform inside a project)

```text
Each project has one app per platform/storefront (one for iOS, one for
Android, etc.).

Read commands `list`, `view`, `public-keys`, `storekit-config` work on
any app. Write commands `create`, `update`, `delete` use the v2 app
endpoints; pass `--file <path>` for any non-trivial body since the
schema is wide and storefront-specific.
```

**Usage**

```sh
revcat apps
```

**Subcommands**: `create`, `delete`, `list`, `public-keys`, `storekit-config`, `update`, `view`

---

## `revcat apps create`

Create an app under the active project

```text
Create a new app. v2's app body is discriminated by --type and the
shape varies per storefront. For the common cases revcat takes
shortcut flags:

  --type app_store    --bundle <com.acme.app>
  --type play_store   --package <com.acme.app>
  --type amazon       --package <com.acme.app>

Anything more (Stripe, rc_billing, paddle, roku, mac_app_store, or any
optional fields like a shared_secret) needs --file <path.json>.

Examples:
    revcat apps create --type app_store --bundle com.acme.app --name "Acme iOS"
    revcat apps create --file ./apps/new-stripe.json
    revcat apps create --file - <<< '{"name":"Web","type":"stripe","stripe":{"stripe_account_id":"acct_..."}}'
```

**Usage**

```sh
revcat apps create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--bundle` | string | - | Bundle id (app_store / mac_app_store) |
| `--file` | string | - | Path to a JSON file with the full v2 body (use - for stdin) |
| `--name` | string | - | App name |
| `--package` | string | - | Package name (play_store / amazon) |
| `--type` | string | - | Storefront type: app_store \| play_store \| amazon \| mac_app_store (use --file for stripe / rc_billing / paddle / roku) |

---

## `revcat apps delete`

Delete an app (hard delete)

```text
Delete an app. This is a hard delete and can return 409 if the app
has dependent resources (offerings, products, etc.). Prefer to drain
those first.
```

**Usage**

```sh
revcat apps delete <app_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the confirmation prompt |

---

## `revcat apps list`

List apps in the active project

**Usage**

```sh
revcat apps list
```

---

## `revcat apps public-keys`

List the public SDK keys for an app

**Usage**

```sh
revcat apps public-keys <app_id>
```

---

## `revcat apps storekit-config`

Print the StoreKit configuration for an iOS app

**Usage**

```sh
revcat apps storekit-config <app_id>
```

---

## `revcat apps update`

Update an app

```text
Update an existing app. Pass --name to rename, or --file <path.json>
for a full v2 body (everything except 'type'). Send a nested field as
null in the JSON body to clear it (e.g. {"app_store":{"shared_secret":null}}).

Examples:
    revcat apps update app_abc --name "Acme iOS (renamed)"
    revcat apps update app_abc --file ./patch.json
```

**Usage**

```sh
revcat apps update <app_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--file` | string | - | Path to a JSON file with the v2 update body (use - for stdin) |
| `--name` | string | - | Rename the app |

---

## `revcat apps view`

Show one app

**Usage**

```sh
revcat apps view <id>
```

---

## `revcat audit-logs`

Inspect the project's audit log

**Usage**

```sh
revcat audit-logs
```

**Aliases**: `audit`

**Subcommands**: `list`

---

## `revcat audit-logs list`

List audit log entries

**Usage**

```sh
revcat audit-logs list
```

---

## `revcat auth`

Manage RevenueCat authentication

```text
revcat authenticates against RevenueCat via OAuth. One browser login
populates a global profile in your OS keychain; a per-repo
.revcat/config.json (written by `revcat init`) carries that credential
into the directory so agents and sandboxes work without keychain access.

Most users only need:

    revcat auth login            # browser OAuth, saves to keychain
    cd ~/your/repo && revcat init   # bind this repo to a project
    revcat auth status

For Linux containers without secret-service, pass --bypass-keychain
(or set REVCAT_BYPASS_KEYCHAIN=1) to use ~/.revcat/config.json instead.

For CI / fresh sandboxes with no browser: set REVCAT_REFRESH_TOKEN
(and REVCAT_PROJECT_ID) to skip both keychain and login flow.
```

**Usage**

```sh
revcat auth
```

**Subcommands**: `doctor`, `list`, `login`, `logout`, `status`, `use`

---

## `revcat auth doctor`

Self-diagnose auth setup

```text
Walk through the most common auth misconfigurations and report what's
working, what isn't, and how to fix it.
```

**Usage**

```sh
revcat auth doctor
```

---

## `revcat auth list`

List stored auth profiles

**Usage**

```sh
revcat auth list
```

---

## `revcat auth login`

Authenticate revcat against RevenueCat via OAuth

```text
Run the browser OAuth flow and store the resulting tokens. The OAuth
client is RevenueCat's public PKCE client by default; override with
REVCAT_OAUTH_CLIENT_ID or --client-id.

Tokens are written to your OS keychain. Pass --bypass-keychain (or set
REVCAT_BYPASS_KEYCHAIN=1) to use ~/.revcat/config.json (HOME) instead,
useful on Linux without secret-service or in containers.

After login, run `revcat init` inside your project directory to bind
a project_id (or set REVCAT_PROJECT_ID per command).

Examples:
  revcat auth login                          # browser, default profile
  revcat auth login --name work              # browser, named profile
  revcat auth login --client-id <id>         # custom OAuth client_id
```

**Usage**

```sh
revcat auth login [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--client-id` | string | - | OAuth client_id (defaults to REVCAT_OAUTH_CLIENT_ID env or the embedded public client) |
| `-n, --name` | string | - | Profile name (default: 'default') |
| `--no-verify` | bool | - | Skip the API check that the credentials work after login |
| `--scope` | stringSlice | - | OAuth scopes (default: revcat's full read/write set) |

---

## `revcat auth logout`

Remove a stored auth profile

```text
Delete a stored auth profile from the keychain (or local file). Pass
--all to wipe every profile.
```

**Usage**

```sh
revcat auth logout [<profile>] [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--all` | bool | - | Delete every stored profile |
| `-y, --yes` | bool | - | Skip confirmation prompt |

---

## `revcat auth status`

Show the active auth profile and resolved project

```text
Print the active credential and where it came from (local
.revcat/config.json, global keychain, file backend, or env hatch). Pass
--validate to also hit the RevenueCat API and confirm the token is
accepted.
```

**Usage**

```sh
revcat auth status [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-n, --name` | string | - | Override the active global profile (ignored when a local config is present) |
| `--validate` | bool | - | Hit the API to confirm the credential is accepted |

---

## `revcat auth use`

Set the default auth profile

```text
Set the active auth profile by writing the name to ~/.revcat/active.
Equivalent to passing --profile <name> on every command, or setting
REVCAT_PROFILE in your shell.
```

**Usage**

```sh
revcat auth use <profile>
```

---

## `revcat charts`

Project charts (revenue, active subs, conversion, etc.)

```text
Charts mirror the dashboard graphs: revenue, active subscribers,
conversion, MRR, churn, etc. Run `revcat charts options <name>` for
the supported filters before requesting data.
```

**Usage**

```sh
revcat charts
```

**Subcommands**: `get`, `options`

---

## `revcat charts get`

Fetch chart data (raw JSON)

```text
Fetch chart data. Filters via --filter key=value (repeatable). Date
range via --start / --end (YYYY-MM-DD). Period via --period (day | week | month).
```

**Usage**

```sh
revcat charts get <chart_name> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--end` | string | - | End date YYYY-MM-DD |
| `--filter` | stringArray | - | key=value (repeatable) |
| `--period` | string | - | day \| week \| month |
| `--start` | string | - | Start date YYYY-MM-DD |

---

## `revcat charts options`

Show the available filters/dimensions for a chart

**Usage**

```sh
revcat charts options <chart_name>
```

---

## `revcat collaborators`

Inspect project collaborators (members)

```text
List the people with access to the active RevenueCat project.

Read-only: v2 doesn't expose invite / role-change / remove via REST.
Manage membership in the dashboard.
```

**Usage**

```sh
revcat collaborators
```

**Aliases**: `members`

**Subcommands**: `list`

---

## `revcat collaborators list`

List collaborators on the active project

**Usage**

```sh
revcat collaborators list
```

---

## `revcat doctor`

Run a top-level health check

**Usage**

```sh
revcat doctor
```

---

## `revcat entitlements`

Manage RevenueCat entitlements

```text
Entitlements are project-level access flags (e.g., "premium", "pro").
Customers gain entitlements via products attached on offerings, or via
promotional grants. Use `revcat subscribers info <user_id>` to see what
a specific customer has.
```

**Usage**

```sh
revcat entitlements
```

**Aliases**: `ent`

**Subcommands**: `archive`, `attach`, `create`, `delete`, `detach`, `list`, `products`, `unarchive`, `update`, `view`

---

## `revcat entitlements archive`

Archive an entitlement

**Usage**

```sh
revcat entitlements archive <id>
```

---

## `revcat entitlements attach`

Attach product(s) to an entitlement

**Usage**

```sh
revcat entitlements attach <id> <product_id> [<product_id> ...]
```

---

## `revcat entitlements create`

Create an entitlement

```text
Create an entitlement. For the common case use shortcut flags:

    revcat entitlements create --id premium --display-name "Premium"

For arbitrary v2 fields, pass --file <path.json>.
```

**Usage**

```sh
revcat entitlements create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--display-name` | string | - | Display name shown in dashboards |
| `-f, --file` | string | - | Body as JSON file (or '-' for stdin) |
| `--id` | string | - | Entitlement lookup_key (e.g., premium) |

---

## `revcat entitlements delete`

Delete an entitlement

**Usage**

```sh
revcat entitlements delete <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat entitlements detach`

Detach product(s) from an entitlement

**Usage**

```sh
revcat entitlements detach <id> <product_id> [<product_id> ...]
```

---

## `revcat entitlements list`

List all entitlements in the active project

**Usage**

```sh
revcat entitlements list
```

---

## `revcat entitlements products`

List products attached to an entitlement

**Usage**

```sh
revcat entitlements products <id>
```

---

## `revcat entitlements unarchive`

Unarchive an entitlement

**Usage**

```sh
revcat entitlements unarchive <id>
```

---

## `revcat entitlements update`

Update an entitlement

**Usage**

```sh
revcat entitlements update <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--display-name` | string | - | New display name |
| `-f, --file` | string | - | Patch body as JSON |

---

## `revcat entitlements view`

Show one entitlement by lookup_key

**Usage**

```sh
revcat entitlements view <id>
```

---

## `revcat init`

Bootstrap project context (revcat.toml + .revcat/config.json)

```text
Bind the current directory to a RevenueCat project. Writes:

  - revcat.toml    (committed): project_id + optional apps
  - .revcat/config.json (gitignored, mode 0600): credentials + project_id

After init, every command run inside this directory inherits the project
context. Agents and sandboxes that have access to the directory can run
revcat without touching the user's keychain.

Interactive (default): lists projects you can access, prompts for one,
then optionally lists apps in that project and lets you tag them.

Scripted: pass --project-id (and optional --app-id, repeated). Skip the
apps block entirely with --no-apps. Skip the local creds copy with
--no-local-creds (writes only revcat.toml).
```

**Usage**

```sh
revcat init [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--app-id` | stringSlice | - | App ids to record (repeatable) |
| `--force` | bool | - | Overwrite an existing revcat.toml or .revcat/config.json |
| `--no-apps` | bool | - | Skip the apps section entirely |
| `--no-local-creds` | bool | - | Don't write .revcat/config.json (only revcat.toml) |
| `--path` | string | - | Where to write files (default: cwd) |

---

## `revcat invoices`

Inspect invoices

**Usage**

```sh
revcat invoices
```

**Subcommands**: `view`

---

## `revcat invoices view`

Show one invoice (raw JSON)

**Usage**

```sh
revcat invoices view <id>
```

---

## `revcat metrics`

Project-level revenue + subscription metrics

**Usage**

```sh
revcat metrics
```

**Subcommands**: `overview`

---

## `revcat metrics overview`

Headline metrics for the active project

```text
Returns the dashboard's headline numbers (active subscribers, MRR,
lifetime revenue, conversion). Shape is preserved verbatim; the table
view flattens top-level keys.
```

**Usage**

```sh
revcat metrics overview
```

---

## `revcat offerings`

Manage RevenueCat offerings

```text
An offering is a presentation grouping of packages displayed on a
paywall. Each project has 0..N offerings; exactly one is "current" and is
returned by SDKs that ask for the current offering.

To set an offering current along with a paywall config in one shot, use
`revcat publish offering`.
```

**Usage**

```sh
revcat offerings
```

**Aliases**: `offer`

**Subcommands**: `archive`, `create`, `delete`, `list`, `preview`, `set-current`, `unarchive`, `update`, `view`

---

## `revcat offerings archive`

Archive an offering

**Usage**

```sh
revcat offerings archive <id>
```

---

## `revcat offerings create`

Create an offering

**Usage**

```sh
revcat offerings create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--display-name` | string | - | Display name |
| `-f, --file` | string | - | Body as JSON file |
| `--id` | string | - | Offering lookup_key |

---

## `revcat offerings delete`

Delete an offering

**Usage**

```sh
revcat offerings delete <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat offerings list`

List all offerings in the active project

**Usage**

```sh
revcat offerings list
```

---

## `revcat offerings preview`

Show what the SDK will receive from /v1/subscribers/{user}/offerings

```text
Hit the v1 SDK-facing endpoint that `Purchases.getOfferings()` calls and
render the response. Useful when the dashboard looks healthy but the SDK
reports zero packages.

Auth is auto-handled: revcat fetches the public SDK key for the project's
app on the chosen platform and sends it as the bearer. `--as` defaults to a
synthetic user id (`revcat_preview_<unix_ms>`) so you don't have to think
about it.

Pass an optional <offering_id> to filter the rendered output to a single
offering (the request still fetches all offerings; v1 has no per-offering
endpoint).
```

**Usage**

```sh
revcat offerings preview [<offering_id>] [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--app-id` | string | - | App id to derive the public SDK key from (default: auto-detect) |
| `--as` | string | - | User id to query as (default: revcat_preview_<unix_ms>) |
| `--platform` | string | ios | Storefront platform to preview (ios\|android\|web) |

---

## `revcat offerings set-current`

Promote an offering to current

```text
Promote an offering so SDKs returning "current offering" hand back this
one. Equivalent to `revcat publish offering <id> --current` minus the
plan/confirm flow.
```

**Usage**

```sh
revcat offerings set-current <id>
```

---

## `revcat offerings unarchive`

Unarchive an offering

**Usage**

```sh
revcat offerings unarchive <id>
```

---

## `revcat offerings update`

Update an offering

**Usage**

```sh
revcat offerings update <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--display-name` | string | - | New display name |
| `-f, --file` | string | - | Patch body as JSON |

---

## `revcat offerings view`

Show one offering by lookup_key

**Usage**

```sh
revcat offerings view <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--packages` | bool | true | Include packages in the rendered card (table mode). On --output json, pass --packages explicitly to opt into the stitched typed shape; default is the verbatim v2 response. |

---

## `revcat packages`

Manage RevenueCat packages (purchasables inside an offering)

```text
Packages are the purchasable units inside an offering. Identifiers
follow RC's $rc_monthly / $rc_annual / custom convention.
```

**Usage**

```sh
revcat packages
```

**Aliases**: `pkg`

**Subcommands**: `attach`, `create`, `delete`, `detach`, `list`, `products`, `update`, `view`

---

## `revcat packages attach`

Attach product(s) to a package

```text
Attach products to a package. RC v2 stores per-attachment eligibility
criteria - the same product can have different eligibility on different
packages. Default is "all" which serves the product to every customer.

Other supported values: "google_sdk_lt_6", "google_sdk_ge_6" (used to
gate attachment by SDK version on Android). Pass --eligibility to
override the default.
```

**Usage**

```sh
revcat packages attach <pkg_id> <product_id> [<product_id> ...] [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--eligibility` | string | all | Eligibility criteria for the attachment: all \| google_sdk_lt_6 \| google_sdk_ge_6 |

---

## `revcat packages create`

Create a package under an offering

```text
Create a package under an offering (passed as id or lookup_key).

Common case via flags:

    revcat packages create default --id '$rc_monthly' --display-name "Monthly"

Or pass an arbitrary v2 body via --file:

    {
      "lookup_key": "$rc_monthly",
      "display_name": "Monthly",
      "position": 0
    }
```

**Usage**

```sh
revcat packages create <offering> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--display-name` | string | - | Display name |
| `-f, --file` | string | - | Body as JSON file |
| `--id` | string | - | Package lookup_key (e.g., $rc_monthly) |
| `--position` | int | 0 | Position in the offering (0-indexed) |

---

## `revcat packages delete`

Delete a package

**Usage**

```sh
revcat packages delete <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat packages detach`

Detach product(s) from a package

**Usage**

```sh
revcat packages detach <id> <product_id> [<product_id> ...]
```

---

## `revcat packages list`

List packages across one offering or all offerings

**Usage**

```sh
revcat packages list [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-o, --offering` | string | - | Restrict to a single offering by lookup_key |

---

## `revcat packages products`

List products attached to a package

```text
List products attached to a package.

The v2 `GET /packages/{id}/products` endpoint returns each
attachment as a binding object: `{eligibility_criteria, product:{...}}`.
The full nested product (display_name, app_id, store_identifier, etc.)
arrives in the same response. `--output json` returns the raw v2
items verbatim; the table view flattens them.
```

**Usage**

```sh
revcat packages products <id>
```

---

## `revcat packages update`

Update a package

**Usage**

```sh
revcat packages update <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-f, --file` | string | - | Patch body as JSON |

---

## `revcat packages view`

Show one package by internal id

**Usage**

```sh
revcat packages view <id>
```

---

## `revcat paywalls`

Manage top-level paywall resources

```text
Manage paywall records in the project's paywall library. To deploy a
paywall config to an offering use `revcat publish offering --paywall <file>`.
```

**Usage**

```sh
revcat paywalls
```

**Subcommands**: `create`, `delete`, `list`, `view`

---

## `revcat paywalls create`

Create a paywall record scoped to an offering

```text
Create a paywall record. v2 only takes {offering_id} - the paywall
content (template, copy, components) is set later via
`revcat publish offering <id> --paywall <file>`.

You usually want publish offering instead of this command.
```

**Usage**

```sh
revcat paywalls create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-f, --file` | string | - | JSON body |
| `--offering` | string | - | Offering id (required if --file not used) |

---

## `revcat paywalls delete`

Delete a paywall

**Usage**

```sh
revcat paywalls delete <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat paywalls list`

List paywalls in the project

**Usage**

```sh
revcat paywalls list
```

---

## `revcat paywalls view`

Show one paywall (raw JSON)

**Usage**

```sh
revcat paywalls view <id>
```

---

## `revcat products`

Manage RevenueCat products (store SKUs)

```text
A product is a project-level catalog entry that mirrors a store SKU
(App Store / Play Store / Stripe / Web Billing). Products are attached
to packages, packages live inside offerings.

Most edits accept a JSON file via --file, since the product schema
differs per store and revcat does not pin a specific shape. Use the
v2 docs to author the body, then `revcat products create -f product.json`.
```

**Usage**

```sh
revcat products
```

**Aliases**: `prod`

**Subcommands**: `archive`, `create`, `delete`, `list`, `push-to-store`, `unarchive`, `update`, `view`

---

## `revcat products archive`

Archive a product

**Usage**

```sh
revcat products archive <id>
```

---

## `revcat products create`

Create a product from a JSON body

```text
Create a product from a JSON body on disk. The body shape follows the
v2 docs - store_identifier, type (subscription | non_subscription | ...),
display_name, app_id are usually required.

Example body:
    {
      "store_identifier": "app.monthly",
      "type": "subscription",
      "display_name": "Monthly",
      "app_id": "app_xxx"
    }
```

**Usage**

```sh
revcat products create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-f, --file` | string | - | Path to JSON body (required) |

---

## `revcat products delete`

Delete a product (most teams should archive instead)

**Usage**

```sh
revcat products delete <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat products list`

List products

**Usage**

```sh
revcat products list
```

---

## `revcat products push-to-store`

Push a product config to the linked store

**Usage**

```sh
revcat products push-to-store <id>
```

---

## `revcat products unarchive`

Unarchive a product

**Usage**

```sh
revcat products unarchive <id>
```

---

## `revcat products update`

Update a product

```text
Update a product. Pass --file <path.json> for an arbitrary patch, or
--display-name <new> for the common case.
```

**Usage**

```sh
revcat products update <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--display-name` | string | - | New display name (shortcut for the common case) |
| `-f, --file` | string | - | Patch body as JSON |

---

## `revcat products view`

Show one product

**Usage**

```sh
revcat products view <id>
```

---

## `revcat projects`

Inspect RevenueCat projects

```text
A project is RevenueCat's top-level container - one per app or app
family. revcat resolves the active project from --project-id,
REVCAT_PROJECT_ID, the local .revcat/config.json written by
`revcat init`, or revcat.toml.

v2 exposes project create + list. There is no v2 update or delete by
id; manage those in the dashboard.
```

**Usage**

```sh
revcat projects
```

**Aliases**: `proj`

**Subcommands**: `create`, `list`, `view`

---

## `revcat projects create`

Create a new project

```text
Create a new project at the account level.

The project id is generated by RevenueCat and printed on success;
quote it (or pipe with --output json) to feed into `revcat init` or
`--project-id` for follow-up commands.

Examples:
    revcat projects create --name "My App"
    revcat projects create --name staging --output json | jq -r .id
```

**Usage**

```sh
revcat projects create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--name` | string | - | Project name (required) |

---

## `revcat projects list`

List projects accessible to this credential

**Usage**

```sh
revcat projects list
```

---

## `revcat projects view`

Show one project by id (defaults to the resolved project)

**Usage**

```sh
revcat projects view [id]
```

---

## `revcat publish`

One-shot deploy verbs (offering, paywall, ...)

```text
Publish-style verbs compose several API calls behind a single command.
The intent is to mirror the dashboard's higher-level actions ("set as
current", "deploy paywall") rather than mirror REST endpoints.
```

**Usage**

```sh
revcat publish
```

**Subcommands**: `offering`

---

## `revcat publish offering`

Set an offering as current and/or push a paywall config

```text
Composes the dashboard's "deploy" workflow:

  1. Verify the offering exists in the active project
  2. (optional) Validate and PUT a paywall config from --paywall <path>
  3. (optional) Promote the offering to current

By default both steps run when their inputs are provided. The plan is
printed before execution; pass --confirm to skip the prompt, or --dry-run
to print the plan without making any changes.

Examples:
    revcat publish offering default --current --confirm
    revcat publish offering pro --paywall ./paywalls/pro.json --current
    revcat publish offering pro --paywall ./paywalls/pro.json --dry-run
```

**Usage**

```sh
revcat publish offering <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the confirmation prompt |
| `--current` | bool | - | Set the offering as current |
| `--dry-run` | bool | - | Print the plan without making changes |
| `--no-current` | bool | - | Do NOT set the offering as current (overrides --current default when --paywall is set alone) |
| `--paywall` | string | - | Path to a paywall config JSON file to PUT |

---

## `revcat purchases`

Inspect non-renewing purchases

```text
A purchase is a one-shot non-renewing transaction (lifetime grants,
consumables, in-app one-offs). For subscriptions see `revcat subscriptions`.
```

**Usage**

```sh
revcat purchases
```

**Aliases**: `purchase`

**Subcommands**: `entitlements`, `refund`, `search`, `view`

---

## `revcat purchases entitlements`

List entitlements granted by a purchase

**Usage**

```sh
revcat purchases entitlements <id>
```

---

## `revcat purchases refund`

Refund a non-renewing purchase (Web Billing)

**Usage**

```sh
revcat purchases refund <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat purchases search`

Find purchases by store id

**Usage**

```sh
revcat purchases search <store_id>
```

---

## `revcat purchases view`

Show one purchase

**Usage**

```sh
revcat purchases view <id>
```

---

## `revcat subscribers`

Inspect and manage RevenueCat subscribers

```text
Subscribers (a.k.a. customers, app users) are the end-users of your app.
revcat treats them as the unit of debugging - one command surfaces their
entitlements, subscriptions, purchases, and aliases in a single card.
```

**Usage**

```sh
revcat subscribers
```

**Aliases**: `customers`, `subs`

**Subcommands**: `attributes`, `create`, `delete`, `grant`, `info`, `invoices`, `list`, `refund`, `revoke`, `transfer`

---

## `revcat subscribers attributes`

Get or set subscriber attributes

```text
With no flags, lists current attributes. With --set key=value (repeatable)
or --file <path.json>, upserts the listed attributes.

The v2 attributes endpoint takes an array of {name, value} objects.
Both the file and --set forms are normalized to that shape before
sending.
```

**Usage**

```sh
revcat subscribers attributes <user_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--file` | string | - | JSON map of attributes to upsert |
| `--set` | stringArray | - | key=value to upsert (repeatable) |

---

## `revcat subscribers create`

Pre-create a customer (migration / import)

```text
Pre-create a customer record. Most apps let the SDK create customers
on first launch; this is for migrations, test seeding, or cases where
you want to seed attributes before the user opens the app.

Pass --file <path.json> for arbitrary v2 fields (attributes, etc.).
```

**Usage**

```sh
revcat subscribers create <user_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-f, --file` | string | - | Optional JSON body merged into the request |

---

## `revcat subscribers delete`

Permanently delete a customer (GDPR / test cleanup)

**Usage**

```sh
revcat subscribers delete <user_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat subscribers grant`

Grant a promotional entitlement to a subscriber

```text
Grant a promotional entitlement (audited) to a subscriber for a
specified duration. v2 stores absolute expiry timestamps; revcat
translates --duration to the right "expires_at".

Duration accepts:
    forever | lifetime         (~100 years)
    7d, 30d, 90d               (days)
    1m, 3m, 6m                 (months ~ 30 days each)
    1y, 2y, 5y                 (years ~ 365 days each)

Examples:
    revcat subscribers grant app_user_123 premium --duration 7d
    revcat subscribers grant app_user_123 pro --duration lifetime --confirm
```

**Usage**

```sh
revcat subscribers grant <user_id> <entitlement> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the confirmation prompt |
| `-d, --duration` | string | - | How long the grant lasts (required) |

---

## `revcat subscribers info`

Show a full debug card for a subscriber

```text
Fan out across the v2 customer endpoints (customer + active_entitlements
+ subscriptions + purchases + aliases) and render a single card.

Pipe the output to a script and you get JSON instead of the card. Each
section appears as a top-level key in the JSON object.

Example:
    revcat subscribers info app_user_123
    revcat subscribers info app_user_123 --output json | jq .entitlements
```

**Usage**

```sh
revcat subscribers info <user_id>
```

---

## `revcat subscribers invoices`

List invoices for a customer

**Usage**

```sh
revcat subscribers invoices <user_id>
```

---

## `revcat subscribers list`

List customers in the active project (paged)

```text
Page through every customer. For support workflows that look up a
specific user by email or order id, prefer the searches under
`revcat purchases search` and `revcat subscriptions search` (faster
and indexed).
```

**Usage**

```sh
revcat subscribers list
```

---

## `revcat subscribers refund`

Refund a transaction on a subscription

```text
Issue a refund through the appropriate store (App Store, Play Store,
Stripe, ...). v2 scopes refunds under the subscription, so both ids
are required.

Find the subscription id with `revcat subscribers info <user_id>` -
each subscription row carries an internal id.

Refund availability depends on the store and the original purchase date.
RC returns the updated transaction status; we surface it in JSON if you
pipe the output.
```

**Usage**

```sh
revcat subscribers refund <subscription_id> <transaction_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the confirmation prompt |

---

## `revcat subscribers revoke`

Revoke a promotional entitlement from a subscriber

```text
Remove a promotional entitlement that was previously granted via
`revcat subscribers grant` or the dashboard. Does not affect
entitlements granted by an active store subscription.
```

**Usage**

```sh
revcat subscribers revoke <user_id> <entitlement> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the confirmation prompt |

---

## `revcat subscribers transfer`

Transfer entitlements/subscriptions from one customer to another

**Usage**

```sh
revcat subscribers transfer <src_user_id> <dst_user_id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat subscriptions`

Inspect and manage subscriptions

```text
A subscription is one customer's ongoing purchase relationship with one
product. Find it via `revcat subscriptions search <store_id>` or by
listing it under a customer (`revcat subscribers info`).
```

**Usage**

```sh
revcat subscriptions
```

**Aliases**: `sub`

**Subcommands**: `cancel`, `entitlements`, `management-url`, `refund`, `search`, `transactions`, `view`

---

## `revcat subscriptions cancel`

Cancel a subscription (Web Billing)

**Usage**

```sh
revcat subscriptions cancel <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat subscriptions entitlements`

List entitlements granted by a subscription

**Usage**

```sh
revcat subscriptions entitlements <id>
```

---

## `revcat subscriptions management-url`

Print the store-specific manage/cancel URL

**Usage**

```sh
revcat subscriptions management-url <id>
```

---

## `revcat subscriptions refund`

Refund the entire subscription (Web Billing)

**Usage**

```sh
revcat subscriptions refund <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat subscriptions search`

Find subscriptions by store id (App Store / Play / Stripe / ...)

**Usage**

```sh
revcat subscriptions search <store_id>
```

---

## `revcat subscriptions transactions`

List billing transactions for a subscription

**Usage**

```sh
revcat subscriptions transactions <id>
```

---

## `revcat subscriptions view`

Show one subscription

**Usage**

```sh
revcat subscriptions view <id>
```

---

## `revcat version`

Print revcat version and build info

**Usage**

```sh
revcat version
```

---

## `revcat virtual-currencies`

Manage virtual currencies (coins / credits)

```text
Project-level virtual currencies (in-game coins, credits, tokens).
v2 keys VCs by their uppercase code (e.g., COIN, GEM) - that's the
identifier you pass to view/update/delete/archive.

Per-customer balances and transactions are NOT exposed by v2 REST.
```

**Usage**

```sh
revcat virtual-currencies
```

**Aliases**: `vc`

**Subcommands**: `archive`, `create`, `delete`, `list`, `unarchive`, `update`, `view`

---

## `revcat virtual-currencies archive`

Archive a virtual currency

**Usage**

```sh
revcat virtual-currencies archive <code>
```

---

## `revcat virtual-currencies create`

Create a virtual currency

```text
Create a virtual currency. Required: name + code. Code is uppercase
and acts as the identifier for view/update/delete.
```

**Usage**

```sh
revcat virtual-currencies create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--code` | string | - | Uppercase code (e.g. COIN) |
| `--description` | string | - | Optional description |
| `-f, --file` | string | - | JSON body |
| `--name` | string | - | Display name |

---

## `revcat virtual-currencies delete`

Delete a virtual currency

**Usage**

```sh
revcat virtual-currencies delete <code> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat virtual-currencies list`

List virtual currencies

**Usage**

```sh
revcat virtual-currencies list
```

---

## `revcat virtual-currencies unarchive`

Unarchive a virtual currency

**Usage**

```sh
revcat virtual-currencies unarchive <code>
```

---

## `revcat virtual-currencies update`

Update a virtual currency

**Usage**

```sh
revcat virtual-currencies update <code> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--description` | string | - | New description |
| `-f, --file` | string | - | Patch body as JSON |
| `--name` | string | - | New name |

---

## `revcat virtual-currencies view`

Show one virtual currency by uppercase code (e.g. COIN)

**Usage**

```sh
revcat virtual-currencies view <code>
```

---

## `revcat webhooks`

Manage webhook integrations

```text
Webhooks are project integrations that receive event POSTs (purchases,
renewals, cancellations, refunds, ...). Each webhook has a name, target
URL, and a list of event_types it subscribes to.

Event values are LOWERCASE in the API config (initial_purchase,
renewal, cancellation, ...) - even though the webhook payload itself
uses screaming case (INITIAL_PURCHASE). revcat lowercases values
passed via --events for you.
```

**Usage**

```sh
revcat webhooks
```

**Subcommands**: `create`, `delete`, `list`, `update`, `view`

---

## `revcat webhooks create`

Create a webhook integration

```text
Create a webhook. Required: name, url, event_types. URL must be a
valid HTTPS endpoint that RC can validate as reachable - localhost and
example.com don't pass.
```

**Usage**

```sh
revcat webhooks create [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--events` | stringSlice | - | Event types (comma-separated). Values are lowercased before send. |
| `-f, --file` | string | - | JSON body |
| `--name` | string | - | Webhook name (required) |
| `--url` | string | - | Target URL (required) |

---

## `revcat webhooks delete`

Delete a webhook

**Usage**

```sh
revcat webhooks delete <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-y, --confirm` | bool | - | Skip the prompt |

---

## `revcat webhooks list`

List webhook integrations

**Usage**

```sh
revcat webhooks list
```

---

## `revcat webhooks update`

Update a webhook

**Usage**

```sh
revcat webhooks update <id> [flags]
```

**Flags**

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--events` | stringSlice | - | New event_types list |
| `-f, --file` | string | - | Patch body as JSON |
| `--name` | string | - | New name |
| `--url` | string | - | New target URL |

---

## `revcat webhooks view`

Show one webhook

**Usage**

```sh
revcat webhooks view <id>
```

---

