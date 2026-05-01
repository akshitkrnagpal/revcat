---
title: API coverage
description: Map every RevenueCat v2 operation to the revcat command that wraps it.
---


revcat aims for full v2 coverage. This file maps every documented operation to the revcat command that wraps it.

Source of truth: <https://www.revenuecat.com/docs/api-v2>.

## Headline

- Most v2 operations reachable with revcat's OAuth scope set are wrapped and smoke-tested.
- A few endpoints don't exist on the v2 customer surface at all - documented at the bottom.
- A small slice (project create, app CRUD, collaborators) isn't exposed by v2 REST.
- RC v2 has no REST events firehose; lifecycle events are webhook-delivered. revcat exposes webhook CRUD; subscribe your own endpoint with `revcat webhooks create`.

## Coverage

### Project management

| API operation | revcat |
| --- | --- |
| `GET /projects` | `revcat projects list`, `revcat auth login` (picker) |
| `GET /projects/{id}` | `revcat projects view` |
| `POST /projects` | partner-tier only |

### Apps

| API operation | revcat |
| --- | --- |
| `GET /apps` | `revcat apps list` |
| `GET /apps/{id}` | `revcat apps view` |
| `GET /apps/{id}/public_api_keys` | `revcat apps public-keys` |
| `GET /apps/{id}/store_kit_config` | `revcat apps storekit-config` |
| `POST /apps`, `POST /apps/{id}`, `DELETE /apps/{id}` | partner-tier only |

### Customers

| API operation | revcat |
| --- | --- |
| `GET /customers` | `revcat subscribers list` |
| `GET /customers/{id}` | `revcat subscribers info` |
| `POST /customers` | `revcat subscribers create` |
| `DELETE /customers/{id}` | `revcat subscribers delete` |
| `POST /customers/{id}/actions/transfer` | `revcat subscribers transfer` |
| `POST /customers/{id}/actions/grant_entitlement` | `revcat subscribers grant` |
| `POST /customers/{id}/actions/revoke_entitlement` | `revcat subscribers revoke` |
| `POST /customers/{id}/actions/override_offering` | - (endpoint 404 on v2; removed in v0.1.2) |
| `POST /customers/{id}/actions/restore_google_play_purchase` | - (endpoint 404 on v2; removed in v0.1.2) |
| `GET /customers/{id}/subscriptions` | `revcat subscribers info` |
| `GET /customers/{id}/purchases` | `revcat subscribers info` |
| `GET /customers/{id}/active_entitlements` | `revcat subscribers info` |
| `GET /customers/{id}/aliases` | `revcat subscribers info` |
| `GET /customers/{id}/invoices` | `revcat subscribers invoices` |
| `GET /customers/{id}/attributes` | `revcat subscribers attributes` |
| `POST /customers/{id}/attributes` | `revcat subscribers attributes --set` |
| `GET /customers/{id}/virtual_currencies_balances` | - (endpoint 404 on v2; removed in v0.1.2) |
| `POST /customers/{id}/virtual_currencies/transactions` | - (endpoint 404 on v2; removed in v0.1.2) |
| `POST /customers/{id}/virtual_currencies_balances` | - (endpoint 404 on v2; removed in v0.1.2) |

### Entitlements

| API operation | revcat |
| --- | --- |
| `GET /entitlements` | `revcat entitlements list` |
| `GET /entitlements/{id}` | `revcat entitlements view` |
| `POST /entitlements` | `revcat entitlements create` |
| `POST /entitlements/{id}` | `revcat entitlements update` |
| `DELETE /entitlements/{id}` | `revcat entitlements delete` |
| `GET /entitlements/{id}/products` | `revcat entitlements products` |
| `POST /entitlements/{id}/actions/archive` | `revcat entitlements archive` |
| `POST /entitlements/{id}/actions/unarchive` | `revcat entitlements unarchive` |
| `POST /entitlements/{id}/actions/attach_products` | `revcat entitlements attach` |
| `POST /entitlements/{id}/actions/detach_products` | `revcat entitlements detach` |

### Offerings

| API operation | revcat |
| --- | --- |
| `GET /offerings` | `revcat offerings list` |
| `GET /offerings/{id}` | `revcat offerings view` |
| `POST /offerings` | `revcat offerings create` |
| `POST /offerings/{id}` | `revcat offerings update`, `revcat offerings set-current`, `revcat publish offering --current` |
| `DELETE /offerings/{id}` | `revcat offerings delete` |
| `POST /offerings/{id}/actions/archive` | `revcat offerings archive` |
| `POST /offerings/{id}/actions/unarchive` | `revcat offerings unarchive` |
| `PUT /offerings/{id}/paywall` | `revcat publish offering --paywall <file>` |

### Packages

| API operation | revcat |
| --- | --- |
| `GET /offerings/{id}/packages` | `revcat packages list --offering <id>` |
| `GET /packages/{id}` | `revcat packages view` |
| `POST /packages` | `revcat packages create` |
| `POST /packages/{id}` | `revcat packages update` |
| `DELETE /packages/{id}` | `revcat packages delete` |
| `GET /packages/{id}/products` | `revcat packages products` |
| `POST /packages/{id}/actions/attach_products` | `revcat packages attach` |
| `POST /packages/{id}/actions/detach_products` | `revcat packages detach` |

