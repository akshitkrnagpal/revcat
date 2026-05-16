# npm packaging

revcat is distributed via npm under six packages: one main (`revcat`) plus five per-platform packages (`revcat-darwin-arm64`, `revcat-darwin-x64`, `revcat-linux-arm64`, `revcat-linux-x64`, `revcat-win32-x64`).

The main package is a thin JS shim. It picks the matching platform package via `optionalDependencies` (npm skips the others using `os` + `cpu` filters) and execs the native binary.

## Layout

```
npm/
├── revcat/                  # main package — shim only, no binary
│   ├── package.json         # optionalDependencies fans out to platforms
│   └── bin/revcat           # JS shim, resolves the right platform pkg + execs it
├── revcat-darwin-arm64/
│   ├── package.json         # os: ["darwin"], cpu: ["arm64"]
│   └── bin/revcat           # native binary, written by the publish script
├── revcat-darwin-x64/        ditto
├── revcat-linux-arm64/       ditto
├── revcat-linux-x64/         ditto
└── revcat-win32-x64/         (binary lives at bin/revcat.exe)
```

The platform `bin/<binary>` files are gitignored and produced by the publish script - the repo only commits the JS shim and the package.json metadata.

## Releasing

After `goreleaser release` (or local `goreleaser build --snapshot`) has produced archives in `dist/`, run:

```sh
scripts/npm-publish.mjs <version>           # e.g. 0.6.0
scripts/npm-publish.mjs <version> --dry-run # smoke-test without publishing
```

The script:

1. Extracts each platform's binary out of the matching `dist/revcat_<version>_<os>_<arch>.{tar.gz,zip}` into `npm/<pkg>/bin/`.
2. Bumps every `package.json` to `<version>` (including the main package's `optionalDependencies` map - they stay in lockstep).
3. Runs `npm publish --access public` on each platform package, then the main package last so its optionalDependency entries already exist on the registry by the time anyone installs `revcat`.

Auth is whatever your local `~/.npmrc` provides - the script doesn't ship credentials. For a one-off CI-less release: `npm login`, then run the script.

## Why not a postinstall download?

The optionalDependencies pattern (esbuild, biome, swc, lefthook, turbo) avoids three classes of failure:

- Corporate proxies / firewalls blocking GitHub releases
- npm `--ignore-scripts` skipping postinstall and leaving the user with no binary
- Offline / cached installs that can't reach the network

With optionalDependencies, the binary is just an npm package - it gets the registry mirror, the lockfile pin, the cache, and the offline-install behavior for free.
