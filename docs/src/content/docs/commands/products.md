---
title: products
description: Manage RevenueCat products (store SKUs)
---

A product is a project-level catalog entry that mirrors a store SKU
(App Store / Play Store / Stripe / Web Billing). Products are attached
to packages, packages live inside offerings.

Most edits accept a JSON file via --file, since the product schema
differs per store and revcat does not pin a specific shape. Use the
v2 docs to author the body, then `revcat products create -f product.json`.

## Subcommands

| Command | Description |
| --- | --- |
| `products archive <id>` | Archive a product |
| `products create` | Create a product from a JSON body |
| `products delete <id>` | Delete a product (most teams should archive instead) |
| `products list` | List products |
| `products push-to-store <id>` | Push a product config to the linked store |
| `products unarchive <id>` | Unarchive a product |
| `products update <id>` | Update a product |
| `products view <id>` | Show one product |

Aliases: `prod`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

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

For App Store / Play Store / Stripe / Web Billing apps:

```json
{
  "store_identifier": "app.monthly",
  "type": "subscription",
  "display_name": "Monthly",
  "app_id": "app_xxx",
  "subscription": { "duration": "P1M" }
}
```

Test Store apps require `title` instead of `display_name`:

```json
{
  "store_identifier": "com.test.monthly",
  "type": "subscription",
  "title": "Monthly",
  "app_id": "app_xxx",
  "subscription": { "duration": "P1M" }
}
```

Type values: `subscription`, `one_time`, `consumable`, `non_consumable`, `non_renewing_subscription`.
