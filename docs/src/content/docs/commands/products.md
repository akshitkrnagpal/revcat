---
title: products
description: Manage RevenueCat products (store SKUs).
---

A product is a project-level catalog entry that mirrors a store SKU (App Store / Play Store / Stripe / Web Billing). Products are attached to packages, packages live inside offerings.

Most edits accept a JSON file via `--file`, since the product schema differs per store and revcat does not pin a specific shape. Use the [v2 docs](https://www.revenuecat.com/docs/api-v2) to author the body, then `revcat products create -f product.json`.

## Subcommands

| Command | Description |
| --- | --- |
| `products list` | List products |
| `products view <id>` | Show one product |
| `products create` | Create a product from a JSON body (`--file`, required) |
| `products update <id>` | Update a product (`--file`, or `--display-name`) |
| `products delete <id>` | Delete (most teams should archive instead) |
| `products archive <id>` | Archive |
| `products unarchive <id>` | Unarchive |
| `products push-to-store <id>` | Push a product config to the linked store |

## Examples

```sh
revcat products list
revcat products create --file ./products/monthly.json
revcat products update prod_xxx --display-name "Monthly"
revcat products archive prod_xxx -y
revcat products push-to-store prod_xxx
```

Aliases: `prod`.

## Body shape (create)

```json
{
  "store_identifier": "app.monthly",
  "type": "subscription",
  "display_name": "Monthly",
  "app_id": "app_xxx"
}
```
