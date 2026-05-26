---
phase: 04-staging-environments-client-content-publish
plan: "06"
subsystem: documentation
tags: [four-layers, staging, content-rail, media, gitignore, publish]

requires:
  - phase: 04-staging-environments-client-content-publish
    plan: "05"
    provides: Publish console at /api/monms/publish and publisher gate
provides:
  - Four-layer and dual-rail lifecycle docs (ENV-01, ENV-02, ENV-03)
  - workspace/MEDIA.md CDN policy (MED-02)
  - workspace/.gitignore for secrets and publish state
  - EDITING-GUIDE Publish to live section for clients
affects:
  - Phase 4 verification and milestone closeout

tech-stack:
  added: []
  patterns:
    - "Cross-link specs/staging.md; summarize operator steps only in workspace docs"
    - "Gitignore .monms/config.json and publish-state.json; commit config.example.json"

key-files:
  created:
    - workspace/MEDIA.md
    - workspace/.gitignore
  modified:
    - README.md
    - workspace/README.md
    - CLAUDE.md
    - workspace/EDITING-GUIDE.md

key-decisions:
  - "Publish UI documented at /api/monms/publish not /_/publish per SPA catch-all"
  - "content/ gitignored by default with note that sites may force-add for audit"

patterns-established:
  - "MEDIA.md: publishable assets = CDN URLs in text fields; export skips file columns"

requirements-completed: [ENV-01, ENV-02, ENV-03, MED-02]

duration: 8min
completed: 2026-05-26
---

# Phase 04 Plan 06: Four-Layer Lifecycle Docs Summary

**Four-layer lifecycle, dual promotion rails, CDN media policy, and workspace gitignore for operators and clients**

## Performance

- **Duration:** 8 min
- **Started:** 2026-05-26T14:00:00Z
- **Completed:** 2026-05-26T14:08:00Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Expanded `workspace/README.md` with L1–L4 table, structure vs content rails, separate `.pb_data/` per environment, and D-56 note (no wholesale `.pb_data/` publish)
- Added root README **Content CLI (v2)** section covering `monms content` subcommands and `POST /api/monms/content/import`
- Created `workspace/MEDIA.md` warning against PocketBase-local file storage for publishable assets
- Extended `EDITING-GUIDE.md` with **Publish to live** workflow for publisher role
- Added `workspace/.gitignore` excluding `.pb_data/`, `.monms/config.json`, `publish-state.json`, and `content/`

## Task Commits

Each task was committed atomically:

1. **Task 1: Four-layer and dual-rail documentation** - `4283904` (docs)
2. **Task 2: MEDIA.md and publish operator guide** - `9e48ba1` (docs)
3. **Task 3: Workspace gitignore for secrets and publish state** - `b66bf9d` (chore)

## Files Created/Modified

- `README.md` - Content CLI table, v2 implemented status, MEDIA cross-link
- `workspace/README.md` - Expanded four layers, dual rails, staging/prod, gitignore table
- `CLAUDE.md` - Phase 4 implemented notes, monms content command, .monms gitignore
- `workspace/MEDIA.md` - CDN URL policy, export file-field warning, consultant setup
- `workspace/EDITING-GUIDE.md` - Publish to live section with publisher prerequisites
- `workspace/.gitignore` - Runtime data, secrets, ephemeral content exports

## Decisions Made

- Document publish UI at `/api/monms/publish` (not `/_/publish`) consistent with implemented routes
- Default-ignore `content/` with documented escape hatch for audit (`git add -f`)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - documentation only. Operators follow existing 04-05 setup for `MONMS_PUBLISH_TOKEN` and `.monms/config.json`.

## Next Phase Readiness

- Phase 4 documentation requirements (ENV-01–03, MED-02) satisfied
- Ready for phase verification / milestone audit

## Self-Check: PASSED

- FOUND: workspace/MEDIA.md
- FOUND: workspace/.gitignore
- FOUND: commits 4283904, 9e48ba1, b66bf9d

---
*Phase: 04-staging-environments-client-content-publish*
*Completed: 2026-05-26*
