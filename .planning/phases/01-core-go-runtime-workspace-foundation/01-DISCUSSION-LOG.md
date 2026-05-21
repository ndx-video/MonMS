# Phase 1: Core Go Runtime & Workspace Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-22
**Phase:** 1-Core Go Runtime & Workspace Foundation
**Areas discussed:** Dev vs production mode, Workspace bootstrap, Route mapping, Error pages, Styling baseline, Workspace path

---

## Dev vs Production Mode

| Option | Description | Selected |
|--------|-------------|----------|
| ENV var only | ENV=production enables cache; unset defaults to dev | |
| ENV + CLI flag | --prod overrides ENV for deploy scripts | |
| Auto-detect | Production via compile flags or runtime heuristics | ✓ |

**User's choice:** Auto-detect via compile-time ldflags; default dev
**Notes:** Dev mode always reads templates from disk (no cache). fsnotify watcher runs in production only.

| Option | Description | Selected |
|--------|-------------|----------|
| No cache, always disk | Every request re-parses templates | ✓ |
| Cache but invalidate | Same code path as prod, ENV differs | |
| Parse once per request | Middle ground for debugging | |

| Option | Description | Selected |
|--------|-------------|----------|
| Detect go run | /tmp/go-build heuristic | |
| Compile-time ldflags | -X main.buildMode=production in release | ✓ |
| CWD heuristic | go.mod + empty pb_data | |

| Option | Description | Selected |
|--------|-------------|----------|
| Always watch | fsnotify in both modes | |
| Watch production only | Skip goroutine in dev | ✓ |
| Watch both, log | Always on with invalidation logs | |

---

## Workspace Bootstrap

| Option | Description | Selected |
|--------|-------------|----------|
| Pre-seeded in repo | workspace/ committed in engine repo | |
| Auto-create on first run | Binary scaffolds if missing | |
| Separate init command | monms init required before serve | ✓ |

**User's choice:** monms init scaffolds workspace; binary fatal-exits if missing

| Option | Description | Selected |
|--------|-------------|----------|
| Always git init | init always runs git init + initial commit | |
| Git init if not repo | Only when .git absent | ✓ |
| No git in init | Operator runs git separately | |

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal skeleton | Empty dirs + base.gohtml only | |
| Include index.gohtml | Placeholder homepage | |
| Full Phase 1 scaffold | base, index, main.css, schema/, fragments/ | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Fatal exit | Print "Run monms init", exit 1 | ✓ |
| Start admin only | PocketBase /_/ without SSR | |
| Auto-init in dev only | Silent init in dev mode | |

---

## Route Mapping

| Option | Description | Selected |
|--------|-------------|----------|
| Flat slugs only | /about → about.gohtml | |
| Nested paths | /press/2024 supported | ✓ |
| Flat + index fallback | Single-segment only | |

| Option | Description | Selected |
|--------|-------------|----------|
| Mirror URL path | templates/press/2024.gohtml | |
| Flatten with dashes | press-2024.gohtml | |
| Mirror + index | /press → press/index.gohtml | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| index.gohtml at root | / → templates/index.gohtml | ✓ |
| index/index.gohtml | Nested homepage | |
| Try both | Fallback chain | |

| Option | Description | Selected |
|--------|-------------|----------|
| Exclude known prefixes | /api, /_/, /assets skip SSR | ✓ |
| Catch-all tries template first | Fall through to PocketBase | |
| Strict router order | Register reserved routes first | |

| Option | Description | Selected |
|--------|-------------|----------|
| Normalize trailing slash | Strip before lookup | ✓ |
| Strict match | /about/ may 404 | |
| Redirect 301 | Canonical without slash | |

| Option | Description | Selected |
|--------|-------------|----------|
| Not SSR routes | fragments/ partials only | |
| /fragments/{name} | HTMX swap targets in Phase 1 | ✓ |
| Defer to Phase 3 | Folder only for now | |

| Option | Description | Selected |
|--------|-------------|----------|
| Lowercase only | /About → 404 | |
| Case-insensitive | Fold case on lookup | |
| Preserve case | Filesystem rules apply | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| 500 descriptive | Safe in prod, detailed in dev | ✓ |
| Treat as 404 | Same as missing slug | |
| 500 always detailed | Show parse error in body | |

