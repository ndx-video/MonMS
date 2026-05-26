# Phase 4: Staging Environments & Client Content Publish - Pattern Map

**Mapped:** 2026-05-23
**Files analyzed:** 28 (17 new Go, 3 modified Go, 8 workspace/docs/fixtures)
**Analogs found:** 24 codebase / 28 total

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/schema/editorial.go` | utility | file-I/O + transform | `internal/schema/sync.go` | exact (schema JSON dir scan) |
| `internal/schema/editorial_test.go` | test | — | `internal/schema/seed_test.go` | role-match |
| `internal/content/export.go` | service | CRUD + file-I/O | `internal/schema/seed.go` (inverse) + `sync.go` | role-match |
| `internal/content/import.go` | service | CRUD | `internal/schema/seed.go` | exact (upsert by ID) |
| `internal/content/schema.go` | utility | file-I/O | `internal/schema/editorial.go` | exact |
| `internal/content/checksum.go` | utility | transform | `internal/schema/sync.go` (json.Marshal) | partial |
| `internal/content/diff.go` | service | transform | `internal/validate/validate.go` (accumulate errors) | role-match |
| `internal/content/state.go` | utility | file-I/O | `internal/scaffold/init.go` (writeScaffoldFile) | role-match |
| `internal/content/auth.go` | middleware | request-response | `internal/router/auth.go` + `internal/testutil/auth.go` | role-match |
| `internal/content/routes.go` | route | request-response | `internal/router/ssr.go` + `main.go` OnServe | role-match |
| `internal/content/publish.go` | service | request-response | `internal/testutil/auth.go` (Bearer outbound) | partial (no outbound HTTP yet) |
| `internal/content/cmd.go` | CLI | request-response + CRUD | `internal/validate/cmd.go` + `internal/schema/seed_test.go` | exact |
| `internal/content/export_test.go` | test | — | `internal/schema/seed_test.go` | exact |
| `internal/content/import_test.go` | test | — | `internal/schema/seed_test.go` | exact |
| `internal/content/routes_test.go` | test | — | `internal/router/handlers_test.go` | exact |
| `internal/testutil/content.go` | test helper | — | `internal/router/inline_edit_test.go` + `testutil/workspace.go` | role-match |
| `main.go` *(modify)* | entry/bootstrap | request-response | self (validate dispatch + OnServe hook) | exact |
| `workspace/schema/hero_content.json` *(modify)* | config | CRUD | self + `internal/scaffold/embed/hero_content.json` | exact |
| `internal/scaffold/embed/hero_content.json` *(modify)* | config | CRUD | `workspace/schema/hero_content.json` | exact |
| `workspace/content/hero_content.json` | config/data | file-I/O | `specs/staging.md` shape (no runtime analog) | format-reference |
| `workspace/.monms/config.example.json` | config | — | none | no analog |
| `workspace/templates/layouts/base.gohtml` *(modify)* | component | request-response | self (editor badge block) | exact |
| `internal/scaffold/embed/base.gohtml` *(modify)* | component | request-response | `workspace/templates/layouts/base.gohtml` | exact |
| `workspace/MEDIA.md` | documentation | — | `workspace/README.md` (media section stubs) | partial |
| `workspace/README.md` *(modify)* | documentation | — | self (four-layer table exists) | exact |
| `EDITING-GUIDE.md` *(modify)* | documentation | — | `workspace/README.md` | partial |
| `CLAUDE.md` *(modify)* | documentation | — | self | partial |
| `README.md` *(modify)* | documentation | — | `workspace/README.md` | partial |

---

## Pattern Assignments

### `internal/schema/editorial.go` (utility, file-I/O + transform)

**Analog:** `internal/schema/sync.go` (`loadSchemaJSONFiles`)

**Imports pattern** — copy from `sync.go` lines 1–13:
```go
package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)
```

**SchemaMeta struct + per-file parse** — mirror `loadSchemaJSONFiles` lines 39–90 (ReadDir, sort, ReadFile, continue-on-error):
```go
type SchemaMeta struct {
	Name      string `json:"name"`
	Editorial bool   `json:"editorial"`
}

func LoadEditorialCollectionNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	var names []string
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue // ← same continue-on-error as sync.go line 64
		}
		data = trimJSON(data)
		if len(data) == 0 || data[0] != '{' {
			continue
		}
		var meta SchemaMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if meta.Name != "" && meta.Editorial {
			names = append(names, meta.Name)
		}
	}
	return names, nil
}
```

**Critical:** Do **not** read editorial flag from live `core.Collection` — PocketBase import strips unknown keys (RESEARCH Pitfall 1). Parse raw JSON files only.

---

### `internal/content/import.go` (service, CRUD)

**Analog:** `internal/schema/seed.go` (`seedHeroHomepage`)

**Imports pattern** — copy from `seed.go` lines 1–7:
```go
package content

import (
	"fmt"
	"log/slog"

	"github.com/pocketbase/pocketbase/core"
)
```

**Idempotent upsert by record ID** — mirror `seed.go` lines 19–44:
```go
func UpsertRecord(app core.App, collectionName string, data map[string]any) error {
	id, _ := data["id"].(string)
	if id == "" {
		return fmt.Errorf("record missing id")
	}

	coll, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return err
	}

	rec, err := app.FindRecordById(collectionName, id)
	if err != nil {
		rec = core.NewRecord(coll)
		rec.Set("id", id)
	}

	for k, v := range data {
		if k == "id" || k == "collectionId" || k == "collectionName" {
			continue
		}
		if coll.Fields.GetByName(k) == nil {
			slog.Warn("content import: unknown field skipped",
				"collection", collectionName, "field", k)
			continue
		}
		rec.Set(k, v)
	}

	if err := app.Save(rec); err != nil {
		return err
	}
	slog.Info("content import: upserted record", "collection", collectionName, "id", id)
	return nil
}
```

**Existence check pattern** — from `seed.go` lines 25–27 (skip if already exists for seed; import always updates):
```go
if _, err := app.FindRecordById(heroCollection, heroRecordID); err == nil {
	return nil // seed: idempotent no-op
}
```

---

### `internal/content/export.go` (service, CRUD + file-I/O)

**Primary Analog:** `internal/schema/seed.go` (record access, inverse direction)
**Secondary Analog:** `internal/schema/sync.go` (write JSON files to workspace)

**Export collection records** — PocketBase API + seed collection lookup:
```go
func exportCollection(app core.App, name string) ([]map[string]any, error) {
	records, err := app.FindAllRecords(name)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		m := rec.PublicExport()
		delete(m, "collectionId")
		delete(m, "collectionName")
		delete(m, "expand")
		out = append(out, m)
	}
	return out, nil
}
```

**Write content file** — mirror `sync.go` error wrapping line 98:
```go
merged, err := json.Marshal(payload)
if err != nil {
	return fmt.Errorf("content export: marshal %s: %w", name, err)
}
dest := filepath.Join(wsAbs, "content", name+".json")
if err := os.WriteFile(dest, merged, 0o644); err != nil {
	return fmt.Errorf("content export: write %s: %w", dest, err)
}
```

**Path guard before read/write** — copy `internal/validate/validate.go` lines 92–98:
```go
cleanF := filepath.Clean(f)
rel, err := filepath.Rel(wsAbs, cleanF)
if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
	return fmt.Errorf("refusing path outside workspace: %s", f)
}
```

---

### `internal/content/cmd.go` (CLI, request-response + CRUD)

**Primary Analog:** `internal/validate/cmd.go`
**Secondary Analog:** `internal/schema/seed_test.go` (ephemeral PocketBase bootstrap)

**RunCLI dispatch** — mirror `validate/cmd.go` lines 15–50:
```go
func RunCLI(args []string) error {
	fs := flag.NewFlagSet("content", flag.ContinueOnError)
	var wsFlag string
	fs.StringVar(&wsFlag, "workspace", "", "workspace path (default: MONMS_WORKSPACE or ./workspace)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	_, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
	if err != nil {
		return err
	}

	sub := fs.Arg(0)
	switch sub {
	case "export", "import", "diff", "publish":
		// bootstrap then dispatch
	default:
		return fmt.Errorf("usage: monms content <export|import|diff|publish>")
	}
}
```

**Ephemeral app bootstrap** — mirror `seed_test.go` lines 9–20 + `main.go` lines 69–74:
```go
app := pocketbase.NewWithConfig(pocketbase.Config{
	DefaultDataDir:  filepath.Join(wsAbs, ".pb_data"),
	DefaultDev:      true,
	HideStartBanner: true,
})
schema.RegisterBootstrapHook(app, wsAbs)
if err := app.Bootstrap(); err != nil {
	return fmt.Errorf("bootstrap: %w", err)
}
```

**Difference from validate:** content CLI **must** Bootstrap — validate never touches PocketBase (Pitfall 4 in RESEARCH).

---

### `internal/content/routes.go` (route, request-response)

**Primary Analog:** `internal/router/ssr.go` (`RegisterRoutes`, `Deps` struct)
**Secondary Analog:** `main.go` lines 77–84 (`OnServe` binding)

**Deps struct pattern** — copy from `ssr.go` lines 17–22:
```go
type Deps struct {
	WsAbs        string
	PublishToken string // from MONMS_PUBLISH_TOKEN via config.envValue pattern
}

func RegisterRoutes(se *core.ServeEvent, deps Deps) {
	// Production import
	se.Router.POST("/api/monms/content/import", importHandler(deps))
	// Staging publish UI + API
	se.Router.GET("/api/monms/publish", publishPageHandler(deps))
	se.Router.GET("/api/monms/publish/diff", publishDiffHandler(deps))
	se.Router.POST("/api/monms/publish", publishPostHandler(deps))
}
```

**OnServe wiring in main.go** — extend existing hook (lines 77–84):
```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
	content.RegisterRoutes(se, content.Deps{
		WsAbs:        abs,
		PublishToken: config.EnvValue("MONMS_PUBLISH_TOKEN"), // or os.Getenv
	})
	router.RegisterRoutes(se, router.Deps{...})
	return se.Next()
})
```

**Register content routes before SSR catch-all** — `/api/monms/*` prefix avoids slug conflicts (D-14). Same `OnServe` hook as router; order within hook: content API first.

**HTML response pattern** — copy from `fragments.go` lines 34–35:
```go
e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
return tmpl.Execute(e.Response, data)
// or: return e.HTML(http.StatusOK, htmlString)
```

**JSON response pattern** — from RESEARCH (PocketBase `e.JSON`, `e.BindBody`):
```go
var body importRequest
if err := e.BindBody(&body); err != nil {
	return e.BadRequestError("invalid JSON", err)
}
return e.JSON(http.StatusOK, report)
```

---

### `internal/content/auth.go` (middleware, request-response)

**Primary Analog:** `internal/router/auth.go` (auth record on RequestEvent)
**Secondary Analog:** `internal/testutil/auth.go` (Bearer header transport)

**Publish token middleware** — new pattern; use stdlib constant-time compare (RESEARCH Pattern 4):
```go
import (
	"crypto/subtle"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func RequirePublishToken(expected string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		auth := e.Request.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			return e.UnauthorizedError("invalid publish token", nil)
		}
		return e.Next()
	}
}
```

**Publisher allowlist** — after superuser session; read email from `e.Auth`:
```go
func RequirePublisher(allowedEmails []string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.UnauthorizedError("authentication required", nil)
		}
		email := e.Auth.GetString("email")
		for _, allowed := range allowedEmails {
			if email == allowed {
				return e.Next()
			}
		}
		return e.ForbiddenError("publisher role required", nil)
	}
}
```

**Superuser test setup** — copy from `inline_edit_test.go` lines 100–102:
```go
user := testutil.NewSuperuser(t, app, "publisher@test.local")
client := testutil.AuthClient(t, app, user)
```

---

### `internal/content/publish.go` (service, request-response)

**Analog:** `internal/testutil/auth.go` lines 10–22 (`bearerTransport`)

**Outbound POST to production** — reuse Bearer transport pattern:
```go
req, err := http.NewRequest(http.MethodPost, productionURL+"/api/monms/content/import", body)
if err != nil {
	return fmt.Errorf("new request: %w", err)
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)

client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Do(req)
```

**Never log token** — match `workspace/SECURITY.md` convention cited in RESEARCH.

---

### `internal/content/state.go` (utility, file-I/O)

**Analog:** `internal/scaffold/init.go` (`writeScaffoldFile`, `ensureUnderWorkspace`)

**Write state file under `.monms/`** — mirror idempotent write lines 164–169:
```go
dest := filepath.Join(wsAbs, ".monms", "publish-state.json")
if err := ensureUnderWorkspace(wsAbs, dest); err != nil {
	return err
}
if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
	return fmt.Errorf("mkdir .monms: %w", err)
}
if err := os.WriteFile(dest, data, 0o644); err != nil {
	return fmt.Errorf("write publish-state: %w", err)
}
```

**Path guard** — copy `ensureUnderWorkspace` from `init.go` lines 173–185:
```go
rel, err := filepath.Rel(wsRoot, dest)
if err != nil {
	return fmt.Errorf("resolve path under workspace: %w", err)
}
if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
	return fmt.Errorf("refusing write outside workspace: %s", dest)
}
```

---

### `internal/content/checksum.go` + `diff.go` (utility/service, transform)

**Analog:** `internal/validate/validate.go` (accumulate errors, continue-on-error)

**Checksum** — stdlib only (RESEARCH):
```go
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func checksum(payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
```

**Diff error accumulation** — mirror `validate.go` lines 87–116:
```go
var errs []string
// compare collections/records, append human-readable deltas
if len(errs) > 0 {
	return fmt.Errorf("content diff:\n%s", strings.Join(errs, "\n"))
}
```

---

### `main.go` *(modify — content early dispatch + route registration)*

**Analog:** self — mirror `validate` arm lines 33–38

**Early-dispatch arm** — insert after validate, before `runServe()`:
```go
if len(os.Args) >= 2 && os.Args[1] == "content" {
	if err := content.RunCLI(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}
```

**Import addition:**
```go
"github.com/monms/monms/internal/content"
```

---

### `internal/content/routes_test.go` (integration test)

**Analog:** `internal/router/handlers_test.go` (`startTestServerWithApp`)

**Test server with content routes** — extend `startTestServerWithApp` pattern lines 32–87:
```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
	content.RegisterRoutes(se, content.Deps{
		WsAbs:        wsAbs,
		PublishToken: "test-publish-token",
	})
	RegisterRoutes(se, deps)
	return se.Next()
})
```

**Admin static handler** — copy lines 67–69 (documents why `/_/publish` fails):
```go
if ui.DistDirFS != nil {
	router.GET("/_/{path...}", apis.Static(ui.DistDirFS, false))
}
```

**HTTP client patterns** — lines 95–104, 256–264:
```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get(ts.URL + "/api/monms/publish")
```

**401/403 assertions** — mirror `inline_edit_test.go` lines 133–153 (guest PUT forbidden):
```go
if resp.StatusCode < 400 {
	t.Fatalf("unauthenticated import status %d, want >= 400", resp.StatusCode)
}
```

---

### `internal/content/export_test.go` + `import_test.go` (integration test)

**Analog:** `internal/schema/seed_test.go`

**Bootstrap + import schema + seed** — lines 9–38:
```go
dir := t.TempDir()
app := pocketbase.NewWithConfig(pocketbase.Config{
	DefaultDataDir:  dir,
	DefaultDev:      true,
	HideStartBanner: true,
})
if err := app.Bootstrap(); err != nil {
	t.Fatalf("bootstrap: %v", err)
}
// Import hero_content with editorial flag in schema file, not in ImportCollections JSON
```

**Idempotency test** — seed_test.go lines 51–61 (run import twice, assert single record):
```go
if err := ImportFiles(app, wsAbs); err != nil {
	t.Fatalf("first import: %v", err)
}
if err := ImportFiles(app, wsAbs); err != nil {
	t.Fatalf("second import: %v", err)
}
list, err := app.FindRecordsByFilter(heroCollection, "id = {:id}", "", 0, 0, map[string]any{"id": "homepage"})
if len(list) != 1 {
	t.Fatalf("expected 1 record, got %d", len(list))
}
```

---

### `internal/testutil/content.go` (test helper)

**Analog:** `internal/router/inline_edit_test.go` (`setupInlineEditWorkspace`) + `testutil/workspace.go`

**Editorial workspace fixture** — mirror `inline_edit_test.go` lines 39–55:
```go
func NewEditorialWorkspace(t *testing.T) string {
	t.Helper()
	ws := NewWorkspace(t)
	WriteFile(t, filepath.Join(ws, "schema/hero_content.json"), heroContentSchemaWithEditorial)
	if err := os.MkdirAll(filepath.Join(ws, ".monms"), 0o755); err != nil {
		t.Fatalf("mkdir .monms: %v", err)
	}
	WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"productionUrl":"http://127.0.0.1:0","publisherEmails":["publisher@test.local"]}`)
	return ws
}
```

**Schema constant** — extend `inline_edit_test.go` lines 15–37 with `"editorial": true`:
```json
{
  "name": "hero_content",
  "type": "base",
  "editorial": true,
  ...
}
```

---

### `workspace/schema/hero_content.json` + `internal/scaffold/embed/hero_content.json` *(modify)*

**Analog:** self — add field to existing schema JSON

**Pattern** — insert after `"type": "base"`:
```json
{
  "name": "hero_content",
  "type": "base",
  "editorial": true,
  "listRule": "",
  ...
}
```

**Dual-write rule:** update both `workspace/schema/hero_content.json` and `internal/scaffold/embed/hero_content.json` (CLAUDE.md scaffold sync).

---

### `workspace/content/hero_content.json` (config/data, file-I/O)

**Analog:** format-reference from `specs/staging.md` §5.1 (no runtime file yet)

**Content file shape:**
```json
{
  "collection": "hero_content",
  "records": [
    {"id": "homepage", "title": "Welcome to MonMS", "body": "..."}
  ]
}
```

---

### `workspace/templates/layouts/base.gohtml` + `internal/scaffold/embed/base.gohtml` *(modify)*

**Analog:** self — editor badge block lines 19–24

**Add Publish link beside admin link** — extend badge:
```html
{{if .IsLoggedIn}}
<div class="editor-badge" role="status" aria-live="polite">
  <span class="editor-badge__dot" aria-hidden="true"></span>
  Live Editor Active
  <a href="/api/monms/publish" class="editor-badge__link">Publish to live</a>
  <a href="/_/" class="editor-badge__link">Full Admin Dashboard</a>
</div>
{{end}}
```

Use `/api/monms/publish` not `/_/publish` — admin SPA catch-all at `/_/{path...}` (handlers_test.go lines 67–69).

---

### `workspace/.monms/config.example.json` (config)

**Analog:** No existing analog.

**Content contract** (RESEARCH Pattern 6):
```json
{
  "productionUrl": "https://production.example.com",
  "publisherEmails": ["publisher@client.com"]
}
```

Commit `config.example.json`; gitignore live `config.json` and `publish-state.json`.

---

### `workspace/MEDIA.md` (documentation)

**Analog:** `workspace/README.md` lines 46–48, 127–133 (existing v2 stubs)

**Content contract (MED-02):**
1. Publishable media = public CDN URLs in text fields
2. Do not use PocketBase `file` fields for cross-env publishable assets
3. Export skips/warns on `file`-type columns
4. Consultant sets CDN base URL once; clients paste full URLs

---

### Documentation updates (`README.md`, `workspace/README.md`, `EDITING-GUIDE.md`, `CLAUDE.md`)

**Analog:** `workspace/README.md` (four-layer table lines 7–14, staging section lines 50+)

Extend existing four-layer and dual-rail sections — do not duplicate `specs/staging.md`. Cross-link `MEDIA.md`, `monms content` CLI, and `/api/monms/publish`.

---

## Shared Patterns

### Early-Dispatch CLI Pattern
**Source:** `main.go` lines 26–38
**Apply to:** `main.go` content arm
```go
if len(os.Args) >= 2 && os.Args[1] == "content" {
	if err := content.RunCLI(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}
```

### Workspace Resolution
**Source:** `internal/config/config.go` lines 14–57
**Apply to:** `internal/content/cmd.go`
```go
_, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
```

### Env Value Lookup
**Source:** `internal/config/config.go` lines 60–71
**Apply to:** `MONMS_PUBLISH_TOKEN`, optional production URL override
```go
func envValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	if env == nil {
		return os.Getenv(key)
	}
	return ""
}
```

### PocketBase Bootstrap Hook
**Source:** `main.go` lines 74–75; `internal/schema/sync.go` lines 15–36
**Apply to:** CLI and tests needing collections + seed
```go
schema.RegisterBootstrapHook(app, wsAbs)
if err := app.Bootstrap(); err != nil {
	return err
}
```

### Idempotent Record Upsert
**Source:** `internal/schema/seed.go` lines 19–44
**Apply to:** `internal/content/import.go`
```go
rec, err := app.FindRecordById(collectionName, id)
if err != nil {
	rec = core.NewRecord(coll)
	rec.Set("id", id)
}
// set fields, app.Save(rec)
```

### Path Traversal Guard
**Source:** `internal/validate/validate.go` lines 92–98; `internal/scaffold/init.go` lines 173–185
**Apply to:** All content file reads/writes under workspace
```go
rel, err := filepath.Rel(wsAbs, cleanF)
if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
	return fmt.Errorf("refusing path outside workspace")
}
```

### slog Structured Logging
**Source:** `internal/schema/seed.go` line 43; `main.go` lines 55–58
**Apply to:** export/import/publish operations
```go
slog.Info("content export: wrote file", "collection", name, "records", len(records))
slog.Warn("content import: unknown field skipped", "collection", name, "field", k)
```

### startTestServerWithApp for Integration Tests
**Source:** `internal/router/handlers_test.go` lines 32–87
**Apply to:** `internal/content/routes_test.go`
```go
ts, app, cache, cleanup := startTestServerWithApp(t, wsAbs, testServerOpts{isDev: true})
defer cleanup()
```
Register `content.RegisterRoutes` in the same `OnServe` bind as `RegisterRoutes`.

### Bearer Auth Client for Tests
**Source:** `internal/testutil/auth.go` lines 43–55
**Apply to:** publish UI tests (superuser session) and import API tests (publish token)
```go
client := testutil.AuthClient(t, app, user)
client.Timeout = 10 * time.Second
```

### RegisterAuthHooks in Tests
**Source:** `handlers_test.go` line 49; `main.go` line 75
**Apply to:** Any integration test using superuser session on MonMS routes
```go
RegisterAuthHooks(app)
```

### Error Wrapping with `%w`
**Source:** `internal/schema/sync.go` line 98; `internal/scaffold/init.go` throughout
**Apply to:** All new content package errors
```go
return fmt.Errorf("content import: %w", err)
```

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/content/auth.go` (`RequirePublishToken`) | middleware | request-response | No custom PocketBase middleware in codebase; use RESEARCH Pattern 4 + PocketBase `e.UnauthorizedError` |
| `internal/content/publish.go` | service | request-response | No outbound HTTP client in MonMS yet; borrow `testutil/auth.go` Bearer header pattern |
| `workspace/.monms/config.example.json` | config | — | New staging-local config artifact; no prior `.monms/` files in repo |
| `workspace/MEDIA.md` | documentation | — | New doc; partial stubs only in `workspace/README.md` |

---

## Metadata

**Analog search scope:** `/home/terence/code/MonMS/internal/**/*.go`, `main.go`, `workspace/**`
**Files scanned:** 18 Go source files + 4 workspace fixtures
**Pattern extraction date:** 2026-05-23
**Primary analogs:** `internal/schema/seed.go` (upsert), `internal/schema/sync.go` (JSON dir I/O), `internal/validate/cmd.go` (CLI dispatch), `internal/router/handlers_test.go` (integration harness), `internal/router/inline_edit_test.go` (hero_content + auth fixtures)
