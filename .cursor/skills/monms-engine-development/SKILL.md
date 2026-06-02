---
name: monms-engine-development
description: >-
  MonMS Go engine patterns — buildMode, PocketBase hooks, route registration,
  SSR/validate parity, schema bootstrap, content rail internals, and integration tests.
  Use when editing internal/, main.go, CLI, or engine tests.
---

# MonMS Engine Development

For L2 site mutations, use [monms-site-shaping](../monms-site-shaping/SKILL.md). For `/_monms/` console UI, use [monms-dash](../monms-dash/SKILL.md). Foundation: [monms-architecture](../monms-architecture/SKILL.md).

## Compile-time mode

Production vs development is **not** an env var:

```bash
# Development (default) — always reads templates from disk, no watcher
go build -o monms .

# Production — template cache + fsnotify watcher on entire site tree
go build -ldflags "-X main.buildMode=production" -o monms .
```

| Mode | Template cache | fsnotify watcher |
|------|----------------|------------------|
| development | Off (re-parse each request) | Off |
| production | On | On (`.gohtml` templates + `documents/**/*.md` sync) |

Production behavior tests need `-ldflags "-X main.buildMode=production"` on packages that read `main.buildMode`.

## Bootstrap and hook order

From `main.go` serve pipeline:

1. Optional shape sync (`content` package)
2. Logging setup
3. `documents.RegisterBootstrapHook` **before** `schema.RegisterBootstrapHook` — markdown sync runs after schema import (outer hook calls `e.Next()` first)
4. `schema.RegisterBootstrapHook(app, siteAbs)` — merge `{siteDir}/schema/*.json`, import collections, seed hero
5. `router.RegisterAuthHooks(app)` — cookie bridge for inline edit
6. `OnServe`: **`content.RegisterRoutes` before `router.RegisterRoutes`** (D-14); production starts `documents.StartWatcher` for `.md` files
7. Wrapped installer banner

Integration tests must mirror this order.

## Routing conventions

**Registration order (D-14):** MonMS JSON API + `/_monms/*` → assets → fragments → SSR catch-all.

**Canonical paths:** `internal/monmsroutes/routes.go`

| Constant | Path |
|----------|------|
| `ContentImportPath` | `/api/monms/content/import` |
| `PublishPath` | `/_monms/publish` |
| `AuthSyncPath` | `/_monms/auth/sync` |
| `AuthLogoutPath` | `/_monms/auth/logout` |

**New reserved URL prefixes:** add to `isReservedSlug` in `internal/router/ssr.go` so SSR does not treat them as page templates. Current reserved: `api`, `assets`, `_`, `_monms`.

**SSR parsing:** `template.ParseFiles(layoutPath, pagePath)` then `ExecuteTemplate(..., "base", data)`. Validation in `internal/validate/` must use the **same** layout+page pairing.

## Package rules

- `internal/` never imports `main` — pass config via function args/setters (e.g. `TemplateCache.SetProductionMode`)
- Use `slog` for structured logging
- Site path: `-s` / `--site` wins over `MONMS_SITE` env (D-26); default `./site` when both omitted — **not a fixed folder name**
- PocketBase data dir: `{siteDir}/.pb_data/` (D-27)

## Schema bootstrap

- Loader merges all `{siteDir}/schema/*.json` (single object or array per file)
- Calls `ImportCollectionsByMarshaledJSON` on bootstrap — self-healing on fresh deploys
- `"editorial": true` is **stripped by PocketBase** — MonMS re-reads raw JSON via `internal/schema/meta.go` for content rail allowlist
- `"monms": { "source": "markdown", ... }` marks Git markdown collections — excluded from JSON publish rail

## Content rail internals

Package: `internal/content/`

| Concern | Detail |
|---------|--------|
| Export/import/diff/publish | CLI: `monms content <subcommand>` |
| Import API gate | `Authorization: Bearer $MONMS_PUBLISH_TOKEN` |
| Allowlist | Collections with `"editorial": true` in schema JSON (PB-native only for export/publish) |
| Upsert | By stable record `id`; deletions not propagated from staging |
| Denied | `_superusers`, `users` collections |
| Publish UI | `/_monms/publish` — gated by `publisherEmails`; shell/notifications: [monms-dash](../monms-dash/SKILL.md) |

Agents do not routine-push editorial copy; clients use publish console.

## Integration test boilerplate

Pattern from `internal/router/handlers_test.go`:

```go
app := pocketbase.NewWithConfig(pocketbase.Config{
    DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
    DefaultDev:      true,
    HideStartBanner: true,
})
schema.RegisterBootstrapHook(app, siteAbs)
RegisterAuthHooks(app)

deps := Deps{SiteAbs: siteAbs, Cache: cache, IsDev: opts.isDev}
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    RegisterAdminUIExtension(se)
    RegisterRoutes(se, deps)
    return se.Next()
})
```

Use `internal/testutil/` helpers: `NewSite`, `NewEditorialSite`, `NewSuperuser`, `AuthClient`.

Main.go also registers `content.RegisterRoutes` before `router.RegisterRoutes` — include both when testing MonMS API routes.

## What to change where

| Change | Package |
|--------|---------|
| SSR behavior | `internal/router/ssr.go` |
| Auth cookie bridge | `internal/router/` auth hooks |
| Cache/watcher | `internal/templates/` |
| Validation rules | `internal/validate/validate.go` |
| Content publish rail | `internal/content/` |
| Markdown document rail | `internal/documents/` — CLI `monms documents`, bootstrap sync, SSR fallback |
| Init scaffold | `internal/scaffold/init.go` + `internal/scaffold/embed/` |
| Site validation | `internal/site/` |
| CLI help | `internal/cli/` |

When updating embedded scaffold files, also update the live site directory copies if the demo should stay in sync (often `./site` in this repo — whatever path that instance uses).

## Testing

```bash
go test ./... -count=1           # Full suite including perf gates
go test ./... -count=1 -short    # Skip TestIdleMemory, TestTTFB
```

Key test files: `internal/router/handlers_test.go`, `inline_edit_test.go`, `press_releases_test.go`, `internal/scaffold/hook_test.go`.

## Out of scope (unless explicitly requested)

- React/Next.js or Node build pipelines
- External databases (PostgreSQL, MySQL)
- Multi-site / Host-header routing
- Real-time WebSocket push
- In-process plugins — see [monms-extensibility](../monms-extensibility/SKILL.md)

## Related docs

- [PROJECT.md](../../../PROJECT.md) — repo layout, D-* decisions
- [docs/reference/monms-api.md](../../../docs/reference/monms-api.md)
- [docs/reference/cli.md](../../../docs/reference/cli.md)
