# Phase 2: Agent Mutation Engine & Safety Guardrails - Pattern Map

**Mapped:** 2026-05-22
**Files analyzed:** 10 (5 new, 2 modified, 3 workspace docs/fixtures)
**Analogs found:** 8 codebase / 10 total

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/validate/validate.go` | service/utility | file-I/O + transform | `internal/router/ssr.go` + `internal/schema/sync.go` | role-match (template parse) + role-match (file-read service) |
| `internal/validate/cmd.go` | CLI/utility | request-response (subprocess) | `internal/scaffold/init.go` + `internal/config/config.go` | exact (same CLI dispatch + flag resolution) |
| `internal/validate/validate_test.go` | test | — | `internal/router/handlers_test.go` | role-match |
| `internal/scaffold/init.go` *(modify)* | service/CLI | file-I/O | self (extend `maybeGitInit`) | exact |
| `main.go` *(modify)* | entry/bootstrap | request-response | self (extend early-dispatch arms) | exact |
| `workspace/agent-guide.md` | documentation | — | none | no analog |
| `workspace/SECURITY.md` | documentation | — | none | no analog |
| `workspace/schema/press_releases.json` | config | CRUD | `internal/schema/sync.go` (format spec) | format-reference |
| `internal/router/press_releases_test.go` | test (integration) | — | `internal/router/handlers_test.go` | exact |
| `internal/scaffold/hook_test.go` | test (integration) | — | `internal/scaffold/init.go` + `internal/testutil/workspace.go` | role-match |

---

## Pattern Assignments

### `internal/validate/validate.go` (service/utility, file-I/O + transform)

**Primary Analog:** `internal/router/ssr.go` (template.ParseFiles pattern)
**Secondary Analog:** `internal/schema/sync.go` (file-read service, log-on-error pattern)

**Imports pattern** — copy from `internal/router/ssr.go` lines 1–14 and `internal/schema/sync.go` lines 1–13:
```go
package validate

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gohtml "golang.org/x/net/html"
)
```

**Template dry-run pattern** — mirrors `internal/router/ssr.go` lines 54–56 (production ParseFiles call):
```go
// ValidateTemplate(wsAbs, filePath string) error
// Source: mirrors ssr.go lines 54-56 loader func exactly
layoutPath := filepath.Join(wsAbs, "templates", "layouts", "base.gohtml")
if _, err := template.ParseFiles(layoutPath, filePath); err != nil {
    return fmt.Errorf("%s: template parse error: %w", filepath.Base(filePath), err)
}
return nil
```
Key: use `template.ParseFiles(layoutPath, filePath)` — two args, layout first, same as ssr.go. A file that passes is guaranteed to parse at serve time (D-37).

**File-read + error-collection pattern** — mirrors `internal/schema/sync.go` lines 53–85 (read per-file, accumulate errors, continue):
```go
// ValidateFiles(wsAbs string, files []string) error
var errs []string
for _, f := range files {
    content, err := os.ReadFile(f)
    if err != nil {
        errs = append(errs, fmt.Sprintf("%s: read error: %v", filepath.Base(f), err))
        continue  // ← same continue-on-error style as sync.go
    }
    if err := ValidateTemplate(wsAbs, f); err != nil {
        errs = append(errs, err.Error())
    }
    if err := ValidateHTML(f, content); err != nil {
        errs = append(errs, err.Error())
    }
}
if len(errs) > 0 {
    return fmt.Errorf("validation failed:\n%s", strings.Join(errs, "\n"))
}
return nil
```

**HTML tokenizer pattern** — new (no codebase analog; use RESEARCH.md Pattern 2 verbatim):
```go
var templateDirectiveRE = regexp.MustCompile(`(?s)\{\{.*?\}\}`)

var voidElements = map[string]bool{
    "area": true, "base": true, "br": true, "col": true,
    "embed": true, "hr": true, "img": true, "input": true,
    "link": true, "meta": true, "param": true, "source": true,
    "track": true, "wbr": true,
}

