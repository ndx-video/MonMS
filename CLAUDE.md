# CLAUDE.md — MonMS Agent Instructions

This file guides AI assistants working in the MonMS repository. Read it before making changes.

## Project summary

MonMS is an **agent-malleable monolithic CMS**: a single Go binary (PocketBase embedded) serves a Git-tracked `workspace/` containing templates, assets, and schema JSON. Structural changes happen in the workspace without rebuilding the binary. Human editors use HTMX inline editing on authenticated pages.

**Core value:** Runtime malleability with < 30 MB RAM idle, no compilation step for UI changes.

**Authoritative docs:**
- Product vision: `specs/monms-prd.md`
- Staging, environments, content publish: `specs/staging.md` (accepted v2 spec)
- Requirements: `.planning/REQUIREMENTS.md`
- Roadmap: `.planning/ROADMAP.md`
- Workspace mutations: `workspace/agent-guide.md`
- Security: `workspace/SECURITY.md`

## Repository layout

```
main.go                    # Entry: init | validate | serve
internal/
  config/                  # --workspace flag, MONMS_WORKSPACE env
  content/                 # Editorial export/import/diff/publish (v2)
  router/                  # SSR, assets, fragments, auth cookie bridge
  schema/                  # Bootstrap import from schema/*.json, seed, editorial flag
  scaffold/                # monms init, embedded templates, pre-commit hook
  templates/               # TemplateCache, ResolveSlug, fsnotify watcher
  validate/                # monms validate CLI (dry-run + HTML lint)
  workspace/               # Workspace directory validation
workspace/                 # Git-tracked deployable state (separate git repo)
specs/monms-prd.md         # PRD
specs/staging.md           # Four layers, staging/prod, content publish (v2)
.planning/                 # GSD planning artifacts (do not treat as runtime config)
```

## Commands

```bash
go build -o monms .                                          # Development binary
go build -ldflags "-X main.buildMode=production" -o monms .  # Production binary
./monms --help                                               # CLI overview (all commands)
./monms <command> --help                                     # Per-command help
./monms init [-w|--workspace PATH]                              # Scaffold workspace
./monms validate [-w|--workspace PATH] [files...]               # Validate templates
./monms content <export|import|diff|publish> [-w|--workspace PATH]  # Editorial sync (v2)
./monms [-w|--workspace PATH]                                   # Start server (port 8090)
go test ./... -count=1                                       # All tests
go test ./... -count=1 -short                                # Skip perf/memory gates
```

## Architecture decisions (follow these)

| Decision | Rule |
|----------|------|
| **D-01** | Production vs development mode is compile-time via `main.buildMode` ldflags — **not** `ENV` or runtime env vars |
| **D-04** | fsnotify watcher runs **only** in production builds |
| **D-10/D-11** | Slug resolution uses mirror+index: `/` → `index.gohtml`, `/press` → `press/index.gohtml` or `press.gohtml` |
| **D-14** | Route registration order: MonMS JSON API + `/_monms/*` tools → assets → fragments → SSR catch-all |
| **D-26** | `-w` / `--workspace` flag wins over `MONMS_WORKSPACE` env |
| **D-27** | PocketBase data dir: `{workspace}/.pb_data/` |
| **D-30** | Watcher monitors entire workspace tree, not just `templates/` |
| **D-32/D-33** | Schema JSON in `workspace/schema/` is audit trail + bootstrap self-healing; live changes via PocketBase API |
| **D-37** | Template validation must mirror production: `html/template.ParseFiles(layoutPath, pagePath)` |
| **D-43** | Pre-commit hook rolls back with `git checkout -- .` on validation failure |

## Four layers & promotion rails (accepted v2 — see `specs/staging.md`)

| Layer | Artifact | Promotion |
|-------|----------|-----------|
| L1 Engine | `monms` binary | Semver release |
| L2 Structure | `workspace/` Git — templates, schema, assets | Git tag → production deploy |
| L3 Content | `.pb_data/` records; export to `content/*.json` | Client **Publish to live** at `/_monms/publish` → JSON upsert via `POST /api/monms/content/import` |
| L4 Audience | Production URL | Read-only |

- **Structure rail** and **content rail** are independent. Git tags do not carry editorial copy.
- **Media:** public CDN URLs in text fields — blobs do not move between staging and production. See `workspace/MEDIA.md`.
- **Roles:** consultants own structure tags; clients own content publish; consultants are not in the routine content loop.
- **Phase 4 (v2) implemented:** `internal/content/` package, `monms content` CLI, `POST /api/monms/content/import` (JSON API), publish console at `/_monms/publish` (HTML tool — not PocketBase admin SPA at `/_/`). Publisher allowlist in gitignored `workspace/.monms/config.json`; commit `config.example.json` only.

## HTTP routing (MonMS namespaces)

| Prefix | Purpose | Examples |
|--------|---------|----------|
| `/api/monms/*` | **JSON REST only** — machine clients, Bearer tokens | `POST /api/monms/content/import` |
| `/_monms/*` | **Operator tools** — HTML pages and browser session helpers | `GET /_monms/publish`, `POST /_monms/auth/sync`, `GET /_monms/auth/logout` |

