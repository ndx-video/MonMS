# Phase 1: Core Go Runtime & Workspace Foundation - Context

**Gathered:** 2026-05-22
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver a working Go + PocketBase single binary that serves SSR templates from a Git-tracked workspace folder, invalidates an in-memory template cache via fsnotify in production, serves static assets safely, scaffolds via `monms init`, and ships a base layout with HTMX/Alpine CDN wiring and hybrid styling. Phase 1 does not include agent mutation guardrails (Phase 2) or inline contextual editing (Phase 3).

</domain>

<decisions>
## Implementation Decisions

### Dev vs Production Mode
- **D-01:** Mode is auto-detected via compile-time ldflags — release builds set `-ldflags "-X main.buildMode=production"`; default/unset is development.
- **D-02:** Development mode never caches parsed templates — every request reads and parses from disk.
- **D-03:** Production mode uses in-memory template cache with `sync.RWMutex` (per PRD/ROADMAP).
- **D-04:** fsnotify watcher runs in production only — skip the goroutine entirely in development.

### Workspace Bootstrap
- **D-05:** Workspace is created by a separate `monms init` command — the serve binary must not auto-scaffold.
- **D-06:** If workspace is missing or incomplete at startup, fatal exit with message directing user to run `monms init` (exit code 1).
- **D-07:** `monms init` generates full Phase 1 scaffold: `templates/layouts/base.gohtml`, `templates/index.gohtml` stub, `templates/errors/errors.gohtml`, `assets/main.css`, `schema/`, `templates/fragments/`, and placeholder files as needed.
- **D-08:** `monms init` runs `git init` in workspace only when `.git` does not already exist.
- **D-09:** Startup validates workspace structure — require `templates/layouts/base.gohtml` and `assets/` to exist before serving.

### Route → Template Mapping
- **D-10:** Nested URL paths are supported with mirror + index convention — `/press` → `templates/press/index.gohtml`, `/press/2024` → `templates/press/2024.gohtml`.
- **D-11:** Homepage `/` and empty slug resolve to `templates/index.gohtml` at workspace root (not nested under `index/`).
- **D-12:** Trailing slashes are normalized — strip trailing slash before template lookup.
- **D-13:** URL slug case is preserved — filesystem case-sensitivity rules apply (Linux default).
- **D-14:** SSR catch-all must not intercept reserved prefixes — `/api/*`, `/_/*`, and `/assets/*` are excluded from SSR routing.
- **D-15:** Fragment partials are served at `/fragments/{name}` as HTMX swap targets in Phase 1.
- **D-16:** Template parse errors return HTTP 500 — show detailed error in dev mode, safe generic message in production.

### Error Pages
- **D-17:** Unknown slugs return styled HTML via workspace template `templates/errors/errors.gohtml` wrapped in base layout, with `Code` and `Message` template variables.
- **D-18:** If workspace error template is missing, fall back to minimal built-in HTML 404.
- **D-19:** 404 pages display the attempted path (e.g., "Page not found: /missing-page") plus a link home.
- **D-20:** 500 errors reuse the same `errors.gohtml` layout with different `Code`/`Message` — no separate 500 template file.

### Styling Baseline
- **D-21:** Hybrid styling — Tailwind preflight via CDN with minimal inline config in `base.gohtml`, plus `workspace/assets/main.css` for component classes (`.btn`, `.card`, `.hero`, etc.).
- **D-22:** Templates use minimal utility classes; branded/component styling lives in `main.css`.
- **D-23:** Alpine.js is included in `base.gohtml` with a minimal demo — simple mobile nav toggle in the index stub proves wiring (full editor UX deferred to Phase 3).
- **D-24:** HTMX CDN tag included in base layout per DEMO-03 (editor overlay block structure only — no auth badge logic until Phase 3).

### Workspace Path & Data Layout
- **D-25:** Default workspace path is `./workspace` relative to process CWD.
- **D-26:** Workspace path is overridable via `MONMS_WORKSPACE` env var and `--workspace` CLI flag — flag overrides env when both set. Same resolution applies to `monms init` and serve.
- **D-27:** PocketBase data directory lives at `workspace/.pb_data/` — co-located with site workspace, not a sibling `./pb_data`.
- **D-28:** Operator manages `.gitignore` for `workspace/.pb_data/` — `monms init` does not auto-gitignore database files.
- **D-29:** Static asset serving uses root-jail validation — resolve absolute path and verify result stays under `workspace/assets/`; reject traversal with 403.
- **D-30:** fsnotify watches the entire workspace tree for `.gohtml` changes in production (not just `templates/` subdirectory).
- **D-31:** Startup logs both configured workspace path (env/flag value) and resolved absolute path at info level.
- **D-32:** `workspace/schema/` holds declarative JSON collection definitions — binary reads `schema/*.json` on startup to sync PocketBase collections in Phase 1.

### Claude's Discretion
- Exact Tailwind CDN script URL and inline config snippet for preflight-only setup.
- Built-in fallback 404/500 HTML markup when workspace error template is absent.
- `monms init` marker file format (if any beyond structural validation).
- fsnotify debouncing/coalescing strategy for rapid agent commits.
- Declarative schema sync error handling (log-and-continue vs fatal on invalid JSON).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `specs/monms-prd.md` — Authoritative PRD with reference `main.go`, template cache, fsnotify, base layout, and inline editing examples (Phase 3 sections inform base layout structure only)
- `.planning/PROJECT.md` — Core value, tech stack constraints (Go + PocketBase + SQLite + HTMX), performance targets (<30MB RAM, <15ms TTFB)
- `.planning/REQUIREMENTS.md` — ENG-*, WRK-*, SEC-01, DEMO-03 traceability for Phase 1
- `.planning/ROADMAP.md` — Phase 1 deliverables list and requirement mapping

### Architecture (from PRD §2–§3)
- `specs/monms-prd.md` §2 — Directory topology: engine binary, workspace Git repo, schema/templates/assets layout
- `specs/monms-prd.md` §3 — TemplateCache struct, getTemplate(), watchWorkspace() reference implementation

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — greenfield repository. No Go source files exist yet.

### Established Patterns
- PRD embeds a complete reference implementation for TemplateCache, SSR catch-all routing, static assets, and fsnotify invalidation — planner should treat `specs/monms-prd.md` §3 as the starting point, adapting per decisions above (ldflags mode detection, nested routing, init command, schema sync).
- Planning config enables comprehensive granularity, research, and verification — plans should include RAM/TTFB validation steps.

### Integration Points
- PocketBase embedded via `pocketbase.New()` — admin at `/_/`, API at `/api/*`
- Workspace Git repo is independent of engine repo — init may create nested git in `workspace/`
- Phase 2 will add agent mutation hooks referencing workspace structure established here
- Phase 3 will extend `base.gohtml` with auth badge, contenteditable, and HTMX PUT wiring

</code_context>

<specifics>
## Specific Ideas

- Release builds distinguish production via `-ldflags "-X main.buildMode=production"` — not ENV-based alone.
- Index stub should include Alpine mobile nav toggle as a wiring proof, not full editor UX.
- Error page shows the missing path explicitly — helpful for agent debugging wrong slugs.
- Schema folder is active in Phase 1 (declarative PocketBase sync on startup), not a passive placeholder.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---
*Phase: 1-Core Go Runtime & Workspace Foundation*
*Context gathered: 2026-05-22*
