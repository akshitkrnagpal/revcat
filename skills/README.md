# revcat skills

Agent Skills for revcat. Each skill is a folder containing a `SKILL.md` with YAML frontmatter (`name` + `description` are required). Drop them into your agent's skills directory and the agent picks the right one based on the description.

## What's here

- `revcat-getting-started/` — first-time install + auth (browser OAuth) + `revcat init` for per-repo project context + command map. Triggers when a user wants to start using revcat.
- `revcat-commands/` — real syntax + examples for every subcommand including `init`. Triggers when constructing a command.
- `revcat-troubleshooting/` — common errors and fixes (401, no profile, no project_id, v0.3-to-v0.4 upgrade paths, toml/local mismatch, Test Store quirks, v1-only endpoints, dashboard-only operations). Triggers on a failed command or error message.
- `revcat-storefront-debug/` — 7-step diagnostic flow for "the SDK sees 0 packages from my offering." Covers the Test Store price gotcha, product/store binding checks, and the v1 `/subscribers/{id}/offerings` verify step. Triggers on "0 packages", "fetchOfferings empty", and similar SDK-side surprises.

## Install for Claude Code

User-level (available in every project):

```sh
mkdir -p ~/.claude/skills
cp -R skills/revcat-* ~/.claude/skills/
```

Project-level (only when you open this repo / a specific project):

```sh
mkdir -p .claude/skills
cp -R skills/revcat-* .claude/skills/
```

Restart Claude Code or open a new session — the skills will appear automatically.

## Install for Cursor

Cursor reads project-scoped rules from `.cursor/rules/`. Wrap each skill into a Cursor rule:

```sh
mkdir -p .cursor/rules
for s in skills/revcat-*; do
  name=$(basename "$s")
  cp "$s/SKILL.md" ".cursor/rules/$name.md"
done
```

Cursor picks them up on next session.

## Install for Codex

Codex CLI reads from `~/.codex/skills/` (when configured). Drop the folders in:

```sh
mkdir -p ~/.codex/skills
cp -R skills/revcat-* ~/.codex/skills/
```

If your Codex setup uses a different path, follow your local install docs — the SKILL.md format is portable across agents.

## Authoring guidelines

- Description should be trigger-focused ("Use when...") and front-load key terms. Keep under ~200 chars where possible.
- Body should be concise and example-heavy. Reference real commands — if you're unsure, check `revcat <group> --help` or the [docs](https://revcat.vercel.app).
- Don't invent flags. If something is genuinely TBD, leave a `TODO:` note for the maintainer.
- After breaking changes (e.g., the v0.4 OAuth-only rewrite), do a sweep across all skills — the troubleshooting skill in particular accumulates legacy error strings users will still search for.
