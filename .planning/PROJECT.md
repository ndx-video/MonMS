# MonMS — The Agent-Malleable Monolithic CMS

## What This Is

MonMS (Monolithic Content Management System) is an agent-malleable, single-binary monolith CMS built on Go and PocketBase. It operates across **four layers** — engine, structure, content, and audience — with **two promotion rails**: consultants and AI agents shape site structure in a Git-tracked workspace; business clients edit copy on staging and publish to production themselves via an admin button.

AI agents mutate the workspace folder tree (templates, schema, assets) with validation and Git versioning. Human editors use HTMX inline editing on staging. Structure deploys by Git tag; editorial content syncs by JSON upsert — not by copying `.pb_data/`.

## Core Value

High-performance runtime malleability without build or compilation overhead, compiled into a single executable that uses less than 30MB of RAM.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] **Single-Binary Monolith**: Production-grade CMS, database, file server, and web server compiled into a single generic Go executable using < 30MB RAM.
- [ ] **Zero-Compilation Malleability**: Dynamically parsed routing, folder watching, and Go HTML templates loaded on the fly without service restart or rebuilding.
- [ ] **Git-Managed Structure**: Track, version, and roll back structural mutations (pages, templates, schema) using workspace Git tags.
- [ ] **Client Content Publish**: Editorial records promote from staging to production via JSON upsert; clients use Publish to live in admin — consultants not required for routine pushes.
- [ ] **Visual Inline Contextual Editing**: contenteditable regions combined with HTMX PUT requests on blur event to update SQLite records seamlessly.
- [ ] **Embedded Access & Security**: PocketBase admin login integration, secure HttpOnly cookie with JWT, and database-level validation to prevent unauthorized visual updates.
- [ ] **Safety Guardrails**: Pre-commit validation including Go HTML template dry-run parser validation, HTML structure linting, and automated git checkout rollbacks on verification failure.

### Out of Scope

- **Multi-Binary/Microservice Layout**: Excluded because it violates the low-overhead monolithic requirement.
- **Local Node.js Compilation Pipeline**: Front-end styling uses native CSS or CDN Tailwind CSS imports, avoiding client-side build steps.
- **Complex Manual DB Migrations**: Managed dynamically via PocketBase's API collections endpoints or declarative schema definitions.
- **Full `.pb_data/` Sync as Publish Path**: Editorial content uses JSON upsert rail, not whole-database backup/restore.

## Context

MonMS addresses the divide between developer-only environments (CI/CD pipelines, build steps) and restricted visual content management. The technical environment leverages:

- Go embedded runtime with PocketBase integration.
- fsnotify for watching file changes in the workspace folder.
- HTMX for asynchronous visual content updates without page reloads.
- Git for auditing and versioning **structure** (templates, schema, assets).
- JSON upsert for **content** promotion between staging and production.
- Public CDN URLs for publishable media (blobs do not move between environments).

**Lifecycle spec:** `specs/staging.md` (accepted 2026-05-23)

## Constraints

- **Tech Stack**: Must use Go + PocketBase + SQLite + HTMX.
- **Performance**: RAM usage must be under 30MB under idle production conditions. Route latency under SQLite reads must be < 15ms.
- **Security**: Strict database-level verification to validate editor PUT calls. Content import API uses scoped publish token — not full superuser.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| SQLite Embedded Database | Embedded database within generic monolithic binary to keep RAM footprint low | — Pending |
| fsnotify Cache Invalidation | Invalidate templates cache immediately on local fs writes/git pulls without server restarts | — Pending |
| HTML5 contenteditable + HTMX | Native browser text-manipulation saving asynchronously on blur for simple visual content updates | — Pending |
| Four-layer lifecycle (D-50) | Engine, structure, content, audience have distinct actors and artifacts | Accepted — specs/staging.md |
| Dual promotion rails (D-51) | Git tags for structure; JSON upsert for content — independent paths | Accepted — specs/staging.md |
| Client Publish button (D-53) | Clients self-serve content to production; consultants not in routine loop | Accepted — Phase 4 |
| CDN URL media policy (D-55) | Publishable media referenced by URL; no blob copy staging→prod | Accepted — specs/staging.md |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-23 after ingest specs/staging.md*
