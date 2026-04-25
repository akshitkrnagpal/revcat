# RevenueCat v2 API coverage

Audit of revcat's coverage of the RevenueCat v2 REST API. Source of truth: <https://www.revenuecat.com/docs/api-v2>.

## Headline

- **~13 of ~95 documented v2 operations have CLI surface (~14%).**
- Coverage skews to high-value workflows (subscriber debugging, event tail, publish offering) rather than full CRUD.
- Two write-paths use v1-style URLs and need verification against the live v2 API. See "implementation issues" below.

## Implementation issues to fix

The CLI ships paths that may be v1-flavored (working in some accounts, broken in others). Verify against the live API and update before public launch.

| File | Method | Current path | v2 docs spelling |
| --- | --- | --- | --- |
| `internal/api/customers.go` | `GrantPromotionalEntitlement` | `POST /customers/{id}/active_entitlements/{ent}/grant_promotional` | `POST /customers/{id}/actions/grant_entitlement` |
| `internal/api/customers.go` | `RevokePromotionalEntitlement` | `POST /customers/{id}/active_entitlements/{ent}/revoke_promotional` | `POST /customers/{id}/actions/revoke_entitlement` |
| `internal/api/customers.go` | `RefundTransaction` | `POST /customers/{id}/transactions/{tid}:refund` | subscription-scoped: `POST /subscriptions/{sid}/transactions/{tid}/actions/refund` |
| `internal/api/offerings.go` | `SetCurrentOffering` | `POST /offerings/{id}/_set_current` | not in v2 docs - likely needs to be modeled as `update offering` with a flag |

## Coverage matrix

