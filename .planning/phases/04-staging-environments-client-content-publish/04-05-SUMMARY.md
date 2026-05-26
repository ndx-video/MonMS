---
phase: 04-staging-environments-client-content-publish
plan: "05"
subsystem: content
tags: [publish-ui, publisher-gate, staging, htmx, pocketbase, diff-preview]

requires:
  - phase: 04-staging-environments-client-content-publish
    plan: "03"
    provides: PublishToProduction, DiffExport, WritePublishState
  - phase: 04-staging-environments-client-content-publish
    plan: "04"
    provides: Production import API for outbound publish POST
provides:
  - GET/POST /api/monms/publish HTML console with diff preview (PUB-06)
  - Publisher email allowlist on publish routes (PUB-07)
  - Publish status and last-published display from publish-state.json (PUB-08)
  - Editor badge Publish to live link for publishers only
affects:
  - 04-06 remaining phase 4 plans

tech-stack:
  added: []
  patterns:
    - "requirePublisherFromWorkspace reloads allowlist per request from .monms/config.json"
    - "Publish page server-rendered via go:embed publish.gohtml and main.css"
    - "SSR IsPublisher from content.LoadMonmsConfig for badge gating"

key-files:
  created:
    - internal/content/publish_handlers.go
    - internal/content/embed/publish.gohtml
    - internal/router/publish_badge_test.go
  modified:
    - internal/content/auth.go
    - internal/content/routes.go
    - internal/content/routes_test.go
    - internal/router/ssr.go
    - workspace/templates/layouts/base.gohtml
    - internal/scaffold/embed/base.gohtml

key-decisions:
  - "GET publish page requires publisher allowlist (not read-only for editors) per plan task 1"
  - "POST publish returns HTML error page on production failure without updating checksum"
  - "Diff grouped by collection in template via parseDiffLine on DiffExport change strings"

patterns-established:
  - "Staging publish routes bind apis.RequireSuperuserAuth then workspace-scoped RequirePublisher"
  - "slog.Info content publish attempt with email and outcome for repudiation acceptance (T-04-14)"

requirements-completed: [PUB-06, PUB-07, PUB-08]

duration: 12min
completed: 2026-05-26
---

# Phase 04 Plan 05: Staging Publish UI Summary

**Client Publish to live console at /api/monms/publish with diff preview, publisher-only POST gate, and editor-badge discoverability**

## Performance

- **Duration:** 12 min
- **Started:** 2026-05-26T13:10:41Z
- **Completed:** 2026-05-26T13:22:00Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- Implemented `LoadMonmsConfig`, `RequirePublisher`, and `IsPublisher` for workspace `.monms/config.json` allowlist
- Added GET/POST `/api/monms/publish` and GET `/api/monms/publish/diff` with superuser + publisher auth
- Server-rendered publish page per 04-UI-SPEC: status labels, semantic diff list, setup mode when production URL unset
- Successful publish POST updates `publish-state.json` checksum only after production import succeeds
- Editor badge shows **Publish to live** link only for allowlisted publisher emails

## Task Commits

Each task was committed atomically:

1. **Task 1: Publisher gate and staging publish routes** - `6254885` (test), `6a0d83f` (feat)
2. **Task 2: Publish HTML template per UI-SPEC** - included in `6a0d83f` (same wave as routes; template required for passing tests)
3. **Task 3: Editor badge Publish link for publishers** - `ec9f112` (feat)

## Files Created/Modified

- `internal/content/auth.go` - MonmsConfig load, RequirePublisher, IsPublisher
- `internal/content/publish_handlers.go` - Publish page/diff/POST handlers
- `internal/content/embed/publish.gohtml` - Publish console HTML per UI-SPEC
- `internal/content/routes.go` - registerPublishRoutes
- `internal/content/routes_test.go` - PublishUI and PublisherGate integration tests
- `internal/router/ssr.go` - IsPublisher in enrichSSRData
- `internal/router/publish_badge_test.go` - Badge visibility tests
- `workspace/templates/layouts/base.gohtml` - Conditional publish link in editor badge
- `internal/scaffold/embed/base.gohtml` - Scaffold parity for publish link

## Decisions Made

- Publisher-only GET on publish routes (plan task 1); non-publishers use inline edit only and do not see badge link
- Dynamic publisher allowlist reload per request via `requirePublisherFromWorkspace`
- HTML error response on failed production POST preserves checksum (04-UI-SPEC)

## Deviations from Plan

None - plan executed exactly as written. Task 2 HTML template shipped in Task 1 feat commit because routes tests require rendered page.

## Issues Encountered

None.

## User Setup Required

Staging deploy needs:

- `MONMS_PUBLISH_TOKEN` env var matching production
- `workspace/.monms/config.json` with `productionUrl` and `publisherEmails` (copy from `config.example.json`)

## Next Phase Readiness

- Plan 04-06 can build on live publish console and badge entry point
- Manual UAT: login as publisher → `/api/monms/publish` → verify diff → Publish now against production with token set

## Self-Check: PASSED

- FOUND: internal/content/embed/publish.gohtml
- FOUND: internal/content/publish_handlers.go
- FOUND: internal/router/publish_badge_test.go
- FOUND: commits 6254885, 6a0d83f, ec9f112

---
*Phase: 04-staging-environments-client-content-publish*
*Completed: 2026-05-26*
