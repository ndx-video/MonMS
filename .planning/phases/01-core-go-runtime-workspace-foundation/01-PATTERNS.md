# Phase 1: Core Go Runtime & Workspace Foundation - Pattern Map

**Mapped:** 2026-05-22
**Files analyzed:** 24 (engine + workspace scaffold + tests)
**Analogs found:** 0 codebase / 24 PRD+RESEARCH reference

> **Greenfield note:** No Go source exists in the repository. All pattern assignments reference `specs/monms-prd.md` §2–§3, §5 (base layout), and `01-RESEARCH.md` code examples. Treat PRD §3 as the starting skeleton; CONTEXT.md decisions override PRD defaults throughout.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `main.go` | entry / bootstrap | request-response | PRD §3 `main()` (lines 75–118) | reference-adapt |
| `go.mod` / `go.sum` | config | — | RESEARCH Standard Stack | reference-only |
| `internal/config/config.go` | utility | config | RESEARCH Pattern 2 (NewWithConfig) | reference-only |
| `internal/workspace/validate.go` | utility | file-I/O | RESEARCH Workspace Validation | reference-only |
| `internal/templates/cache.go` | service | transform | PRD §3 `TemplateCache` + `getTemplate()` (62–146) | reference-adapt |
| `internal/templates/resolver.go` | utility | file-I/O | RESEARCH Pattern 3 (Mirror + Index) | reference-only |
| `internal/templates/watcher.go` | service | event-driven | PRD §3 `watchWorkspace()` (148–172) + RESEARCH debounce | reference-adapt |
| `internal/router/ssr.go` | route / handler | request-response | PRD §3 SSR catch-all (88–110) | reference-adapt |
| `internal/router/assets.go` | route / handler | file-I/O | PRD §3 `/assets/{path...}` (83–86) + RESEARCH root-jail | reference-adapt |
| `internal/router/fragments.go` | route / handler | request-response | RESEARCH OnServe fragment route | reference-only |
| `internal/schema/sync.go` | service | CRUD | RESEARCH Pattern 5 (OnBootstrap import) | reference-only |
| `internal/scaffold/init.go` | service / CLI | file-I/O | RESEARCH Pattern 1 (early dispatch) | reference-only |
| `internal/scaffold/embed/*` | config / embed | file-I/O | RESEARCH embed scaffold | reference-only |
| `workspace/templates/layouts/base.gohtml` | template | request-response | PRD §5 base layout (235–279) | reference-adapt |
| `workspace/templates/index.gohtml` | template | request-response | PRD §5 index (285–314) | reference-adapt |
| `workspace/templates/errors/errors.gohtml` | template | request-response | CONTEXT D-17–D-20 | no-prd-analog |
| `workspace/assets/main.css` | static asset | file-I/O | CONTEXT D-21–D-22 | no-prd-analog |
| `workspace/schema/` | config | CRUD | PRD §2 topology (35) | reference-only |
| `internal/templates/*_test.go` | test | — | RESEARCH Validation Architecture | reference-only |
| `internal/router/*_test.go` | test | — | RESEARCH Validation Architecture | reference-only |
| `internal/scaffold/*_test.go` | test | — | RESEARCH Validation Architecture | reference-only |

## Pattern Assignments

### `main.go` (entry, request-response)

**Analog:** PRD §3 `main()` — `specs/monms-prd.md` lines 75–118

**Imports pattern** (lines 48–60):
```go
import (
    "html/template"
    "log"
    "os"
    "path/filepath"
    "sync"

    "github.com/fsnotify/fsnotify"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)
```

