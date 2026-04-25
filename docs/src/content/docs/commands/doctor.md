---
title: doctor
description: Run a top-level health check.
---

`revcat doctor` walks the most common breakage points in one shot: platform, build version, credential store reachable, active profile resolves, API reachable.

```sh
revcat doctor
```

Sample output:

```
status  check             detail
OK      platform          darwin/arm64 go1.26.2
OK      revcat            0.0.1-dev
OK      credential store  keychain
OK      active profile    default
OK     api reach         ok, 2 project access
```

Pipe to JSON to script against the result:

```sh
revcat doctor --output json | jq '.[] | select(.status == "FAIL")'
```

For auth-specific diagnostics use `revcat auth doctor` instead.
