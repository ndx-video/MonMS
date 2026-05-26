# MonMS Workspace

This directory is the **Git-tracked site structure** for a MonMS deployment. The Go binary (`monms`) stays generic and frozen; everything that defines how the site looks and behaves lives here.

Editorial **content** (copy clients write) lives in `.pb_data/` at runtime and promotes to production separately — see [../specs/staging.md](../specs/staging.md).

## Four layers (what this folder is)

| Layer | In this folder? | Promoted how |
|-------|-----------------|--------------|
| Engine (binary) | No | Semver release |
| **Structure** (templates, schema, assets) | **Yes** — this repo | Git tag → production deploy |
| Content (editorial records) | Runtime: `.pb_data/`; export: `content/` (v2) | Publish button → JSON upsert |
| Audience | No | Production URL |

## Directory layout

```
workspace/
├── schema/                 # L2 — PocketBase collection definitions (JSON)
│   ├── hero_content.json
│   └── press_releases.json
├── content/                # L3 — editorial record exports (v2, optional in Git)
│   └── hero_content.json
├── templates/
│   ├── layouts/
│   │   └── base.gohtml     # Global HTML shell, HTMX, editor badge
│   ├── fragments/          # HTMX partials (served at /fragments/{name})
│   ├── errors/
│   │   └── errors.gohtml   # Styled 404/500 pages
│   ├── index.gohtml        # Homepage (route: /)
│   └── {slug}.gohtml       # Additional route templates
├── assets/
│   └── main.css            # Site styles (L2 — deploys with structure tag)
├── .pb_data/               # L3 runtime — PocketBase SQLite (DO NOT COMMIT)
└── .git/                   # Version control for structure mutations
```

## What lives where

| Concern | Location | Changed by | Promotion |
|---------|----------|------------|-----------|
| Page layout and routing | `templates/*.gohtml` | Consultants, agents | Structure rail (Git tag) |
| Collection shape | `schema/*.json` + live PocketBase | Consultants, agents | Structure rail |
| Site CSS/static files | `assets/` | Consultants, agents | Structure rail |
| Editorial copy | `.pb_data/` collections | Clients (inline HTMX) | Content rail (Publish button) |
| Media URLs in content | Text fields pointing at CDN | Clients | Content rail (URL strings only) |
| Auth sessions | `.pb_data/` | PocketBase admin at `/_/` | Never synced between envs |

## Staging vs production

In a full deployment you typically have **two workspaces** (two MonMS instances):

| | Staging | Production |
|---|---------|------------|
| **Structure** | `main` branch, active development | Tagged release (e.g. `v1.2.0`) |
| **Content** | Clients edit here | Updated via **Publish to live** |
| **Audience** | Internal preview | Public |

Consultants tag structure releases. Clients publish content themselves — consultants are not involved in every copy change.

## Getting started

If this workspace was not created yet:

```bash
# From the MonMS repo root
../monms init --workspace .
```

Re-running `init` on an existing workspace skips files that already exist (safe to run again).

Add `.pb_data/` to `.gitignore` immediately:

```bash
echo ".pb_data/" >> .gitignore
git add .gitignore
git commit -m "chore: exclude .pb_data from git"
```

## Template conventions

### Page templates

Every route template defines a `body` block — the base layout wraps it automatically:

```html
{{define "body"}}
<section>
  <h1>My Page</h1>
</section>
{{end}}
```

Do **not** re-declare `<!DOCTYPE html>`, `<html>`, `<head>`, or `<body>` in page templates.

### Inline editing (authenticated users)

When `.IsLoggedIn` is true, expose `contenteditable` with HTMX blur-save:

```html
<h1
  {{if .IsLoggedIn}}
  contenteditable="true"
  hx-put="/api/collections/hero_content/records/homepage"
  hx-trigger="blur"
  hx-ext="json-enc"
  hx-vals='js:{"title": event.target.innerText}'
  {{end}}>
  {{.Hero.Title}}
</h1>
```

The base layout injects the PocketBase Bearer token into HTMX requests automatically.

### Fragment partials

Files in `templates/fragments/` are served at `/fragments/{name}` without the base layout — use for HTMX partial swaps. Do not use `{{define "body"}}` in fragments.

## Schema dual-write (structure rail)

When creating a new PocketBase collection:

1. `POST /api/collections` — takes effect immediately, no server restart
2. Write matching JSON to `schema/{name}.json` — audit trail and bootstrap self-healing on next start

For collections clients will publish, add `"editorial": true` to the schema JSON (v2).

See [agent-guide.md](agent-guide.md) for curl examples.

## Content publish (content rail, v2)

Editorial records export to `content/{collection}.json` and upsert to production by record ID. Clients trigger publish from the PocketBase admin **Publish to live** button.

Media on public CDNs: store the **URL in a text field** — blobs are not copied between staging and production.

Full spec: [../specs/staging.md](../specs/staging.md)

## Validation and commits (structure rail)

A pre-commit hook (installed by `monms init`) validates staged `*.gohtml` files:

```bash
# Manual validation before commit
monms validate --workspace .

# Commit triggers hook automatically
git add templates/my-page.gohtml
git commit -m "agent: add my-page template"
```

On validation failure, the hook runs `git checkout -- .` to restore the last stable state.

Use `agent:` prefix in commit messages for AI mutations. Tag releases for production structure deploys.

## Guides

| Guide | Purpose |
|-------|---------|
| [EDITING-GUIDE.md](EDITING-GUIDE.md) | Human inline editing walkthrough |
| [agent-guide.md](agent-guide.md) | AI agent structure mutation workflow |
| [SECURITY.md](SECURITY.md) | SSH keys, API tokens, git hygiene |
| [../specs/staging.md](../specs/staging.md) | Environments, roles, content publish |

## Environment variables

| Variable | Purpose |
|----------|---------|
| `MONMS_URL` | Running server URL (e.g. `http://localhost:8090`) |
| `POCKETBASE_ADMIN_TOKEN` | Admin JWT for collection management |
| `MONMS_BIN` | Path to the `monms` binary (for pre-commit hook) |
| `MONMS_WORKSPACE` | Override workspace path (default: `./workspace`) |

Publish credentials (v2): production URL and scoped publish token — configured once by consultant, used by client Publish button.
