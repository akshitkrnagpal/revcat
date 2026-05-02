# Releasing

One-page runbook. Keep it current — when the steps shift, edit this file in the same PR.

## Prereqs

- `goreleaser` on `$PATH` (`brew install goreleaser`)
- `gh` CLI authenticated (`gh auth status`)
- Write access to `akshitkrnagpal/revcat` and `akshitkrnagpal/homebrew-tap`
- Working tree on `main`, fully merged, fully pushed

## Steps

```sh
# 1. Sync
git checkout main && git pull --ff-only

# 2. Verify everything still passes
make verify          # typecheck + test + drift-check

# 3. Pick the version (semver). Bug fixes -> patch (v0.5.1).
#    New features -> minor (v0.6.0). Breaking -> major.
VERSION=v0.5.1

# 4. Add a CHANGELOG entry under "## [VERSION] - YYYY-MM-DD".
$EDITOR CHANGELOG.md

# 5. Open a CHANGELOG-only PR, merge it.
git checkout -b changelog-$VERSION
git add CHANGELOG.md && git commit -m "CHANGELOG: $VERSION entry"
git push -u origin changelog-$VERSION
gh pr create --title "CHANGELOG: $VERSION entry" --body "..."
gh pr merge --squash --delete-branch

# 6. Sync, tag, push tag
git checkout main && git pull --ff-only
git tag -a $VERSION -m "$VERSION: <one-line summary>"
git push origin $VERSION

# 7. Run goreleaser locally (no GH Actions release flow per repo policy)
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean

# 8. Mark prerelease if alpha/beta/rc; otherwise leave as Latest
gh release edit $VERSION --prerelease   # only for prereleases
```

## What gets published automatically

- GitHub Release with binaries for darwin / linux / windows × amd64 / arm64 + `checksums.txt`
- Homebrew tap formula at `akshitkrnagpal/homebrew-tap` is updated

## Verification after release

```sh
# Brew tap users may need an explicit update on newer Homebrew
brew update && brew upgrade revcat
revcat version    # should print the new VERSION + commit + build time
```

## When NOT to release

If commits since the last tag are docs-only, scripts-only, or Makefile-only, the binary is functionally identical. Skip the release. Batch into the next real feature.

## Rollback

```sh
# Re-tag a prior commit if a release ships broken
gh release delete $VERSION --cleanup-tag
# Then either re-cut from a known-good commit or fix forward.
```

## Prerelease cadence

Use `vX.Y.Z-alpha.N` for early access; mark prerelease so it doesn't show as Latest. Promote to `vX.Y.Z` when smoke-tested on a real project.
