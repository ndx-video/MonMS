# MonMS — The Agent-Malleable Monolithic CMS

## What This Is

MonMS (Monolithic Content Management System) is an agent-malleable, single-binary monolith CMS built on Go and PocketBase that treats database schemas, UI structures, and content as a fluid, singular continuum. AI agents directly mutate a Git-managed folder tree to update schemas, layouts, templates, and styles dynamically, while authenticated human editors can safely edit content visual inline using HTMX.

## Core Value

High-performance runtime malleability without build or compilation overhead, compiled into a single executable that uses less than 30MB of RAM.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] **Single-Binary Monolith**: Production-grade CMS, database, file server, and web server compiled into a single generic Go executable using < 30MB RAM.
- [ ] **Zero-Compilation Malleability**: Dynamically parsed routing, folder watching, and Go HTML templates loaded on the fly without service restart or rebuilding.
- [ ] **Git-Managed State**: Track, version, and roll back all AI structural mutations (new pages, template modifications, schema changes) using Git.
- [ ] **Visual Inline Contextual Editing**: contenteditable regions combined with HTMX PUT requests on blur event to update SQLite records seamlessly.
- [ ] **Embedded Access & Security**: PocketBase admin login integration, secure HttpOnly cookie with JWT, and database-level validation to prevent unauthorized visual updates.
- [ ] **Safety Guardrails**: Pre-commit validation including Go HTML template dry-run parser validation, HTML structure linting, and automated git checkout rollbacks on verification failure.

### Out of Scope

- **Multi-Binary/Microservice Layout**: Excluded because it violates the low-overhead monolithic requirement.
- **Local Node.js Compilation Pipeline**: Front-end styling uses native CSS or CDN Tailwind CSS imports, avoiding client-side build steps.
- **Complex Manual DB Migrations**: Managed dynamically via PocketBase's API collections endpoints or declarative schema definitions.

## Context

MonMS is designed to address the divide between developer-only environments (CI/CD pipelines, build steps) and restricted visual content management. The technical environment leverages:
- Go embedded runtime with PocketBase integration.
- fsnotify for watching file changes in the workspace folder.
- HTMX for asynchronous visual content updates without page reloads.
- Git for auditing and versioning layout/schema updates.

## Constraints

- **Tech Stack**: Must use Go + PocketBase + SQLite + HTMX.
- **Performance**: RAM usage must be under 30MB under idle production conditions. Route latency under SQLite reads must be < 15ms.
- **Security**: Strict database-level verification to validate editor PUT calls.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| SQLite Embedded Database | Embedded database within generic monolithic binary to keep RAM footprint low | — Pending |
| fsnotify Cache Invalidation | Invalidate templates cache immediately on local fs writes/git pulls without server restarts | — Pending |
| HTML5 contenteditable + HTMX | Native browser text-manipulation saving asynchronously on blur for simple visual content updates | — Pending |

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
*Last updated: 2026-05-22 after initialization*
