---
title: collaborators
description: Inspect project collaborators (members).
---

List the people with access to the active RevenueCat project.

Read-only: v2 doesn't expose invite / role-change / remove via REST. Manage membership in the [dashboard](https://app.revenuecat.com).

## Subcommands

| Command | Description |
| --- | --- |
| `collaborators list` | List collaborators on the active project |

## Examples

```sh
revcat collaborators list
revcat collaborators list --output json | jq '.[] | select(.role == "admin")'

# Spot pending invites
revcat collaborators list --output json | jq '.[] | select(.accepted_at == null)'
```

Aliases: `members`.

## Output

Table columns:

- `id` — collaborator id
- `name` — display name (`-` if the invite hasn't been accepted)
- `email`
- `role` — free-form per v2 (commonly `admin`, `developer`, `billing`, `viewer`)
- `accepted` — invite acceptance date, or `pending` if the user hasn't accepted yet
- `mfa` — whether the collaborator has multi-factor auth enabled
