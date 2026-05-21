# Phase 3: Inline Contextual Editing & Demonstration Content - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-22
**Phase:** 3-Inline Contextual Editing & Demonstration Content
**Mode:** `--auto` (autonomous selection, no user prompts)
**Areas discussed:** Auth session & HTMX authorization, Editor overlay UX, Hero content seeding, PocketBase API rules, Inline edit markup & HTMX save, Verification approach

---

## Auth Session & HTMX Authorization

| Option | Description | Selected |
|--------|-------------|----------|
| Client reads `pb_auth` cookie in JS | PRD §5 script pattern; breaks SEC-04 if HttpOnly | |
| Cookie-only HTMX (no Bearer) | Rely on same-origin cookie sent automatically | |
| Server-injected Bearer token | Go renders token into `htmx:configRequest` when logged in | ✓ |

**Auto-selected choice:** Server-injected Bearer token
**Notes:** `[auto] Auth session & HTMX authorization — Q: "How should HTMX authorize PocketBase PUT requests?" → Selected: "Server-injected Bearer token" (recommended default). Resolves ICE-05 vs SEC-04: HttpOnly cookie never read by JS; Bearer still sent on HTMX requests. IsLoggedIn stays e.Auth != nil (D-50).

---

## Editor Overlay UX

| Option | Description | Selected |
|--------|-------------|----------|
| Keep hidden placeholder always | Phase 1 DOM structure unchanged | |
| Conditional badge only when logged in | UI-SPEC .editor-badge, link to /_/ | ✓ |
| Persistent badge with login CTA when logged out | Extra UI not in requirements | |

**Auto-selected choice:** Conditional badge when logged in
**Notes:** `[auto] Editor overlay UX — Q: "When should the Live Editor badge appear?" → Selected: "Conditional badge only when logged in" (recommended default). ICE-01. Reuses .editor-badge classes from 01-UI-SPEC. Logged-out visitors see no overlay DOM (D-55).

---

## Hero Content Seeding

| Option | Description | Selected |
|--------|-------------|----------|
| Template `.App.FindRecordById` | PRD §5 pattern | |
| Handler-loaded Hero map | Go loads record for index route only | ✓ |
| Manual seed via admin UI only | No automated demo | |

**Auto-selected choice:** Handler-loaded Hero map + schema JSON + idempotent bootstrap seed
**Notes:** `[auto] Hero content seeding — Q: "How is hero_content loaded and seeded?" → Selected: "Handler-loaded Hero map + schema JSON + idempotent bootstrap seed" (recommended default). workspace/schema/hero_content.json per D-57; fixed id homepage per D-58; no App in templates per D-59.

---

## PocketBase API Rules

| Option | Description | Selected |
|--------|-------------|----------|
| Go middleware validates PUT | Custom auth layer in router | |
| Collection rules in schema JSON | Public read, auth-only update | ✓ |
| Open update (insecure) | Fails SEC-02 | |

**Auto-selected choice:** Collection rules in schema JSON
**Notes:** `[auto] PocketBase API rules — Q: "How is SEC-02 enforced?" → Selected: "Collection rules in schema JSON" (recommended default). list/view public; update requires auth. No custom middleware (D-62).

---

## Inline Edit Markup & HTMX Save

| Option | Description | Selected |
|--------|-------------|----------|
| Always contenteditable, JS hides | Fails ICE-06 | |
| Conditional attributes in Go template | {{if .IsLoggedIn}} wraps attrs | ✓ |
| Separate editor-only template file | Duplicate templates | |

**Auto-selected choice:** Conditional attributes in Go template
**Notes:** `[auto] Inline edit markup — Q: "How do unauthenticated users avoid contenteditable?" → Selected: "Conditional attributes in Go template" (recommended default). hx-put to /api/collections/hero_content/records/homepage, blur trigger, partial hx-vals per field (D-63–D-66).

---

## Verification Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Manual EDITING-GUIDE only | No automated tests | |
| Integration tests + EDITING-GUIDE | Unauth/authed HTML + PUT rejection | ✓ |
| E2E browser automation | Heavier than needed for v1 | |

**Auto-selected choice:** Integration tests + EDITING-GUIDE
**Notes:** `[auto] Verification approach — Q: "How is Phase 3 validated?" → Selected: "Integration tests + EDITING-GUIDE" (recommended default). testutil harness; assert no contenteditable when logged out (D-71–D-72).

---

## Claude's Discretion

- PocketBase rule string exact syntax, seed file placement, default homepage copy, HTMX error UX — delegated per CONTEXT.md Claude's Discretion section.

## Deferred Ideas

- Automatic git commits on human inline edits (PRD NFR §7.2)
- RICH-01 Markdown, RICH-02 image upload (v2)
- Custom /admin/login route
- Save confirmation toast UI
