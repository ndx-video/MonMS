# MonMS

**Monolithic Management System** — an agent-malleable, single-binary CMS built on Go and PocketBase.

MonMS treats database schemas, UI structures, and content as related but separable layers. AI agents and consultants **shape** the site in a Git-tracked `site/`; business clients **stage** copy edits on a staging instance and publish to production themselves — no rebuild, no consultant in the loop for routine updates.

## Phases of work

| Phase | Who | Focus |
|-------|-----|-------|
| **Development** | MonMS engine developers | Building this open-source Go project |
| **Shaping** | Consultants, AI agents | Templates, schema, assets in `site/` Git |
| **Staging** | Client editors/publishers | Preparing editorial copy on the staging instance |
| **Production** | Audience; clients publish into | Live site |

Full terminology and lifecycle: [specs/staging.md](specs/staging.md)

## Four layers

MonMS has four distinct layers. Each has different actors, artifacts, and promotion paths:

| Layer | Who | What changes | Promoted how |
|-------|-----|--------------|--------------|
| **1 — Engine** | MonMS developers | Go runtime, router, validation | Semver `monms` binary release |
| **2 — Structure** | Consultants, AI agents | Templates, CSS, schema definitions | **Git tag** → staging + production checkouts (operator policy) |
| **3 — Content** | Business clients (editors/publishers) | Copy, records, page text | **JSON upsert outside Git** from stage site to production site via Publish button |
| **4 — Audience** | End users | Nothing — read/interact only | Production URL |

The binary (engine) is **frozen at deploy**. end users of this product are not concerned with Layer 1. Structure and content promote on **separate rails**, ie. "staging" to "production" site instances — Git carries shape only (between instances); editorial copy lives in `.pb_data/` and never commits. Content does not live in the repo.

Full lifecycle spec: [specs/staging.md](specs/staging.md)

## Why MonMS

| Pillar | What it means |
|--------|---------------|
| **Single-binary monolith** | CMS, API server, file server, and web server in one executable targeting **< 30 MB RAM** idle |
| **Zero-compilation malleability** | Templates and schemas change on disk and load on the next request — no Node build step, no binary rebuild |
| **Git-managed structure** | Layout and schema changes are versioned, validated, and tagged for production |
| **Client-driven content publish** | Editors push approved copy from staging to live via an admin button — consultants aren't pinged every time |
| **Inline contextual editing** | Logged-in editors click and type directly on the page; HTMX saves to PocketBase on blur |

## Staging and production

Typical usage runs **two MonMS instances** for client content work:

| Instance | Phase | Purpose |
|----------|-------|---------|
| **Staging** | Staging | Clients edit and preview editorial copy |
| **Production** | Production | Audience sees the live site |

**Shaping** (templates, schema) happens on a **site** Git checkout. When ready, the consultant tags the repo; an operator-chosen policy pulls that tag into **both** instances — for example GitHub Actions, cron calling `monms site sync`, or optional `shapeSync` in config at serve startup.

```
Structure rail:  site Git tag  ──→  staging + production checkouts (operator policy)
Content rail:    editorial JSON upsert (outside Git)  ──→  production records (Publish button)
Media rail:      shared public CDN URLs  ──→  no blob copy (URLs in content only)
```

- **Consultants** shape structure and tag releases — infrequent.
- **Clients** use **Publish to live** at `/_monms/publish` after editing on staging — frequent.
- **Media** on public buckets: content records store CDN URLs; blobs stay put, only strings sync.

