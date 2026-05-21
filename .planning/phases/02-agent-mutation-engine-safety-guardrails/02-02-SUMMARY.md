---
phase: "02"
plan: "02"
subsystem: scaffold
tags: [pre-commit, git-hook, idempotency, rollback, AGT-05]
dependency_graph:
  requires: ["02-01"]
  provides: ["workspace/.git/hooks/pre-commit", "installPreCommitHook"]
  affects: ["internal/scaffold/init.go", "RunInit"]
tech_stack:
  added: []
  patterns: ["idempotency via bytes.Contains marker", "os.WriteFile with 0o755 mode", "MkdirAll for A4 hooks-dir guard"]
key_files:
  created: ["internal/scaffold/hook_test.go"]
  modified: ["internal/scaffold/init.go"]
decisions:
  - "D-40: marker-based idempotency (bytes.Contains) not bare file-existence check"
  - "D-41: hook discovers monms binary via $MONMS_BIN → PATH → ../../monms relative"
  - "D-42: git diff --cached --name-only --diff-filter=ACM | grep .gohtml for staged-file filter"
  - "D-43: git checkout -- . rollback on validation failure, then exit 1"
  - "A4 guard: os.MkdirAll(hooksDir) when hooks dir absent (some git versions omit it)"
  - ".git absence check: silently skip hook install when .git does not exist (git not in PATH)"
metrics:
  duration: "~3 min"
  completed: "2026-05-21T16:37:22Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 1
  files_modified: 1
---

# Phase 02 Plan 02: Pre-Commit Hook Installation Summary

**One-liner:** Pre-commit hook with `#!/bin/sh`, `monms-validate-hook` marker, 3-path binary discovery, staged-gohtml filter, and `git checkout -- .` rollback installed idempotently into `workspace/.git/hooks/pre-commit` by `monms init`.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add installPreCommitHook to init.go | 199eae0 | internal/scaffold/init.go |
| 2 | Create hook_test.go | 0a1f1dd | internal/scaffold/hook_test.go |

## Verification

```
go test ./internal/scaffold/... -count=1 -race -v
```

Result: **PASS** — all 7 tests pass (4 new hook tests + 3 existing scaffold tests), no race conditions.

## Deviations from Plan

None — plan executed exactly as written.

The only implementation judgment call was the additional `.git` existence check before the hooks-dir check: when `maybeGitInit` skips because git is not in PATH, there is no `.git` directory. Rather than let `os.MkdirAll` create a bare `.git/hooks` (surprising behavior), `installPreCommitHook` checks for `.git` first and logs a warning + returns nil. The plan implies this via `t.Skip("git not in PATH")` in the test, but does not make the guard explicit. The test still passes because the skip guard matches the implementation behavior.

## Known Stubs

None.

## Threat Flags

No new security-relevant surface beyond what is documented in the plan's threat model (T-02-05, T-02-06, T-02-07). The hook path derivation is `filepath.Join(wsAbs, ".git", "hooks", "pre-commit")` where `wsAbs` is validated by `config.ResolveWorkspace` before `RunInit` calls `installPreCommitHook` — satisfying T-02-06 mitigation.

## Self-Check: PASSED

- `internal/scaffold/init.go` — exists and contains `installPreCommitHook` ✓
- `internal/scaffold/hook_test.go` — exists and contains all 4 test functions ✓
- Commit `199eae0` — present in git log ✓
- Commit `0a1f1dd` — present in git log ✓
- `go test ./internal/scaffold/... -count=1 -race` exits 0 ✓
