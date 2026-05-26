---
phase: 04-staging-environments-client-content-publish
plan: "03"
subsystem: content
tags: [content-cli, pocketbase, export, import, diff, publish, operator-fallback]

requires:
  - phase: 04-staging-environments-client-content-publish
    plan: "02"
    provides: ExportAll, ImportFiles, DiffExport, CollectionPayload
provides:
  - monms content export|import|diff|publish CLI (PUB-03, PUB-04, PUB-09)
  - PublishToProduction outbound POST for plan 04-04/04-05 reuse
  - main.go early dispatch before runServe
affects:
  - 04-04 production import API
  - 04-05 staging publish UI

tech-stack:
  added: []
  patterns:
    - "Early-dispatch CLI like validate with PocketBase bootstrap unlike validate"
    - "Subcommand-first args: monms content export --workspace PATH"
    - "Diff returns ErrPendingChanges for exit code 1 without logging secrets"
    - "Publish HTTP body uses staging.md name field mapped from CollectionPayload"

key-files:
  created:
    - internal/content/cmd.go
    - internal/content/publish.go
    - internal/content/cmd_test.go
  modified:
    - main.go

key-decisions:
  - "Publish subcommand strips --workspace flags before publish FlagSet parse"
  - "MONMS_PUBLISH_TOKEN read via os.Getenv; errors never include token value (T-04-06)"
  - "Diff prints human-readable field deltas then returns ErrPendingChanges"

patterns-established:
  - "bootstrapApp: ephemeral PocketBase with RegisterBootstrapHook for all content CLI DB ops"
  - "PublishToProduction POST /api/monms/content/import with 30s timeout and Bearer auth"

requirements-completed: [PUB-03, PUB-04, PUB-09]

duration: 12min
completed: 2026-05-26
---

# Phase 04 Plan 03: Content CLI Subcommands Summary

**monms content export/import/diff/publish with ephemeral PocketBase bootstrap and production POST fallback**

## Performance

- **Duration:** 12 min
- **Started:** 2026-05-26T22:52:00Z
- **Completed:** 2026-05-26T23:04:00Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Implemented `content.RunCLI` with export, import, diff, and publish subcommands
- Export writes `workspace/content/*.json`; import is idempotent; diff exits non-zero via `ErrPendingChanges`
- `PublishToProduction` POSTs canonical export payload to production import endpoint with Bearer token
- `main.go` dispatches `content` before `runServe`, matching validate early-dispatch pattern

## Task Commits

Each task was committed atomically:

1. **Task 1: RunCLI bootstrap and export/import/diff subcommands** - `5f890ae` (test), `36927e7` (feat)
2. **Task 2: publish subcommand and main.go dispatch** - `ad0048f` (feat)

## Files Created/Modified

- `internal/content/cmd.go` - RunCLI, bootstrapApp, export/import/diff/publish handlers
- `internal/content/publish.go` - PublishToProduction HTTP client (staging.md §5.4 body shape)
- `internal/content/cmd_test.go` - CLI integration tests for export, import, diff, publish
- `main.go` - Early dispatch arm for `monms content`

## Decisions Made

- Publish flag parsing strips `--workspace` args already consumed by `config.ResolveWorkspace`
- Missing publish token fails before bootstrap or network I/O
- Diff prints pending changes to stdout; caller (main) exits 1 on `ErrPendingChanges`

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required. Production publish requires `MONMS_PUBLISH_TOKEN` at runtime (documented in plan 04-04 user setup).

## Next Phase Readiness

- Plan 04-04 can wire `POST /api/monms/content/import` and reuse `PublishToProduction` payload shape
- Plan 04-05 staging publish UI can call `DiffExport` and `PublishToProduction` from handlers

## Self-Check: PASSED

- FOUND: internal/content/cmd.go
- FOUND: internal/content/publish.go
- FOUND: internal/content/cmd_test.go
- FOUND: main.go (content dispatch)
- FOUND: commits 5f890ae, 36927e7, ad0048f

---
*Phase: 04-staging-environments-client-content-publish*
*Completed: 2026-05-26*