### Products

| API operation | revcat |
| --- | --- |
| `GET /products` | `revcat products list` |
| `GET /products/{id}` | `revcat products view` |
| `POST /products` | `revcat products create` |
| `POST /products/{id}` | `revcat products update` |
| `DELETE /products/{id}` | `revcat products delete` |
| `POST /products/{id}/actions/archive` | `revcat products archive` |
| `POST /products/{id}/actions/unarchive` | `revcat products unarchive` |
| `POST /products/{id}/actions/push_to_store` | `revcat products push-to-store` |

### Paywalls

| API operation | revcat |
| --- | --- |
| `GET /paywalls` | `revcat paywalls list` |
| `GET /paywalls/{id}` | `revcat paywalls view` |
| `POST /paywalls` | `revcat paywalls create` |
| `DELETE /paywalls/{id}` | `revcat paywalls delete` |

### Subscriptions

| API operation | revcat |
| --- | --- |
| `GET /subscriptions/{id}` | `revcat subscriptions view` |
| `GET /subscriptions/{id}/transactions` | `revcat subscriptions transactions` |
| `GET /subscriptions/{id}/entitlements` | `revcat subscriptions entitlements` |
| `POST /subscriptions/{id}/actions/cancel` | `revcat subscriptions cancel` |
| `POST /subscriptions/{id}/actions/refund` | `revcat subscriptions refund` |
| `POST /subscriptions/{sid}/transactions/{tid}/actions/refund` | `revcat subscribers refund` |
| `GET /subscriptions/{id}/management_url` | `revcat subscriptions management-url` |
| `GET /subscriptions/search` | `revcat subscriptions search` |

### Purchases

| API operation | revcat |
| --- | --- |
| `GET /purchases/{id}` | `revcat purchases view` |
| `GET /purchases/{id}/entitlements` | `revcat purchases entitlements` |
| `POST /purchases/{id}/actions/refund` | `revcat purchases refund` |
| `GET /purchases/search` | `revcat purchases search` |

### Invoices

| API operation | revcat |
| --- | --- |
| `GET /invoices/{id}` | `revcat invoices view` |

### Audit log

| API operation | revcat |
| --- | --- |
| `GET /audit_logs` | `revcat audit-logs list` |

### Webhooks

| API operation | revcat |
| --- | --- |
| `GET /integrations/webhooks` | `revcat webhooks list` |
| `GET /integrations/webhooks/{id}` | `revcat webhooks view` |
| `POST /integrations/webhooks` | `revcat webhooks create` |
| `POST /integrations/webhooks/{id}` | `revcat webhooks update` |
| `DELETE /integrations/webhooks/{id}` | `revcat webhooks delete` |

### Virtual currencies

| API operation | revcat |
| --- | --- |
| `GET /virtual_currencies` | `revcat virtual-currencies list` |
| `GET /virtual_currencies/{code}` | `revcat virtual-currencies view <CODE>` |
| `POST /virtual_currencies` | `revcat virtual-currencies create` |
| `POST /virtual_currencies/{code}` | `revcat virtual-currencies update <CODE>` |
| `DELETE /virtual_currencies/{code}` | `revcat virtual-currencies delete <CODE>` |
| `POST /virtual_currencies/{code}/actions/archive` | `revcat virtual-currencies archive <CODE>` |
| `POST /virtual_currencies/{code}/actions/unarchive` | `revcat virtual-currencies unarchive <CODE>` |

### Charts & metrics

| API operation | revcat |
| --- | --- |
| `GET /metrics/overview` | `revcat metrics overview` |
| `GET /charts/{name}` | `revcat charts get <name>` |
| `GET /charts/{name}/options` | `revcat charts options <name>` |

## Out of scope (partner-tier keys only)

- `POST /projects` - project create
- App CRUD (`POST /apps`, `POST /apps/{id}`, `DELETE /apps/{id}`)
- `GET /collaborators`

## Not exposed by v2 REST

These endpoints either return 404 across the v2 customer surface or are gated behind a higher-tier key. Smoke-tested 2026-04-25:

- An events firehose. RC delivers lifecycle events (purchases, renewals, cancellations, refunds, ...) via webhooks. revcat covers webhook CRUD; subscribe your endpoint with `revcat webhooks create`.
- `POST /customers/{id}/actions/override_offering`
- `POST /customers/{id}/actions/restore_google_play_purchase`
- `GET /customers/{id}/virtual_currencies_balances`
- `POST /customers/{id}/virtual_currencies/transactions`
- `POST /customers/{id}/virtual_currencies_balances`
- `POST /customers/{id}/actions/revoke_entitlement` (no-op in revcat: implemented as "grant with near-future expires_at")

These aren't exposed by the v2 REST API. They live in the dashboard.
