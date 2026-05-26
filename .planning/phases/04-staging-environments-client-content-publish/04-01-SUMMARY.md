---
phase: 04-staging-environments-client-content-publish
plan: "01"
subsystem: schema
tags: [editorial, schema-json, pocketbase, test-fixtures, content-publish]

requires: []
provides:
  - LoadEditorialCollectionNames parser reading raw schema JSON per D-54
  - hero_content marked editorial in workspace and scaffold embed
  - NewEditorialWorkspace test helper with staging config
  - workspace/.monms/config.example.json staging template
affects:
  - 04-02 content export/import plans
  - 04-04 production import allowlist

tech-stack:
  added: []
  patterns:
    - "Parse MonMS-specific schema metadata from raw JSON files, not PocketBase Collection model"
    - "Editorial test fixtures via NewEditorialWorkspace temp workspace"

key-files:
  created:
    - internal/schema/editorial.go
    - internal/schema/editorial_test.go
    - internal/testutil/content.go
    - internal/testutil/content_test.go
    - workspace/.monms/config.example.json
  modified:
    - workspace/schema/hero_content.json
    - internal/scaffold/embed/hero_content.json

key-decisions:
  - "Editorial flag read only from workspace/schema/*.json because ImportCollections strips unknown keys (D-54 / Pitfall 1)"
  - "config.example.json uses _comment field for gitignore documentation since JSON has no file comments"

patterns-established:
  - "LoadEditorialCollectionNames: sorted dir scan, continue-on-error, nil slice for missing dir"
  - "NewEditorialWorkspace: temp dir with editorial hero_content + .monms/config.json for downstream content tests"

requirements-completed: [PUB-01]

duration: 8min
completed: 2026-05-26
---

# Phase 04 Plan 01: Editorial Schema Parser Summary

**Editorial collection discovery from raw schema JSON with hero_content flagged and Wave 0 test fixtures for content publish plans**

## Performance

- **Duration:** 8 min
- **Started:** 2026-05-26T20:30:00Z
- **Completed:** 2026-05-26T20:38:00Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Implemented `LoadEditorialCollectionNames` parsing `editorial: true` from workspace schema JSON files per D-54
- Marked `hero_content` as editorial in both workspace and scaffold embed (dual-write)
- Added `NewEditorialWorkspace` test helper and committed `config.example.json` staging template

## Task Commits

Each task was committed atomically:

1. **Task 1: Editorial flag parser from schema JSON** - `2c2f7e8` (test RED) + `6cdf28c` (feat GREEN)
2. **Task 2: hero_content editorial flag in schema fixtures** - `d82a74d` (feat)
3. **Task 3: Editorial test fixtures and config template** - `d6b09aa` (feat)

## Files Created/Modified

- `internal/schema/editorial.go` - SchemaMeta struct and LoadEditorialCollectionNames dir scanner
- `internal/schema/editorial_test.go` - Unit tests for allowlist, missing dir, malformed skip
- `workspace/schema/hero_content.json` - Added `"editorial": true`
- `internal/scaffold/embed/hero_content.json` - Scaffold mirror of editorial flag
- `internal/testutil/content.go` - NewEditorialWorkspace helper
- `internal/testutil/content_test.go` - ValidateWorkspace integration test
- `workspace/.monms/config.example.json` - Staging config template with productionUrl and publisherEmails

## Decisions Made

- Editorial metadata must come from raw JSON files, not PocketBase collections, because import strips unknown keys
- Used `_comment` key in config.example.json to document gitignored live config paths (PUB-08 setup)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed unused workspace import from content.go**
- **Found during:** Task 3
- **Issue:** ValidateWorkspace import placed in content.go but only used in content_test.go, breaking build
- **Fix:** Moved validation to test file only
- **Files modified:** internal/testutil/content.go
- **Committed in:** d6b09aa

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Minor compile fix; no scope change.

## Issues Encountered

None beyond the unused-import compile error resolved inline.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 04-02 can use `LoadEditorialCollectionNames` and `NewEditorialWorkspace` for content export/import tests
- hero_content editorial flag available in workspace and init scaffold

## Self-Check: PASSED

- FOUND: internal/schema/editorial.go
- FOUND: internal/schema/editorial_test.go
- FOUND: internal/testutil/content.go
- FOUND: workspace/.monms/config.example.json
- FOUND: editorial in hero_content
- FOUND: 2c2f7e8, 6cdf28c, d82a74d, d6b09aa

---
*Phase: 04-staging-environments-client-content-publish*
*Completed: 2026-05-26*
