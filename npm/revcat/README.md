# revcat

The RevenueCat CLI. Run RevenueCat from the terminal instead of clicking through the dashboard.

```sh
npm install -g revcat
revcat auth login
revcat metrics overview
```

This is a thin npm wrapper around the native [revcat](https://github.com/akshitkrnagpal/revcat) Go binary. On install, npm picks the matching platform package (`revcat-darwin-arm64`, `revcat-linux-x64`, etc.) via `optionalDependencies` and the JS shim execs it with your argv.

Supported platforms: `darwin-arm64`, `darwin-x64`, `linux-arm64`, `linux-x64`, `win32-x64`.

For Homebrew, `go install`, or pre-built tarballs see the [main repo](https://github.com/akshitkrnagpal/revcat).
