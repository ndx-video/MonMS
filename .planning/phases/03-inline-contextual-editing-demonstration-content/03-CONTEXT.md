# Phase 3: Inline Contextual Editing & Demonstration Content - Context

**Gathered:** 2026-05-22
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver authenticated inline contextual editing on live SSR pages: `IsLoggedIn` drives conditional `contenteditable` + HTMX blur-save to PocketBase, a floating "Live Editor Active" badge linking to `/_/`, and demonstration `hero_content` collection with seeded `homepage` record rendered on `index.gohtml`. Phase 3 does not include agent mutation guardrails (Phase 2), automatic git commits on human edits (deferred), chatbox orchestration, markdown/rich media editing (v2 RICH-*), or REST webhook rollback (v2 EXT-01).

</domain>

<decisions>
## Implementation Decisions

### Auth Session & Template Context
- **D-50:** `IsLoggedIn` remains `e.Auth != nil` in `ssrData` (already wired in `internal/router/ssr.go`) — no alternate cookie-parsing path in templates.
- **D-51:** Pass `AuthToken` string in SSR data **only when** `IsLoggedIn` is true — value from `e.Auth.Token()` (or equivalent PocketBase auth record token accessor). Empty/absent when logged out. Used for HTMX Authorization header injection; never expose full user object to templates beyond existing `User` key.
- **D-52:** Resolve ICE-05 vs SEC-04 by **server-side token injection**: inline script sets `htmx:configRequest` Bearer header from a Go-rendered `{{.AuthToken}}` variable — **do not** read `pb_auth` from `document.cookie` in JavaScript (HttpOnly cookie stays inaccessible to JS per SEC-04). ICE-05 satisfied because HTMX requests still carry `Authorization: Bearer <token>`.
- **D-53:** Login path for human editors documented as PocketBase admin at `/_/` (not a custom `/admin/login` route). `EDITING-GUIDE.md` walkthrough uses admin UI login → navigate to `/` → edit inline.

### Editor Overlay & Badge UX
- **D-54:** Replace Phase 1 hidden `#editor-overlay` placeholder with `{{if .IsLoggedIn}}` block rendering `.editor-badge` markup per `01-UI-SPEC.md` (fixed top-right, pulsing dot, "Live Editor Active", link to `/_/` labeled "Full Admin Dashboard").
- **D-55:** When logged out, omit editor overlay from DOM entirely (no hidden placeholder) — satisfies ICE-06 visually and keeps DOM clean for public visitors.
- **D-56:** Reuse existing `main.css` classes `.editor-badge`, `.editor-badge__dot`, `.editor-badge__link` — no new design system; match indigo badge spec from UI-SPEC.

### Hero Content Data & Seeding
- **D-57:** Add `workspace/schema/hero_content.json` declarative collection (`title` text, `body` text) imported via existing `RegisterBootstrapHook` / D-32 sync — same pattern as `press_releases.json`.
- **D-58:** Seed one record with **fixed id** `homepage` (matches PRD `FindRecordById` / HTMX PUT URL). Idempotent bootstrap seed in Go: after schema import, if `hero_content` collection exists and no `homepage` record, create it with default title/body copy from UI-SPEC hero stub.
- **D-59:** Do **not** expose `core.App` to templates via `.App` — load homepage record in Go SSR handler (index slug only) and pass `Hero` map `{"Title": ..., "Body": ..., "ID": "homepage"}` in template data. Safer, testable, matches current `ssrData` pattern.
- **D-60:** Index route (`/` / empty slug) requires hero record — if missing after seed failure, render index with fallback static copy and `log/slog` warning (no 500 panic).

### PocketBase API Rules (SEC-02)
- **D-61:** `hero_content` collection rules in schema JSON: **public read** (`listRule`/`viewRule` empty or `""`), **authenticated update only** (`updateRule`: `@request.auth.id != ""`), **create/delete** restricted to admin (`createRule`/`deleteRule`: `@request.auth.id != ""` or admin-only equivalent). Unauthenticated PUT returns 403/401 at API layer.
- **D-62:** No custom Go middleware for PUT auth in Phase 3 — rely on PocketBase collection rules + Bearer header from D-52.

### Inline Edit Markup & HTMX Save
- **D-63:** `contenteditable`, `hx-put`, `hx-trigger="blur"`, and `hx-vals` attributes render **only inside** `{{if .IsLoggedIn}}` blocks on title and body elements — unauthenticated HTML must not contain those strings (ICE-03, ICE-06). Integration test asserts absence.
- **D-64:** HTMX PUT targets: `/api/collections/hero_content/records/homepage` for both fields (PRD canonical URL).
- **D-65:** Field payload via `hx-vals='js:{"title": event.target.innerText}'` and `hx-vals='js:{"body": event.target.innerText}'` respectively — partial field update on blur, no full-record replace.
- **D-66:** Title field uses `hx-ext="json-enc"` per PRD; body field uses standard `hx-vals` (no json-enc required for single text field).
- **D-67:** HTMX 1.9.12 CDN version unchanged from Phase 1 (D-24 / UI-SPEC) — no version bump in Phase 3.

### Index Template & Scaffold
- **D-68:** Update `internal/scaffold/embed/index.gohtml` and post-init workspace `templates/index.gohtml` to render `{{.Hero.Title}}` / `{{.Hero.Body}}` with inline-edit attributes when logged in; retain site header + Alpine mobile nav from Phase 1 stub.
- **D-69:** Update `internal/scaffold/embed/base.gohtml` with live editor badge + auth script blocks (uncomment Phase 3 sections); `monms init` skip-if-exists behavior unchanged — operators with existing workspace may need manual merge or re-init guidance in EDITING-GUIDE.
- **D-70:** Add `workspace/EDITING-GUIDE.md` (not embedded in binary) documenting: login at `/_/`, verify badge, edit title/body on `/`, blur-save behavior, logout verification (no contenteditable), troubleshooting failed saves (401 → re-login).

