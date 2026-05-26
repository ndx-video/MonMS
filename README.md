# MonMS

**Monolithic Management System** — an agent-malleable, single-binary CMS built on Go and PocketBase.

MonMS treats database schemas, UI structures, and content as related but separable layers. AI agents and consultants shape the site in a Git-tracked `workspace/`; business clients edit copy in place on staging and publish to production themselves — no rebuild, no consultant in the loop for routine updates.

## Four layers

MonMS has four distinct layers. Each has different actors, artifacts, and promotion paths:

| Layer | Who | What changes | Promoted how |
|-------|-----|--------------|--------------|
| **1 — Engine** | MonMS developers | Go runtime, router, validation | Semver `monms` binary release |
| **2 — Structure** | Consultants, AI agents | Templates, CSS, schema definitions | **Git tag** on workspace repo |
| **3 — Content** | Business clients (editors/publishers) | Copy, records, page text | **JSON upsert** via Publish button |
| **4 — Audience** | End users | Nothing — read/interact only | Production URL |

The binary (L1) is **frozen at deploy**. Structure and content promote on **separate rails** — Git does not carry editorial copy.

Full lifecycle spec: [specs/staging.md](specs/staging.md)

## Why MonMS

| Pillar | What it means |
|--------|---------------|
| **Single-binary monolith** | CMS, SQLite database, file server, and web server in one executable targeting **< 30 MB RAM** idle |
| **Zero-compilation malleability** | Templates and schemas change on disk and load on the next request — no Node build step, no binary rebuild |
| **Git-managed structure** | Layout and schema changes are versioned, validated, and tagged for production |
| **Client-driven content publish** | Editors push approved copy from staging to live via an admin button — consultants aren't pinged every time |
| **Inline contextual editing** | Logged-in editors click and type directly on the page; HTMX saves to PocketBase on blur |

## Staging and production

Typical usage runs **two MonMS instances**:

| Environment | Purpose |
|-------------|---------|
| **Staging** | Consultants shape structure; clients edit content |
| **Production** | Audience sees the live site |

```
Structure rail:  workspace Git tag  ──→  production deploy
Content rail:    editorial JSON upsert  ──→  production records (Publish button)
Media rail:      shared public CDN URLs  ──→  no blob copy (URLs in content only)
```

- **Consultants** tag structure releases (new pages, collections, layouts) — infrequent.
- **Clients** use **Publish to live** at `/api/monms/publish` after editing on staging — frequent.
- **Media** on public buckets: content records store CDN URLs; blobs stay put, only strings sync.