**PocketBase bootstrap + OnServe registration** (lines 75–113):
```go
func main() {
    app := pocketbase.New()

    go watchWorkspace(tplCache)

    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        se.Router.GET("/assets/{path...}", assetsHandler)
        se.Router.GET("/{slug...}", ssrHandler)
        return se.Next()
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

**Adaptation notes (CONTEXT):**
- **D-01:** Replace PRD `os.Getenv("ENV") == "production"` with compile-time ldflags: `var buildMode = "development"` and `-ldflags "-X main.buildMode=production"`.
- **D-05/D-06:** Early-dispatch `init` before PocketBase construction — do not auto-scaffold on serve.
- **D-27:** Use `pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: filepath.Join(wsAbs, ".pb_data")})` instead of `pocketbase.New()`.
- **D-04:** Start fsnotify goroutine only when `buildMode == "production"`.
- **D-09/D-31:** Call `workspace.ValidateWorkspace(ws)` before serve; log configured + absolute workspace path via `slog.Info`.
- **D-32:** Register `app.OnBootstrap()` hook for schema sync (see `internal/schema/sync.go`).
- **Route order:** Register `/assets/{path...}` → `/fragments/{name}` → `/{slug...}` inside `OnServe()` (RESEARCH Phase-Specific #2).

**Integration points:**
- Imports and wires `internal/config`, `internal/workspace`, `internal/templates`, `internal/router`, `internal/schema`, `internal/scaffold`.
- Passes resolved workspace path and `*TemplateCache` into router handlers.

---

### `internal/config/config.go` (utility, config)

**Analog:** RESEARCH Pattern 2 + `--workspace` integration

**Core pattern:**
```go
// ResolveWorkspace parses --workspace from os.Args, then MONMS_WORKSPACE env,
// then defaults to "./workspace". Flag wins over env (D-26).
func ResolveWorkspace(args []string, env []string) (configured string, absolute string, err error) {
    configured = defaultWorkspace // "./workspace"
    // parse --workspace from args (strip before PB eager parse)
    // override from MONMS_WORKSPACE if flag absent
    absolute, err = filepath.Abs(configured)
    return configured, absolute, err
}
```

**Adaptation notes:**
- Same resolution for both `monms init` and `monms serve`.
- Return both configured path (for logging D-31) and absolute path (for jail checks D-29, template paths).

**Integration points:**
- Called from `main.go` (serve path) and `scaffold.RunInit()` (init path).
- Absolute path feeds `DefaultDataDir`, asset root, template root, fsnotify root.

---

### `internal/workspace/validate.go` (utility, file-I/O)

**Analog:** RESEARCH Workspace Validation snippet

**Core pattern** (D-06, D-09):
```go
func ValidateWorkspace(ws string) error {
    checks := []string{
        filepath.Join(ws, "templates/layouts/base.gohtml"),
        filepath.Join(ws, "assets"),
    }
    for _, p := range checks {
        if _, err := os.Stat(p); err != nil {
            return fmt.Errorf("workspace incomplete: missing %s\nRun: monms init", p)
        }
    }
    return nil
}
```

**Error handling:**
- Caller (`main.go`) prints error to stderr and `os.Exit(1)` — never auto-scaffold (D-05).

---

### `internal/templates/cache.go` (service, transform)

**Analog:** PRD §3 `TemplateCache` + `getTemplate()` — lines 62–146

**Struct + global instance** (lines 62–73):
```go
type TemplateCache struct {
    mu     sync.RWMutex
    cache  map[string]*template.Template
    active bool
}

