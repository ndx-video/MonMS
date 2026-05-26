---
gsd_state_version: 1.0
milestone: v2.0
milestone_name: Staging & Client Content Publish
status: executing
stopped_at: Completed 04-05-PLAN.md
last_updated: "2026-05-26T13:22:00Z"
progress:
  total_phases: 4
  completed_phases: 3
  total_plans: 18
  completed_plans: 17
  percent: 94
---

# MonMS — Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-23)

**Core value:** High-performance runtime malleability without build or compilation overhead, compiled into a single executable that uses less than 30MB of RAM.
**Current focus:** Phase 04 — staging-environments-client-content-publish

## Active Phase

**Phase 4:** Staging Environments & Client Content Publish
**Status:** Executing Phase 04
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
- Editorial flag parsed from raw workspace/schema/*.json only — PocketBase ImportCollections strips unknown keys (D-54 Pitfall 1)
- Content diff baseline uses workspace/content/*.json on disk when publish-state checksum is stale
- Export strips PocketBase file-type columns; publishable media uses CDN URLs in text fields (MED-01 / D-55)
- Publish subcommand strips --workspace before publish FlagSet parse
- MONMS_PUBLISH_TOKEN via os.Getenv; CLI errors never log token value (T-04-06)
- Diff returns ErrPendingChanges for monms content diff exit code 1
- Import handler maps HTTP collections[].name to CollectionPayload for ImportPayload reuse
- content.RegisterRoutes before router.RegisterRoutes in OnServe for /api/monms/* (D-14)
- RequirePublishToken fails closed when MONMS_PUBLISH_TOKEN unset on production

## Last Session

**Stopped at:** Completed 04-04-PLAN.md
**Resume file:** None

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
*State updated: 2026-05-26 after 04-04 plan execution*