// ValidateHTML(filePath string, content []byte) error
stripped := templateDirectiveRE.ReplaceAll(content, []byte(" "))
z := gohtml.NewTokenizer(bytes.NewReader(stripped))
var stack []string
var errs []string
for {
    tt := z.Next()
    switch tt {
    case gohtml.ErrorToken:
        if z.Err() == io.EOF {
            goto done
        }
        errs = append(errs, fmt.Sprintf("tokenizer error: %v", z.Err()))
        goto done
    case gohtml.StartTagToken:
        rawName, selfClose := z.TagName()
        name := string(rawName)
        if !selfClose && !voidElements[name] {
            stack = append(stack, name)
        }
    case gohtml.EndTagToken:
        rawName, _ := z.TagName()
        name := string(rawName)
        if len(stack) == 0 {
            errs = append(errs, fmt.Sprintf("unexpected </%s>: no open tag", name))
        } else if stack[len(stack)-1] != name {
            errs = append(errs, fmt.Sprintf("mismatched tag: open <%s>, close </%s>", stack[len(stack)-1], name))
        } else {
            stack = stack[:len(stack)-1]
        }
    }
}
done:
for _, tag := range stack {
    errs = append(errs, fmt.Sprintf("unclosed <%s>", tag))
}
```

**Error format** (Claude's discretion): `"<filename>: <category>:\n  <detail1>\n  <detail2>"` — matches `internal/schema/sync.go` slog.Error style but formatted for stderr.

---

### `internal/validate/cmd.go` (CLI/utility, request-response via subprocess)

**Primary Analog:** `internal/scaffold/init.go` (CLI dispatch, workspace resolution, slog logging)
**Secondary Analog:** `internal/config/config.go` (ResolveWorkspace signature)

**Package + imports pattern** — mirror `internal/scaffold/init.go` lines 1–13:
```go
package validate

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/config"
)
```

**RunCLI dispatch pattern** — mirrors `internal/scaffold/init.go` `RunInit()` lines 36–70:
```go
// RunCLI(args []string) error — called from main.go early dispatch
func RunCLI(args []string) error {
    fs := flag.NewFlagSet("validate", flag.ContinueOnError)
    var wsFlag string
    fs.StringVar(&wsFlag, "workspace", "", "workspace path")
    if err := fs.Parse(args); err != nil {
        return err
    }

    // Inject --workspace into args slice for ResolveWorkspace (same as RunInit)
    _, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
    if err != nil {
        return err
    }

    files := fs.Args() // positional args = explicit file paths

    if len(files) == 0 {
        files, err = getStagedGohtml(wsAbs)
        if err != nil {
            return fmt.Errorf("get staged files: %w", err)
        }
    }
    if len(files) == 0 {
        return nil // nothing to validate; exit 0
    }
    return ValidateFiles(wsAbs, files)
}
```

**Subprocess pattern** — `os/exec.Command` with `.Dir` set (same as `internal/scaffold/init.go` lines 158–163 git init call):
```go
// getStagedGohtml(wsAbs string) ([]string, error)
cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
cmd.Dir = wsAbs                    // ← same .Dir pattern as scaffold/init.go line 159
out, err := cmd.Output()
if err != nil {
    return nil, fmt.Errorf("git diff: %w", err)
}
var result []string
for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
    if strings.HasSuffix(line, ".gohtml") && line != "" {
        result = append(result, filepath.Join(wsAbs, line))
    }
}
return result, nil
```

**Logging pattern** — use `slog.Info`/`slog.Warn` matching `internal/scaffold/init.go` lines 68, 147:
```go
slog.Info("validate: no staged .gohtml files, skipping")
slog.Warn("validate: git not available, --files required", "err", err)
```

---

### `internal/validate/validate_test.go` (test, unit)

**Primary Analog:** `internal/router/handlers_test.go` (table-driven tests, testutil.NewWorkspace, testutil.WriteFile)

**Test file setup pattern** — mirrors `handlers_test.go` lines 84–98:
```go
package validate_test  // external test package (same as handlers_test)

import (
    "testing"
    "path/filepath"

    "github.com/monms/monms/internal/testutil"
    "github.com/monms/monms/internal/validate"
)

func TestValidateTemplate(t *testing.T) {
    ws := testutil.NewWorkspace(t)
    // write a valid template
    pagePath := filepath.Join(ws, "templates/press/index.gohtml")
    testutil.WriteFile(t, pagePath, `{{define "body"}}<h1>Press</h1>{{end}}`)

    if err := validate.ValidateTemplate(ws, pagePath); err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}
