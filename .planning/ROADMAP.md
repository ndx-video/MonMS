# Roadmap: MonMS — The Agent-Malleable Monolithic CMS

**Milestone 1:** v1 — Working Malleable Monolith
**Goal:** A production-grade single-binary CMS with Git-tracked workspace, dynamic template serving, AI mutation engine with safety guardrails, and inline contextual editing for authenticated users.

---

## Phase 1: Core Go Runtime & Workspace Foundation

**Goal:** A working Go + PocketBase binary that serves templates from the workspace folder, supports dynamic cache invalidation via fsnotify, and provides static asset serving with a correct base layout.

**Requirements Covered:** ENG-01, ENG-02, ENG-03, ENG-04, ENG-05, ENG-06, WRK-01, WRK-02, WRK-03, WRK-04, WRK-05, SEC-01, DEMO-03

**Deliverables:**
- `main.go` with PocketBase embedded, catch-all `/{slug...}` SSR route, and `/assets/{path...}` static route
- `TemplateCache` struct with `sync.RWMutex`, production/dev mode toggle via `ENV` env var
- `watchWorkspace()` goroutine using fsnotify to clear template cache on write/create
- `workspace/` Git repository with `templates/layouts/base.gohtml`, `templates/fragments/`, `assets/main.css`, and `schema/` directories
- `workspace/templates/layouts/base.gohtml` with HTMX + Alpine.js CDN script tags, auth badge block, and `{{template "body" .}}` render slot
- 404 handling for unknown slugs (no Go panics)
- Go module setup with `go.mod`, PocketBase and fsnotify dependencies

**Plans:** 5 plans

Plans:
**Wave 1**
- [x] 01-01-PLAN.md — Wave 0: Go module, config, validation, test infrastructure
- [x] 01-02-PLAN.md — Wave 1: PocketBase bootstrap, schema sync, TemplateCache skeleton

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 01-03-PLAN.md — Wave 1: Template cache, fsnotify watcher, slug resolver
- [x] 01-05-PLAN.md — Wave 2: monms init scaffold and UI-SPEC templates

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 01-04-PLAN.md — Wave 3: SSR, assets, fragments, error pages, perf gates

**Status:** Pending

---

## Phase 2: Agent Mutation Engine & Safety Guardrails

**Goal:** Enable an AI agent to safely mutate the workspace — creating PocketBase collections, modifying templates — with pre-commit validation and automatic git rollback on failure.

**Requirements Covered:** AGT-01, AGT-02, AGT-03, AGT-04, AGT-05, AGT-06, SEC-03

**Deliverables:**
- Agent Mutation workflow documentation in `workspace/agent-guide.md` covering: PocketBase collection creation via REST API, template editing conventions
- Go HTML template dry-run validator: a standalone script/Go test that parses modified `*.gohtml` files against the layout without starting the full server
- HTML linter integration (e.g., invoking `html-validator` or equivalent) as a pre-commit hook
- Git pre-commit hook script in `workspace/.git/hooks/pre-commit` that runs both validators and aborts commit on failure
- `git checkout -- .` rollback integration in the hook's failure path
- Agent SSH key + PocketBase API token scope documentation in `workspace/SECURITY.md`
- Sample agent operation: create `press_releases` collection via POST to `/api/collections`

**Plans:** 2/3 plans executed

Plans:
**Wave 1**
- [x] 02-01-PLAN.md — internal/validate package (ValidateTemplate, ValidateHTML) + monms validate CLI

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 02-02-PLAN.md — pre-commit hook install in monms init + rollback integration + scaffold tests

**Wave 3** *(blocked on Waves 1 + 2 completion)*
- [ ] 02-03-PLAN.md — workspace docs (agent-guide.md, SECURITY.md), press_releases fixture + integration test

**Status:** Pending

---

## Phase 3: Inline Contextual Editing & Demonstration Content

**Goal:** Authenticated PocketBase users can see and use contenteditable HTMX-powered inline editing on live pages. Demonstration hero_content collection and index template are seeded.

**Requirements Covered:** ICE-01, ICE-02, ICE-03, ICE-04, ICE-05, ICE-06, SEC-02, SEC-04, DEMO-01, DEMO-02, DEMO-03

