---
title: Installation
description: Install revcat via Homebrew, go install, or a pre-built binary.
---

revcat is a single static Go binary. Install it however you prefer.

## Homebrew (planned, v0.1+)

```sh
brew install akshitkrnagpal/tap/revcat
```

> The Homebrew tap will land alongside v0.1. Until then, use `go install` or grab a binary from GitHub Releases.

## `go install` (today)

Requires Go 1.23 or later.

```sh
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

The binary lands in `$(go env GOPATH)/bin`; make sure that's on your `PATH`.

## Pre-built binaries (planned, v0.1+)

Each release will publish binaries for macOS / Linux / Windows on amd64 and arm64. Grab the matching archive from [GitHub Releases](https://github.com/akshitkrnagpal/revcat/releases), unpack, and put `revcat` on your `PATH`.

## Verify

```sh
revcat version
revcat doctor
```

`revcat doctor` runs a top-level health check and tells you what's missing if you haven't logged in yet.

## Next

[Quickstart →](/getting-started/quickstart/)
