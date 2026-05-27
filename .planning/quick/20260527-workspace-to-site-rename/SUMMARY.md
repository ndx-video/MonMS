---
status: complete
slug: workspace-to-site-rename
date: 2026-05-27
---

# Quick Task Summary: workspace → site rename

Renamed MonMS terminology from **workspace** to **site** across runtime code, CLI, env vars, and active documentation.

## Changes

- CLI: `--site` / `-s`, env `MONMS_SITE`; removed `--workspace`, `-w`, `MONMS_WORKSPACE`
- Package `internal/workspace` → `internal/site`; command `monms site sync`
- Directory `workspace/` → `site/` (git mv)
- Identifiers: `ResolveSite`, `ValidateSite`, `SiteAbs`, `NewSite`, etc.
- CLAUDE.md migration note; `.planning/PROJECT.md` evolution entry

## Verification

- `go test ./... -count=1 -short` — pass
- `go build -o monms .` — pass
- `./monms site sync --help` — pass
- `./monms init -s /tmp/monms-site-smoke` — pass

## Out of scope

Historical `.planning/phases/` artifacts retain "workspace" (map via PROJECT.md / CLAUDE.md).
