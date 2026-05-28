---
status: complete
task: config-driven file logging with SCHEMA level
date: 2026-05-27
---

# Config-driven file logging

## Summary

PocketBase does **not** expose a dedicated schema-change log channel — collection events land in generic WARN/INFO. Added MonMS file logging under `{site}/.monms/logs/` with config-driven levels and a custom **SCHEMA** slog level + `schema.log`.

## Changes

- **`internal/logging/`** — level bitset (`error`, `warn`, `info`, `debug`, `schema`), lumberjack rotation, slog file handler, PocketBase tee handler, `logging.Schema()` API
- **`logging` config** — `logging: ["error"]` default; optional `loggingRotation` in `config.json`
- **PocketBase** — all PB logs → `pocketbase.log` (always) + enabled level files; SQLite aux log persistence disabled when file logging active
- **MonMS slog** — `slog.SetDefault` routes to level files; ERROR always on
- **Schema** — bootstrap import/seed success uses `logging.Schema()` → `schema.log` when enabled
- **Daemon** — no more `serve.log` stdout redirect; reports `.monms/logs/` path
- **Banner** — startup TTY shows logs directory
- **Scaffold** — wizard sets `"logging": ["error"]`; embed/example `_fieldDocs` updated

## Verification

```bash
go test ./... -count=1
```

## Usage

```json
"logging": ["error", "warn", "schema"]
```

Log files: `pocketbase.log`, `error.log`, and optionally `warn.log`, `info.log`, `debug.log`, `schema.log`.
