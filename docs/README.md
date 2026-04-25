# revcat docs

The Astro Starlight site that powers <https://revcat.vercel.app>.

## Run locally

```sh
cd docs
pnpm install
pnpm dev          # http://localhost:4321
```

## Build

```sh
pnpm build        # outputs to docs/dist
pnpm preview      # serve the built site
```

## Deploy to Vercel

The `docs/` folder is a self-contained Vercel project.

1. Sign in to <https://vercel.com> and import this repo.
2. Set **Root Directory** to `docs`.
3. Framework Preset is auto-detected as **Astro**. The build / install / output commands are pinned in `vercel.json`:
   - Build: `pnpm build`
   - Install: `pnpm install --frozen-lockfile`
   - Output: `dist`
4. Add the production domain (`revcat.vercel.app` or your custom one) in Project Settings -> Domains.
5. Deploy.

Pushes to `main` redeploy automatically once linked. Preview deployments are produced for every other branch / PR.

## Structure

- `astro.config.mjs` - Starlight config (sidebar, social, dark mode default).
- `src/content/docs/` - all docs pages.
  - `index.mdx` - landing.
  - `getting-started/` - install + quickstart.
  - `commands/` - one page per top-level subcommand.
  - `guides/` - task-oriented walkthroughs.
  - `reference/` - configuration + lower-level details.
- `src/styles/custom.css` - opt-in stylesheet (currently sets dark color-scheme by default).

## Adding a command page

When a new subcommand lands in the CLI, drop a new `.md` file in `src/content/docs/commands/<name>.md` and add an entry to the `commands` array in `astro.config.mjs`.
