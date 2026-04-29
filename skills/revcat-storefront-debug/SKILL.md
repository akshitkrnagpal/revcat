---
name: revcat-storefront-debug
description: Use when the SDK sees 0 packages in an offering, when a Test Store / App Store / Play / Stripe product isn't reaching the client, or when a freshly-created offering isn't returning what the user expects. Triggers on "0 packages", "no packages", "SDK not seeing", "test store", "products not in offering", "offering not working", "current_offering_id but empty packages", "fetchOfferings empty", "Purchases.getOfferings returned nothing".
---

# revcat - storefront / offering delivery debugging

The SDK call (`Purchases.getOfferings()` on iOS/Android, `purchases.getOfferings()` on web) returns whatever the v1 endpoint `/v1/subscribers/{user_id}/offerings` returns. **revcat itself only wraps v2 endpoints**, so if you need to see the v1 view directly, you fall back to curl. The diagnostic flow below is the canonical "why isn't my offering working" walk.

## When to reach for this skill

The user says any of:

- "the SDK is seeing zero packages"
- "my offering returns empty"
- "I added products to my package but they're not showing up"
- "Test Store products aren't loading"
- "App Store products aren't reaching the SDK"
- "`fetchOfferings` returns the offering id but no packages"

If the user is just constructing a command, route to `revcat-commands` instead. This skill is for diagnosing why the SDK isn't seeing what the dashboard shows.

## Diagnostic flow (in order)

Run these top-down. Stop at the first level that's wrong — that's the broken layer.

### 1. Confirm the offering exists and is current

```sh
revcat offerings list
```

The current offering is marked with `*`. The SDK's `current_offering_id` only returns this one. If you see your offering but it isn't current, set it:

```sh
revcat offerings set-current <id>
```

### 2. Confirm packages are attached to the offering

```sh
revcat packages list --offering <offering_id>
```

If this is empty, the offering has no packages. Create them:

```sh
revcat packages create --file ./package.json
```

(Body shape: `{identifier, offering_id}`. Use `$rc_monthly`, `$rc_annual`, etc. for RC's standard identifiers.)

### 3. Confirm products are attached to each package

```sh
revcat packages products <pkg_id>
```

If this is empty, the package isn't bound to a product yet. Attach:

```sh
revcat packages attach <pkg_id> <product_id>
```

If the output shows products but `display_name` is empty, that's the v2 API's normal shape — the display name lives on the product, not the binding. Run `revcat products view <product_id>` to see it.

### 4. Confirm each product is bound to a Store app and is reachable

```sh
revcat products view <product_id>
```

Look at `app_id`. The product must be attached to an app entry that matches the runtime store (App Store, Play Store, Stripe, **Test Store**). One product per store + one app entry per store.

```sh
revcat apps list
```

Confirm the product's `app_id` matches the app for the runtime store you're testing against.

### 5. (App Store / Play only) Push the product to the live store

```sh
revcat products push-to-store <product_id>
```

This is the v2 endpoint that triggers RC's sync to the underlying store. Must be run after creating a product if you want the App Store / Play to see it.

### 6. (Test Store only) Set a price in the RC dashboard UI

**The v2 API does not expose Test Store prices.** This is a documented RC limitation as of late 2025/2026. If the offering targets the Test Store app, every product must have a price set via the dashboard UI:

> https://app.revenuecat.com/projects/{project_id}/apps/{test_store_app_id}/products → click each product → set price.

Without prices, `/v1/subscribers/{user_id}/offerings` returns the offering with `packages: []`, which is exactly the "SDK sees 0 packages" symptom. revcat cannot fix this from the CLI.

### 7. Verify what the SDK will actually receive (v1 offerings preview)

```sh
# Default platform is iOS, default user id is a synthetic revcat_preview_<unix_ms>.
revcat offerings preview

# Other platforms / a specific user / one offering:
revcat offerings preview --platform android
revcat offerings preview --as cust_abc123
revcat offerings preview default --platform ios

# Raw v1 JSON shape:
revcat offerings preview --output json
```

`revcat offerings preview` calls the v1 SDK-facing endpoint
(`/v1/subscribers/{user_id}/offerings`) using the public SDK key for the
project's app on the chosen platform. This is the same payload the
on-device SDK sees from `Purchases.getOfferings()`.

A healthy response looks like:

```json
{
  "current_offering_id": "default",
  "offerings": [
    {
      "identifier": "default",
      "packages": [
        { "identifier": "$rc_monthly", "platform_product_identifier": "..." }
      ]
    }
  ]
}
```

If `packages` is empty here, the SDK will also see empty. The cause is upstream — re-check steps 1-6.

## Common root causes (with the layer they show up at)

| Symptom in v1 offerings response | Most likely cause | Layer |
|---|---|---|
| Offering missing entirely | Not marked current | Step 1 |
| Offering exists, `packages: []` | No packages attached, or none attached to a product, or Test Store product has no price | Steps 2 / 3 / 6 |
| Packages exist but `platform_product_identifier` is null | Product not pushed to store, or wrong `X-Platform` header | Steps 4 / 5 |
| Different products on iOS vs Android | One store missing the product binding | Step 4 |

## Where revcat does NOT cover the flow

revcat primarily wraps the v2 RC API (with one v1 read in `offerings preview`). It does not cover:

- **Test Store price setting** (dashboard UI only, no v2 endpoint exists)
- **Sync status to App Store Connect / Play Console** beyond `push-to-store`

If the user reaches for curl on any v2 endpoint, that's a revcat bug worth filing. If they reach for curl on a dashboard-only operation, that's expected and this skill should call it out.