| Resource | API operation | API path | revcat command | Notes |
|---|---|---|---|---|
| Project | list | `GET /projects` | `revcat auth login` (picker), `revcat auth status`, `revcat auth doctor`, `revcat doctor` | covered (no top-level `revcat projects list`) |
| Project | get | `GET /projects/{project_id}` | - | gap |
| Project | create | `POST /projects` | - | partner-tier only |
| App | list | `GET /projects/{project_id}/apps` | - | gap |
| App | get | `GET /apps/{app_id}` | - | gap |
| App | create | `POST /apps` | - | partner-tier only |
| App | update | `POST /apps/{app_id}` | - | partner-tier only |
| App | delete | `DELETE /apps/{app_id}` | - | partner-tier only |
| App | list public API keys | `GET /apps/{app_id}/public_api_keys` | - | gap |
| App | get StoreKit config | `GET /apps/{app_id}/store_kit_config` | - | gap |
| Audit Log | list | `GET /projects/{id}/audit_logs` | - | partner-tier only |
| Charts & Metrics | overview | `GET /projects/{id}/metrics/overview` | - | gap (high TTY value) |
| Charts & Metrics | get chart | `GET /charts/{chart_name}` | - | gap |
| Charts & Metrics | chart options | `GET /charts/{chart_name}/options` | - | gap |
| Collaborator | list | `GET /collaborators` | - | partner-tier only |
| Customer | list | `GET /customers` | - | gap |
| Customer | get | `GET /customers/{id}` | `revcat subscribers info` | covered |
| Customer | create | `POST /customers` | - | gap |
| Customer | delete | `DELETE /customers/{id}` | - | gap (GDPR/test cleanup) |
| Customer | transfer | `POST .../actions/transfer` | - | gap |
| Customer | grant entitlement | `POST .../actions/grant_entitlement` | `revcat subscribers grant` | covered (path spelling fix needed) |
| Customer | revoke entitlement | `POST .../actions/revoke_entitlement` | `revcat subscribers revoke` | covered (path spelling fix needed) |
| Customer | override offering | `POST .../actions/override_offering` | - | gap |
| Customer | restore Google Play purchase | `POST .../actions/restore_google_play_purchase` | - | gap |
| Customer | list subscriptions | `GET /customers/{id}/subscriptions` | `revcat subscribers info` | covered |
| Customer | list purchases | `GET /customers/{id}/purchases` | `revcat subscribers info` | covered |
| Customer | list active entitlements | `GET /customers/{id}/active_entitlements` | `revcat subscribers info` | covered |
| Customer | list aliases | `GET /customers/{id}/aliases` | `revcat subscribers info` | covered |
| Customer | list invoices | `GET /customers/{id}/invoices` | - | gap |
| Customer | list virtual currency balances | `GET /customers/{id}/virtual_currencies_balances` | - | gap |
| Customer | create VC transaction | `POST /customers/{id}/virtual_currencies/transactions` | - | gap |
| Customer | update VC balance | `POST /customers/{id}/virtual_currencies_balances` | - | gap |
| Customer | list attributes | `GET /customers/{id}/attributes` | - | gap |
| Customer | set attributes | `POST /customers/{id}/attributes` | - | gap |
| Entitlement | list | `GET /entitlements` | `revcat entitlements list` | covered |
| Entitlement | get | `GET /entitlements/{id}` | `revcat entitlements view` | covered |
| Entitlement | create | `POST /entitlements` | - | gap |
| Entitlement | update | `POST /entitlements/{id}` | - | gap |
| Entitlement | delete | `DELETE /entitlements/{id}` | - | gap |
| Entitlement | list attached products | `GET /entitlements/{id}/products` | - | gap |
| Entitlement | archive / unarchive | `POST /entitlements/{id}/actions/archive\|unarchive` | - | gap |
| Entitlement | attach / detach products | `POST /entitlements/{id}/actions/attach_products\|detach_products` | - | gap (high value) |
| Offering | list | `GET /offerings` | `revcat offerings list` | covered |
| Offering | get | `GET /offerings/{id}` | `revcat offerings view` | covered |
| Offering | create | `POST /offerings` | - | gap |
| Offering | update | `POST /offerings/{id}` | - | gap |
| Offering | delete | `DELETE /offerings/{id}` | - | gap |
| Offering | archive / unarchive | `POST /offerings/{id}/actions/archive\|unarchive` | - | gap |
| Offering | set current | (no v2 docs endpoint) | `revcat publish offering` | covered with v1-style path; needs verification |
| Package | list | `GET /offerings/{id}/packages` | `revcat packages list` | covered |
| Package | get | `GET /packages/{id}` | - | gap |
| Package | create | `POST /packages` | - | gap |
| Package | update | `POST /packages/{id}` | - | gap |
| Package | delete | `DELETE /packages/{id}` | - | gap |
| Package | list attached products | `GET /packages/{id}/products` | - | gap |
| Package | attach / detach products | `POST /packages/{id}/actions/attach_products\|detach_products` | - | gap |
| Product | list | `GET /products` | - | gap (no `revcat products` group at all) |
| Product | get | `GET /products/{id}` | - | gap |
| Product | create | `POST /products` | - | gap |
| Product | update | `POST /products/{id}` | - | gap |
| Product | delete | `DELETE /products/{id}` | - | gap |
| Product | archive / unarchive | `POST /products/{id}/actions/archive\|unarchive` | - | gap |
| Product | push to store | `POST /products/{id}/actions/push_to_store` | - | gap |
| Virtual Currency | list | `GET /virtual_currencies` | - | gap |
| Virtual Currency | get | `GET /virtual_currencies/{id}` | - | gap |
| Virtual Currency | create | `POST /virtual_currencies` | - | gap |
| Virtual Currency | update | `POST /virtual_currencies/{id}` | - | gap |
| Virtual Currency | delete | `DELETE /virtual_currencies/{id}` | - | gap |
| Virtual Currency | archive / unarchive | `POST /virtual_currencies/{id}/actions/archive\|unarchive` | - | gap |
| Purchase | get | `GET /purchases/{id}` | - | gap |
| Purchase | list entitlements | `GET /purchases/{id}/entitlements` | - | gap |
| Purchase | refund (Web Billing) | `POST /purchases/{id}/actions/refund` | - | gap |
| Purchase | search by store id | `GET /purchases/search` | - | gap (high support value) |
| Subscription | get | `GET /subscriptions/{id}` | - | gap |
| Subscription | list transactions | `GET /subscriptions/{id}/transactions` | - | gap |
| Subscription | refund transaction | `POST /subscriptions/{sid}/transactions/{tid}/actions/refund` | `revcat subscribers refund` | covered (path spelling fix needed) |
| Subscription | list entitlements | `GET /subscriptions/{id}/entitlements` | - | gap |
| Subscription | cancel (Web Billing) | `POST /subscriptions/{id}/actions/cancel` | - | gap |
| Subscription | refund subscription | `POST /subscriptions/{id}/actions/refund` | - | gap |
| Subscription | management URL | `GET /subscriptions/{id}/management_url` | - | gap |
| Subscription | search by store id | `GET /subscriptions/search` | - | gap (high support value) |
| Invoice | get | `GET /invoices/{id}` | - | gap |
| Paywall | list | `GET /paywalls` | - | gap |
| Paywall | get | `GET /paywalls/{id}` | (partial) `revcat publish offering` reads `/offerings/{id}/paywall` | covered for offering-scoped flavor only |
| Paywall | create | `POST /paywalls` | - | gap |
| Paywall | delete | `DELETE /paywalls/{id}` | - | gap |
| Paywall | put offering paywall | `PUT /offerings/{id}/paywall` | `revcat publish offering --paywall <file>` | covered |
| Webhook | list | `GET /integrations/webhooks` | - | gap |
| Webhook | get | `GET /integrations/webhooks/{id}` | - | gap |
| Webhook | create | `POST /integrations/webhooks` | - | gap |
| Webhook | update | `POST /integrations/webhooks/{id}` | - | gap |
| Webhook | delete | `DELETE /integrations/webhooks/{id}` | - | gap |
| Event | list | `GET /events` | `revcat events list`, `revcat events tail` | covered |

