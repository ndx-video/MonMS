---
name: monms-dash
description: >-
  MonMS operator console at /_monms/ — embedded dashboard templates, auth gates,
  Tailwind offline build, page extension patterns, and the monms.msg notification
  strip for operator feedback. Use when editing internal/monmsdash or adding console pages.
---

# MonMS Operator Dashboard (`/_monms/`)

Foundation: [monms-architecture](../monms-architecture/SKILL.md). Engine wiring: [monms-engine-development](../monms-engine-development/SKILL.md).

The **MonMS Console** is an engine-embedded operator UI — separate from the public site’s `{siteDir}/templates/`. It is inspired by Sentinel’s ops console layout and notification strip.

## Boundaries

| Belongs here | Does **not** belong here |
|--------------|---------------------------|
| `internal/monmsdash/` — HTML pages, publish console UI | `{siteDir}/templates/` — public SSR pages |
| `/_monms/static/*` — offline Tailwind, HTMX, Alpine, JS | CDN assets in dashboard (must stay offline-capable) |
| Operator workflows (publish, future console pages) | PocketBase admin SPA (`/_/` — extend via `internal/router/embed/admin/`) |
| Browser session helpers consumed by dash | JSON content API (`internal/content/` — machine clients) |

Auth cookie bridge (`/_monms/auth/sync`, `/_monms/auth/logout`) lives in **`internal/router/auth.go`**, not `monmsdash`. Dashboard handlers only **consume** `LoadAuthFromCookie`.

## Package layout

```
internal/monmsdash/
├── register.go          # RegisterRoutes — static + dashboard + publish
├── handlers.go          # PageData, home, auth gates
├── publish.go           # Publish page + POST + diff JSON
├── templates.go         # embed parse + renderPage()
├── handlers_test.go     # integration tests
└── ui/
    ├── embed.go         # go:embed static + templates
    ├── static/
    │   ├── monms-dash.css      # committed Tailwind build output
    │   ├── components.css      # hand-written component tokens
    │   ├── src/input.css       # Tailwind v4 @source for templates
    │   └── js/
    │       ├── messages.js     # notification strip + modal
    │       └── auth-sync.js    # PB localStorage → cookie bridge helper
    └── templates/
        ├── base.gohtml         # shell: topbar, sidebar, content slot
        ├── partials/           # sidebar, nav-item
        └── pages/              # one file per route body
```

Path constants: `internal/monmsroutes/routes.go`.

## Serve pipeline

From `main.go` (order matters):

1. `router.RegisterAuthHooks` — cookie bridge + auth routes
2. `monmsdash.RegisterRoutes` — `/_monms/*` HTML + static
3. `content.RegisterRoutes` — `/api/monms/content/import`
4. `router.RegisterRoutes` — public SSR catch-all (D-14: after MonMS routes)

Unauthenticated HTML under `/_monms/*` → **303 to `/_/`** (`monmsroutes.AdminPath`). Operators sign in via PocketBase admin, then open **MonMS Console** from the PB header link (`internal/router/embed/admin/main.js`).

## Shell layout

```
┌─────────────────────────────────────────────────────────────┐
│ topbar: logo │ #monms-msg-strip (notifications) │ status pill│
├──────────┬──────────────────────────────────────────────────┤
│ sidebar  │ main → {{template "content" .}}                  │
│ nav      │                                                  │
│ …        │                                                  │
│ sign out │                                                  │
│ email    │  ← user email footer (sidebar bottom)            │
└──────────┴──────────────────────────────────────────────────┘
```

- **Topbar center** — notification strip only (not user email).
- **Sidebar footer** — signed-in email (`PageData.UserEmail`).
- **Role gates** — `IsSuperuser` (any `_superusers` auth), `IsPublisher` (`publisherEmails` in `{siteDir}/.monms/config.json`).

## PageData contract

Shared context in `handlers.go`:

| Field | Use |
|-------|-----|
| `ActivePage` | Sidebar highlight (`"home"`, `"publish"`, …) |
| `Title` | `<title>` suffix |
| `UserEmail` | Sidebar footer |
| `IsSuperuser` / `IsPublisher` | Nav sections and gates |
| `SiteURL` / `AdminURL` | Links |
| `FlashMessage` / `FlashError` | Server → navbar notification on load |

Embed extra fields on a page struct (see `publishPageData` in `publish.go`). `renderPage` passes the whole struct to `base`.

## Adding a console page

1. **Route** — register in `register.go` or a dedicated `registerXRoutes` helper:
   - Bind `bindLoadAuth(deps.LoadAuth)` on all authenticated pages
   - Use `requireAuthenticatedRedirect()` for browser HTML (303 to `/_/` when logged out)
   - Use `requirePublisherFromSite` or custom gate when needed
2. **Handler** — `buildPageData(...)` then set `ActivePage`, `Title`, optional `FlashMessage`
3. **Template** — `ui/templates/pages/{name}.gohtml` with `{{define "content"}}…{{end}}`
4. **Nav** — add `partials/nav-item.gohtml` entry in `sidebar.gohtml` with role checks
5. **Tailwind** — if new utility classes, ensure `@source` in `ui/static/src/input.css` includes `../../templates/**/*.gohtml`, then rebuild CSS (below)
6. **Test** — extend `handlers_test.go` (auth redirect, 200 for role, flash attrs if applicable)

