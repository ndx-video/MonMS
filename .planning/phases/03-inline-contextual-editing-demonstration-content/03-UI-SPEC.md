---
phase: 3
slug: inline-contextual-editing-demonstration-content
status: draft
shadcn_initialized: false
preset: none
created: 2026-05-22
---

# Phase 3 — UI Design Contract

> Visual and interaction contract for authenticated inline contextual editing: Live Editor badge, HTMX blur-save hero fields, server-rendered auth token injection, and demonstration `hero_content` seed copy. Extends Phase 1 scaffold — no new design system.

**Sources:** D-50 through D-72 (03-CONTEXT.md); 01-UI-SPEC.md (spacing, typography, color, `.editor-badge`, `.hero`); PRD §5 (markup reference, adapted per D-52); ICE-01–ICE-06, DEMO-01–DEMO-02, SEC-02, SEC-04.

---

## Design System

| Property | Value |
|----------|-------|
| Tool | none (Go SSR + Go HTML templates — not React) |
| Preset | not applicable |
| Component library | none — semantic HTML + `main.css` component classes (Phase 1) |
| Icon library | none (inline SVG hamburger unchanged from Phase 1) |
| Font | system-ui stack via Tailwind preflight |
| Styling approach | Hybrid: Tailwind v3 Play CDN + `workspace/assets/main.css` |
| JS runtime | HTMX 1.9.12 (active PUT on blur), Alpine.js 3.14.8 (mobile nav only) |

**shadcn gate:** Skipped — unchanged from Phase 1.

**Phase 3 delta:** Activate Phase 1 deferred `.editor-badge` styles; add `.hero__title--editable` / `.hero__body--editable` focus affordances; add `.editor-badge__dot` pulse animation. No new CDN dependencies.

---

## Spacing Scale

Inherited unchanged from 01-UI-SPEC.md:

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Badge dot size, inline flex gaps |
| sm | 8px | Badge vertical padding, nav padding |
| md | 16px | Hero title/body margin, badge offset (`top-4 right-4`) |
| lg | 24px | Card padding, hero body bottom margin |
| xl | 32px | Main horizontal padding |
| 2xl | 48px | Hero section vertical padding |
| 3xl | 64px | Error page vertical padding |

**Phase 3 exceptions:**
- Editor badge: `fixed top-4 right-4` (16px viewport inset) — same as Phase 1 placeholder spec
- Contenteditable focus outline offset: 2px (`outline-offset: 2px`)
- Mobile hamburger touch target: 44×44px (unchanged)

---

## Typography

Inherited unchanged from 01-UI-SPEC.md — four sizes, two weights only:

| Role | Size | Weight | Line Height | CSS / Class |
|------|------|--------|-------------|-------------|
| Body | 16px | 400 | 1.5 | `.hero__body`, card copy |
| Label | 14px | 600 | 1.4 | `.editor-badge`, nav links, `.btn` |
| Heading | 24px | 600 | 1.2 | `.hero__title`, error `<h1>` |

**Phase 3 rule:** Hero title and body use the same typography whether logged in or out. Do not switch to PRD §5 `text-4xl` / `text-lg` utilities — keep `.hero__title` and `.hero__body` from `main.css` for visual continuity.

**Logged-in edit affordance:** `cursor: text` on contenteditable hero elements only — no font-size change on edit mode.

---

## Color

Inherited unchanged from 01-UI-SPEC.md:

| Role | Value | Usage |
|------|-------|-------|
| Dominant (60%) | `#f8fafc` (slate-50) | `<body>` canvas |
| Secondary (30%) | `#ffffff`, `#f1f5f9` | Header, cards, mobile nav |
| Accent (10%) | `#4f46e5` (indigo-600) | Primary buttons, nav hover, editor badge bg, contenteditable focus ring |
| Text primary | `#0f172a` (slate-900) | Hero title |
| Text muted | `#475569` (slate-600) | Hero body, card copy |
| Border | `#e2e8f0` (slate-200) | Header, cards |
| Status live | `#22c55e` (green-500) | Editor badge dot |
| Destructive | `#dc2626` (red-600) | Reserved — no destructive UI in Phase 3 |

**Accent reserved for (Phase 3 additions in bold):**
- `.btn--primary`, nav link hover/active states
- **`.editor-badge` background**
- **`.editor-badge__link` hover**
- **Contenteditable `:focus-visible` outline on hero title/body**