**Deliverables:**
- `IsLoggedIn` context populated from PocketBase auth session cookie in the SSR route handler
- `pb_auth` cookie → Bearer token JavaScript injection in `base.gohtml` for HTMX request authorization headers
- Floating "Live Editor Active" overlay badge with PocketBase admin dashboard link (only rendered when `IsLoggedIn = true`)
- `workspace/templates/index.gohtml` reading from `hero_content` collection with `contenteditable` + `hx-put` + `hx-trigger="blur"` on title and body fields
- PocketBase seed migration: `hero_content` collection with `title` (text) and `body` (text) fields, one `homepage` record
- Validation: unauthenticated page load renders no `contenteditable` attributes
- Manual test walkthrough in `workspace/EDITING-GUIDE.md` documenting the login → edit → save flow

**Plans:** 4/4 plans complete

Plans:
**Wave 1** *(parallel — no shared files)*
- [x] 03-01-PLAN.md — hero_content schema, API rules, bootstrap homepage seed
- [x] 03-02-PLAN.md — auth cookie bridge, SSR AuthToken + Hero enrichment

**Wave 2** *(blocked on Wave 1)*
- [x] 03-03-PLAN.md — base/index templates, editor badge, HTMX inline edit, CSS

**Wave 3** *(blocked on Waves 1 + 2)*
- [x] 03-04-PLAN.md — integration tests (ICE/SEC), EDITING-GUIDE.md

**Status:** Confirmed (UAT 2026-05-26)

---

## Milestone 2: v2 — Staging & Client Content Publish

**Goal:** Staging and production environments with dual promotion rails; clients publish editorial content to live via admin UI without consultant involvement.

---

## Phase 4: Staging Environments & Client Content Publish

**Goal:** Clients publish editorial content from staging to production via admin UI; structure continues to promote via Git tags; media uses shared CDN URLs.

**Requirements Covered:** ENV-01, ENV-02, ENV-03, PUB-01, PUB-02, PUB-03, PUB-04, PUB-05, PUB-06, PUB-07, PUB-08, PUB-09, MED-01, MED-02

**Depends on:** Phase 3 (inline editing)

**Deliverables:**
- `workspace/content/` convention + `editorial: true` flag in schema JSON
- `internal/content/` — export, import, diff, upsert by record ID
- `monms content export|import|diff|publish` CLI subcommands
- `POST /api/monms/content/import` with scoped publish token on production
- Admin publish page (`/api/monms/publish`) with diff preview and **Publish to live**
- Publisher role / permission model
- Lifecycle docs (`specs/staging.md`, README updates)

**Plans:** 4/6 plans executed

Plans:
**Wave 0**
- [x] 04-01-PLAN.md — Editorial schema parse, test fixtures, hero editorial flag

**Wave 1** *(blocked on Wave 0)*
- [x] 04-02-PLAN.md — Core content export/import/diff/checksum/state engine

**Wave 2** *(blocked on Wave 1)*
- [x] 04-03-PLAN.md — monms content CLI (export, import, diff, publish)

**Wave 3** *(blocked on Waves 1–2)*
- [x] 04-04-PLAN.md — Production POST /api/monms/content/import + publish token

**Wave 4** *(blocked on Waves 2–3)*
- [ ] 04-05-PLAN.md — Staging publish UI, publisher gate, editor badge link

**Wave 5** *(blocked on Wave 4)*
- [ ] 04-06-PLAN.md — Four-layer docs, MEDIA.md, gitignore (ENV/MED)

**Status:** Ready to execute

---

## Backlog (v2)

The following are tracked but not scheduled for v1:

- **EXT-01**: Git rollback via REST webhook or chatbox command
- **EXT-02**: Per-field cache invalidation
- **EXT-03**: Agent-managed CSS variable theming
- **MULT-01**: Multi-workspace routing (single binary)
- **MULT-02**: Per-workspace domain routing via Host header
- **RICH-01**: Markdown rendering in inline edit regions
- **RICH-02**: Image drag-and-drop upload in inline edit mode

---
*Roadmap created: 2026-05-22*
*Last updated: 2026-05-26 — Phase 3 UAT confirmed*