```

**Negative test pattern** — mirrors `handlers_test.go` `TestProduction500Generic` lines 242–272 (write broken file, assert error):
```go
func TestValidateTemplateBadSyntax(t *testing.T) {
    ws := testutil.NewWorkspace(t)
    pagePath := filepath.Join(ws, "templates/broken.gohtml")
    testutil.WriteFile(t, pagePath, `{{define "body"}}{{if}}{{end}}`)  // ← same broken content as handlers_test line 245

    err := validate.ValidateTemplate(ws, pagePath)
    if err == nil {
        t.Fatal("expected template parse error, got nil")
    }
}
```

**Table-driven HTML test pattern** (consistent with handlers_test style):
```go
func TestValidateHTML(t *testing.T) {
    cases := []struct {
        name    string
        content string
        wantErr bool
    }{
        {"valid div", `<div><p>text</p></div>`, false},
        {"unclosed div", `<div><p>text</div>`, true},
        {"void self-close", `<br><img src="x.png">`, false},
        {"template directives stripped", `{{define "body"}}<div></div>{{end}}`, false},
        {"multiline range directive", "{{range .Items}}\n<li>{{.Title}}</li>\n{{end}}", false},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            err := validate.ValidateHTML("test.gohtml", []byte(tc.content))
            if (err != nil) != tc.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
            }
        })
    }
}
```

---

### `internal/scaffold/init.go` *(modify — add installPreCommitHook)*

**Analog:** self — extend existing `maybeGitInit` pattern (lines 144–164)

**Idempotency guard pattern** — mirrors `writeScaffoldFile` lines 89–94 (stat + skip-if-exists with marker check):
```go
// installPreCommitHook writes workspace/.git/hooks/pre-commit (D-40).
// Idempotent: skips if file contains "monms-validate-hook" marker.
func installPreCommitHook(wsRoot string) error {
    hooksDir := filepath.Join(wsRoot, ".git", "hooks")
    if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
        // ← same IsNotExist check pattern as maybeGitInit line 146
        slog.Warn("hooks dir not found, skipping pre-commit hook install", "dir", hooksDir)
        return nil
    }
    if err := os.MkdirAll(hooksDir, 0o755); err != nil {  // A4: ensure hooks/ exists
        return fmt.Errorf("mkdir hooks dir: %w", err)
    }
    hookPath := filepath.Join(hooksDir, "pre-commit")
    if existing, err := os.ReadFile(hookPath); err == nil {
        if bytes.Contains(existing, []byte("monms-validate-hook")) {
            slog.Info("pre-commit hook already installed, skipping")  // ← same slog.Info style line 147
            return nil
        }
    }
    if err := os.WriteFile(hookPath, []byte(preCommitHookScript), 0o755); err != nil {
        return fmt.Errorf("install pre-commit hook: %w", err)
    }
    slog.Info("pre-commit hook installed", "path", hookPath)  // ← same slog.Info("workspace initialized") line 68
    return nil
}
```

**Embed constant pattern** — follows existing `embed.go`/`embed/` approach:
```go
// preCommitHookScript is the bash script installed into workspace/.git/hooks/pre-commit
const preCommitHookScript = `#!/bin/sh
# monms-validate-hook — DO NOT REMOVE THIS COMMENT (idempotency marker)
...`
```

**RunInit extension point** — insert after `maybeGitInit` call (line 64):
```go
// current (lines 64-68):
if err := maybeGitInit(wsAbs); err != nil {
    return err
}
// ADD after maybeGitInit:
if err := installPreCommitHook(wsAbs); err != nil {
    return err
}
slog.Info("workspace initialized", "path", wsAbs)
```

---

### `main.go` *(modify — add validate early-dispatch arm)*

**Analog:** self — mirror existing `init` dispatch (lines 25–31)

**Early-dispatch pattern** — insert before `runServe()` call, following lines 25–31 exactly:
```go
// current (lines 24-33):
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