---

## Error Pages

| Option | Description | Selected |
|--------|-------------|----------|
| Styled workspace template | errors.gohtml in base layout | ✓ |
| Minimal built-in HTML | Hardcoded in Go | |
| Plain text | text/plain descriptive 404 | |

| Option | Description | Selected |
|--------|-------------|----------|
| Fallback built-in | If errors template missing | ✓ |
| Required in init | Startup error if missing | |
| Always built-in Phase 1 | No workspace 404 yet | |

| Option | Description | Selected |
|--------|-------------|----------|
| Show attempted path | "Page not found: /path" + home link | ✓ |
| Generic prod only | Path in dev mode | |
| Branded minimal | No path disclosure | |

| Option | Description | Selected |
|--------|-------------|----------|
| Separate 500.gohtml | Dedicated template | |
| Built-in 500 only | Go-rendered inline | |
| Reuse errors layout | Single errors.gohtml with Code/Message | ✓ |

---

## Styling Baseline

| Option | Description | Selected |
|--------|-------------|----------|
| Native CSS only | main.css hand-written | |
| CDN Tailwind | Utility classes in templates | |
| Hybrid | CDN preflight + main.css overrides | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Tailwind v4 CDN | cdn.tailwindcss.com | |
| Tailwind v3 play CDN | Stable v3 play CDN | |
| Preflight only | CDN preflight + inline config | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Design tokens | CSS variables in main.css | |
| Component classes | .btn, .card, .hero in main.css | ✓ |
| Reset + typography | Tailwind utilities for layout | |

| Option | Description | Selected |
|--------|-------------|----------|
| Include inert | Alpine loaded, no x-data in Phase 1 | |
| Skip Phase 1 | Add in Phase 3 | |
| Minimal demo | Mobile nav toggle in index stub | ✓ |

---

## Workspace Path

| Option | Description | Selected |
|--------|-------------|----------|
| Fixed ./workspace | Relative to CWD | ✓ |
| Relative to binary | Next to executable | |
| Try both | Binary dir then CWD | |

| Option | Description | Selected |
|--------|-------------|----------|
| MONMS_WORKSPACE env only | Environment override | |
| CLI flag only | --workspace flag | |
| Both env and flag | Flag overrides env | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Fixed ./pb_data | Sibling to CWD | |
| MONMS_DATA env | Configurable data dir | |
| Inside workspace | workspace/.pb_data/ | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Strict path clean | Reject .. segments | |
| Root jail | Verify under workspace/assets/ | ✓ |
| Allow symlinks | Trust operator | |

| Option | Description | Selected |
|--------|-------------|----------|
| Same flags | init uses --workspace and env | ✓ |
| Init CWD only | Always ./workspace in CWD | |
| Init requires explicit path | No default | |

| Option | Description | Selected |
|--------|-------------|----------|
| templates/ only | Watch templates subdirectory | |
| templates/ recursively | layouts, fragments, errors | |
| Whole workspace | All .gohtml under workspace | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Gitignore pb_data | init adds workspace/.gitignore | |
| Commit .gitkeep | Structure tracked, data ignored | |
| Operator manages | No gitignore from init | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Validate structure | Require base.gohtml and assets/ | ✓ |
| Directory only | Exists as directory | |
| Init marker file | .monms-init.json required | |

| Option | Description | Selected |
|--------|-------------|----------|
| Log absolute paths | Resolved paths only | |
| Log configured only | Env/flag value | |
| Log both | Configured + resolved absolute | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| JSON backups only | No runtime read Phase 1 | |
| Declarative load | Sync PocketBase from schema/*.json on startup | ✓ |
| Placeholder only | Empty until Phase 2 | |

---

## Claude's Discretion

- Tailwind CDN URL and inline preflight config details
- Built-in fallback error HTML markup
- fsnotify debouncing for rapid commits
- Declarative schema sync failure handling (log vs fatal)

## Deferred Ideas

None captured during this session.
