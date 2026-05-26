---
phase: 04-staging-environments-client-content-publish
plan: "04"
subsystem: content
tags: [content-api, publish-token, pocketbase, import, auth-middleware]

requires:
  - phase: 04-staging-environments-client-content-publish
    plan: "02"
    provides: ImportPayload, CollectionPayload, editorial allowlist
  - phase: 04-staging-environments-client-content-publish
    plan: "03"
    provides: PublishToProduction HTTP body shape for staging POST
provides:
  - POST /api/monms/content/import with MONMS_PUBLISH_TOKEN Bearer auth (PUB-05)
  - RequirePublishToken middleware fail-closed when token unset
  - main.go OnServe wiring before SSR catch-all
affects:
  - 04-05 staging publish UI
  - staging PublishToProduction client calls

tech-stack:
  added: []
  patterns:
    - "RequirePublishToken with crypto/subtle.ConstantTimeCompare and empty-token fail closed"
    - "Import API accepts staging.md collections[].name JSON shape"
    - "Denied system collections (_superusers, users) before editorial allowlist"
    - "maxRecordsPerCollection=1000 guard on import handler"

key-files:
  created:
    - internal/content/auth.go
    - internal/content/routes.go
    - internal/content/routes_test.go
  modified:
    - main.go

key-decisions:
  - "Import handler maps HTTP name field to CollectionPayload.Collection for ImportPayload reuse"
  - "ImportReport returns upserted/collections counts; unknown-field warnings remain slog-only from UpsertRecord"
  - "MONMS_PUBLISH_TOKEN read via os.Getenv in main.go matching content CLI publish subcommand"

patterns-established:
  - "content.RegisterRoutes called before router.RegisterRoutes in OnServe per D-14"
  - "Import API integration tests use local httptest harness without RegisterAuthHooks"

requirements-completed: [PUB-05]

duration: 8min
completed: 2026-05-26
---

# Phase 04 Plan 04: Production Import API Summary

**Scoped Bearer publish token on POST /api/monms/content/import with editorial-only upsert and fail-closed auth**

## Performance

- **Duration:** 8 min
- **Started:** 2026-05-26T23:10:00Z
- **Completed:** 2026-05-26T23:18:00Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Implemented `RequirePublishToken` middleware with constant-time compare and 401 when token unset
- Added `POST /api/monms/content/import` handler returning `ImportReport` JSON after idempotent upsert
- Rejected system collections and non-editorial names; capped records per collection at 1000
- Wired `content.RegisterRoutes` in `main.go` OnServe before SSR routes

## Task Commits

Each task was committed atomically:

1. **Task 1: Publish token middleware and import handler** - `9c2deda` (test), `1d6f7c1` (feat)
2. **Task 2: Wire RegisterRoutes in main.go OnServe** - `f6c6629` (feat)

## Files Created/Modified

- `internal/content/auth.go` - `RequirePublishToken` middleware (PUB-05)
- `internal/content/routes.go` - `RegisterRoutes`, import handler, `ImportReport`
- `internal/content/routes_test.go` - ImportAPI auth, upsert, and guard integration tests
- `main.go` - OnServe registers content routes with `MONMS_PUBLISH_TOKEN`

## Decisions Made

- HTTP request uses `collections[].name` per staging.md; mapped internally to `CollectionPayload`
- Warnings in API response omitted when UpsertRecord only logs unknown fields (counts sufficient for v1)
- Empty publish token always 401 even if client sends a Bearer value (fail closed)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

Production deploy must set `MONMS_PUBLISH_TOKEN` to a long random string. Import API returns 401 until configured.

## Next Phase Readiness

- Plan 04-05 staging publish UI can POST via existing `PublishToProduction` client to this endpoint
- Production instance needs `MONMS_PUBLISH_TOKEN` env var at deploy time

## Self-Check: PASSED

- FOUND: internal/content/auth.go
- FOUND: internal/content/routes.go
- FOUND: internal/content/routes_test.go
- FOUND: main.go (content.RegisterRoutes)
- FOUND: commits 9c2deda, 1d6f7c1, f6c6629

---
*Phase: 04-staging-environments-client-content-publish*
*Completed: 2026-05-26*
