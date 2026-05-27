---
status: complete
slug: serve-auto-scaffold
date: 2026-05-27
---

# Serve auto-scaffold + setup wizard

## Summary

Implemented interactive site bootstrap on `monms serve` and idempotent `monms init`:

- **Pre-serve check** (`site.CheckSite` / `site.EnsureReady`): detects missing or incomplete site; prompts to scaffold on TTY; exits with actionable error when non-interactive.
- **Scaffold** (`scaffold.InitAt`): idempotent file creation with absolute-path `Created` list on stdout.
- **Setup wizard** (`scaffold.RunSetupWizard`): port (default 8090), allowed hosts (default localhost → localhost + 127.0.0.1), bind host (default 0.0.0.0); merge-writes `{site}/.monms/config.json`.
- **Start choice**: foreground serve, background daemon (`-d`), or exit without starting.
- **`monms init`**: same wizard on TTY anytime (idempotent re-run for operators); non-TTY remains scaffold-only for CI/scripts.

## Files

- `main.go` — EnsureReady + start-mode branching
- `internal/config/config.go` — `ResolveSiteMeta`
- `internal/site/validate.go`, `ensure.go`
- `internal/scaffold/init.go`, `wizard.go`, `config_save.go`
- `internal/cli/prompt/prompt.go`
- Tests across config, site, scaffold, prompt packages
- `README.md`, `internal/cli/help.go` — quick start / init help updates

## Verification

```bash
go test ./... -count=1   # pass
go build -o monms .      # pass
```
