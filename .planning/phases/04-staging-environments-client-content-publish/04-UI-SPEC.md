---
phase: 4
slug: staging-environments-client-content-publish
status: draft
shadcn_initialized: false
preset: none
created: 2026-05-23
---

# Phase 4 — UI Design Contract

> Server-rendered **Publish to live** console at `/api/monms/publish` (MonMS HTML page, not PocketBase admin SPA). Extends Phase 3 editor badge with publish link.

**Sources:** D-50–D-56, PUB-06–PUB-08; 04-RESEARCH.md Pattern 5; 01-UI-SPEC.md + 03-UI-SPEC.md tokens.

---

## Design System

| Property | Value |
|----------|-------|
| Tool | Go SSR — embedded HTML template or inline `e.HTML` |
| Component library | Semantic HTML + existing `main.css` utility classes |
| Styling | Reuse Phase 1/3 slate palette; no new CDN deps |
| JS | Minimal vanilla JS or HTMX for diff refresh + POST publish (optional) |

**Route:** `GET /api/monms/publish` — HTML console. Link from editor badge: "Publish to live".

---

## Page Layout

```
┌─────────────────────────────────────────────────────┐
│  Publish to live                          [Back to /] │
├─────────────────────────────────────────────────────┤
│  Last published: 2026-05-20 14:32 UTC               │
│  Status: ● Unpublished changes (or ✓ Up to date)    │
├─────────────────────────────────────────────────────┤
│  Pending changes                                    │
│  ┌───────────────────────────────────────────────┐  │
│  │ hero_content / homepage                       │  │
│  │   title: "Welcome…" → "Welcome to Acme"       │  │
│  │   body: (unchanged)                           │  │
│  └───────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────┤
│  [ Preview diff ]              [ Publish now ]      │
└─────────────────────────────────────────────────────┘
```

---

## Components

### Publish status badge

- Green dot + "Up to date" when checksum matches `publish-state.json`
- Amber dot + "Unpublished changes" when diff non-empty
- Typography: 14px label (`.editor-badge` scale from Phase 3)

### Diff list

- One card per collection with changed records
- Field-level before → after for text fields
- Empty state: "No pending changes — production matches staging editorial content."

### Actions

| Control | Behavior |
|---------|----------|
| Preview diff | `GET /api/monms/publish/diff` → JSON or inline refresh |
| Publish now | `POST /api/monms/publish` → export + POST production import → update state → success message |
| Back to / | Link to public homepage |

### Error states

- Production URL unset: setup instructions + link to `.monms/config.example.json`
- Non-publisher (403): "You can edit content but cannot publish. Contact your site administrator."
- Import failed: show error message from production API; do not update checksum

---

## Auth & permissions

- Page requires superuser session (same as inline edit)
- **Publish now** additionally requires publisher email in allowlist (PUB-07)
- Editors see page read-only or 403 on POST only

---

## Editor badge extension (base.gohtml)

Add link next to "Full Admin Dashboard":

```html
<a href="/api/monms/publish" class="editor-badge__link">Publish to live</a>
```

Only visible when `.IsLoggedIn` and user is in publisher allowlist (or show to all superusers with 403 on POST for non-publishers — prefer hide link for non-publishers).

---

## Copywriting

| Element | Text |
|---------|------|
| Page title | Publish to live |
| Publish button | Publish now |
| Success | Content published successfully. |
| Up to date | Production matches your staging content. |
| Unpublished | You have unpublished changes. |

---

## Accessibility

- Publish button: `type="submit"`, focus visible ring
- Status indicators: text label alongside color dot (not color-only)
- Diff cards: semantic `<ul>` / `<dl>` for field changes

---

*Phase 4 UI-SPEC — 2026-05-23*
