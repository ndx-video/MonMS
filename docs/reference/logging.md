# File logging

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

MonMS writes structured logs to `{site}/.monms/logs/`. Both MonMS (`slog`) and PocketBase (`app.Logger()`) tee into these files. PocketBase’s built-in SQLite log store is disabled when file logging is active — the log directory is the source of truth.

Console output is separate: run `monms serve --dev` for verbose PocketBase/SQL **console** logs (`devLogging` in `_fieldDocs` — CLI only, not stored in config).

## Log directory

```
site/.monms/logs/
├── pocketbase.log   # always — full PocketBase stream
├── error.log        # always — ERROR and above
├── warn.log         # when warn level enabled
├── info.log         # when info level enabled
├── debug.log        # when debug level enabled
└── schema.log       # when schema level enabled
```

The serve startup banner prints the absolute logs path on a TTY. Background (`-d`) processes write here instead of `.monms/serve.log`.

Rotation is optional via `loggingRotation` in config (lumberjack: size, backups, age, compress).

## Configuration

In `site/.monms/config.json`:

```json
{
  "logging": ["error", "warn", "schema"],
  "loggingRotation": {
    "maxSizeMB": 10,
    "maxBackups": 5,
    "maxAgeDays": 30,
    "compress": true
  }
}
```

| Key | Meaning |
|-----|---------|
| `logging` | Which optional level files to open. **ERROR is always on** regardless of this list. |
| `loggingRotation` | Optional per-file rotation settings. |

### Build-mode defaults

If `logging` is **omitted** from config, MonMS uses compile-time `buildMode`:

| Binary | Default `logging` |
|--------|-------------------|
| **Production** (`-ldflags "-X main.buildMode=production"`) | `error`, `warn`, `schema` |
| **Development** (plain `go build`) | `error`, `warn`, `info`, `debug`, `schema` |

An explicit `"logging": []` means **error only** (empty list does not fall back to build-mode defaults).

Production binary:

```bash
go build -ldflags "-X main.buildMode=production" -o monms .
```

Development binary (engine work):

```bash
go build -o monms .
```

## Log levels

| Level | Config value | File | What it captures |
|-------|--------------|------|------------------|
| **ERROR** | `error` (always on) | `error.log` | Failures, panics, unrecoverable errors from MonMS and PocketBase |
| **WARN** | `warn` | `warn.log` | PocketBase warnings (collection cache, file cleanup, rate limits, etc.) |
| **INFO** | `info` | `info.log` | Routine operational messages |
| **DEBUG** | `debug` | `debug.log` | Verbose diagnostics (auth traces, realtime, query-adjacent noise) |
| **SCHEMA** | `schema` | `schema.log` | MonMS-only custom level: schema bootstrap import from `site/schema/`, hero seed success |

**SCHEMA** is not a PocketBase level. PocketBase does not expose a dedicated schema-change channel; collection events appear as generic WARN/INFO in `pocketbase.log`. MonMS schema sync uses `logging.Schema()` so operators can tail structure changes without enabling full DEBUG.

## Examples

Production staging (recommended explicit config):

```json
"logging": ["error", "warn", "schema"]
```

Engine debugging on a dev binary (omit `logging` key — all files enabled by default):

```json
{}
```

Minimal logging (errors only):

```json
"logging": []
```

Full audit including debug:

```json
"logging": ["error", "warn", "info", "debug", "schema"]
```

## Related

- [CLI reference](cli.md) — `monms serve`, `--dev`
- [Getting started](../operators/getting-started.md) — `.monms/config.json` overview
- [Deploy with Docker](../operators/deploy-docker.md) — production binary build
