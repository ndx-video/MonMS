---
phase: "02"
plan: "01"
subsystem: "validate"
tags: [validation, cli, html, template, safety]
dependency_graph:
  requires: [internal/router/ssr.go, internal/config/config.go, internal/testutil/workspace.go]
  provides: [internal/validate package, monms validate CLI subcommand]
  affects: [main.go, pre-commit hook (Wave 2), integration test (Wave 3)]
tech_stack:
  added: [golang.org/x/net/html (already transitive dep — no go.mod change)]
  patterns: [tag-balance stack tokenizer, (?s) regexp directive stripping, early-dispatch CLI arm]
key_files:
  created:
    - internal/validate/validate.go
    - internal/validate/cmd.go
    - internal/validate/validate_test.go
  modified:
    - main.go
decisions:
  - "D-37: ValidateTemplate uses template.ParseFiles(layoutPath, filePath) — identical two-arg call as ssr.go loader; guarantees serve-time parse compatibility"
  - "D-38: ValidateHTML uses (?s) dotall regexp to strip multi-line {{...}} directives before x/net/html tokenization; avoids false positives on range/if blocks spanning lines"
  - "D-39: Positional args to RunCLI serve as explicit --files paths; absence triggers getStagedGohtml fallback"
  - "D-42: getStagedGohtml failure is graceful (slog.Warn); git unavailable is not a fatal error"
  - "T-02-01: ValidateFiles verifies filepath.Rel(wsAbs, f) before os.ReadFile — rejects paths outside workspace"
  - "T-02-04: getStagedGohtml sets cmd.Dir=wsAbs; no user-supplied git args injected"
metrics:
  duration: "~15 minutes"
  completed: "2026-05-22"
  tasks_completed: 2
  files_changed: 4
---

# Phase 02 Plan 01: validate package + CLI subcommand Summary

**One-liner:** Go-native template dry-run + HTML tag-balance validator wired as `monms validate` CLI subcommand via early-dispatch in main.go, before PocketBase bootstrap.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create internal/validate/validate.go and validate_test.go | f2ca0f4 | validate.go, validate_test.go |
| 2 | Create internal/validate/cmd.go and wire main.go validate dispatch | 2fb1121 | cmd.go, main.go |

## Implementation Details

### ValidateTemplate
Mirrors `internal/router/ssr.go` lines 54–56 exactly:
```go
template.ParseFiles(layoutPath, filePath)  // layout first, same two-arg call as ssr.go
```
Any `.gohtml` passing `ValidateTemplate` is guaranteed to parse at serve time (D-37).

### ValidateHTML
- Package-level `(?s)\{\{.*?\}\}` regexp strips multi-line directives before tokenization
- `golang.org/x/net/html` tokenizer with a `[]string` tag stack
- Void-element map covers all 14 HTML5 void elements (area, base, br, col, embed, hr, img, input, link, meta, param, source, track, wbr)
- Reports: unexpected close tags, mismatched tags, unclosed tags

### ValidateFiles
- Calls both validators per file; accumulates all errors (continue-on-error, sync.go pattern)
- T-02-01 path traversal guard via `filepath.Rel` before `os.ReadFile`

### RunCLI / getStagedGohtml
- `flag.NewFlagSet("validate", flag.ContinueOnError)` + `config.ResolveWorkspace` — identical to `RunInit` pattern
- Positional args = explicit file paths (D-39 `--files` equivalent)
- `getStagedGohtml`: `git diff --cached --name-only --diff-filter=ACM` with `cmd.Dir=wsAbs`
- D-42: git failure → `slog.Warn`, treat as empty (graceful)

### main.go dispatch
Validate arm inserted after `init` arm, before `runServe()` — PocketBase never constructed on `monms validate`.

## Deviations from Plan

None — plan executed exactly as written. Threat mitigations T-02-01 and T-02-04 applied as directed by plan threat model.

## Test Results

```
go test ./internal/validate/... -count=1 -race -v
PASS: TestValidateTemplate
PASS: TestValidateTemplateBadSyntax
PASS: TestValidateHTML/valid_div
PASS: TestValidateHTML/unclosed_div
PASS: TestValidateHTML/void_elements_ok
PASS: TestValidateHTML/template_directives_stripped
PASS: TestValidateHTML/multiline_range_directive
PASS: TestValidateHTML/gt_in_directive
PASS: TestValidateHTML/mismatched_close
PASS: TestValidateFilesAggregatesErrors
ok  github.com/monms/monms/internal/validate  1.011s

go test ./... -count=1 -short
ok  github.com/monms/monms/internal/config
ok  github.com/monms/monms/internal/router
ok  github.com/monms/monms/internal/scaffold
ok  github.com/monms/monms/internal/templates
ok  github.com/monms/monms/internal/validate
ok  github.com/monms/monms/internal/workspace
```

All Phase 1 tests remain green.

## Known Stubs

None.

## Threat Flags

None — no new network endpoints, auth paths, or trust boundaries introduced. All T-02-0x mitigations applied as specified in plan threat model.

## Self-Check: PASSED

- `internal/validate/validate.go` — FOUND
- `internal/validate/cmd.go` — FOUND
- `internal/validate/validate_test.go` — FOUND
- `main.go` validate dispatch — FOUND
- Commit f2ca0f4 — FOUND
- Commit 2fb1121 — FOUND