**Not accent:** Badge label text (white on indigo), hero body at rest, card section.

---

## Copywriting Contract

| Element | Copy |
|---------|------|
| Live Editor badge label | **Live Editor Active** |
| Live Editor badge link | **Full Admin Dashboard** — `href="/_/"` |
| Hero seed title (`homepage` record) | **Welcome to MonMS** |
| Hero seed body (`homepage` record) | **This headline and paragraph are stored in the hero_content collection. Sign in via Admin, then click here to edit in place — changes save when you click away.** |
| Hero fallback title (seed missing) | **MonMS is running** — Phase 1 static copy |
| Hero fallback body (seed missing) | **Your workspace is live. Templates load from disk — no build step required.** |
| Primary CTA (index) | **Explore Admin** — links to `/_/` (unchanged) |
| Card section heading | **Workspace ready** (unchanged) |
| Card section body | **Templates in `workspace/templates/` render on the next request. Edit files and refresh — no restart needed.** (unchanged) |
| Empty state | N/A — hero always renders record or fallback copy |
| Save confirmation | None — silent blur-save per deferred toast UX |
| Save error (401/403) | No in-page UI — operator re-authenticates at `/_/` per EDITING-GUIDE |
| Destructive confirmation | None in Phase 3 |

**Tone:** Operator-friendly demo copy that teaches inline editing without jargon. Badge copy is status-only, not instructional.

**EDITING-GUIDE.md** (workspace doc, not embedded): references same strings above for walkthrough steps.

---

## Registry Safety

| Registry | Blocks Used | Safety Gate |
|----------|-------------|-------------|
| shadcn official | none | not applicable |
| Third-party | none | not applicable |

**CDN versions (unchanged from Phase 1 — D-67):**

| CDN | Version pin |
|-----|-------------|
| Tailwind Play | v3 Play CDN |
| HTMX | 1.9.12 |
| Alpine.js | 3.14.8, `defer` |

---

## Layout

### Base layout (`base.gohtml`)

**Logged out:** No editor overlay in DOM. Main content unchanged.

**Logged in:** Fixed badge precedes `<main>`; no `#editor-overlay` wrapper, no `hidden` placeholder.

```
<body>
  [if IsLoggedIn] .editor-badge (fixed top-right)
  <main max-w-6xl mx-auto px-6 py-12>
    {{template "body" .}}
  </main>
  [if IsLoggedIn] HTMX auth script
  Alpine.js defer
</body>
```

**Z-index stack:**

| Layer | z-index | Element |
|-------|---------|---------|
| Editor badge | 50 | `.editor-badge` |
| Site header | 40 | `.site-header` (sticky) |
| Main content | auto | `<main>` |

Badge must not overlap mobile nav drawer when open — badge stays top-right; nav panel is below header (no conflict).

### Index page (`index.gohtml`)

Structure unchanged from Phase 1: site header → hero section → card section.

**Hero data binding:**

| Template key | Source | Notes |
|--------------|--------|-------|
| `.Hero.Title` | PocketBase `hero_content` record `homepage` | Handler-loaded map (D-59) |
| `.Hero.Body` | same | |
| `.Hero.ID` | `"homepage"` | For tests/docs; not rendered |
| `.IsLoggedIn` | `e.Auth != nil` | Drives badge + edit attrs |
| `.AuthToken` | `e.Auth.Token()` when logged in | Script injection only; never render in HTML body |

**Non-index routes:** Omit `.Hero`; no inline editing markup on other pages in Phase 3.

---

## Visuals

### Live Editor badge (logged in only)

Render inside `{{if .IsLoggedIn}}` — omit entire block when logged out (D-55).

```gohtml
{{if .IsLoggedIn}}
<div class="editor-badge" role="status" aria-live="polite">
  <span class="editor-badge__dot" aria-hidden="true"></span>
  Live Editor Active
  <a href="/_/" class="editor-badge__link">Full Admin Dashboard</a>
</div>
{{end}}
```

| Property | Value |
|----------|-------|
| Position | `fixed`, top 16px, right 16px |
| Shape | Pill (`border-radius: 9999px`) |
| Background | `#4f46e5` |
| Text | 14px semibold white |
| Shadow | `0 10px 15px -3px rgba(15, 23, 42, 0.1)` |
| Dot | 8×8px green circle with pulse animation |
| Link | `#c7d2fe`, underline; hover `#ffffff` |

