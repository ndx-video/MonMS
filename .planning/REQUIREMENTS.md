# Requirements: MonMS — The Agent-Malleable Monolithic CMS

**Defined:** 2026-05-22
**Core Value:** High-performance runtime malleability without build or compilation overhead, compiled into a single executable that uses less than 30MB of RAM.

## v1 Requirements

### Core Runtime Engine

- [x] **ENG-01**: Go binary starts and serves HTTP with PocketBase embedded without any configuration files.
- [x] **ENG-02**: Template cache is populated on first access and invalidated automatically when workspace files change.
- [x] **ENG-03**: fsnotify watches the `workspace/templates/` folder and clears in-memory template cache on write/create events.
- [x] **ENG-04**: Production mode activates template caching; development mode always reads from disk.
- [ ] **ENG-05**: Binary's idle RAM footprint stays under 30MB in production mode.
- [ ] **ENG-06**: Server-side rendered routes under SQLite reads produce a TTFB under 15ms.

### Workspace Structure

- [x] **WRK-01**: The workspace folder contains `schema/`, `templates/layouts/`, `templates/fragments/`, and `assets/` subdirectories.
- [x] **WRK-02**: The workspace folder is a Git repository and all agent mutations are tracked as commits.
- [ ] **WRK-03**: Static assets (`/assets/{path}`) are served directly from `workspace/assets/` without compilation.
- [ ] **WRK-04**: Route templates are resolved as `workspace/templates/{slug}.gohtml` and merged with `workspace/templates/layouts/base.gohtml`.
- [ ] **WRK-05**: An unknown slug returns a descriptive 404 rather than a Go panic.

### Agent Mutation Engine

- [ ] **AGT-01**: Agent can create a new PocketBase collection by POSTing to `/api/collections` without restarting the binary.
- [ ] **AGT-02**: Agent can modify an existing `*.gohtml` template file and the change is visible on the next browser request without restart.
- [ ] **AGT-03**: Before committing, agent runs Go HTML template dry-run validation on modified templates.
- [ ] **AGT-04**: Before committing, agent runs an HTML structure linter on modified templates.
- [ ] **AGT-05**: On validation failure, agent performs `git checkout -- .` to restore the last stable workspace state.
- [ ] **AGT-06**: All agent file mutations are committed to git with descriptive commit messages.

### Inline Contextual Editing

- [ ] **ICE-01**: Authenticated users see a floating "Live Editor Active" badge with a link to the PocketBase admin dashboard.
- [ ] **ICE-02**: `IsLoggedIn` context variable is correctly set when a PocketBase session cookie is present.
- [ ] **ICE-03**: Template regions can expose `contenteditable="true"` conditionally when `.IsLoggedIn` is true.
- [ ] **ICE-04**: HTMX PUT requests on `blur` event transmit updated field content to PocketBase REST endpoint.
- [ ] **ICE-05**: HTMX request includes Authorization Bearer token extracted from `pb_auth` cookie via JavaScript.
- [ ] **ICE-06**: Unauthenticated users see the page normally with no `contenteditable` attributes rendered.

### Authentication & Security

- [x] **SEC-01**: PocketBase admin dashboard (`/_/`) is accessible for full management fallback.
- [ ] **SEC-02**: Unauthenticated PUT requests to PocketBase collection endpoints are rejected at the database layer.
- [ ] **SEC-03**: Agent operates with SSH keys and REST API tokens scoped strictly to the active workspace subdirectory.
- [ ] **SEC-04**: HttpOnly cookie contains the session JWT; it is not accessible from JavaScript.

### Demonstration Content

- [ ] **DEMO-01**: A working `hero_content` PocketBase collection with `title` and `body` fields is seeded.
- [ ] **DEMO-02**: An `index.gohtml` template renders hero content from the collection with inline editing when authenticated.
- [x] **DEMO-03**: A `base.gohtml` layout includes global HTMX, Alpine.js script tags, and the editor overlay block.

## v2 Requirements

### Extended Malleability

- **EXT-01**: Agent can roll back to any previous git commit state via REST webhook or chatbox command.
- **EXT-02**: Granular per-field cache invalidation instead of full template cache flush.
- **EXT-03**: Agent-managed CSS variables for global theming without file edits.

### Multi-Tenancy

- **MULT-01**: Multiple workspace folders with independent schemas can be served by a single binary instance.
- **MULT-02**: Per-workspace domain routing via Host header.

### Rich Content Editing

- **RICH-01**: Support for Markdown rendering in `contenteditable` regions with HTMX preview fragment.
- **RICH-02**: Image upload via drag-and-drop in inline edit mode, persisted via PocketBase file fields.

## Out of Scope

| Feature | Reason |
|---------|--------|
| React / Next.js frontend | Violates zero-compilation monolith principle |
| Kubernetes / containerized deploy | Contradicts single-binary low-overhead goal |
| Node.js local build pipeline | Asset footprint must use native CSS or CDN imports only |
| External database (PostgreSQL, MySQL) | SQLite is embedded and sufficient for target deployment scale |
| Real-time WebSocket push | HTMX polling or SSE are sufficient; full WS adds complexity |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| ENG-01 | Phase 1 | Complete |
| ENG-02 | Phase 1 | Complete |
| ENG-03 | Phase 1 | Complete |
| ENG-04 | Phase 1 | Complete |
| ENG-05 | Phase 1 | Pending |
| ENG-06 | Phase 1 | Pending |
| WRK-01 | Phase 1 | Complete |
| WRK-02 | Phase 1 | Complete |
| WRK-03 | Phase 1 | Pending |
| WRK-04 | Phase 1 | Pending |
| WRK-05 | Phase 1 | Pending |
| AGT-01 | Phase 2 | Pending |
| AGT-02 | Phase 2 | Pending |
| AGT-03 | Phase 2 | Pending |
| AGT-04 | Phase 2 | Pending |
| AGT-05 | Phase 2 | Pending |
| AGT-06 | Phase 2 | Pending |
| ICE-01 | Phase 3 | Pending |
| ICE-02 | Phase 3 | Pending |
| ICE-03 | Phase 3 | Pending |
| ICE-04 | Phase 3 | Pending |
| ICE-05 | Phase 3 | Pending |
| ICE-06 | Phase 3 | Pending |
| SEC-01 | Phase 1 | Complete |
| SEC-02 | Phase 3 | Pending |
| SEC-03 | Phase 2 | Pending |
| SEC-04 | Phase 3 | Pending |
| DEMO-01 | Phase 3 | Pending |
| DEMO-02 | Phase 3 | Pending |
| DEMO-03 | Phase 1 | Complete |

**Coverage:**
- v1 requirements: 29 total
- Mapped to phases: 29
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-22*
*Last updated: 2026-05-22 after initial definition*
