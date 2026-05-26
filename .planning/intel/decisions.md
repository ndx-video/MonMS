# Synthesized Decisions — ingest specs/staging.md

## Locked (from SPEC — accepted 2026-05-23)

- **D-50**: MonMS has four layers — engine (L1), structure (L2), content (L3), audience (L4) — with distinct actors and promotion paths.
- **D-51**: Structure (templates, schema, assets) promotes via workspace Git tags; content (editorial records) promotes via JSON upsert — independent rails.
- **D-52**: Staging and production are separate MonMS instances; each has its own `.pb_data/`.
- **D-53**: Primary content publish UX is client **Publish to live** in PocketBase admin; consultants are not in the routine content loop.
- **D-54**: Editorial records export/import via `workspace/content/*.json` with upsert by fixed record ID.
- **D-55**: Publishable media uses public CDN URLs in content fields; blobs are not copied between staging and production.
- **D-56**: Full `.pb_data/` backup/restore is not the primary content publish mechanism.