var tplCache = &TemplateCache{
    cache:  make(map[string]*template.Template),
    active: os.Getenv("ENV") == "production", // → replace with buildMode check
}
```

**Cache read + lazy populate** (lines 124–145):
```go
func (c *TemplateCache) getTemplate(slug string, loader func() (*template.Template, error)) (*template.Template, error) {
    c.mu.RLock()
    if c.Active() {
        if cached, exists := c.cache[slug]; exists {
            c.mu.RUnlock()
            return cached, nil
        }
    }
    c.mu.RUnlock()

    tmpl, err := loader()
    if err != nil {
        return nil, err
    }

    c.mu.Lock()
    if c.Active() {
        c.cache[slug] = tmpl
    }
    c.mu.Unlock()
    return tmpl, nil
}
```

**Parse pattern** (lines 120–134):
```go
layoutPath := filepath.Join(ws, "templates/layouts/base.gohtml")
pagePath := resolvedPagePath // from resolver
tmpl, err := template.ParseFiles(layoutPath, pagePath)
```

**Adaptation notes:**
- **D-01/D-02/D-03:** `Active()` returns `buildMode == "production"`; dev mode skips cache entirely (D-02).
- **D-16:** Parse errors bubble to SSR handler for dev-detailed vs prod-generic 500.
- Extract `Flush()` method: `c.cache = make(map[string]*template.Template)` (PRD line 165).
- Cache key: slug string (RESEARCH Template Cache Key Strategy).

**Integration points:**
- Used by `router/ssr.go`, `router/fragments.go` (fragments may bypass layout — see fragments handler).
- Invalidated by `watcher.go` calling `Flush()`.

---

### `internal/templates/resolver.go` (utility, file-I/O)

**Analog:** RESEARCH Pattern 3 (Mirror + Index) — no PRD analog for nested paths

**Core algorithm** (D-10, D-11, D-12):
```go
// ResolveSlug maps URL path to workspace template file path.
// 1. Strip trailing slash (D-12)
// 2. Empty slug → templates/index.gohtml (D-11)
// 3. Try templates/{slug}.gohtml (flat)
// 4. Try templates/{slug}/index.gohtml (directory index)
// Returns ("", ErrNotFound) if neither exists.
func ResolveSlug(ws, slug string) (pagePath string, err error)
```

**Examples (locked in CONTEXT):**
| URL | Resolved file |
|-----|---------------|
| `/` | `templates/index.gohtml` |
| `/press` | `templates/press/index.gohtml` |
| `/press/2024` | `templates/press/2024.gohtml` |

**Adaptation notes:**
- **D-13:** Preserve slug case; rely on filesystem case sensitivity.
- Flat file wins if both flat and directory index exist (RESEARCH Open Question #1).

---

### `internal/templates/watcher.go` (service, event-driven)

**Analog:** PRD §3 `watchWorkspace()` — lines 148–172 + RESEARCH recursive debounce

**PRD baseline (adapt heavily)** (lines 148–171):
```go
func watchWorkspace(c *TemplateCache) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return
    }
    defer watcher.Close()

    _ = watcher.Add("./workspace/templates") // → recursive walk entire workspace

    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
                c.mu.Lock()
                c.cache = make(map[string]*template.Template)
                c.mu.Unlock()
            }
        case <-watcher.Errors:
            return
        }
    }
}
```

**RESEARCH upgrade — recursive + debounce** (01-RESEARCH.md lines 562–619):
```go
func watchWorkspace(ctx context.Context, root string, onChange func()) error {
    // filepath.WalkDir: watcher.Add(path) for every directory
    // On Create dir event: w.Add(newDir)
    // Filter: .gohtml suffix only (D-30)
    // Debounce 100ms per path before calling onChange() → tplCache.Flush()
}
```

**Adaptation notes:**
- **D-04:** Only start watcher in production (`buildMode == "production"`).
- **D-30:** Watch entire workspace tree, not just `templates/` (PRD watches templates only).
- Use `context.Context` for clean shutdown on serve stop.

---

### `internal/router/ssr.go` (route, request-response)

**Analog:** PRD §3 SSR catch-all — lines 88–110

**Core handler pattern** (lines 88–110):
```go
se.Router.GET("/{slug...}", func(e *core.RequestEvent) error {
    slug := e.Request.PathValue("slug")
    if slug == "" {
        slug = "index" // → D-11: resolve to templates/index.gohtml, not slug "index"
    }

    tmpl, err := getTemplate(slug)
    if err != nil {
        return e.NotFoundError("Template variant not found or syntax error", err)
    }

    data := map[string]any{
        "IsLoggedIn": e.Auth != nil,
        "User":       e.Auth,
        "Slug":       slug,
        "App":        app,
    }

    e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
    return tmpl.Execute(e.Response, data) // → ExecuteTemplate(w, "base", data)
})
```

**Adaptation notes:**
- **D-11:** Empty slug maps to `templates/index.gohtml` via resolver, not `index.gohtml` under slug name.
- **D-14:** Reserved prefixes `/api/*`, `/_/*`, `/assets/*` handled by PocketBase or earlier routes — SSR is last registered.
- **D-16/D-17/D-19/D-20:** Missing template → render `errors.gohtml` with `{Code: 404, Message: "Page not found: /path"}`; parse error → 500 with dev detail vs prod generic.
- **D-18:** Fallback minimal built-in HTML if error template missing.
- Always `ExecuteTemplate(w, "base", data)` — PRD uses `Execute()` which is incorrect for multi-file layouts (RESEARCH Pitfall 4).

**Integration points:**
- Calls `templates.ResolveSlug()`, `templates/cache.Get()`, renders via layout+page parse.

---

### `internal/router/assets.go` (route, file-I/O)

**Analog:** PRD §3 static assets — lines 83–86 + RESEARCH root-jail

**PRD baseline** (lines 83–86):
```go
se.Router.GET("/assets/{path...}", func(e *core.RequestEvent) error {
    filePath := filepath.Join("./workspace/assets", e.Request.PathValue("path"))
    return e.File(filePath)
})
```

**RESEARCH root-jail upgrade** (01-RESEARCH.md lines 541–555):
```go
func safeAssetPath(assetsRoot, requestPath string) (string, error) {
    clean := filepath.Clean(filepath.Join(assetsRoot, requestPath))
    absRoot, _ := filepath.Abs(assetsRoot)
    absClean, _ := filepath.Abs(clean)
    if absClean != absRoot && !strings.HasPrefix(absClean, absRoot+string(os.PathSeparator)) {
        return "", errForbidden
    }
    return absClean, nil
}
```

**Adaptation notes:**
- **D-29:** Reject traversal with HTTP 403 before `e.File()`.
- Use resolved workspace absolute path, not hardcoded `./workspace`.

---

### `internal/router/fragments.go` (route, request-response)

**Analog:** RESEARCH OnServe fragment route — no PRD analog

**Core pattern** (D-15):
```go
se.Router.GET("/fragments/{name}", func(e *core.RequestEvent) error {
    name := e.Request.PathValue("name")
    path := filepath.Join(ws, "templates/fragments", name+".gohtml")
    // Parse and render fragment only — no base layout (HTMX swap target)
    // 404 if missing
})
```

**Integration points:**
- Shares `TemplateCache` with SSR but renders partial without `ExecuteTemplate("base")`.

---

### `internal/schema/sync.go` (service, CRUD)

**Analog:** RESEARCH Pattern 5 — no PRD §3 analog

**Core pattern** (D-32):
```go
app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
    if err := e.Next(); err != nil {
        return err
    }
    raw, err := loadSchemaJSONFiles(filepath.Join(ws, "schema"))
    if err != nil || len(raw) == 0 {
        return nil // ImportCollections rejects empty array
    }
    return e.App.ImportCollectionsByMarshaledJSON(raw, false)
})
```

**Adaptation notes:**
- Log-and-continue on invalid JSON per file (RESEARCH Implementation C — Claude's discretion).
- Load files in sorted name order; skip empty `schema/` directory.
- JSON format: PocketBase `migrate collections` snapshot (`name`, `type`, `fields` keys).

**Integration points:**
- Registered in `main.go` before `app.Start()`.
- Reads from workspace; writes to `workspace/.pb_data/` SQLite via PocketBase.

---

### `internal/scaffold/init.go` (service / CLI, file-I/O)

**Analog:** RESEARCH Pattern 1 (Early-Dispatch CLI) — no PRD analog

**Core pattern** (D-05, D-07, D-08):
```go
// Called from main.go before PocketBase construction
func RunInit(args []string) error {
    ws, _, err := config.ResolveWorkspace(args, os.Environ())
    // Write scaffold from embed.FS:
    //   templates/layouts/base.gohtml
    //   templates/index.gohtml
    //   templates/errors/errors.gohtml
    //   templates/fragments/ (dir)
    //   assets/main.css
    //   schema/ (empty or .gitkeep)
    if !dirExists(filepath.Join(ws, ".git")) {
        exec.Command("git", "init", ws).Run() // warn if git missing
    }
    return nil
}
```

**main.go early dispatch:**
```go
func main() {
    if len(os.Args) >= 2 && os.Args[1] == "init" {
        if err := scaffold.RunInit(os.Args[2:]); err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        return
    }
    runServe()
}
```

**Adaptation notes:**
- **D-28:** Do not auto-gitignore `.pb_data/`.
- **D-07:** Full Phase 1 scaffold list is locked.
- Use `embed` for scaffold templates (RESEARCH Standard Stack).

---

### `workspace/templates/layouts/base.gohtml` (template)

**Analog:** PRD §5 base layout — lines 235–279

**Layout define pattern** (lines 235–279):
```html
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}MonMS{{end}}</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link rel="stylesheet" href="/assets/main.css">
</head>
<body>
    <main>
        {{template "body" .}}
    </main>
</body>
</html>
{{end}}
```

**Adaptation notes (Phase 1 scaffold):**
- **D-21:** Add Tailwind Play CDN + minimal inline config for preflight-only; keep `/assets/main.css` for component classes.
- **D-23:** Add Alpine.js CDN with `x-data` mobile nav toggle demo (not full editor UX).
- **D-24:** Include HTMX CDN + editor overlay block structure only — **omit** `{{if .IsLoggedIn}}` auth badge until Phase 3.
- **Omit** HTMX JWT cookie script from PRD §5 (lines 265–276) until Phase 3 inline editing.
- Use `{{define "title"}}` / `{{define "body"}}` blocks in page templates; execute via `ExecuteTemplate(w, "base", data)`.

---

### `workspace/templates/index.gohtml` (template)

**Analog:** PRD §5 index — lines 285–314

**Page template pattern:**
```html
{{define "title"}}Welcome to MonMS{{end}}

{{define "body"}}
<div class="hero">
    <!-- Alpine mobile nav toggle demo (D-23) -->
    <nav x-data="{ open: false }">...</nav>
    <h1>Welcome</h1>
</div>
{{end}}
```

**Adaptation notes:**
- **Do not** include PRD PocketBase record reads (`$record := .App.FindRecordById`) — Phase 1 index is static stub (RESEARCH Phase-Specific #5, keeps ENG-06 measurable).
- Minimal utility classes in template; branded styles in `main.css` (D-22).

---

### `workspace/templates/errors/errors.gohtml` (template)

**Analog:** None in PRD — derive from CONTEXT D-17–D-20

**Pattern:**
```html
{{define "title"}}Error {{.Code}}{{end}}

{{define "body"}}
<div class="error-page">
    <h1>{{.Code}}</h1>
    <p>{{.Message}}</p>
    <a href="/">Return home</a>
</div>
{{end}}
```

**Usage from SSR handler:**
```go
data := map[string]any{
    "Code":    404,
    "Message": fmt.Sprintf("Page not found: %s", r.URL.Path),
}
// Parse base.gohtml + errors.gohtml; ExecuteTemplate(w, "base", data)
```

**Adaptation notes:**
- Same template for 404 and 500 (D-20); only `Code`/`Message` differ.
- **D-18:** Built-in minimal HTML fallback if this file missing at render time.

---

### `workspace/assets/main.css` (static asset)

**Analog:** PRD §5 references `/assets/main.css` — no content spec

**Pattern:**
```css
/* Component classes per D-21/D-22 */
.btn { /* ... */ }
.card { /* ... */ }
.hero { /* ... */ }
.error-page { /* ... */ }
```

**Adaptation notes:**
- Hybrid styling: Tailwind preflight via CDN handles reset; this file holds branded component classes.

---

### Test files (`internal/**/*_test.go`)

**Analog:** RESEARCH Validation Architecture — Phase Requirements → Test Map

**Pattern:**
```go
func TestResolveSlug_PressIndex(t *testing.T) {
    ws := t.TempDir()
    // scaffold minimal workspace fixture
    // assert ResolveSlug(ws, "press") → templates/press/index.gohtml
}
```

**Key test helpers needed (Wave 0):**
- Temp workspace fixture with scaffold files
- Ldflags or build-tag helper for production mode cache tests
- `httptest` for TTFB (ENG-06) and handler integration (ENG-01, WRK-03, WRK-05)

---

## Shared Patterns

### Compile-Time Production Mode (replaces PRD ENV check)

**Source:** CONTEXT D-01; supersedes PRD line 72
**Apply to:** `main.go`, `internal/templates/cache.go`, `internal/templates/watcher.go`

```go
var buildMode = "development" // overridden by -ldflags "-X main.buildMode=production"

