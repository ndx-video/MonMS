# Ingest Synthesis — specs/staging.md

**Source:** `specs/staging.md` (SPEC, accepted 2026-05-23)  
**Mode:** merge into existing `.planning/`  
**Milestone:** v2 — Phase 4

## Summary

Extends v1 with a four-layer lifecycle and dual promotion model. v1 single-instance inline editing remains valid; staging/production split and client-driven content publish shipped in Phase 4 (2026-05-26).

## Phase 4 — complete

**Phase 4: Staging Environments & Client Content Publish**

Requirements: ENV-01–03, PUB-01–09, MED-01–02 — all marked complete in `.planning/REQUIREMENTS.md`.

## PROJECT.md updates — done (2026-05-26)

- Refine "What This Is" to distinguish structure vs content layers ✓
- Add Key Decisions D-50 through D-56 ✓
- Clarify Git-managed state applies to structure (L2), not editorial content (L3) ✓

## Non-goals confirmed

- Full `.pb_data/` sync as publish path
- Consultant on every content push
- Blob replication staging → production
