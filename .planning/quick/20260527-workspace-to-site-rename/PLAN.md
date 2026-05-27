# Quick Task: workspace → site rename

Breaking terminology rename across runtime code, CLI, env, and active docs.

## Waves

1. Config & CLI: `--site`/`-s`, `MONMS_SITE`, `ResolveSite`, `StripSiteFlags`
2. Package `internal/workspace` → `internal/site`; `monms site sync`
3. Identifiers: `SiteAbs`, `ensureUnderSite`, `NewSite`, `NewEditorialSite`
4. `git mv workspace site`
5. Docs: CLAUDE.md migration note, README, specs, site/*.md, PROJECT.md
6. `go test ./...`, rebuild, smoke test

## Out of scope

Bulk edit of `.planning/phases/` historical artifacts.
