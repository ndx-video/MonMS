# Ingest Conflict Report — specs/staging.md

**Ingested:** 2026-05-23  
**Mode:** merge  
**Source:** `specs/staging.md` (SPEC)

## Auto-resolved

### Git-managed state scope (SPEC > PROJECT narrative)

- **Found:** SPEC separates structure (Git tag) from content (JSON upsert); `.pb_data/` never in Git.
- **Existing:** PROJECT.md Active requirement "Git-Managed State" implied all mutations including content.
- **Resolution:** Refined to **Git-managed structure** (L2). Editorial content (L3) uses content rail. No contradiction with WRK-02 (workspace Git for agent mutations).

### Single-environment v1 vs dual-environment v2

- **Found:** SPEC defines staging + production instances.
- **Existing:** v1 phases assumed single host for dev/demo.
- **Resolution:** v1 remains complete on single instance; ENV-03 applies to Phase 4 / v2 only. Extension, not replacement.

## Competing variants (approved — proceed)

### RICH-02 vs MED-02 (image upload approach)

- **Found (existing v2 backlog):** RICH-02 — image drag-and-drop via PocketBase file fields.
- **Found (ingested SPEC):** MED-02 — avoid PocketBase-local file storage for publishable assets; use CDN URLs.
- **Impact:** RICH-02 implementation must upload to shared public bucket and store URL in content field — not `.pb_data/storage/` as publish source.
- **Action:** Annotate RICH-02 in REQUIREMENTS.md with MED-02 constraint when scheduled.

## Informational

### New milestone phase

- Phase 4 added to ROADMAP under Milestone 2 (v2).
- 14 new requirements (ENV-*, PUB-*, MED-*) appended to REQUIREMENTS.md traceability.

## Blockers

None.
