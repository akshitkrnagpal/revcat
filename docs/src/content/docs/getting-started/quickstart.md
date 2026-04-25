---
title: Quickstart
description: Sign in and run your first command in 60 seconds.
---

## 1. Get a v2 secret key

In the RevenueCat dashboard, open your project and create a new **Secret API key** (v2). Copy it - it starts with `sk_`.

## 2. Sign in

```sh
revcat auth login --name my-app --secret-key sk_xxx
```

revcat verifies the key, lists the projects it can reach, and stores the credentials in your OS keychain. If your key has access to more than one project, it will prompt you to pick one.

For CI / Docker without a keychain, pass `--bypass-keychain` (or set `REVCAT_BYPASS_KEYCHAIN=1`) to write to `./.revcat/config.json` instead. A `.gitignore` is created on first write.

## 3. Verify

```sh
revcat auth doctor
revcat doctor
```

Both should report `OK` for every check. If something fails, the doctor commands tell you what to fix.

## 4. Run your first command

Inspect a customer:

```sh
revcat subscribers info app_user_123
```

Show your headline numbers:

```sh
revcat metrics overview
```

Promote an offering with a fresh paywall:

```sh
revcat publish offering pro --paywall ./paywalls/pro.json
```

## Multiple profiles

Have a dev / staging / prod project? Log in to each with a different `--name` and switch:

```sh
revcat auth login --name dev      --secret-key sk_dev_xxx
revcat auth login --name prod     --secret-key sk_prod_xxx
revcat auth use prod
revcat --profile dev subscribers info app_user_123   # one-shot override
```

## Where to go next

- [All commands →](/commands/)
- [Configuration →](/reference/configuration/)
