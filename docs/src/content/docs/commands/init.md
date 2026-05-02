---
title: init
description: Bootstrap project context (revcat.toml + .revcat/config.json)
---

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

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## What it writes

| File | Committed? | Mode | Purpose |
| --- | --- | --- | --- |
| `revcat.toml` | yes | 0644 | `project_id` + optional `apps[]`. Documents which RC project this repo belongs to. |
| `.revcat/config.json` | no (gitignored) | 0600 | OAuth credential blob + `project_id` + `apps`. Walked up from cwd by every revcat command. |

`.revcat/` is auto-appended to `.gitignore` (idempotent). The committed half (`revcat.toml`) is non-secret; the gitignored half (`.revcat/config.json`) carries the refresh token and must not be committed.

## Synopsis

```sh
revcat init [flags]
```

Requires an authenticated global profile (run `revcat auth login` first), or `REVCAT_REFRESH_TOKEN` set in the environment.

## Flags

| Flag | Description |
| --- | --- |
| `--project-id <id>` | Skip the project picker and use this id (also reads from `REVCAT_PROJECT_ID`) |
| `--app-id <id>` | Record an app id (repeatable). Skips the app picker. |
| `--no-apps` | Skip the apps section entirely (still writes `project_id`) |
| `--no-local-creds` | Write only `revcat.toml`; skip `.revcat/config.json` |
| `--force` | Overwrite an existing `revcat.toml` or `.revcat/config.json` |
| `--path <dir>` | Where to write files (default: cwd) |

## Examples

Interactive — pick project + apps from a list:

```sh
cd ~/your-repo
revcat init
```

Scripted — for CI / agents / re-init in another shell:

```sh
revcat init --project-id projaac376d8 --no-apps --force
```

Re-init after switching active global profile (the local config is rewritten with the new credential):

```sh
revcat auth use work
revcat init --force
```

Just record project context without copying credentials (e.g., when committing a fresh repo and you want the next dev to log in themselves):

```sh
revcat init --project-id projaac376d8 --no-local-creds
```

## After init

Inside the directory, project-scoped commands work without flags:

```sh
revcat offerings list
revcat entitlements list
revcat publish offering pro --paywall ./paywalls/pro.json
```

`cd` out of the directory and the binding goes away — commands fall back to the global keychain, and project-scoped commands need `--project-id`.

## Verifying

```sh
revcat auth status --validate
```

`source` should be `local` and `source_path` should point at your `.revcat/config.json`. `project_source` should be that same path.

## Mismatch detector

If `revcat.toml.project_id` and `.revcat/config.json.project_id` disagree (e.g., someone hand-edited the toml), `revcat auth doctor` flags it:

```text
FAIL  toml/local mismatch  revcat.toml says proj_x, .revcat/config.json says proj_y
      hint: rerun `revcat init --force` to realign, or edit revcat.toml to match
```

## Related

- [`auth login`](/commands/auth/) — required before init
- [`auth status`](/commands/auth/) — show resolved credential + project context
- [`auth doctor`](/commands/auth/) — diagnose auth + project binding issues
- [Configuration](/reference/configuration/) — full env var list and resolution order
