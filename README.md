# revcat

The RevenueCat CLI. Run your RevenueCat project from the terminal instead of clicking through the dashboard.

> Pre-MVP.

## Install

```sh
# from source (Go 1.23+)
go install github.com/akshitkrnagpal/revcat/cmd/revcat@latest
```

Binary releases via Homebrew + GitHub Releases land with v0.1.

## Quick start

```sh
revcat auth login --name my-app --secret-key sk_xxx
revcat auth doctor
revcat subscribers info app_user_123
```

## Commands

```
revcat auth          login | status | doctor | use | logout | list
revcat doctor                                          # top-level health check
revcat version
```

More commands land daily through the MVP window.

## Auth

revcat reads a RevenueCat v2 secret key (`sk_...`) from one of:

1. `REVCAT_API_KEY` env (highest precedence, one-shot)
2. `--profile <name>` flag
3. `REVCAT_PROFILE` env
4. `~/.revcat/active` (set via `revcat auth use <name>`)
5. profile named `default`

Profiles live in your OS keychain by default. Pass `--bypass-keychain` (or set `REVCAT_BYPASS_KEYCHAIN=1`) to write them to `./.revcat/config.json` instead, useful in containers.

## License

MIT
