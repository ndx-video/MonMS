# MonMS Workspace

This directory is the **Git-tracked site structure** for a MonMS deployment. The Go binary (`monms`) stays generic and frozen; everything that defines how the site looks and behaves lives here.

Editorial **content** (copy clients write) lives in `.pb_data/` at runtime and promotes to production separately — see [../specs/staging.md](../specs/staging.md).

## Four layers (what this folder is)

MonMS separates **engine**, **structure**, **content**, and **audience** (D-50). This Git repo is the **structure** layer (L2). Editorial copy lives in PocketBase at runtime (L3) and promotes independently.

| Layer | Name | In this folder? | Artifact | Promoted how |
|-------|------|-----------------|----------|--------------|
| **L1** | Engine | No | `monms` binary | Semver release — same binary on staging and production |
| **L2** | Structure | **Yes** — this repo | `templates/`, `schema/`, `assets/` | **Git tag** → production checkout |
| **L3** | Content | Runtime only | `.pb_data/` records; export: `content/` | **Publish to live** → JSON upsert (see below) |
| **L4** | Audience | No | Production URL | Read-only — end users |

Full lifecycle spec: [../specs/staging.md](../specs/staging.md)

## Structure rail vs content rail

Two promotion rails operate **independently** (D-51):

| Rail | What moves | Mechanism | Who |
|------|------------|-----------|-----|
| **Structure** | Templates, schema JSON, CSS/assets | Workspace **Git tag** → production deploy | Consultants, agents |
| **Content** | Editorial PocketBase records | Export → `POST /api/monms/content/import` on production | Clients (Publish button) |
| **Media** | CDN URL strings in text fields | Upserted with content — **no blob copy** | Clients (paste URLs) |

- Git tags carry **structure only** — they do not include editorial copy from `.pb_data/`.
- Structure must reach production **before** content that depends on new collections or fields.
- **Full `.pb_data/` backup/restore is not the primary publish mechanism** (D-56). Content promotes via JSON upsert only; auth sessions and local file storage stay environment-local.

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

Staging and production are **separate MonMS instances** with **separate `.pb_data/` directories** (D-52). Both run the same pinned `monms` binary unless an engine upgrade is intentional.

Optional Docker packaging uses the same model: a thin engine image plus git-managed workspace mounts and persistent `.pb_data/` volumes — see [DEPLOY-DOCKER.md](DEPLOY-DOCKER.md).

| | Staging | Production |
|---|---------|------------|
| **Instance** | Own MonMS process + workspace checkout | Own MonMS process + tagged workspace checkout |
| **`.pb_data/`** | Staging SQLite, sessions, local uploads | Production SQLite — never synced wholesale |
| **Structure (L2)** | `main` branch, active development | Tagged release (e.g. `v1.2.0`) |
| **Content (L3)** | Clients edit inline here | Updated via **Publish to live** only |
| **Audience (L4)** | Internal preview URL | Public production URL |

**One-time consultant setup** (per site):

1. Edit `.monms/config.json` (created by `monms init` with `_fieldDocs` for each option): set `productionUrl`, `publisherEmails`, and optional `allowedHosts` (CORS origins for `monms serve` when `--origins` is omitted).
2. Set `MONMS_PUBLISH_TOKEN` on **both** staging (outbound publish) and production (import API gate) — same secret, never commit.

Consultants tag structure releases. Clients publish content themselves — consultants are not in the routine content loop.

## Getting started

If this workspace was not created yet:

```bash
# From the MonMS repo root
../monms init --workspace .
```

Re-running `init` on an existing workspace skips files that already exist (safe to run again). Fresh workspaces get `.monms/config.json` and `config.example.json` with documented options in `_fieldDocs`.

This workspace ships a `.gitignore` that excludes runtime data and secrets:

| Path | Why ignored |
|------|-------------|
| `.pb_data/` | PocketBase SQLite, sessions, local uploads (L3 runtime) |
| `.monms/config.json` | Production URL, publisher emails, optional `allowedHosts` → serve `--origins` |
| `.monms/publish-state.json` | Last-publish checksum (environment-specific) |
| `content/` | Export snapshots — ephemeral by default |

Keep `.monms/config.example.json` committed as the template. Sites may force-add `content/` to Git for audit if desired (`git add -f content/`).

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

## Content publish (content rail)

Editorial records export to `content/{collection}.json` and upsert to production by fixed record ID. Clients trigger publish from **`/_monms/publish`** (Publish to live console) or the editor badge link.

```bash
# Operator / CI fallback (same payload as the Publish button)
monms content export --workspace .
monms content diff --workspace .
monms content publish --workspace .   # POST to productionUrl from .monms/config.json
```

Production accepts inbound payloads at `POST /api/monms/content/import` (Bearer `MONMS_PUBLISH_TOKEN`).

Media on public CDNs: store the **URL in a text field** — see [MEDIA.md](MEDIA.md). Blobs are not copied between staging and production.

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
| [EDITING-GUIDE.md](EDITING-GUIDE.md) | Human inline editing + Publish to live walkthrough |
| [MEDIA.md](MEDIA.md) | CDN URL policy for publishable assets |
| [agent-guide.md](agent-guide.md) | AI agent structure mutation workflow |
| [SECURITY.md](SECURITY.md) | SSH keys, API tokens, git hygiene |
| [DEPLOY-DOCKER.md](DEPLOY-DOCKER.md) | Optional Docker deploy (git-on-volume, L1 image + L2/L3 volumes) |
| [../specs/staging.md](../specs/staging.md) | Environments, roles, content publish |

## Environment variables

| Variable | Purpose |
|----------|---------|
| `MONMS_URL` | Running server URL (e.g. `http://localhost:8090`) |
| `POCKETBASE_ADMIN_TOKEN` | Admin JWT for collection management |
| `MONMS_PUBLISH_TOKEN` | Bearer token for production `POST /api/monms/content/import` (staging publish + production gate) |
| `MONMS_BIN` | Path to the `monms` binary (for pre-commit hook) |
| `MONMS_WORKSPACE` | Override workspace path (default: `./workspace`) |

Publish credentials: `productionUrl` in `.monms/config.json` and `MONMS_PUBLISH_TOKEN` env — configured once by consultant, used by client Publish button.
