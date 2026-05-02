---
title: Quickstart
description: Sign in and run your first command in 60 seconds.
---

## 1. Sign in

```sh
revcat auth login
```

revcat opens your browser for the OAuth flow, you authorize against RevenueCat, and the tokens land in `~/.revcat/config.json` (mode 0600).

## 2. Bind a project to your repo

```sh
cd ~/your-repo
revcat init
```

![revcat init](https://raw.githubusercontent.com/akshitkrnagpal/revcat/main/demo/init.gif)

`revcat init` lists projects you can access, prompts for one (and optionally apps), then writes:

- `revcat.toml` (committed): records `project_id` + apps so a `git clone` documents which RC project this repo belongs to.
- `.revcat/config.json` (gitignored, mode 0600): copies your credential into the directory so subsequent commands and any agent or sandbox in the directory inherit it without touching `~/.revcat/config.json`.

`.revcat/` is auto-appended to `.gitignore`.

## 3. Verify

```sh
revcat auth doctor
revcat doctor
```

Both should report `OK` for every check.

## 4. Run your first command

Inside the repo, project context is automatic:

```sh
revcat metrics overview
revcat subscribers info app_user_123
revcat publish offering pro --paywall ./paywalls/pro.json
```

## Multiple accounts

Juggling several RC accounts? Log in to each with a different `--name` and switch:

```sh
revcat auth login --name work
revcat auth login --name personal
revcat auth use personal                 # default for global commands
revcat --profile work auth status        # one-shot override
```

When you `revcat init` inside a repo, it copies whichever profile is active at that moment. If you need to switch credentials for an existing repo, rerun `revcat init --force`.

## Headless / CI

For fresh sandboxes with no browser:

```sh
export REVCAT_REFRESH_TOKEN=rtk_...
export REVCAT_PROJECT_ID=proj_...
revcat offerings list
```

revcat synthesizes a virtual profile, refreshes tokens in-memory, and skips the login flow entirely. Pull the refresh token from your CI secret manager.

## Where to go next

- [All commands →](/commands/)
- [Configuration →](/reference/configuration/)
