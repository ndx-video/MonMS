---
phase: 01-core-go-runtime-workspace-foundation
plan: 05
subsystem: scaffold
tags: [go, embed, monms-init, tailwind, htmx, alpine, git]

requires:
  - phase: 01-02
    provides: early init dispatch in main.go, ResolveWorkspace
  - phase: 01-01
    provides: ValidateWorkspace, test patterns
provides:
  - monms init with embedded UI-SPEC scaffold templates
  - Idempotent workspace bootstrap with optional git init
  - DEMO-03 automated CDN/layout verification
affects: [01-04]

tech-stack:
  added: [go:embed]
  patterns: [embed.FS scaffold copy, skip-existing idempotent init, workspace path containment]

key-files:
  created:
    - internal/scaffold/embed.go
    - internal/scaffold/embed/base.gohtml
    - internal/scaffold/embed/index.gohtml
    - internal/scaffold/embed/errors.gohtml
    - internal/scaffold/embed/main.css
  modified:
    - internal/scaffold/init.go
    - internal/scaffold/init_test.go

key-decisions:
  - "Skip existing scaffold files on re-init with info log (no force flag in Phase 1)"
  - "git init warns and continues when git missing from PATH"
  - "No .gitignore for .pb_data per D-28"

patterns-established:
  - "Scaffold assets live in internal/scaffold/embed/ copied to workspace on init"
  - "ensureUnderWorkspace guards all writes under resolved abs path (T-01-12)"

requirements-completed: [WRK-01, WRK-02, DEMO-03]

duration: 1min
completed: 2026-05-22
---

# Phase 1 Plan 05: monms init Scaffold Summary

**Embedded UI-SPEC workspace bootstrap with pinned CDN layout, idempotent file writes, and conditional git init**

## Performance

- **Duration:** 1 min
- **Started:** 2026-05-21T15:55:29Z
- **Completed:** 2026-05-21T15:56:12Z
- **Tasks:** 3/3
- **Files modified:** 7

## Accomplishments

- `monms init` writes full Phase 1 workspace tree (templates, assets, schema, fragments) from `go:embed`
- Base layout matches 01-UI-SPEC: Tailwind Play CDN, HTMX 1.9.12, Alpine 3.14.8, hidden editor overlay
- WRK-01/WRK-02/DEMO-03 covered by `TestInitScaffold`, `TestInitGit`, `TestBaseLayoutCDN`

## Task Commits

1. **Task 1: Embedded scaffold assets per UI-SPEC** - `73b7503` (feat)
2. **Task 2: monms init command and git bootstrap** - `cd30eee` (test), `df63d4c` (feat)
3. **Task 3: DEMO-03 CDN and layout verification test** - `9c053c4` (test)

**Plan metadata:** `2f3f237` (docs: complete plan)

## Files Created/Modified

- `internal/scaffold/embed.go` - embed.FS for scaffold assets
- `internal/scaffold/embed/base.gohtml` - base layout with CDN order and editor overlay placeholder
- `internal/scaffold/embed/index.gohtml` - index stub with Alpine mobile nav and hero copy
- `internal/scaffold/embed/errors.gohtml` - styled error page template
- `internal/scaffold/embed/main.css` - component classes (.btn, .card, .hero, .error-page, etc.)
- `internal/scaffold/init.go` - RunInit implementation with path containment and git bootstrap
- `internal/scaffold/init_test.go` - scaffold, git, and DEMO-03 CDN tests

## Decisions Made

- Skip existing files on re-init rather than overwrite (Phase 1 has no --force)
- Continue without git when binary not in PATH (log warning)
- Do not create `.gitignore` entries for `.pb_data/`

## Deviations from Plan

None - plan executed exactly as written.

## TDD Gate Compliance

- RED: `cd30eee` test(01-05) failing init tests
- GREEN: `df63d4c` feat(01-05) RunInit implementation
- Task 3 test: `9c053c4` TestBaseLayoutCDN (passes against GREEN output)

## Self-Check: PASSED

- FOUND: internal/scaffold/embed.go
- FOUND: internal/scaffold/embed/base.gohtml
- FOUND: internal/scaffold/init.go
- FOUND: internal/scaffold/init_test.go
- FOUND: 73b7503
- FOUND: cd30eee
- FOUND: df63d4c
- FOUND: 9c053c4
