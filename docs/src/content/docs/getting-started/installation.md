---
title: Installation
description: Install revcat via Homebrew, npm, go install, or a pre-built binary.
---

revcat is a single static Go binary. Install it however you prefer.

## Homebrew

```sh
brew install akshitkrnagpal/tap/revcat
```

The tap lives at <https://github.com/akshitkrnagpal/homebrew-tap> and is updated by every release.

## npm

```sh
npm install -g revcat
```

Works on macOS / Linux / Windows on x64 and arm64 (no Windows arm64 yet). The npm package is a thin shim that picks the matching prebuilt binary via `optionalDependencies`. Useful when you already have Node in your toolchain (CI, dev containers, web monorepos) and don't want to add Homebrew or Go.

## `go install`

Requires Go 1.26 or later.

```sh
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

The binary lands in `$(go env GOPATH)/bin`; make sure that's on your `PATH`.

## Pre-built binaries

Every release publishes binaries for macOS / Linux / Windows on amd64 and arm64. Grab the matching archive from [GitHub Releases](https://github.com/akshitkrnagpal/revcat/releases), unpack, and put `revcat` on your `PATH`.

## Verify

```sh
revcat version
revcat doctor
```

`revcat doctor` runs a top-level health check and tells you what's missing if you haven't logged in yet.

## Next

[Quickstart →](/getting-started/quickstart/)
