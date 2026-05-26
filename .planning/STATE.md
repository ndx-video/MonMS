---
gsd_state_version: 1.0
milestone: v2.0
milestone_name: Staging & Client Content Publish
status: ready_to_execute
stopped_at: Phase 4 planned — 6 plans verified
last_updated: "2026-05-23T00:00:00.000Z"
progress:
  total_phases: 4
  completed_phases: 3
  total_plans: 18
  completed_plans: 8
  percent: 44
---

# MonMS — Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-23)

**Core value:** High-performance runtime malleability without build or compilation overhead, compiled into a single executable that uses less than 30MB of RAM.
**Current focus:** Phase 04 — staging-environments-client-content-publish

## Active Phase

**Phase 4:** Staging Environments & Client Content Publish
**Status:** Ready to execute — 6 plans in 6 waves
**Next action:** `/gsd-execute-phase 4`

## Decisions

- filepath.Clean only on flag/env workspace overrides; default `./workspace` preserved for logging
- Wave 0 skeleton tests skip with plan ownership until feature plans land
- TemplateCache production mode set via SetProductionMode from main to avoid internal importing main
- Integration tests register /_/ static UI route like apis.Serve because NewRouter alone omits admin assets
- Flat templates/{slug}.gohtml wins when both flat and directory index exist
- StartWatcher watches entire workspace tree per D-30, not templates/ only
- Skip existing scaffold files on re-init with info log (no force flag in Phase 1)
- git init warns and continues when git missing from PATH
- Traversal blocked in safeAssetPath; unit test covers ../../../ because ServeMux redirects .. in URL paths
- RegisterRoutes order: assets, fragments, SSR catch-all (D-14)
- **D-50:** Four-layer lifecycle (engine, structure, content, audience)
- **D-51:** Dual promotion rails — Git structure, JSON content
- **D-52:** Separate staging/production instances
- **D-53:** Client Publish button primary UX
- **D-54:** workspace/content/ JSON upsert by record ID
- **D-55:** CDN URL media — no blob copy
- **D-56:** No full .pb_data/ as publish path

## Last Session

**Stopped at:** Phase 4 ingested from specs/staging.md via gsd-ingest-docs merge
**Resume file:** .planning/phases/04-staging-environments-client-content-publish/04-CONTEXT.md

## Ingest

**2026-05-23:** Merged `specs/staging.md` — 14 v2 requirements, Phase 4 added. Conflict report: `.planning/INGEST-CONFLICTS.md` (0 blockers).

## Workflow State

| Setting | Value |
|---------|-------|
| Mode | YOLO |
| Granularity | Comprehensive |
| Parallelization | Enabled |
| Model Profile | Budget |
| Research | Enabled |
| Plan Check | Enabled |
| Verifier | Enabled |

---
*State updated: 2026-05-23 after ingest specs/staging.md*
