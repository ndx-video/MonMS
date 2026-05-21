---
phase: 01-core-go-runtime-workspace-foundation
plan: 02
subsystem: api
tags: [go, pocketbase, schema-sync, template-cache, serve-bootstrap]

requires:
  - phase: 01-01
    provides: ResolveWorkspace, ValidateWorkspace, testutil.NewWorkspace, buildMode ldflags
provides:
  - PocketBase serve bootstrap with workspace/.pb_data and early init dispatch
  - TemplateCache skeleton with production gate via SetProductionMode
  - Schema sync from workspace/schema/*.json on OnBootstrap
  - ENG-01/SEC-01 integration smoke tests for health and admin routes
affects: [01-03, 01-04, 01-05]

tech-stack:
  added: []
  patterns: [early init dispatch before PocketBase, DefaultDev aligned with buildMode, schema log-and-continue import, TemplateCache Get/Flush with RWMutex]

key-files:
  created:
    - internal/scaffold/init.go
    - internal/schema/sync.go
    - internal/templates/cache.go
  modified:
    - main.go
    - go.mod
    - go.sum
    - internal/templates/cache_test.go
    - internal/router/handlers_test.go

key-decisions:
  - "TemplateCache production mode set via SetProductionMode from main to avoid internal importing main"
  - "Integration tests register /_/ static UI route like apis.Serve because NewRouter alone omits admin assets"

patterns-established:
  - "runServe resolves workspace, validates, logs configured+absolute paths, then NewWithConfig with workspace/.pb_data"
  - "RegisterBootstrapHook runs after e.Next() and skips empty schema directories per Pitfall 6"

requirements-completed: [ENG-01, SEC-01]

duration: 2min
completed: 2026-05-22
---

# Phase 1 Plan 02: PocketBase Serve Bootstrap Summary

**Embedded PocketBase serve with workspace-scoped pb_data, declarative schema bootstrap, TemplateCache skeleton, and automated health/admin smoke tests**

## Performance

- **Duration:** 2 min
- **Started:** 2026-05-22T01:51:00+10:00
- **Completed:** 2026-05-22T02:00:00+10:00
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Wired `main.go` with early `init` dispatch, `runServe()`, workspace validation, structured startup logging, and PocketBase `NewWithConfig` pointing at `workspace/.pb_data`
- Implemented `TemplateCache` with dev no-cache and production flush semantics plus `schema/sync.go` bootstrap import from sorted `schema/*.json`
- Replaced skipped tests with passing `TestDevNoCache`, `TestCacheFlush`, `TestServeStarts`, and `TestAdminDashboard`

## Task Commits

Each task was committed atomically:

1. **Task 1: main.go CLI dispatch and PocketBase serve bootstrap** - `ed8d2e0` (feat)
2. **Task 2: TemplateCache skeleton and schema bootstrap hook** - `9083435` (feat)
3. **Task 3: Prove serve starts and admin route exists** - `9ec39c7` (test)

**Plan metadata:** pending (docs commit follows)

## Files Created/Modified

- `main.go` — Early init dispatch, runServe, OnServe stub, global tplCache
- `internal/scaffold/init.go` — Stub RunInit until plan 01-05
- `internal/schema/sync.go` — OnBootstrap schema JSON merge and import
- `internal/templates/cache.go` — RWMutex cache with production gate
- `internal/templates/cache_test.go` — Dev no-cache and flush unit tests
- `internal/router/handlers_test.go` — Serve health and admin dashboard integration tests
- `go.mod` / `go.sum` — Direct PocketBase dependency for main package

## Decisions Made

- Use `SetProductionMode` on TemplateCache instead of importing `main.buildMode` from internal packages
- Integration tests manually register `/_/{path...}` static route because `apis.NewRouter` does not include embedded admin UI (only `apis.Serve` does)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Register admin UI static route in integration test harness**
- **Found during:** Task 3 (TestAdminDashboard)
- **Issue:** `GET /_/` returned 404 when using `apis.NewRouter` alone; admin assets are registered in `apis.Serve`, not `NewRouter`
- **Fix:** Register `router.GET("/_/{path...}", apis.Static(ui.DistDirFS, false))` in test helper before BuildMux
- **Files modified:** internal/router/handlers_test.go
- **Verification:** `go test ./internal/router/... -run 'TestAdminDashboard' -count=1` passes
- **Committed in:** 9ec39c7

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Test harness mirrors production Serve behavior for SEC-01 verification. No production code scope change.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 01-03 can wire fsnotify watcher against tplCache.Flush in production mode
- Plan 01-04 can register SSR/assets/fragments routes in OnServe using tplCache
- Schema sync and empty schema directory behavior verified; invalid JSON log-and-continue implemented

## Self-Check: PASSED

- All key files FOUND
- Commits ed8d2e0, 9083435, 9ec39c7 verified
- `go test ./internal/router/... -run 'TestServeStarts|TestAdminDashboard' -count=1` green
- `go test ./internal/templates/... -count=1 -short` green
- `go build -o /dev/null .` passes

---
*Phase: 01-core-go-runtime-workspace-foundation*
*Completed: 2026-05-22*
