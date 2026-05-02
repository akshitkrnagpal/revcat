# revcat skills

Agent Skills for revcat. Each skill is a folder containing a `SKILL.md` with YAML frontmatter (`name` + `description` are required). Drop them into your agent's skills directory and the agent picks the right one based on the description.

## What's here

- `revcat-getting-started/` — first-time install + auth (browser OAuth) + `revcat init` for per-repo project context + command map. Triggers when a user wants to start using revcat.
- `revcat-commands/` — real syntax + examples for every subcommand including `init`. Triggers when constructing a command.
- `revcat-troubleshooting/` — common errors and fixes (401, no profile, no project_id, v0.3-to-v0.4 upgrade paths, toml/local mismatch, Test Store quirks, v1-only endpoints, dashboard-only operations). Triggers on a failed command or error message.
- `revcat-storefront-debug/` — 7-step diagnostic flow for "the SDK sees 0 packages from my offering." Covers the Test Store price gotcha, product/store binding checks, and the v1 `/subscribers/{id}/offerings` verify step. Triggers on "0 packages", "fetchOfferings empty", and similar SDK-side surprises.

## Install via skills.sh (recommended)

[skills.sh](https://skills.sh) is a registry for Agent Skills. The CLI auto-detects your agent (Claude Code, Cursor, Codex, etc.) and writes to the right config dir.

All four revcat skills:

```sh
npx skills add akshitkrnagpal/revcat
```

A single skill from the bundle:

```sh
npx skills add akshitkrnagpal/revcat/revcat-getting-started
npx skills add akshitkrnagpal/revcat/revcat-commands
npx skills add akshitkrnagpal/revcat/revcat-troubleshooting
npx skills add akshitkrnagpal/revcat/revcat-storefront-debug
```

Browse on the web: <https://skills.sh/akshitkrnagpal/revcat>.

Restart your agent (or open a new session) and the skills are live.

## Install manually

If you can't or don't want to use skills.sh, copy the folders directly.

### Claude Code

User-level (every project):

```sh
mkdir -p ~/.claude/skills
cp -R skills/revcat-* ~/.claude/skills/
```

Project-level (this repo only):

```sh
mkdir -p .claude/skills
cp -R skills/revcat-* .claude/skills/
```

### Cursor

Cursor reads project-scoped rules from `.cursor/rules/`:

```sh
mkdir -p .cursor/rules
for s in skills/revcat-*; do
  name=$(basename "$s")
  cp "$s/SKILL.md" ".cursor/rules/$name.md"
done
```

### Codex

Codex CLI reads from `~/.codex/skills/`:

```sh
mkdir -p ~/.codex/skills
cp -R skills/revcat-* ~/.codex/skills/
```

If your agent uses a different path, follow its local install docs — the SKILL.md format is portable.

## Authoring guidelines

- Description should be trigger-focused ("Use when...") and front-load key terms. Keep under ~200 chars where possible.
- Body should be concise and example-heavy. Reference real commands — if you're unsure, check `revcat <group> --help` or the [docs](https://revcat.vercel.app).
- Don't invent flags. If something is genuinely TBD, leave a `TODO:` note for the maintainer.
- After breaking changes (e.g., the v0.4 OAuth-only rewrite), do a sweep across all skills — the troubleshooting skill in particular accumulates legacy error strings users will still search for.