// ADD: second arm before runServe(), same structure as init arm
if len(os.Args) >= 2 && os.Args[1] == "validate" {
    if err := validate.RunCLI(os.Args[2:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    return
}
```

**Import addition** — add to existing imports block (lines 3–18):
```go
"github.com/monms/monms/internal/validate"
```

**Critical:** dispatch BEFORE `runServe()` — never via PocketBase cobra (Pitfall 7 in RESEARCH.md: validate must not trigger Bootstrap/DB).

---

### `workspace/agent-guide.md` (documentation)

**Analog:** No existing analog in codebase.

**Content contract** (D-46, D-47): Six required sections per CONTEXT.md:
1. Overview — what agent mutation enables
2. Prerequisites — MONMS_BIN, admin token, SSH key setup
3. Dual-write schema workflow — curl POST + write JSON (RESEARCH Pattern 5)
4. Template editing conventions — `{{define "body"}}`, `base.gohtml`, HTMX patterns, mirror+index slug rules
5. Pre-commit validation lifecycle — `monms validate`, hook, rollback behavior
6. `press_releases` sample walkthrough — end-to-end from POST to browser verify

**Template example to embed** (RESEARCH lines 736–749):
```html
{{define "body"}}
<section class="press-list">
  <h1>Press Releases</h1>
  <div hx-get="/api/collections/press_releases/records"
       hx-trigger="load"
       hx-target="#press-list">
    Loading...
  </div>
  <ul id="press-list"></ul>
</section>
{{end}}
```

---

### `workspace/SECURITY.md` (documentation)

**Analog:** No existing analog.

**Content contract** (D-48, SEC-03): Four required sections:
1. SSH key scope — dedicated key restricted to workspace subdirectory
2. PocketBase admin token — env/vault storage, never in git, token scope
3. Git history safety — `.pb_data/` excluded, credential commit prevention
4. Operator escape hatches — `git log`, `git revert`, `--no-verify` limitation warning

---

### `workspace/schema/press_releases.json` (config, CRUD)

**Analog:** `internal/schema/sync.go` (defines the format this file must satisfy)

**Format pattern** — must match `sync.go` `loadSchemaJSONFiles` lines 68–82 (single JSON object, `{` first byte):
```json
{
  "name": "press_releases",
  "type": "base",
  "fields": [
    {"name": "title", "type": "text"},
    {"name": "body",  "type": "text"}
  ]
}
```
This is the canonical D-35 integration test fixture. On server restart, `schema/sync.go` imports it via `ImportCollectionsByMarshaledJSON` — same self-healing bootstrap as any other schema file.

---

### `internal/router/press_releases_test.go` (integration test)

**Primary Analog:** `internal/router/handlers_test.go` — copy `startTestServer` invocation pattern exactly

**Full test structure** — mirrors `TestFragmentPartial` (lines 210–240) + `TestHomepageSSR` (lines 184–208):
```go
package router

import (
    "io"
    "net/http"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "github.com/monms/monms/internal/testutil"
    "github.com/monms/monms/internal/validate"
)

func TestPressReleasesOperation(t *testing.T) {
    ws := testutil.NewWorkspace(t)

    // Step 1: Write press_releases schema JSON (D-33 dual-write fixture)
    testutil.WriteFile(t, filepath.Join(ws, "schema/press_releases.json"),
        `{"name":"press_releases","type":"base","fields":[{"name":"title","type":"text"},{"name":"body","type":"text"}]}`)

    // Step 2: Start server — Bootstrap() imports schema/*.json via sync.go
    ts, cache, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
    defer cleanup()

    // Step 3: Write template (AGT-02 mutation)
    testutil.WriteFile(t, filepath.Join(ws, "templates/press/index.gohtml"),
        `{{define "body"}}<h1>Press Releases</h1>{{end}}`)
    cache.Flush() // simulate fsnotify invalidation (same as used in watcher tests)

    // Step 4: Validate template (AGT-03, AGT-04)
    pagePath := filepath.Join(ws, "templates/press/index.gohtml")
    if err := validate.ValidateTemplate(ws, pagePath); err != nil {
        t.Fatalf("template validation: %v", err)
    }
    content, _ := os.ReadFile(pagePath)
    if err := validate.ValidateHTML(pagePath, content); err != nil {
        t.Fatalf("HTML validation: %v", err)
    }

    // Step 5: Verify page renders without restart (AGT-02)
    client := &http.Client{Timeout: 10 * time.Second}   // ← same client pattern as handlers_test lines 89, 106
    resp, err := client.Get(ts.URL + "/press")
    if err != nil {
        t.Fatalf("GET /press: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("status %d, want 200", resp.StatusCode)
    }
    body, _ := io.ReadAll(resp.Body)
    if !strings.Contains(string(body), "Press Releases") {
        t.Fatalf("body missing Press Releases, got: %s", string(body))
    }
}
```

---

### `internal/scaffold/hook_test.go` (integration test)

**Primary Analog:** `internal/scaffold/init.go` (RunInit behavior) + `internal/testutil/workspace.go` (t.TempDir pattern)

**Test setup pattern** — use `t.TempDir()` directly (scaffold tests don't need full workspace):
```go
package scaffold_test

import (
    "os"
    "path/filepath"
    "testing"
)

func TestInstallPreCommitHook(t *testing.T) {
    ws := t.TempDir()
    // Pre-create .git/hooks (maybeGitInit would do this in production)
    hooksDir := filepath.Join(ws, ".git", "hooks")
    if err := os.MkdirAll(hooksDir, 0o755); err != nil {
        t.Fatal(err)
    }

    // Call RunInit equivalent or directly test installPreCommitHook
    if err := RunInit([]string{"--workspace=" + ws}); err != nil {
        t.Fatalf("RunInit: %v", err)
    }

    hookPath := filepath.Join(hooksDir, "pre-commit")
    data, err := os.ReadFile(hookPath)
    if err != nil {
        t.Fatalf("hook not created: %v", err)
    }
    if string(data[:10]) != "#!/bin/sh\n" {
        t.Fatalf("hook missing shebang, got: %s", string(data[:20]))
    }
}

func TestInstallPreCommitHookIdempotent(t *testing.T) {
    ws := t.TempDir()
    // Write existing hook with marker
    hooksDir := filepath.Join(ws, ".git", "hooks")
    os.MkdirAll(hooksDir, 0o755)
    hookPath := filepath.Join(hooksDir, "pre-commit")
    original := "# monms-validate-hook\noriginal content"
    os.WriteFile(hookPath, []byte(original), 0o755)

    // RunInit should NOT overwrite
    RunInit([]string{"--workspace=" + ws})
    data, _ := os.ReadFile(hookPath)
    if string(data) != original {
        t.Fatalf("idempotency broken: hook was overwritten")
    }
}
```

---

## Shared Patterns

### Early-Dispatch CLI Pattern
**Source:** `main.go` lines 24–31
**Apply to:** `main.go` (validate arm addition)
```go
if len(os.Args) >= 2 && os.Args[1] == "<subcommand>" {
    if err := pkg.RunCLI(os.Args[2:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    return
}
```
Rule: Always before `runServe()` / PocketBase construction.

### Workspace Resolution
**Source:** `internal/config/config.go` lines 14–57
**Apply to:** `internal/validate/cmd.go` `RunCLI`
```go
_, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
if err != nil {
    return err
}
```
Pass raw `args []string` and `os.Environ()` — same signature as `RunInit`.

### Idempotent File-Write Guard
**Source:** `internal/scaffold/init.go` lines 89–94 (`writeScaffoldFile`)
**Apply to:** `installPreCommitHook`
```go
if _, err := os.Stat(dest); err == nil {
    slog.Info("skip existing scaffold file", "path", destRel)
    return nil
} else if !os.IsNotExist(err) {
    return fmt.Errorf("stat %s: %w", destRel, err)
}
```
For hook: check **content** for marker (`bytes.Contains`), not just file existence (D-40 requirement).

### os/exec with .Dir
**Source:** `internal/scaffold/init.go` lines 158–163 (`maybeGitInit`)
**Apply to:** `internal/validate/cmd.go` `getStagedGohtml`
```go
cmd := exec.Command("git", ...)
cmd.Dir = wsAbs     // ← always set Dir; never rely on CWD
out, err := cmd.Output()
```

### Error Wrapping with `%w`
**Source:** `internal/scaffold/init.go` lines 43–65, `internal/schema/sync.go` lines 93
**Apply to:** All new error returns in validate package
```go
return fmt.Errorf("install pre-commit hook: %w", err)
return fmt.Errorf("get staged files: %w", err)
```

### slog Structured Logging
**Source:** `main.go` lines 47–51; `internal/scaffold/init.go` lines 68, 147, 154
**Apply to:** `internal/validate/cmd.go`, `installPreCommitHook`
```go
slog.Info("key noun", "field", value, "field2", value2)
slog.Warn("key noun", "err", err)
```

### testutil.WriteFile for Test Fixtures
**Source:** `internal/testutil/workspace.go` lines 46–48; `internal/router/handlers_test.go` lines 213
**Apply to:** All test files in Phase 2
```go
testutil.WriteFile(t, filepath.Join(ws, "relative/path.gohtml"), `content`)
```

### startTestServer for Integration Tests
**Source:** `internal/router/handlers_test.go` lines 27–82
**Apply to:** `internal/router/press_releases_test.go`
```go
ts, cache, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
defer cleanup()
```
This is package-internal in `router` — integration tests that need it must live in `package router`.

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `workspace/agent-guide.md` | documentation | — | No existing Markdown docs in codebase; write per D-46/D-47 content contract |
| `workspace/SECURITY.md` | documentation | — | No existing security docs; write per D-48/SEC-03 content contract |

---

## Metadata

**Analog search scope:** `/home/terence/code/MonMS/internal/**/*.go`, `main.go`
**Files scanned:** `main.go`, `internal/scaffold/init.go`, `internal/router/ssr.go`, `internal/router/handlers_test.go`, `internal/schema/sync.go`, `internal/config/config.go`, `internal/testutil/workspace.go`, `internal/testutil/buildmode.go`
**Pattern extraction date:** 2026-05-22
**Primary analogs:** `internal/scaffold/init.go` (CLI dispatch + file-write patterns), `internal/router/ssr.go` (ParseFiles template pattern), `internal/router/handlers_test.go` (integration test harness)