Example handler pattern:

```go
func settingsHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
    return func(e *core.RequestEvent) error {
        data, err := buildPageData(e, deps.SiteAbs, "settings", "Settings")
        if err != nil {
            return e.InternalServerError("dashboard", err)
        }
        return tmpl.renderPage(e.Response, "settings", data)
    }
}
```

Reserved URL prefixes for SSR are unrelated — dashboard routes register on PocketBase router **before** SSR catch-all.

## Notification system (operator feedback)

Cloned from Sentinel’s `messages.js`. **Always use the strip** for operator-visible outcomes — not inline page alerts (except persistent instructional copy like publish warnings).

### Navbar slot

`#monms-msg-strip` in `base.gohtml` center column. Empty state: “No messages”. Latest message shown with type-colored styling; click opens history modal (100 items, session-scoped, copy-to-clipboard).

### Server-rendered flash (full page POST/redirect)

Set on `PageData` before `renderPage`:

```go
data.FlashMessage = "Content published successfully."
data.FlashError = false // true → error styling
```

Template emits attributes consumed on load:

```html
<div id="monms-msg-strip" … data-flash="{{.FlashMessage}}" data-flash-type="error|success"></div>
```

`messages.js` calls `consumeServerFlash()` on `DOMContentLoaded`. Prefer this for form POST results (see `buildPublishPageData` in `publish.go`).

### Client-side API (HTMX, fetch, inline scripts)

Loaded globally from `base.gohtml`. Use after `messages.js`:

```javascript
monms.msg('Saved draft', 'success');
monms.msgOk('Done');
monms.msgErr('Publish failed', { detail: 'Connection refused' });
monms.msgWarn('Production URL not set');
monms.msgInfo('Diff refreshed');
monms.msgClear();
monms.confirm('Publish now?', function () { /* yes */ }, function () { /* no */ });
```

| Type | When |
|------|------|
| `success` | Completed action |
| `error` | Failed action, validation blockers |
| `warn` | Recoverable / setup incomplete |
| `info` | Neutral status |

Optional `meta`: `{ detail, trace, copyText }` — shown expandable in history modal.

**Do not** duplicate flash as a separate `monms-alert` banner on the same page unless the message must remain visible while scrolling long content (rare). Default: strip only.

### HTMX pattern

After swap or request completion:

```javascript
document.body.addEventListener('htmx:afterRequest', function (ev) {
  if (!ev.detail.successful) {
    monms.msgErr('Request failed');
    return;
  }
  monms.msgOk('Updated');
});
```

Or return a fragment that runs a small inline script calling `monms.msgOk` — prefer event listeners in page-specific JS under `ui/static/js/` when logic grows.

## Styling rules

- **Offline-first:** Inter + Material Symbols + Tailwind build committed as `monms-dash.css`. No CDN links in dashboard templates.
- **Rebuild CSS** after template class changes:

```bash
./scripts/build-monms-dash-css.sh
```

- **Component tokens** — reusable classes in `components.css` (`nav-link`, `monms-panel`, `monms-btn-*`, `monms-alert-*`). Match existing ops-console aesthetic (dark, zero radius, uppercase tracking).
- **New `@theme` colors** — edit `ui/static/src/input.css`, rebuild, commit both `input.css` and `monms-dash.css`.

## Auth and roles (quick reference)

| Check | Mechanism |
|-------|-----------|
| Logged in | `e.Auth != nil` after `LoadAuthFromCookie` |
| Superuser | `_superusers` record → `IsSuperuser` in `buildPageData` |
| Publisher | Email in `publisherEmails` → `IsPublisher` |
| Publish POST | `requirePublisherFromSite` + token env |

Cookie name: `monms_auth` (HttpOnly). PB admin link calls `monmsSyncAuth()` when needed (`auth-sync.js`).

## Testing

```bash
go test ./internal/monmsdash/... -count=1 -short
```

Patterns in `handlers_test.go`:

- Unauthenticated GET → 303 to `/_/` 
- Static assets under `/_monms/static/` return 200 (proves embed + offline bundle)
- Role-gated pages (publisher vs editor)
- `data-flash="..."` present after successful publish POST

Mirror production wiring: `monmsdash.RegisterRoutes` + `content.RegisterRoutes` + `schema.RegisterBootstrapHook`.

## Anti-patterns

- Putting dashboard templates in `{siteDir}/templates/`
- Adding CDN Tailwind/HTMX to `base.gohtml`
- Inline success/error banners instead of `FlashMessage` / `monms.msg*`
- Assuming `./site` — use `deps.SiteAbs` for config paths
- Registering dashboard routes after `router.RegisterRoutes` (SSR may swallow paths)
- Duplicating auth route handlers inside `monmsdash` — extend `internal/router/auth.go` instead

## Related docs

- [docs/reference/monms-api.md](../../../docs/reference/monms-api.md) — auth sync/logout, content import
- [docs/user-guide/publish-to-live.md](../../../docs/user-guide/publish-to-live.md) — publish console behavior
- Sentinel reference: `sentinel/dashboard/static/js/messages.js` (upstream for notification UX)
