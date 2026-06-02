---
name: monms-site-shaping
description: >-
  Executes MonMS L2 site mutations — schema dual-write, Go HTML templates,
  HTMX inline-edit patterns, monms validate, and pre-commit guardrails.
  Use when editing templates, schema, or assets in the configured site directory
  (not necessarily ./site).
---

# MonMS Site Shaping

L2 structure rail only. Agents shape templates, schema, and assets — **not** routine editorial copy (clients use `/_monms/publish`).

Foundation: [monms-architecture](../monms-architecture/SKILL.md). Full walkthrough: [reference.md](reference.md).

## Which directory?

All paths below are **inside the site directory** MonMS is using for this instance — resolved by `-s` / `--site`, then `MONMS_SITE`, then default `./site`. Do not assume the folder is named `site`; `./site-stage` and `./site-prod` are common.

Before shaping, set or confirm the path:

```bash
# Examples — use whatever path this instance actually runs
export SITE=./site          # default demo name
# export SITE=./site-stage  # staging checkout
monms validate -s "$SITE"
```

When `cd` into the site directory for Git commits, use that same `$SITE` path.

**Doctree folders** (`{site}/{stub}/` markdown trees): renaming or moving leaf folders changes the engine-managed `dt_*` collection name. After filesystem edits, use `/_monms/doctrees` alignment panel → re-scan → confirm; see [monms-doctree](../monms-doctree/SKILL.md). Do not edit `schema/dt_*.json` names by hand without matching `DoctreeCollectionName`.

## Prerequisites

| Variable | Purpose |
|----------|---------|
| `MONMS_URL` | Running server base URL |
| `POCKETBASE_ADMIN_TOKEN` | Admin JWT for collection management (never commit) |
| `MONMS_BIN` | Path to `monms` binary (optional; else PATH or `../../monms`) |

Obtain token:

```bash
curl -s -X POST "$MONMS_URL/api/collections/_superusers/auth-with-password" \
  -H "Content-Type: application/json" \
  -d '{"identity":"admin@example.com","password":"your-admin-password"}' \
  | jq -r '.token'
export POCKETBASE_ADMIN_TOKEN="<token>"
```

## Schema dual-write checklist

When creating a collection:

- [ ] **Step 1 — Live API:** `POST $MONMS_URL/api/collections` with `Authorization: Bearer $POCKETBASE_ADMIN_TOKEN`
- [ ] **Step 2 — Audit file:** Write `{siteDir}/schema/{name}.json` mirroring the collection definition
- [ ] **Step 3 — Editorial flag:** Add `"editorial": true` if clients should publish this collection to production (MonMS-only key; PocketBase strips it on import)
- [ ] **Step 4 — Commit:** `git commit -m "agent: …"` (D-45 prefix)

Schema-only commits skip pre-commit template validation.

## Template routing decision tree

| URL | Template path |
|-----|---------------|
| `/` | `templates/index.gohtml` |
| `/press` | `templates/press.gohtml` **or** `templates/press/index.gohtml` (flat wins) |
| `/press/2024` | `templates/press/2024.gohtml` |
| `/about` | `templates/about.gohtml` or `templates/about/index.gohtml` |

**404 after new page?** Template path does not mirror URL slug — check mirror+index rule first.

**Reserved first segments** (not page templates): `api`, `assets`, `_`, `_monms`.

### Page templates

- Use `{{define "body"}}…{{end}}` only — no `<!DOCTYPE>`, `<html>`, `<head>`, or `<body>`
- Shell lives in `templates/layouts/base.gohtml`

### Fragments

- Path: `templates/fragments/{name}.gohtml`
- Served at `/fragments/{name}` **without** base layout
- **No** `{{define "body"}}`

## HTMX inline-edit cookbook

Gate all edit attributes on `{{if .IsLoggedIn}}`.

Required pattern:

```html
<h1
  {{if .IsLoggedIn}}
  contenteditable="true"
  hx-patch="/api/collections/hero_content/records/homepage"
  hx-trigger="blur"
  hx-swap="none"
  hx-ext="json-enc"
  hx-vals='js:{"title": event.target.innerText}'
  {{end}}>{{.Hero.Title}}</h1>
```

Rules:

- **`hx-swap="none"`** — PocketBase returns JSON; without this, response replaces element text
- **`hx-ext="json-enc"`** — required when using `hx-vals` for JSON PATCH body
- **Auth** — Bearer via `htmx:configRequest` in base layout (`Authorization: Bearer {{.AuthToken}}`); never read `document.cookie` (HttpOnly)
- **`.Hero`** — injected only on homepage slug (`/`); new pages need their own SSR data enrichment in engine code
- Auth bridge: PocketBase admin at `/_/` → `POST /_monms/auth/sync` → `monms_auth` cookie

## Validation loop

```bash
monms validate -s "$SITE"
# from inside the site directory:
monms validate --site .
```

Pre-commit hook (installed by `monms init`):

- Validates **staged `*.gohtml` only**
- On failure: `git checkout -- .` and abort commit
- **Rollback caveat:** does not remove newly added untracked files — clean up manually

Never use `git commit --no-verify` routinely.

## Scaffold sync rule

When updating defaults in `internal/scaffold/embed/`, also update the **live site directory** copies if the demo should stay in sync (whatever path `$SITE` points at — often `./site` in this repo). `monms init` **skips existing files** — re-init does not upgrade old sites.

## Never commit

| Path (under `{siteDir}/`) | Reason |
|------|--------|
| `.pb_data/` | PocketBase SQLite runtime |
| `.monms/config.json` | URLs, publisher emails |
| `.monms/publish-state.json` | Last-publish checksum |
| `content/` | Ephemeral editorial exports |
| Tokens, `.env`, secrets | Credentials |

Commit `.monms/config.example.json` only.

## Common pitfalls

1. Schema JSON only **or** API only — must dual-write both
2. Full HTML in page templates — breaks layout pattern
3. Missing `hx-swap="none"` — inline edit replaces text with JSON
4. Pushing editorial copy via Git — use content rail instead
5. Forgetting `"editorial": true` — collection excluded from publish rail
6. Editing `./site` when the running instance uses `-s ./site-stage` (or another path)

## Related docs

- [docs/operators/shaping-and-agents.md](../../../docs/operators/shaping-and-agents.md)
- [docs/operators/templates-and-routing.md](../../../docs/operators/templates-and-routing.md)
- [docs/operators/security.md](../../../docs/operators/security.md)