func isProduction() bool { return buildMode == "production" }
```

```bash
go build -ldflags "-X main.buildMode=production" -o monms .
```

### PocketBase Workspace-Scoped Config

**Source:** RESEARCH Pattern 2; supersedes PRD `pocketbase.New()` and §2 `pb_data/` sibling layout
**Apply to:** `main.go`

```go
app := pocketbase.NewWithConfig(pocketbase.Config{
    DefaultDataDir: filepath.Join(wsAbs, ".pb_data"),
})
```

### Nested Layout Rendering

**Source:** PRD §3 parse pattern + RESEARCH Pattern 4
**Apply to:** All full-page SSR and error responses

```go
tmpl, err := template.ParseFiles(layoutPath, pagePath)
if err != nil {
    return renderError(w, 500, err, !isProduction())
}
e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
return tmpl.ExecuteTemplate(e.Response, "base", data)
```

### OnServe Route Registration Order

**Source:** RESEARCH Code Examples + Pitfall 3
**Apply to:** `main.go` / `internal/router/*`

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    se.Router.GET("/assets/{path...}", assetsHandler)
    se.Router.GET("/fragments/{name}", fragmentsHandler)
    se.Router.GET("/{slug...}", ssrHandler) // last — catch-all
    return se.Next()
})
```

### Template Context Injection

**Source:** PRD §3 lines 100–106
**Apply to:** SSR handler (Phase 1 subset)

```go
data := map[string]any{
    "IsLoggedIn": e.Auth != nil, // keep for forward compat; badge deferred Phase 3
    "User":       e.Auth,
    "Slug":       slug,
    "App":        app,
}
```

Phase 1 index stub does not use `App` for DB reads.

### Startup Logging

**Source:** CONTEXT D-31

```go
slog.Info("workspace configured",
    "path", configuredPath,
    "absolute", absolutePath,
    "mode", buildMode,
)
```

### Error Response Strategy

**Source:** CONTEXT D-16–D-20
**Apply to:** `internal/router/ssr.go`

| Condition | Dev mode | Production mode |
|-----------|----------|-----------------|
| Template parse error | Detailed error in 500 body | Generic "Internal server error" |
| Unknown slug | 404 via errors.gohtml | Same |
| Missing error template | Built-in minimal HTML | Same |

---

## Integration Points Between Packages

```
main.go
├── config.ResolveWorkspace()          → ws path for all packages
├── workspace.ValidateWorkspace()        → fatal exit if incomplete
├── scaffold.RunInit() [init path only]  → writes workspace; no PocketBase
└── runServe()
    ├── pocketbase.NewWithConfig()       → DefaultDataDir = ws/.pb_data
    ├── schema.RegisterBootstrapHook()   → OnBootstrap: schema/*.json → SQLite
    ├── templates.NewCache()           → shared by router handlers
    ├── templates.StartWatcher()       → production only → Flush()
    └── router.RegisterRoutes()          → OnServe: assets, fragments, SSR

Request flow (GET /press/2024):
  router/ssr → templates/resolver.ResolveSlug("press/2024")
            → templates/cache.Get(slug, loader)
            → template.ParseFiles(layout, page)
            → ExecuteTemplate("base", data)

Request flow (GET /assets/main.css):
  router/assets → safeAssetPath() → e.File()

Request flow (GET /fragments/nav):
  router/fragments → parse fragments/nav.gohtml → Execute (no base)

Filesystem flow (production):
  watcher → debounce .gohtml events → cache.Flush()
```

**External system boundaries (do not reimplement):**
- PocketBase admin: `/_/` (SEC-01)
- PocketBase REST: `/api/*`
- PocketBase auth/session: `e.Auth` on RequestEvent

---

## PRD → CONTEXT Override Matrix

| PRD Default | Phase 1 Decision | Package Affected |
|-------------|------------------|------------------|
| `os.Getenv("ENV") == "production"` | ldflags `buildMode` (D-01) | `main.go`, `cache.go`, `watcher.go` |
| `pb_data/` sibling to workspace | `workspace/.pb_data/` (D-27) | `main.go`, `config.go` |
| Flat `{slug}.gohtml` only | Mirror + index resolver (D-10) | `resolver.go`, `ssr.go` |
| Empty slug → slug `"index"` | Resolve to `templates/index.gohtml` (D-11) | `resolver.go` |
| `tmpl.Execute()` | `ExecuteTemplate("base")` | `ssr.go` |
| Watch `./workspace/templates` only | Watch entire workspace `.gohtml` (D-30) | `watcher.go` |
| Watcher always on | Production only (D-04) | `main.go`, `watcher.go` |
| Auto-implied workspace | `monms init` required (D-05) | `scaffold/init.go`, `validate.go` |
| Simple asset join | Root-jail validation (D-29) | `assets.go` |
| No schema sync in §3 | `ImportCollectionsByMarshaledJSON` on bootstrap (D-32) | `schema/sync.go` |
| Auth badge in base layout | HTMX/Alpine CDN only; badge Phase 3 (D-24) | `base.gohtml` scaffold |
| Index with PocketBase reads | Static stub + Alpine nav demo (D-23) | `index.gohtml` scaffold |

---

## No Analog Found

Files with no PRD §3 reference — planner should use RESEARCH.md patterns:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/config/config.go` | utility | config | Workspace flag/env resolution not in PRD |
| `internal/workspace/validate.go` | utility | file-I/O | Init/validate flow is Phase 1 addition |
| `internal/templates/resolver.go` | utility | file-I/O | PRD only supports flat `{slug}.gohtml` |
| `internal/router/fragments.go` | route | request-response | `/fragments/{name}` is D-15 addition |
| `internal/schema/sync.go` | service | CRUD | Schema bootstrap not in PRD §3 |
| `internal/scaffold/init.go` | service | file-I/O | Separate init command not in PRD |
| `workspace/templates/errors/errors.gohtml` | template | request-response | Error page contract is CONTEXT-only |
| `workspace/assets/main.css` | static | file-I/O | Content unspecified in PRD |

---

## Metadata

**Analog search scope:** `/home/terence/code/MonMS/**/*.go` (0 files), `specs/monms-prd.md`, `01-RESEARCH.md`, `01-CONTEXT.md`
**Files scanned:** 0 Go source; 3 planning/spec documents
**Pattern extraction date:** 2026-05-22
**Primary reference:** `specs/monms-prd.md` §3 (lines 46–172), §5 (lines 235–314)
