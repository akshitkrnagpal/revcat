# revcat skills

Agent Skills for revcat, distributable via [skills.sh](https://skills.sh) (planned). Each skill is a folder containing a `SKILL.md` with YAML frontmatter (`name` + `description` are required).

## What's here

- `revcat-getting-started/` - first-time install + auth + command map. Triggers when a user wants to start using revcat.
- `revcat-commands/` - real syntax + examples for every subcommand. Triggers when constructing a command.
- `revcat-troubleshooting/` - common errors and fixes. Triggers on a failed command or error message.

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

Restart Claude Code or open a new session - the skills will appear automatically.

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

If your Codex setup uses a different path, follow your local install docs - the SKILL.md format is portable across agents.

## Publishing to skills.sh

These skills will be published to <https://skills.sh> once revcat ships v0.1. Until then, install locally per the above. The `name` and `description` frontmatter is the trigger surface, so keep edits to the description trigger-focused.

## Authoring guidelines

- Description should be trigger-focused ("Use when...") and front-load key terms. Keep under ~200 chars.
- Body should be concise and example-heavy. Reference real commands - if you're unsure, check `revcat <group> --help` or the [docs](https://revcat.vercel.app).
- Don't invent flags. If something is genuinely TBD, leave a `TODO:` note for the maintainer.
