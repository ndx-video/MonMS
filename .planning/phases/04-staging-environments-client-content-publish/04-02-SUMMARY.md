---
phase: 04-staging-environments-client-content-publish
plan: "02"
subsystem: content
tags: [content-rail, pocketbase, export, import, checksum, publish-state, editorial]

requires:
  - phase: 04-staging-environments-client-content-publish
    plan: "01"
    provides: LoadEditorialCollectionNames and NewEditorialWorkspace fixtures
provides:
  - Editorial export to workspace/content/*.json (PUB-02)
  - Idempotent import by record ID with editorial allowlist (PUB-04)
  - Canonical checksum and publish-state.json tracking (PUB-08)
  - DiffExport pending-change detection vs publish state (PUB-09)
  - File-field skip on export with slog.Warn (MED-01)
affects:
  - 04-03 content CLI
  - 04-04 production import API
  - 04-05 publish UI

tech-stack:
  added: []
  patterns:
    - "CollectionFile JSON shape {collection, records[]} per specs/staging.md §5.1"
    - "UpsertRecord mirrors seed.go FindRecordById + Save idempotency"
    - "Diff baseline from workspace/content/*.json when checksum differs from publish-state"
    - "ensureUnderWorkspace path guard on all content and .monms state paths"

key-files:
  created:
    - internal/content/export.go
    - internal/content/import.go
    - internal/content/checksum.go
    - internal/content/diff.go
    - internal/content/state.go
    - internal/content/schema.go
    - internal/content/path.go
    - internal/content/testdata/hero_content.example.json
  modified: []

key-decisions:
  - "Field-level diff uses on-disk content/ exports as baseline when publish-state checksum is stale"
  - "File-type PocketBase columns stripped from PublicExport; CDN URLs remain in text fields (D-55)"

patterns-established:
  - "ExportSnapshot + ChecksumExport for stable sha256:hex publish fingerprints"
  - "ImportPayload accepts []CollectionPayload for HTTP reuse in plan 04-04"

requirements-completed: [PUB-01, PUB-02, PUB-04, PUB-08, PUB-09, MED-01]

duration: 15min
completed: 2026-05-26
---

# Phase 04 Plan 02: Core Content Rail Engine Summary

**Editorial JSON export/import with stable checksums, publish-state tracking, and field-level diff — no CLI or HTTP yet**

## Performance

- **Duration:** 15 min
- **Started:** 2026-05-26T21:00:00Z
- **Completed:** 2026-05-26T21:15:00Z
- **Tasks:** 3
- **Files modified:** 12

## Accomplishments

- Implemented `internal/content` package: export, import, checksum, diff, and publish-state persistence
- Export writes `workspace/content/{collection}.json` with editorial allowlist and skips file-type fields (MED-01)
- Double `ImportFiles` is idempotent; non-editorial collections rejected at import boundary (T-04-04)
- `DiffExport` detects title changes when publish-state checksum is stale

## Task Commits

Each task was committed atomically:

1. **Task 1: Export and import with editorial allowlist** - `3a7e166` (feat)
2. **Task 2: Checksum, diff, and publish state** - `0712a4e` (feat)
3. **Task 3: File-field skip integration test (MED-01)** - `aabe672` (test)

## Files Created/Modified

- `internal/content/export.go` - ExportCollection, ExportAll, ExportSnapshot, LoadContentFiles
- `internal/content/import.go` - UpsertRecord, ImportFiles, ImportPayload
- `internal/content/checksum.go` - ChecksumExport with sorted canonical JSON
- `internal/content/diff.go` - DiffExport and field-level delta strings
- `internal/content/state.go` - ReadPublishState, WritePublishState under `.monms/`
- `internal/content/schema.go` - Editorial allowlist wrapper
- `internal/content/path.go` - Workspace path traversal guard
- `internal/content/testdata/hero_content.example.json` - PUB-02 shape reference
- `internal/content/*_test.go` - Export/import, diff/state, MED-01 coverage

## Decisions Made

- Diff field deltas compare live export against `workspace/content/` files on disk when checksum differs (baseline from last export)
- MED-01 test omits setting file field values (PocketBase rejects invalid file names without real uploads)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 04-03 can wire `monms content` CLI to ExportAll, ImportFiles, DiffExport
- Plan 04-04 can reuse ImportPayload for production HTTP import
- Plan 04-05 can call DiffExport and WritePublishState from publish handlers

## Self-Check: PASSED

- FOUND: internal/content/export.go
- FOUND: internal/content/import.go
- FOUND: internal/content/checksum.go
- FOUND: internal/content/diff.go
- FOUND: internal/content/state.go
- FOUND: internal/content/testdata/hero_content.example.json
- FOUND: commits 3a7e166, 0712a4e, aabe672 (git log)

---
*Phase: 04-staging-environments-client-content-publish*
*Completed: 2026-05-26*
