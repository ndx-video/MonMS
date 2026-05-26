# Phase 4 Context: Staging Environments & Client Content Publish

**Phase:** 04-staging-environments-client-content-publish  
**Milestone:** v2  
**Source spec:** `specs/staging.md` (accepted 2026-05-23)  
**Depends on:** Phase 3 complete

## Goal

Clients publish editorial content from staging to production via PocketBase admin **Publish to live** button. Structure promotes via workspace Git tags (consultant-driven). Media uses public CDN URLs — no blob copy between environments.

## Locked decisions

<decisions>

- **D-50:** Four layers — engine (L1), structure (L2), content (L3), audience (L4).
- **D-51:** Dual promotion rails — Git tag for structure; JSON upsert for content.
- **D-52:** Staging and production are separate MonMS instances with separate `.pb_data/`.
- **D-53:** Primary publish UX is client **Publish to live** in admin; consultants not in routine content loop.
- **D-54:** Editorial records in `workspace/content/*.json`; upsert by fixed record ID.
- **D-55:** Publishable media = public CDN URLs in content fields; blobs do not move.
- **D-56:** Full `.pb_data/` backup/restore is not the primary publish mechanism.

</decisions>

## Requirements

ENV-01–03, PUB-01–09, MED-01–02 — see `.planning/REQUIREMENTS.md`

## Key deliverables

1. `workspace/content/` + `editorial: true` on schema JSON
2. `internal/content/` package (export, import, diff, upsert)
3. `monms content` CLI
4. `POST /api/monms/content/import` + publish token
5. Admin `/_/publish` page (diff + publish)
6. Publisher role

## Non-goals

- Whole-database sync between environments
- Consultant required for every content push
- Blob replication staging → production

## Next action

Run `/gsd-plan-phase 4` or `/gsd-import --from specs/staging.md` to generate execution plans.
