---
phase: 01-core-go-runtime-workspace-foundation
plan: 03
subsystem: templates
tags: [go, html-template, fsnotify, slug-resolver, cache]

requires:
  - phase: 01-02
    provides: TemplateCache skeleton, SetProductionMode wiring, testutil.NewWorkspace
provides:
  - ResolveSlug mirror+index algorithm with ErrNotFound
  - Production-only recursive fsnotify watcher with 100ms debounce
  - Verified dev vs production cache semantics (ENG-02, ENG-04)
affects: [01-04, 01-05]

tech-stack:
  added: [github.com/fsnotify/fsnotify]
  patterns: [mirror+index slug resolution, debounced whole-workspace .gohtml watch, flat-file-wins-over-index]

key-files:
  created:
    - internal/templates/resolver.go
    - internal/templates/watcher.go
  modified:
    - internal/templates/resolver_test.go
    - internal/templates/watcher_test.go
    - main.go

key-decisions:
  - "Flat templates/{slug}.gohtml wins when both flat and directory index exist"
  - "StartWatcher watches entire workspace tree per D-30, not templates/ only"
  - "Watcher started from main only when buildMode == production (D-04)"

patterns-established:
  - "ResolveSlug: strip trailing slash, empty → index.gohtml, flat before index lookup"
  - "StartWatcher: WalkDir all dirs, re-Add on Create dir, 100ms debounce per .gohtml path"

requirements-completed: [ENG-02, ENG-03, ENG-04]

duration: 8min
completed: 2026-05-22
---

# Phase 1 Plan 03: Template Cache, Watcher, and Slug Resolver Summary

**Mirror+index slug resolver, production-only debounced fsnotify invalidation, and verified dev-no-cache / production-cache semantics**

## Performance

- **Duration:** 8 min
- **Started:** 2026-05-22T02:05:00+10:00
- **Completed:** 2026-05-22T02:13:00+10:00
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments

- Implemented `ResolveSlug` with D-10/D-11/D-12/D-13 behavior and table-driven unit tests
- Verified `TemplateCache` dev-no-cache and production flush semantics (implemented in plan 01-02)
- Added recursive workspace `StartWatcher` with 100ms debounce and wired it in `main.go` for production builds only

## Task Commits

Each task was committed atomically:

1. **Task 1: Slug resolver mirror+index algorithm** - `e24f76e` (test), `ccd62bb` (feat)
2. **Task 2: Template cache dev vs production semantics** - verified from `9083435` (01-02); no new commit (tests already green)
3. **Task 3: Production fsnotify watcher and main wiring** - `c60c16e` (test), `577640d` (feat)

**Plan metadata:** pending (docs commit follows)

## Files Created/Modified

- `internal/templates/resolver.go` — ResolveSlug mirror+index algorithm, ErrNotFound
- `internal/templates/resolver_test.go` — Table-driven tests for all D-10–D-13 cases
- `internal/templates/watcher.go` — StartWatcher with recursive walk, debounce, dir re-add
- `internal/templates/watcher_test.go` — Invalidation, non-.gohtml ignore, new subdirectory tests
- `main.go` — Production-only watcher startup calling tplCache.Flush

## Decisions Made

- Flat file wins when both `templates/{slug}.gohtml` and `templates/{slug}/index.gohtml` exist
- Watcher scope is entire workspace (D-30), filtering events to `.gohtml` suffix only
- Cache production mode remains controlled via SetProductionMode from main, not ENV

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 01-04 can call ResolveSlug + tplCache.Get in SSR and fragment handlers
- Production builds invalidate cache on agent .gohtml commits within debounce window
- WRK-04 slug resolution logic complete and tested

## Self-Check: PASSED

- FOUND: internal/templates/resolver.go
- FOUND: internal/templates/watcher.go
- FOUND: internal/templates/cache.go
- FOUND: commits e24f76e, ccd62bb, c60c16e, 577640d
- `go test ./internal/templates/... -count=1 -timeout 30s` green
- `go build -ldflags "-X main.buildMode=production" -o /tmp/monms .` succeeds

---
*Phase: 01-core-go-runtime-workspace-foundation*
*Completed: 2026-05-22*