> **v2 (Phase 4):** staging/production split, content CLI, and Publish console are implemented — see [Content CLI](#content-cli-v2) below.

## Quick start

### Prerequisites

- Go 1.25+
- Git (for workspace versioning and pre-commit validation)

### Build and initialize

```bash
git clone <repo-url> monms && cd monms
go build -o monms .

# Scaffold a Git-tracked workspace (templates, assets, schema, pre-commit hook)
./monms init
```

### Run (development)

```bash
./monms
```

The server starts on **http://127.0.0.1:8090** (PocketBase default).

| URL | Purpose |
|-----|---------|
| `/` | Public homepage with hero content |
| `/_/` | PocketBase admin dashboard |
| `/assets/*` | Static files from `workspace/assets/` |

### Run (production)

Build with compile-time production mode to enable template caching and fsnotify invalidation:

```bash
go build -ldflags "-X main.buildMode=production" -o monms .
./monms
```

### First-time admin setup

1. Open `http://127.0.0.1:8090/_/` and create a superuser account.
2. Sign in, then visit `/` — you should see the **Live Editor Active** badge.
3. Click the hero headline or paragraph, edit inline, and blur to save.

See [workspace/EDITING-GUIDE.md](workspace/EDITING-GUIDE.md) for the full walkthrough.

## CLI commands

```bash
monms                          # Start the server
monms init [--workspace PATH]  # Scaffold workspace (default: ./workspace)
monms validate [--workspace PATH] [files...]  # Dry-run template + HTML validation
monms content <subcommand> [--workspace PATH]  # Editorial export/import/diff/publish (v2)
```

### Content CLI (v2)

Editorial content promotes separately from structure Git tags. Operators and CI use `monms content`; clients normally use the **Publish to live** console.

| Subcommand | Purpose |
|------------|---------|
| `monms content export` | Snapshot editorial collections → `workspace/content/*.json` |
| `monms content import` | Upsert from `workspace/content/*.json` → local `.pb_data/` |
| `monms content diff` | Show records/fields changed since last publish (exit 1 if pending) |
| `monms content publish` | Export staging + POST to production (CI/consultant fallback) |

Production import endpoint (Bearer `MONMS_PUBLISH_TOKEN`):

```
POST /api/monms/content/import
```

Staging publish UI: `GET/POST /api/monms/publish` (publisher allowlist in `workspace/.monms/config.json`).

See [specs/staging.md](specs/staging.md) and [workspace/README.md](workspace/README.md) for four-layer lifecycle, dual rails, and environment setup.

### Configuration

| Input | Precedence | Default |
|-------|------------|---------|
| `--workspace` flag | Highest | `./workspace` |
| `MONMS_WORKSPACE` env | Second | `./workspace` |
| (unset) | — | `./workspace` |

PocketBase data (SQLite, uploads, logs) lives at `{workspace}/.pb_data/` — **never commit this directory**.

## Architecture

```
monms/                          # Generic Go binary (frozen at deploy)
├── main.go                     # PocketBase bootstrap, route registration
├── internal/
│   ├── config/                 # Workspace path resolution
│   ├── router/                 # SSR, assets, fragments, auth
│   ├── schema/                 # Declarative schema sync + seed
│   ├── scaffold/               # monms init + embedded templates
│   ├── templates/              # TemplateCache, slug resolver, fsnotify watcher
│   ├── validate/               # monms validate CLI
│   └── workspace/              # Workspace structure validation
└── workspace/                  # Git-tracked site structure (mutable)
    ├── schema/                 # Collection definitions (L2 — structure rail)
    ├── content/                # Editorial record exports (L3 — content rail, v2)
    ├── templates/
    │   ├── layouts/base.gohtml  # Global layout, HTMX, editor badge
    │   ├── fragments/          # HTMX partials (no base layout)
    │   └── {slug}.gohtml       # Route templates
    ├── assets/                 # CSS, fonts, static media (L2)
    └── .pb_data/               # PocketBase SQLite (L3 runtime — gitignored)
```

### Request flow

```
Browser GET /{slug}
  → ResolveSlug (mirror+index rule)
  → TemplateCache.Get (production) or disk read (development)
  → ParseFiles(base.gohtml, page.gohtml)
  → Inject IsLoggedIn, AuthToken, Hero, …
  → Render HTML

Agent commits *.gohtml
  → fsnotify detects change (production)
  → TemplateCache.Flush()
  → Next request loads fresh templates
```

### Slug → template mapping

| URL | Template |
|-----|----------|
| `/` | `templates/index.gohtml` |
| `/about` | `templates/about.gohtml` or `templates/about/index.gohtml` |
| `/press/2024` | `templates/press/2024.gohtml` |

Unknown slugs return a styled 404 — never a Go panic.

## Agent mutations (structure rail)

AI agents modify site **structure** by editing the workspace:

1. **Schema** — `POST /api/collections` (live) + write matching JSON to `schema/` (audit/bootstrap)
2. **Templates** — edit `*.gohtml` files; changes appear on the next request
3. **Validate** — `monms validate` runs Go template dry-run + HTML structure checks
4. **Commit & tag** — pre-commit hook validates; consultant tags for production deploy

Full workflow: [workspace/agent-guide.md](workspace/agent-guide.md)  
Security policy: [workspace/SECURITY.md](workspace/SECURITY.md)

## Development

```bash
# Run all tests
go test ./... -count=1

# Short mode (skips perf/memory gates)
go test ./... -count=1 -short

# Production-mode tests
go test -ldflags "-X main.buildMode=production" ./...
```

### Key packages

| Package | Role |
|---------|------|
| `internal/router` | `/assets`, `/fragments/{name}`, `/{slug...}` SSR |
| `internal/templates` | Cache, slug resolver, fsnotify watcher |
| `internal/schema` | Import `schema/*.json` on bootstrap; seed demo content |
| `internal/validate` | Template dry-run mirroring production `ParseFiles` path |
| `internal/scaffold` | `monms init` workspace bootstrap + pre-commit hook |

### Non-functional targets (v1)

- Idle heap **< 30 MB** (production build)
- SSR TTFB **< 15 ms** p50 (cached, SQLite reads)
- No Node.js build pipeline — native CSS or CDN imports only

## Project status

**v1 (complete):** engine, workspace/Git structure mutation, inline editing on a single instance.

**v2 (Phase 4 — implemented):** staging/production environments, `workspace/content/` JSON sync, client Publish console at `/api/monms/publish`, publisher role — [specs/staging.md](specs/staging.md).

Requirements: [.planning/REQUIREMENTS.md](.planning/REQUIREMENTS.md)  
Roadmap: [.planning/ROADMAP.md](.planning/ROADMAP.md)

## Documentation

| Document | Audience |
|----------|----------|
| [specs/monms-prd.md](specs/monms-prd.md) | Product vision and architecture |
| [specs/staging.md](specs/staging.md) | Four layers, environments, content publish |
| [workspace/README.md](workspace/README.md) | Workspace layout, four layers, dual rails |
| [workspace/MEDIA.md](workspace/MEDIA.md) | CDN URL policy for publishable media |
| [workspace/EDITING-GUIDE.md](workspace/EDITING-GUIDE.md) | Human inline editing walkthrough |
| [workspace/agent-guide.md](workspace/agent-guide.md) | AI agent structure mutation workflow |
| [workspace/SECURITY.md](workspace/SECURITY.md) | SSH scope, token policy, git hygiene |

## Out of scope (v1)

- React/Next.js frontend or Node build pipelines
- External databases (PostgreSQL, MySQL)
- Multi-workspace / multi-tenant routing
- Kubernetes-style container orchestration

See v2 backlog in [.planning/ROADMAP.md](.planning/ROADMAP.md) and [specs/staging.md](specs/staging.md).

## License

See repository license file.
