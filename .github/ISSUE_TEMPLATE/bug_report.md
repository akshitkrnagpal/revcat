---
name: Bug report
about: A revcat command behaved unexpectedly
title: "[bug] "
labels: bug
---

## What I ran

```sh
revcat ...
```

## What happened

Paste the actual output. If it was a 4xx/5xx from RC, include the full error including the `doc_url`. Re-running with `REVCAT_DEBUG=api` and pasting the redacted request/response is the most useful possible report.

## What I expected

What did you think would happen?

## Environment

- revcat version: `revcat version`
- OS / arch:
- Install method: brew / go install / binary download

## doctor output

```
revcat doctor --output table
revcat auth doctor --output table
```