### Verification & Testing
- **D-71:** Integration tests in `internal/router/`: (1) unauthenticated GET `/` has no `contenteditable` in body; (2) authenticated admin session GET `/` includes `contenteditable` and `Live Editor Active`; (3) unauthenticated PUT to hero_content record returns non-2xx (SEC-02).
- **D-72:** Use existing `testutil` workspace + `startTestServer` pattern from Phase 2 press_releases test; create admin user + auth cookie in test harness for authenticated cases.

### Claude's Discretion
- Exact PocketBase schema JSON rule string syntax if import rejects shorthand — verify against PocketBase v0.22+ collection import format.
- Bootstrap seed implementation file placement (`internal/schema/seed.go` vs extended bootstrap hook).
- Whether to pass `Hero` on non-index routes (default: omit; only index handler enriches data).
- HTMX error feedback UX (toast vs silent fail) — minimal: rely on browser network tab; optional `hx-on::after-request` logging deferred.
- Default seed copy text for homepage record.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `specs/monms-prd.md` §5 — Inline contextual editing: base layout badge, pb_auth/HTMX auth script, index.gohtml contenteditable + hx-put pattern, UX flow
- `specs/monms-prd.md` §7 — NFRs: performance, security (database-level PUT validation)
- `.planning/PROJECT.md` — Visual inline editing core value, HTMX + PocketBase constraints
- `.planning/REQUIREMENTS.md` — ICE-01 through ICE-06, SEC-02, SEC-04, DEMO-01, DEMO-02 traceability
- `.planning/ROADMAP.md` — Phase 3 deliverables list

### Phase 1–2 Foundation (locked)
- `.planning/phases/01-core-go-runtime-workspace-foundation/01-CONTEXT.md` — D-01 through D-32, especially D-24 (HTMX CDN, overlay placeholder), D-32 (schema sync)
- `.planning/phases/01-core-go-runtime-workspace-foundation/01-UI-SPEC.md` — Editor overlay classes, CDN order, copywriting contract, Phase 3 deferred items now in scope
- `.planning/phases/02-agent-mutation-engine-safety-guardrails/02-CONTEXT.md` — D-33 through D-49; dual-write schema pattern for hero_content.json

### Implementation anchors
- `internal/router/ssr.go` — `ssrData`, `IsLoggedIn`, extension point for Hero + AuthToken
- `internal/schema/sync.go` — Collection import hook; extend for record seed
- `internal/scaffold/embed/base.gohtml` — Phase 3 badge + auth script targets
- `internal/scaffold/embed/index.gohtml` — Hero + inline edit markup target
- `workspace/schema/press_releases.json` — Schema JSON format reference for hero_content.json

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/router/ssr.go` — `IsLoggedIn` and `User` already in SSR data; extend `ssrData` or index-specific enrichment
- `internal/schema/sync.go` — Bootstrap collection import; add idempotent record seed after import
- `internal/scaffold/embed/base.gohtml` — Commented Phase 3 blocks for badge and HTMX auth script
- `internal/scaffold/embed/index.gohtml` — Phase 1 hero stub + Alpine nav; replace hero section with dynamic content
- `workspace/assets/main.css` — `.editor-badge*` classes pre-defined in UI-SPEC
- `internal/router/press_releases_test.go` + `internal/testutil` — Integration test harness pattern

### Established Patterns
- Schema-as-JSON in `workspace/schema/*.json` synced on bootstrap (D-32, D-33)
- Template execute with `"base"` layout and page-specific `body` define
- Phase 1 UI-SPEC CDN order is authoritative — do not reorder scripts
- Conditional template attributes use `{{if .IsLoggedIn}}` wrapping attribute lists (validator strips for AGT-04)

### Integration Points
- PocketBase `e.Auth` on `core.RequestEvent` for session detection
- HTMX PUT → `/api/collections/hero_content/records/homepage` (same-origin, Bearer header)
- Admin UI at `/_/` for login and full CRUD fallback
- `monms init` embed templates copied to workspace — scaffold embed must be updated for new installs

</code_context>

<specifics>
## Specific Ideas

- PRD §5 is the canonical markup reference for badge, auth script, and index inline-edit attributes — adapt auth script per D-52 (server token, not cookie parse).
- SEC-04 / ICE-05 tension explicitly resolved: HttpOnly `pb_auth` stays; Bearer comes from server-rendered token only when logged in.
- Fixed record id `homepage` keeps HTMX URLs stable and matches PRD examples.
- Phase 2 deferred "automatic git commits on human inline edits" remains out of scope — human edits persist in SQLite only until operator/agent commits workspace separately.

</specifics>

<deferred>
## Deferred Ideas

- **Automatic git commits on human inline edits** — PRD NFR §7.2; not in Phase 3 ROADMAP deliverables; operator/agent manual git workflow
- **RICH-01 Markdown in contenteditable** — v2 backlog
- **RICH-02 Image drag-and-drop upload** — v2 backlog
- **Custom `/admin/login` route** — use PocketBase `/_/` only
- **Toast/inline save confirmation UI** — nice-to-have; not required for ICE acceptance
- **Passing `core.App` into templates** — rejected in favor of handler-loaded `Hero` map (D-59)

</deferred>

---

*Phase: 3-Inline Contextual Editing & Demonstration Content*
*Context gathered: 2026-05-22*