**Logged out:** Zero editor DOM nodes — no hidden placeholder, no `aria-hidden` stub (ICE-06 visual cleanliness).

### Hero section — authenticated vs public

**Public (logged out):** Identical to Phase 1 rendered output. Plain `<h1 class="hero__title">` and `<p class="hero__body">` with database/fallback text. No edit chrome.

**Authenticated (logged in):** Same visual weight; add focus ring on interaction only:

```gohtml
<h1 class="hero__title{{if .IsLoggedIn}} hero__title--editable{{end}}"
  {{if .IsLoggedIn}}
  contenteditable="true"
  role="textbox"
  aria-label="Hero headline"
  aria-multiline="false"
  spellcheck="true"
  hx-put="/api/collections/hero_content/records/homepage"
  hx-trigger="blur"
  hx-ext="json-enc"
  hx-vals='js:{"title": event.target.innerText}'
  {{end}}>{{.Hero.Title}}</h1>

<p class="hero__body{{if .IsLoggedIn}} hero__body--editable{{end}}"
  {{if .IsLoggedIn}}
  contenteditable="true"
  role="textbox"
  aria-label="Hero description"
  aria-multiline="true"
  spellcheck="true"
  hx-put="/api/collections/hero_content/records/homepage"
  hx-trigger="blur"
  hx-vals='js:{"body": event.target.innerText}'
  {{end}}>{{.Hero.Body}}</p>
```

**Visual edit affordances (logged in only):**

| State | Title | Body |
|-------|-------|------|
| Default | `cursor: text` | `cursor: text` |
| `:focus-visible` | 2px indigo outline, 2px offset, 4px radius | same |
| Hover | no background change | no background change |

Do not add dashed borders, pencil icons, or "click to edit" hints — badge carries edit-mode signal.

### CSS additions (`main.css`)

Extend existing classes only:

```css
/* Pulse on live dot */
.editor-badge__dot {
  position: relative;
  flex-shrink: 0;
}
.editor-badge__dot::after {
  content: "";
  position: absolute;
  inset: -2px;
  border-radius: 9999px;
  background: #22c55e;
  opacity: 0.5;
  animation: editor-dot-ping 2s cubic-bezier(0, 0, 0.2, 1) infinite;
}
@keyframes editor-dot-ping {
  0%, 100% { transform: scale(1); opacity: 0.5; }
  50% { transform: scale(1.8); opacity: 0; }
}

.hero__title--editable,
.hero__body--editable {
  cursor: text;
}
.hero__title--editable:focus-visible,
.hero__body--editable:focus-visible {
  outline: 2px solid #4f46e5;
  outline-offset: 2px;
  border-radius: 4px;
}
```

---

## Interaction

### Auth token injection (D-52, SEC-04, ICE-05)

Render **only when** `{{if .IsLoggedIn}}`. Do **not** read `document.cookie` or `pb_auth`.

```gohtml
{{if .IsLoggedIn}}
<script>
  document.body.addEventListener('htmx:configRequest', function (event) {
    event.detail.headers['Authorization'] = 'Bearer {{.AuthToken}}';
  });
</script>
{{end}}
```

| Rule | Detail |
|------|--------|
| Placement | End of `<body>`, before Alpine defer script |
| Token source | Go template `{{.AuthToken}}` from SSR handler |
| Logged out | Script block omitted entirely |
| HTMX version | 1.9.12 (head, unchanged CDN order) |

Go must HTML-escape token in template context to prevent script injection if token contains special characters.

### Inline save flow

1. User focuses contenteditable field (keyboard or pointer).
2. User edits text in place.
3. On `blur`, HTMX issues `PUT /api/collections/hero_content/records/homepage`.
4. Payload: partial field only — `{"title": ...}` or `{"body": ...}` via `hx-vals`.
5. Title uses `hx-ext="json-enc"`; body uses standard `hx-vals` (D-66).
6. Success: no DOM swap, no toast — content stays as edited (ICE-04).
7. Failure (401/403): silent; network tab shows error — EDITING-GUIDE documents re-login.

**No `hx-target` / `hx-swap`** — in-place edit, no fragment replacement.

### Conditional attribute rule (ICE-03, ICE-06)

Strings `contenteditable`, `hx-put`, `hx-trigger`, `hx-vals`, `hx-ext` must appear in HTML **only** inside `{{if .IsLoggedIn}}` blocks. Integration test asserts absence for unauthenticated GET `/`.