`workspace/.monms/config.json` fields: `productionUrl`, `publisherEmails`, `allowedHosts` (injects `monms serve --origins` when CLI flag omitted; CLI wins if set), `bind` (injects `--http=host:port` when CLI `--http` omitted).
| `/api/` (other) | PocketBase collection REST | `/api/collections/...` |
| `/_/` | PocketBase admin SPA | Full management fallback |

Canonical path constants live in `internal/monmsroutes/routes.go`. Register `/_monms/*` and `/api/monms/*` in `content.RegisterRoutes` and auth hooks **before** the SSR catch-all (D-14). Add new reserved slug prefixes to `isReservedSlug` in `internal/router/ssr.go` so SSR does not treat them as page templates.

## Code conventions

### Go

- Keep `internal/` packages free of imports from `main` — pass config via function args/setters (e.g. `TemplateCache.SetProductionMode`)
- Use `slog` for structured logging
- Match existing test patterns in `internal/router/handlers_test.go` and `internal/testutil/`
- Integration tests must register auth hooks like `main.go` does (`router.RegisterAuthHooks`)

### Templates (`workspace/templates/`)

- Page templates: `{{define "body"}}...{{end}}` only — base layout adds HTML shell
- Fragments: no `{{define "body"}}`, served without layout at `/fragments/{name}`
- Inline edit attrs gated on `{{if .IsLoggedIn}}`
- HTMX auth: base layout sets Bearer token via `htmx:configRequest` from server-injected `AuthToken` (not `document.cookie` — cookie is HttpOnly)

### Agent workspace mutations

When modifying the workspace as an agent:

1. **Schema:** dual-write — POST `/api/collections` then write `schema/{name}.json`
2. **Templates:** edit `*.gohtml`, run `monms validate`, commit with `agent:` prefix
3. **Never commit:** `.pb_data/`, `.monms/config.json`, `.monms/publish-state.json`, tokens, secrets, `.env`
4. **Rollback caveat:** `git checkout -- .` does not remove newly added untracked files after failed validation — clean up manually

## What to change vs leave alone

| Change type | Where |
|-------------|-------|
| New page/route | `workspace/templates/{slug}.gohtml` |
| Global layout/HTMX | `workspace/templates/layouts/base.gohtml` + `internal/scaffold/embed/base.gohtml` |
| New collection | PocketBase API + `workspace/schema/{name}.json` (add `"editorial": true` for client-publishable collections) |
| Content publish rail | `internal/content/` — agents do not routine-push; clients use `/_monms/publish`; production import at `POST /api/monms/content/import` |
| SSR behavior | `internal/router/ssr.go` |
| Cache/watcher | `internal/templates/` |
| Validation rules | `internal/validate/validate.go` |
| Init scaffold | `internal/scaffold/init.go` + `internal/scaffold/embed/` |

When updating embedded scaffold files, also update the live `workspace/` copies if they should stay in sync for the demo workspace.

## Testing expectations

- Run `go test ./... -count=1` before claiming work is complete
- Auth/inline-edit tests live in `internal/router/inline_edit_test.go`, `auth_test.go`
- Agent workflow tests: `internal/router/press_releases_test.go`, `internal/scaffold/hook_test.go`
- Perf gates (`TestIdleMemory`, `TestTTFB`) may skip in `-short` mode

## Security reminders

- PocketBase admin at `/_/` — full management fallback
- Unauthenticated PUT to collections must be rejected (SEC-02) — enforced by PocketBase API rules
- Agent SSH keys and admin tokens scoped to workspace only — see `workspace/SECURITY.md`
- Never log or commit `POCKETBASE_ADMIN_TOKEN`

## Out of scope (do not implement without explicit request)

- React/Next.js or Node build pipelines
- External databases (PostgreSQL, MySQL)
- Multi-workspace / Host-header routing (v2)
- Real-time WebSocket push
- Modifying `.planning/` unless the user asks for planning work

## Common pitfalls

1. **404 after adding a page** — template path must mirror URL slug exactly (check mirror+index rule)
2. **Template changes not visible in dev** — development mode always reads from disk (no cache); if stale, check you're editing the right workspace path
3. **Production cache not invalidating** — watcher only runs with `buildMode=production` ldflag
4. **Validation passes but serve fails** — ensure validate uses same `ParseFiles` paths as `ssr.go`
5. **Inline edit 401** — user must re-authenticate at `/_/`; Bearer token injected server-side
6. **Re-init overwrites** — `monms init` skips existing files; manual merge needed for scaffold updates on old workspaces

## Planning context

GSD milestone v1 has three phases (all verified):

1. Core Go runtime & workspace foundation
2. Agent mutation engine & safety guardrails
3. Inline contextual editing & demonstration content

**v2 Phase 4 (implemented):** staging environments, `workspace/content/` JSON sync, `/_monms/publish` console, publisher role — see `specs/staging.md` and `.planning/ROADMAP.md`.

Other v2 backlog (EXT-*, MULT-*, RICH-*) is in `.planning/ROADMAP.md` — do not implement unless asked.
