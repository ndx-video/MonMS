# Requirements: MonMS — The Agent-Malleable Monolithic CMS

**Defined:** 2026-05-22
**Core Value:** High-performance runtime malleability without build or compilation overhead, compiled into a single executable that uses less than 30MB of RAM.

## v1 Requirements

### Core Runtime Engine

- [x] **ENG-01**: Go binary starts and serves HTTP with PocketBase embedded without any configuration files.
- [x] **ENG-02**: Template cache is populated on first access and invalidated automatically when workspace files change.
- [x] **ENG-03**: fsnotify watches the workspace tree for `.gohtml` changes and clears in-memory template cache on write/create events (D-30; entire workspace, not `templates/` only).
- [x] **ENG-04**: Production mode activates template caching; development mode always reads from disk.
- [x] **ENG-05**: Binary's idle RAM footprint stays under 30MB in production mode.
- [x] **ENG-06**: Server-side rendered routes under SQLite reads produce a TTFB under 15ms.

### Workspace Structure

- [x] **WRK-01**: The workspace folder contains `schema/`, `templates/layouts/`, `templates/fragments/`, and `assets/` subdirectories.
- [x] **WRK-02**: The workspace folder is a Git repository and all agent mutations are tracked as commits.
- [x] **WRK-03**: Static assets (`/assets/{path}`) are served directly from `workspace/assets/` without compilation.
- [x] **WRK-04**: Route templates are resolved as `workspace/templates/{slug}.gohtml` and merged with `workspace/templates/layouts/base.gohtml`.
- [x] **WRK-05**: An unknown slug returns a descriptive 404 rather than a Go panic.

### Agent Mutation Engine

- [x] **AGT-01**: Agent can create a new PocketBase collection by POSTing to `/api/collections` without restarting the binary.
- [x] **AGT-02**: Agent can modify an existing `*.gohtml` template file and the change is visible on the next browser request without restart.
- [x] **AGT-03**: Before committing, agent runs Go HTML template dry-run validation on modified templates.
- [x] **AGT-04**: Before committing, agent runs an HTML structure linter on modified templates.
- [x] **AGT-05**: On validation failure, agent performs `git checkout -- .` to restore the last stable workspace state.
- [x] **AGT-06**: All agent file mutations are committed to git with descriptive commit messages.

### Inline Contextual Editing

- [x] **ICE-01**: Authenticated users see a floating "Live Editor Active" badge with a link to the PocketBase admin dashboard.
- [x] **ICE-02**: `IsLoggedIn` context variable is correctly set when a PocketBase session cookie is present.
- [x] **ICE-03**: Template regions can expose `contenteditable="true"` conditionally when `.IsLoggedIn` is true.
- [x] **ICE-04**: HTMX PUT requests on `blur` event transmit updated field content to PocketBase REST endpoint.
- [x] **ICE-05**: HTMX requests include Authorization Bearer token via server-injected `AuthToken` in `htmx:configRequest` (HttpOnly session cookie — not read from JavaScript).
- [x] **ICE-06**: Unauthenticated users see the page normally with no `contenteditable` attributes rendered.

### Authentication & Security

- [x] **SEC-01**: PocketBase admin dashboard (`/_/`) is accessible for full management fallback.
- [x] **SEC-02**: Unauthenticated PUT requests to PocketBase collection endpoints are rejected at the database layer.
- [x] **SEC-03**: Agent operates with SSH keys and REST API tokens scoped strictly to the active workspace subdirectory.
- [x] **SEC-04**: HttpOnly cookie contains the session JWT; it is not accessible from JavaScript.

### Demonstration Content

- [x] **DEMO-01**: A working `hero_content` PocketBase collection with `title` and `body` fields is seeded.
- [x] **DEMO-02**: An `index.gohtml` template renders hero content from the collection with inline editing when authenticated.
- [x] **DEMO-03**: A `base.gohtml` layout includes global HTMX, Alpine.js script tags, and the editor overlay block.

## v2 Requirements

### Environment & Lifecycle (Phase 4)

- [x] **ENV-01**: Documentation and tooling distinguish four layers (engine, structure, content, audience).
- [x] **ENV-02**: Structure promotion uses workspace Git tags; content promotion is a separate rail.
- [x] **ENV-03**: Staging and production are separate MonMS instances with separate `.pb_data/` directories.

### Content Publish (Phase 4)

