---
phase: 01-core-go-runtime-workspace-foundation
plan: 01
subsystem: testing
tags: [go, pocketbase, fsnotify, workspace, test-scaffold]

requires: []
provides:
  - Go module github.com/monms/monms with PocketBase v0.38.1 and fsnotify v1.9.0
  - Workspace path resolution (flag-over-env) and structural validation
  - testutil.NewWorkspace temp fixture for downstream integration tests
  - Wave 0 skeleton tests named per 01-VALIDATION.md verification map
affects: [01-02, 01-03, 01-04, 01-05]

tech-stack:
  added: [github.com/pocketbase/pocketbase@v0.38.1, github.com/fsnotify/fsnotify@v1.9.0]
  patterns: [ResolveWorkspace flag-over-env precedence, ValidateWorkspace before serve, ldflags buildMode default development]

key-files:
  created:
    - go.mod
    - main.go
    - internal/config/config.go
    - internal/workspace/validate.go
    - internal/testutil/workspace.go
    - internal/testutil/buildmode.go
    - internal/templates/*_test.go
    - internal/router/*_test.go
    - internal/scaffold/init_test.go
  modified: []

key-decisions:
  - "filepath.Clean applied only to flag/env overrides; default ./workspace preserved for D-31 logging"
  - "Skeleton tests use t.Skip with plan ownership comments until feature plans land"

patterns-established:
  - "ResolveWorkspace(args, env) returns (configured, absolute, err) for logging vs file ops"
  - "ValidateWorkspace errors include 'Run: monms init' for D-06 operator guidance"
  - "testutil.NewWorkspace creates minimal valid workspace passing ValidateWorkspace"

requirements-completed: [WRK-01]

duration: 1min
completed: 2026-05-22
---

# Phase 1 Plan 01: Wave 0 Foundation Summary

**Go module with PocketBase/fsnotify deps, workspace resolution and validation packages, and full Wave 0 test scaffold for Nyquist verification targets**

## Performance

- **Duration:** 1 min
- **Started:** 2026-05-22T01:49:18+10:00
- **Completed:** 2026-05-22T01:49:55Z
- **Tasks:** 3
- **Files modified:** 15

## Accomplishments

- Initialized `github.com/monms/monms` with stub `main.go` declaring `buildMode = "development"` (D-01)
- Implemented `config.ResolveWorkspace` with `--workspace` flag, `MONMS_WORKSPACE` env, and `./workspace` default (D-25, D-26)
- Implemented `workspace.ValidateWorkspace` requiring base layout and assets with `monms init` guidance (D-06, D-09)
- Created `testutil.NewWorkspace` fixture and seven skeleton test files matching 01-VALIDATION.md test names

## Task Commits

Each task was committed atomically:

1. **Task 1: Initialize Go module and package skeleton** - `c07ba04` (feat)
2. **Task 2: Workspace config, validation, and testutil fixture** - `adf5c3d` (feat)
3. **Task 3: Wave 0 test scaffolds per validation map** - `b440391` (feat)

**Plan metadata:** pending (docs commit follows)

## Files Created/Modified

- `go.mod` / `go.sum` — Module with PocketBase v0.38.1 and fsnotify v1.9.0
- `main.go` — Stub entry with `buildMode` defaulting to development
- `internal/config/config.go` — `ResolveWorkspace` with flag-over-env precedence
- `internal/workspace/validate.go` — Structural workspace validation before serve
- `internal/testutil/workspace.go` — Temp workspace fixture for tests
- `internal/testutil/buildmode.go` — Production-mode test helper documenting ldflags pattern
- `internal/templates/*_test.go` — Cache, resolver, watcher skeleton tests
- `internal/router/*_test.go` — Handler and perf skeleton tests
- `internal/scaffold/init_test.go` — Init scaffold test (partial WRK-01) plus skipped git/CDN tests

## Decisions Made

- Apply `filepath.Clean` only to operator-provided paths (flag/env), preserving `./workspace` literal for startup logging (D-31)
- Skeleton tests skip with plan ownership messages; `TestInitScaffold` runs real assertions against testutil fixture

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Preserve default `./workspace` configured path after Clean**
- **Found during:** Task 2 (config tests)
- **Issue:** `filepath.Clean("./workspace")` returned `workspace`, failing configured-path assertion
- **Fix:** Apply Clean only when path comes from flag or env override (T-01-03 still mitigates untrusted input)
- **Files modified:** internal/config/config.go
- **Verification:** `go test ./internal/config/... -short` passes
- **Committed in:** adf5c3d

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary for D-25/D-31 configured-path contract. No scope creep.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 01-02 can implement PocketBase bootstrap and wire `buildMode` into cache/watcher
- All Wave 0 verification targets exist; downstream plans replace `t.Skip` with real assertions
- `testutil.NewWorkspace` ready for router and scaffold integration tests

## Self-Check: PASSED

- All key files FOUND
- Commits c07ba04, adf5c3d, b440391 verified
- `go test ./... -short -count=1` green
- `go mod verify` passes

---
*Phase: 01-core-go-runtime-workspace-foundation*
*Completed: 2026-05-22*