### User flows

| Actor | Flow |
|-------|------|
| Public visitor | GET `/` → static hero from DB → no badge → no edit attrs |
| Editor | Login at `/_/` → GET `/` → badge visible → click hero text → blur saves |
| Editor logout | Session cleared → refresh `/` → badge gone → plain hero HTML |

---

## Accessibility

| Requirement | Implementation |
|-------------|----------------|
| Document language | `<html lang="en">` (unchanged) |
| Live Editor status | `role="status"` + `aria-live="polite"` on `.editor-badge` |
| Decorative dot | `aria-hidden="true"` on `.editor-badge__dot` |
| Editable headline | `role="textbox"`, `aria-label="Hero headline"`, `aria-multiline="false"` |
| Editable body | `role="textbox"`, `aria-label="Hero description"`, `aria-multiline="true"` |
| Keyboard access | Native `contenteditable` focus order — title then body; no `tabindex` override |
| Focus visible | `:focus-visible` indigo outline on editable hero elements |
| Spellcheck | `spellcheck="true"` on both editable fields |
| Color contrast | Badge: white on indigo-600 ≥ 4.5:1; link indigo-200 on indigo-600 ≥ 4.5:1 |
| Motion | Dot pulse is decorative; respect `prefers-reduced-motion: reduce` — disable `editor-dot-ping` animation |
| Mobile nav | Unchanged Phase 1 `aria-expanded` / `aria-controls` pattern |
| Save feedback | No `aria-live` save toast (deferred) — badge remains sole persistent status indicator |

**Reduced motion CSS:**

```css
@media (prefers-reduced-motion: reduce) {
  .editor-badge__dot::after {
    animation: none;
  }
}
```

**Screen reader note:** Contenteditable regions announce as textboxes; blur-save has no announcement — acceptable for v1 per deferred toast UX.

---

## Template Context (Phase 3)

| Key | Type | When present |
|-----|------|--------------|
| `IsLoggedIn` | bool | All SSR pages |
| `User` | auth record | When logged in (existing) |
| `AuthToken` | string | When logged in only — empty/absent when logged out (D-51) |
| `Hero` | map `Title`, `Body`, `ID` | Index route only |
| `Slug` | string | All pages (existing) |
| `Code`, `Message`, `Path` | error pages only | Unchanged |

---

## CDN & Script Placement

**Strict order preserved from 01-UI-SPEC.md.** Phase 3 changes body-end scripts only:

1–7. Head: unchanged (meta, title, Tailwind, config, main.css, HTMX)

8. `{{if .IsLoggedIn}}` HTMX auth Bearer script

9. Alpine.js defer

Remove Phase 1 `#editor-overlay` hidden placeholder and commented badge block — replace with live conditional markup above.

---

## Scaffold File Checklist

| File | Phase 3 changes |
|------|-----------------|
| `templates/layouts/base.gohtml` | Conditional `.editor-badge`; conditional auth script; remove hidden overlay |
| `templates/index.gohtml` | `{{.Hero.Title}}` / `{{.Hero.Body}}`; conditional edit attrs + a11y |
| `assets/main.css` | Pulse animation; editable focus styles |
| `schema/hero_content.json` | Collection schema (not UI — referenced for demo) |
| `EDITING-GUIDE.md` | Operator walkthrough copy |

Update `internal/scaffold/embed/*` mirrors for `monms init` (D-68, D-69).

---

## Phase Boundaries

| In scope (Phase 3) | Deferred |
|--------------------|----------|
| Live Editor badge when authenticated | Save confirmation toast |
| Server-rendered Bearer token for HTMX | Cookie parsing in JS |
| Hero contenteditable + blur PUT | Markdown/rich media (RICH-*) |
| Demo seed copy for `homepage` | Automatic git commits on edit |
| Editable focus/accessibility attrs | Custom `/admin/login` route |
| Fallback static hero if seed fails | Passing `.App` to templates |

---

## Checker Sign-Off

- [ ] Dimension 1 Copywriting: PASS
- [ ] Dimension 2 Visuals: PASS
- [ ] Dimension 3 Color: PASS
- [ ] Dimension 4 Typography: PASS
- [ ] Dimension 5 Layout: PASS
- [ ] Dimension 6 Interaction: PASS
- [ ] Dimension 7 Accessibility: PASS
- [ ] Dimension 8 Registry Safety: PASS

**Approval:** pending