- [x] **PUB-01**: Editorial collections marked `"editorial": true` in schema JSON.
- [x] **PUB-02**: `workspace/content/*.json` holds exported editorial records with stable IDs.
- [x] **PUB-03**: `monms content export` writes editorial snapshots to `workspace/content/`.
- [x] **PUB-04**: `monms content import` upserts records idempotently by ID.
- [x] **PUB-05**: Production exposes `POST /api/monms/content/import` with scoped publish token.
- [x] **PUB-06**: Staging admin UI includes **Publish to live** with diff preview.
- [x] **PUB-07**: Publisher role gates the publish action; editors may edit without publishing.
- [x] **PUB-08**: Staging tracks last-published checksum for unpublished-changes indicator.
- [x] **PUB-09**: `monms content diff` shows pending changes before publish.

### Media (Phase 4)

- [x] **MED-01**: Publishable media uses public CDN URLs stored in content fields; blobs are not copied between environments.
- [x] **MED-02**: Documentation warns against PocketBase-local file storage for publishable assets.

### Extended Malleability

- **EXT-01**: Agent can roll back to any previous git commit state via REST webhook or chatbox command.
- **EXT-02**: Granular per-field cache invalidation instead of full template cache flush.
- **EXT-03**: Agent-managed CSS variables for global theming without file edits.

### Multi-Tenancy

- **MULT-01**: Multiple workspace folders with independent schemas can be served by a single binary instance.
- **MULT-02**: Per-workspace domain routing via Host header.

### Rich Content Editing

- **RICH-01**: Support for Markdown rendering in `contenteditable` regions with HTMX preview fragment.
- **RICH-02**: Image upload via drag-and-drop in inline edit mode — must store public CDN URL in content field (see MED-02; not `.pb_data/storage/` as publish source).

## Out of Scope

| Feature | Reason |
|---------|--------|
| React / Next.js frontend | Violates zero-compilation monolith principle |
| Kubernetes / multi-container orchestration | Contradicts single-binary low-overhead goal; optional single-container Docker recipes (see `workspace/DEPLOY-DOCKER.md`) are supported |
| Node.js local build pipeline | Asset footprint must use native CSS or CDN imports only |
| External database (PostgreSQL, MySQL) | SQLite is embedded and sufficient for target deployment scale |
| Real-time WebSocket push | HTMX polling or SSE are sufficient; full WS adds complexity |
| Full `.pb_data/` backup as content publish | Editorial content uses JSON upsert rail (D-56) |
| Consultant on every content push | Clients self-serve via Publish button (D-53) |
| Blob replication staging → production | Media uses shared CDN URLs (D-55) |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| ENG-01 | Phase 1 | Complete |
| ENG-02 | Phase 1 | Complete |
| ENG-03 | Phase 1 | Complete |
| ENG-04 | Phase 1 | Complete |
| ENG-05 | Phase 1 | Complete |
| ENG-06 | Phase 1 | Complete |
| WRK-01 | Phase 1 | Complete |
| WRK-02 | Phase 1 | Complete |
| WRK-03 | Phase 1 | Complete |
| WRK-04 | Phase 1 | Complete |
| WRK-05 | Phase 1 | Complete |
| AGT-01 | Phase 2 | Complete |
| AGT-02 | Phase 2 | Complete |
| AGT-03 | Phase 2 | Complete |
| AGT-04 | Phase 2 | Complete |
| AGT-05 | Phase 2 | Complete |
| AGT-06 | Phase 2 | Complete |
| ICE-01 | Phase 3 | Complete |
| ICE-02 | Phase 3 | Complete |
| ICE-03 | Phase 3 | Complete |
| ICE-04 | Phase 3 | Complete |
| ICE-05 | Phase 3 | Complete |
| ICE-06 | Phase 3 | Complete |
| SEC-01 | Phase 1 | Complete |
| SEC-02 | Phase 3 | Complete |
| SEC-03 | Phase 2 | Complete |
| SEC-04 | Phase 3 | Complete |
| DEMO-01 | Phase 3 | Complete |
| DEMO-02 | Phase 3 | Complete |
| DEMO-03 | Phase 1 | Complete |
| ENV-01 | Phase 4 | Complete |
| ENV-02 | Phase 4 | Complete |
| ENV-03 | Phase 4 | Complete |
| PUB-01 | Phase 4 | Complete |
| PUB-02 | Phase 4 | Complete |
| PUB-03 | Phase 4 | Complete |
| PUB-04 | Phase 4 | Complete |
| PUB-05 | Phase 4 | Complete |
| PUB-06 | Phase 4 | Complete |
| PUB-07 | Phase 4 | Complete |
| PUB-08 | Phase 4 | Complete |
| PUB-09 | Phase 4 | Complete |
| MED-01 | Phase 4 | Complete |
| MED-02 | Phase 4 | Complete |

**Coverage:**
- v1 requirements: 29 total — mapped to phases 1–3
- v2 requirements (Phase 4): 14 total — ENV, PUB, MED
- Unmapped: 0 ✓

---
*Last updated: 2026-05-26 after Phase 4 execution and doc reconciliation*