## High-value gaps (ranked)

1. **`revcat subscribers list / search`** - `GET /customers`, `GET /subscriptions/search`, `GET /purchases/search`. The single most-asked support flow ("find user by email or order id") is unbuildable today.
2. **`revcat products` group** - the entire `/products` resource is missing. Catalog-as-code is the natural revcat moat and you can't do it without products.
3. **`revcat offerings create|update|delete` + package CRUD + entitlement attach/detach** - turns revcat into a catalog deployer instead of a read-only viewer. Pairs with #2.
4. **`revcat metrics overview` / `revcat charts`** - one-shot revenue snapshot in the terminal. Cheap to build, very high "wow" factor for a demo.
5. **`revcat webhooks` group** - list/create/update/delete `/integrations/webhooks`. Easy CRUD, very high pain today (everyone clicks the dashboard for this).
6. **Customer attributes get/set** - `GET|POST /customers/{id}/attributes`. Heavy support-rep use case, trivial to add.
7. **`revcat subscribers transfer`** - `POST .../actions/transfer`. Common manual-fix flow that today requires the dashboard.
8. **Subscription management** - `cancel`, `refund` at subscription level, `management_url`, `list transactions`. Today `revcat subscribers refund` only handles a single transaction by id; the realistic flow is "find the active sub, refund the latest charge", which needs these.

## Out of scope

- **Project create, App CRUD, Audit Logs, Collaborators** - all behind project-management / partner-tier keys per RC's permissions matrix. revcat targets the per-project secret key, so these should be explicitly punted in the README rather than implemented.
- **Virtual Currency CRUD** - niche, mostly mobile-game studios. Worth a stub later, not a priority.
- **Web Billing-specific verbs** (`purchases/{id}/actions/refund`, `subscriptions/{id}/actions/cancel`) - only meaningful for RC's Web Billing customers. Add when user-base overlaps.
- **Paywalls top-level resource** (`/paywalls` list/get/create/delete) - the offering-scoped `PUT /offerings/{id}/paywall` already covers the publish-as-code use case. Standalone resource is for dashboard's paywall library, lower priority.
