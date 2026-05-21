---
phase: 01-core-go-runtime-workspace-foundation
plan: 04
subsystem: router
tags: [go, pocketbase, ssr, assets, htmx-fragments, path-jail, perf-gates]

requires:
  - phase: 01-02
    provides: TemplateCache, integration test server pattern
  - phase: 01-03
    provides: ResolveSlug, production cache semantics
  - phase: 01-05
    provides: Scaffold templates with hero and errors.gohtml
provides:
  - GET /assets/{path...} with D-29 root-jail
  - GET /fragments/{name} partial HTML without base layout
  - GET /{slug...} SSR with styled 404/500 error pages
  - ENG-05/ENG-06 automated perf gates
affects: [phase-2, phase-3]

tech-stack:
  added: []
  patterns: [route order assets→fragments→catch-all, safeAssetPath Abs+prefix jail, ExecuteTemplate base for SSR]

key-files:
  created:
    - internal/router/assets.go
    - internal/router/ssr.go
    - internal/router/fragments.go
    - internal/router/assets_test.go
  modified:
    - main.go
    - internal/router/handlers_test.go
    - internal/router/perf_test.go
    - internal/testutil/workspace.go

key-decisions:
  - "Traversal security verified via TestSafeAssetPath; ServeMux canonicalizes .. paths before handler (301)"
  - "Production 500 uses generic copy; dev mode surfaces parse error detail (D-16)"
  - "Reserved slug prefixes api/assets/_ return 404 via defensive SSR check (D-14)"

patterns-established:
  - "RegisterRoutes registers GET-only handlers in fixed order before catch-all"
  - "renderErrorPage uses errors.gohtml + base with fallback minimal HTML (D-18)"

requirements-completed: [ENG-05, ENG-06, WRK-03, WRK-04, WRK-05]

duration: 12min
completed: 2026-05-22
---

# Phase 1 Plan 04: HTTP Router (Assets, SSR, Errors) Summary

**Root-jailed static assets, SSR catch-all with styled error pages, HTMX fragments, and sub-30MB / sub-15ms perf gates**

## Performance

- **Duration:** 12 min
- **Started:** 2026-05-21T15:58:59Z
- **Completed:** 2026-05-21T15:59:00Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Implemented `AssetsHandler` with `safeAssetPath` Abs+prefix jail (WRK-03, D-29)
- Implemented `SSRHandler`, `FragmentsHandler`, and `RegisterRoutes` wired in `main.go` (WRK-04, WRK-05, D-14–D-20)
- Enabled integration tests for homepage SSR, styled 404, fragments, production 500 generic message
- Added `TestIdleMemory` and `TestTTFB` perf gates (ENG-05, ENG-06)

## Task Commits

Each task was committed atomically:

1. **Task 1: Static assets handler with path jail** - `012624b` (test), `0733527` (feat)
2. **Task 2: SSR catch-all, fragments, and error pages** - `9f9fa90` (feat)
3. **Task 3: Performance gates ENG-05 and ENG-06** - `26adef6` (test)

**Plan metadata:** pending (docs commit follows)

## Files Created/Modified

- `internal/router/assets.go` — Root-jailed static file serving from workspace/assets
- `internal/router/assets_test.go` — Unit tests for safeAssetPath traversal block
- `internal/router/ssr.go` — SSR catch-all, error rendering, RegisterRoutes
- `internal/router/fragments.go` — HTMX partial handler without base layout
- `main.go` — OnServe calls RegisterRoutes with workspace cache deps
- `internal/router/handlers_test.go` — Integration tests for assets, SSR, 404, fragments, prod 500
- `internal/router/perf_test.go` — Heap and TTFB threshold tests
- `internal/testutil/workspace.go` — Hero title and errors template for SSR tests

## Decisions Made

- Path traversal blocked in `safeAssetPath` before file read; unit test covers `../../../etc/passwd` because ServeMux redirects `..` in URL paths
- SSR uses `ExecuteTemplate(..., "base", data)` with `IsLoggedIn`, `Slug`, `Path` context
- Fragment handler uses `Execute` (not base) for HTMX swap targets

## Deviations from Plan

### Auto-fixed Issues

None — plan executed as written.

### Test Infrastructure Note

Integration traversal test uses `TestSafeAssetPath` unit test instead of httptest GET with `..` segments because Go ServeMux returns 301 on canonical path cleanup before the handler runs.

## Self-Check

- FOUND: internal/router/assets.go
- FOUND: internal/router/ssr.go
- FOUND: internal/router/fragments.go
- FOUND: internal/router/assets_test.go
- FOUND: 012624b
- FOUND: 0733527
- FOUND: 9f9fa90
- FOUND: 26adef6

**Self-Check: PASSED**