> **v2 (Phase 4):** staging/production split, content CLI, and Publish console are implemented — see [Content CLI](#content-cli-v2) below.

## Quick start

### Prerequisites

- Go 1.25+
- Git (for site versioning and pre-commit validation)

### Build and run

```bash
git clone <repo-url> monms && cd monms
go build -o monms .

# First run: scaffolds site/ interactively when missing, configures listen settings,
# and offers to start the server (or use monms init explicitly)
./monms serve

# Non-interactive / CI: scaffold first, then serve
./monms init -s ./site
./monms serve
```

### Run (local engine build)

```bash
./monms
```

This starts a server using a **development build** of the binary (`buildMode=development` — no template cache). That compile-time flag is unrelated to the product **development** phase (contributing to the MonMS engine repo).

The server starts on **http://127.0.0.1:8090** (PocketBase default).

| URL | Purpose |
|-----|---------|
| `/` | Public homepage with hero content |
| `/_/` | PocketBase admin dashboard |
| `/assets/*` | Static files from `site/assets/` |

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

See [docs/user-guide/inline-editing.md](docs/user-guide/inline-editing.md) for the full walkthrough.

## CLI commands

```bash
monms                          # Start the server
monms init [--site PATH]  # Scaffold site (default: ./site)
monms validate [--site PATH] [files...]  # Dry-run template + HTML validation
monms content <subcommand> [--site PATH]  # Editorial export/import/diff/publish (v2)
monms site sync --ref TAG [--site PATH]  # Shape sync (fetch + checkout)
```

### Content CLI (v2)

Editorial content promotes separately from structure Git tags. Operators and CI use `monms content`; clients normally use the **Publish to live** console.

| Subcommand | Purpose |
|------------|---------|
| `monms content export` | Snapshot editorial collections → `site/content/*.json` |
| `monms content import` | Upsert from `site/content/*.json` → local `.pb_data/` |
| `monms content diff` | Show records/fields changed since last publish (exit 1 if pending) |
| `monms content publish` | Export staging + POST to production (CI/consultant fallback) |

Production import endpoint (Bearer `MONMS_PUBLISH_TOKEN`):

```
POST /api/monms/content/import
```

Staging publish UI: `GET/POST /_monms/publish` (publisher allowlist in `site/.monms/config.json`).

See [docs/operators/getting-started.md](docs/operators/getting-started.md) and [site/README.md](site/README.md) for four-layer lifecycle, dual rails, and environment setup.

### Configuration

| Input | Precedence | Default |
|-------|------------|---------|
| `-s`, `--site` flag | Highest | `./site` |
| `MONMS_SITE` env | Second | `./site` |
| (unset) | — | `./site` |

PocketBase data (SQLite, uploads, logs) lives at `{site}/.pb_data/` — **never commit this directory**.

### File logging

MonMS writes rotated logs to `{site}/.monms/logs/`. Configure levels in `site/.monms/config.json`:

```json
"logging": ["error", "warn", "schema"]
```

| Build | Default when `logging` omitted |
|-------|-------------------------------|
| Production (`buildMode=production`) | `error`, `warn`, `schema` |
| Development (`go build`) | all levels (`error`, `warn`, `info`, `debug`, `schema`) |

ERROR is always written to `error.log`; PocketBase output always goes to `pocketbase.log`. Full level reference: [docs/reference/logging.md](docs/reference/logging.md).

Console/SQL verbosity is separate: `monms serve --dev` (CLI flag only).

## Architecture

```
monms/                          # Generic Go binary (frozen at deploy)
├── main.go                     # PocketBase bootstrap, route registration
├── internal/
│   ├── config/                 # Site path resolution
│   ├── content/                # Editorial export/import/diff/publish (v2)
│   ├── router/                 # SSR, assets, fragments, auth
│   ├── schema/                 # Declarative schema sync, seed, editorial flag
│   ├── scaffold/               # monms init + embedded templates
│   ├── templates/              # TemplateCache, slug resolver, fsnotify watcher
│   ├── validate/               # monms validate CLI
│   └── site/                   # Site structure validation
└── site/                  # Git-tracked site structure (mutable)
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

AI agents modify site **structure** by editing the `site/` tree:

1. **Schema** — `POST /api/collections` (live) + write matching JSON to `schema/` (audit/bootstrap)
2. **Templates** — edit `*.gohtml` files; changes appear on the next request
3. **Validate** — `monms validate` runs Go template dry-run + HTML structure checks
4. **Commit & tag** — pre-commit hook validates; consultant tags for shape deploy to staging + production

Full workflow: [docs/operators/shaping-and-agents.md](docs/operators/shaping-and-agents.md)  
Security policy: [docs/operators/security.md](docs/operators/security.md)

## Engine development

Contributing to the MonMS **engine** (this Go repository):

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
| `internal/content` | Editorial export/import, diff, publish API + UI (v2) |
| `internal/router` | `/assets`, `/fragments/{name}`, `/{slug...}` SSR |
| `internal/templates` | Cache, slug resolver, fsnotify watcher |
| `internal/schema` | Import `schema/*.json` on bootstrap; seed demo content |
| `internal/validate` | Template dry-run mirroring production `ParseFiles` path |
| `internal/scaffold` | `monms init` site bootstrap + pre-commit hook |

### Non-functional targets (v1)

- Idle heap **< 30 MB** (production build)
- SSR TTFB **< 15 ms** p50 (cached, SQLite reads)
- No Node.js build pipeline — native CSS or CDN imports only

## Project status

**v1 (complete):** engine, site/Git structure mutation, inline editing on a single instance.

**v2 (Phase 4 — implemented):** staging/production environments, `site/content/` JSON sync, client Publish console at `/_monms/publish`, publisher role — [docs/operators/getting-started.md](docs/operators/getting-started.md).

Requirements: [.planning/REQUIREMENTS.md](.planning/REQUIREMENTS.md)  
Roadmap: [.planning/ROADMAP.md](.planning/ROADMAP.md)

## Documentation

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

Official documentation lives in [`docs/`](docs/README.md). A GitHub Pages site will mirror this tree (URL TBD).

| Document | Audience |
|----------|----------|
| [docs/README.md](docs/README.md) | Documentation roadmap |
| [docs/user-guide/](docs/user-guide/inline-editing.md) | Content editors and publishers |
| [docs/operators/](docs/operators/getting-started.md) | Shapers, consultants, administrators |
| [docs/reference/monms-api.md](docs/reference/monms-api.md) | MonMS HTTP API (not PocketBase) |
| [site/README.md](site/README.md) | Site directory layout only |

Legacy specs in [`specs/`](specs/staging.md) are deprecated.

## Out of scope

The following are not part of the core MonMS model (see v2 backlog in [.planning/ROADMAP.md](.planning/ROADMAP.md)):

- React/Next.js frontend or Node build pipelines
- External databases (PostgreSQL, MySQL)
- Multi-site / multi-tenant routing (MULT-* backlog)
- Kubernetes-style container orchestration

See [docs/operators/getting-started.md](docs/operators/getting-started.md) for implemented v2 lifecycle vs remaining backlog.

## License

See repository license file.
