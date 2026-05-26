---
phase: 01-core-go-runtime-workspace-foundation
verified: 2026-05-22T06:00:00Z
status: passed
score: 12/12 must-haves verified
overrides_applied: 0
re_verification: false
---

# Phase 1: Core Go Runtime & Workspace Foundation Verification Report

**Phase Goal:** A working Go + PocketBase binary that serves templates from the workspace folder, supports dynamic cache invalidation via fsnotify, and provides static asset serving with a correct base layout.

**Verified:** 2026-05-22T06:00:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #   | Truth | Status | Evidence |
| --- | ----- | ------ | -------- |
| 1 | Go + PocketBase binary starts and serves HTTP without config files (ENG-01) | ✓ VERIFIED | `main.go` embeds PocketBase; `TestServeStarts` returns non-5xx from `/api/health` |
| 2 | Templates served from workspace via SSR catch-all `/{slug...}` (WRK-04) | ✓ VERIFIED | `SSRHandler` calls `ResolveSlug` + `ParseFiles`; `TestHomepageSSR` renders index content |
| 3 | Template cache populated on first access (ENG-02) | ✓ VERIFIED | `TemplateCache.Get` loads via loader on miss; `TestCacheFlush` confirms single load then hit |
| 4 | fsnotify clears cache on `.gohtml` write/create (ENG-03) | ✓ VERIFIED | `StartWatcher` debounces 100ms → `Flush()`; `TestWatcherInvalidates` confirms reload within 500ms |
| 5 | Production caches templates; development re-reads from disk (ENG-04) | ✓ VERIFIED | `buildMode` gates `SetProductionMode` + watcher; `TestDevNoCache` (2 loads) vs cached prod (1 load) |
| 6 | Static assets served from `/assets/{path}` with traversal blocked (WRK-03) | ✓ VERIFIED | `AssetsHandler` + `safeAssetPath`; `TestAssetsHandler` serves CSS; `TestSafeAssetPath` blocks `../../../etc/passwd` |
| 7 | `base.gohtml` includes HTMX 1.9.12, Alpine 3.14.8, Tailwind CDN, `{{template "body" .}}`, editor overlay block (DEMO-03) | ✓ VERIFIED | `internal/scaffold/embed/base.gohtml`; `TestBaseLayoutCDN` asserts all CDN pins |
| 8 | Unknown slug returns styled 404 with attempted path, no panic (WRK-05) | ✓ VERIFIED | `renderErrorPage` via `errors.gohtml`; `Test404NoPanic` asserts `Page not found: /does-not-exist` |
| 9 | `monms init` scaffolds full workspace tree and runs `git init` when absent (WRK-01, WRK-02) | ✓ VERIFIED | `scaffold.RunInit`; `TestInitScaffold`, `TestInitGit`; live `./monms init` produced `.git` + dirs |
| 10 | PocketBase admin dashboard reachable at `/_/` (SEC-01) | ✓ VERIFIED | `TestAdminDashboard` returns 200 or 302 |
| 11 | Production/dev mode via `buildMode` ldflags (D-01) | ✓ VERIFIED | `var buildMode = "development"` in `main.go`; gates cache + watcher at lines 53–58 |
| 12 | Idle heap < 30MB and TTFB p50 < 15ms (ENG-05, ENG-06) | ✓ VERIFIED | `TestIdleMemory` PASS; `TestTTFB` p50 PASS (short mode) |

**Score:** 12/12 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `main.go` | PocketBase bootstrap, init dispatch, route registration | ✓ VERIFIED | 82 lines; `RunInit` + `runServe` with watcher, schema hook, router |
| `go.mod` | PocketBase v0.38.1, fsnotify v1.9.0 | ✓ VERIFIED | Direct require on pocketbase; fsnotify in dependency tree |
| `internal/templates/cache.go` | `sync.RWMutex` cache with Get/Flush/Active | ✓ VERIFIED | 70 lines; production gate; wired from SSR + fragments |
| `internal/templates/watcher.go` | Recursive fsnotify with debounce | ✓ VERIFIED | Watches entire workspace; `.gohtml` filter; calls `Flush` |
| `internal/templates/resolver.go` | Mirror+index slug resolution | ✓ VERIFIED | Empty→index, flat file, directory index, trailing slash strip |
| `internal/router/ssr.go` | Catch-all SSR with error pages | ✓ VERIFIED | Reserved slug guard; base layout execution |
| `internal/router/assets.go` | Root-jailed static handler | ✓ VERIFIED | Path prefix check; `http.ServeFile` |
| `internal/router/fragments.go` | HTMX partials without layout | ✓ VERIFIED | `TestFragmentPartial` confirms no DOCTYPE |
| `internal/scaffold/init.go` | Workspace bootstrap | ✓ VERIFIED | Embedded files; git init; no `.pb_data` gitignore |
| `internal/scaffold/embed/base.gohtml` | Phase 1 UI contract layout | ✓ VERIFIED | CDN pins, editor overlay div, body slot |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| `main.go` | `templates.StartWatcher` | production buildMode gate | ✓ WIRED | Lines 55–58 |
| `watcher.go` | `cache.go` | `onChange` → `Flush()` | ✓ WIRED | Debounced callback clears map |
| `ssr.go` | `resolver.go` | `ResolveSlug` before render | ✓ WIRED | Line 40 |
| `ssr.go` | `cache.go` | `Get` + `ExecuteTemplate("base")` | ✓ WIRED | Lines 58–65 |
| `main.go` | `scaffold.RunInit` | `os.Args[1]=="init"` | ✓ WIRED | Lines 25–30 |
| `main.go` | `router.RegisterRoutes` | `OnServe` assets→fragments→slug | ✓ WIRED | Lines 68–74 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| `SSRHandler` | `pagePath` | `ResolveSlug(wsAbs, slug)` + `os.Stat` | Yes — filesystem lookup | ✓ FLOWING |
| `SSRHandler` | `tmpl` | `cache.Get(cacheKey, loader)` → `ParseFiles` | Yes — parsed layout+page | ✓ FLOWING |
| `AssetsHandler` | `absPath` | `safeAssetPath` + `os.Stat` | Yes — real CSS bytes served | ✓ FLOWING |
| `FragmentsHandler` | `tmpl` | `ParseFiles(fragPath)` | Yes — fragment HTML | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full test suite | `go test ./... -count=1 -short` | exit 0, all packages ok | ✓ PASS |
| Binary builds | `go build .` | exit 0, produces `monms` (~33MB) | ✓ PASS |
| Init scaffolds workspace | `./monms init --workspace /tmp/monms-verify-ws` | `.git`, `templates/`, `assets/`, `schema/` created | ✓ PASS |
| Perf gates | `go test ./internal/router/... -run 'TestIdleMemory\|TestTTFB' -short -v` | both PASS | ✓ PASS |

### Probe Execution

Step 7c: SKIPPED — no `probe-*.sh` scripts declared or present in repository.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| ENG-01 | 01-02 | Go binary serves HTTP with embedded PocketBase | ✓ SATISFIED | `TestServeStarts` |
| ENG-02 | 01-02, 01-03 | Cache populated on first access, invalidated on change | ✓ SATISFIED | `cache.go`, `TestCacheFlush`, `TestWatcherInvalidates` |
| ENG-03 | 01-03 | fsnotify clears cache on write/create | ✓ SATISFIED | `watcher.go` watches workspace tree for `.gohtml` (D-30 supersedes REQUIREMENTS literal `templates/` path) |
| ENG-04 | 01-02, 01-03 | Production caches; dev reads disk | ✓ SATISFIED | `TestDevNoCache`, `buildMode` gate |
| ENG-05 | 01-04 | Idle RAM < 30MB | ✓ SATISFIED | `TestIdleMemory` PASS |
| ENG-06 | 01-04 | TTFB p50 < 15ms | ✓ SATISFIED | `TestTTFB` PASS |
| WRK-01 | 01-01, 01-05 | Workspace directory structure | ✓ SATISFIED | `TestInitScaffold` |
| WRK-02 | 01-05 | Workspace is Git repo | ✓ SATISFIED | `TestInitGit`, live init creates `.git` |
| WRK-03 | 01-04 | Static assets from `/assets/{path}` | ✓ SATISFIED | `AssetsHandler`, `TestAssetsHandler` |
| WRK-04 | 01-03, 01-04 | Slug→template + base layout merge | ✓ SATISFIED | `ResolveSlug`, `ParseFiles(layout, page)` |
| WRK-05 | 01-04 | Unknown slug → descriptive 404 | ✓ SATISFIED | `Test404NoPanic` |
| SEC-01 | 01-02 | Admin dashboard at `/_/` | ✓ SATISFIED | `TestAdminDashboard` |
| DEMO-03 | 01-05 | Base layout CDN + editor overlay block | ✓ SATISFIED | `embed/base.gohtml`, `TestBaseLayoutCDN` |

No orphaned Phase 1 requirements — all 13 mapped IDs have implementation evidence.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | None | — | No TBD/FIXME/stub markers in Phase 1 source |

### Human Verification Required

None required for phase goal achievement. Automated integration tests cover fsnotify invalidation (`TestWatcherInvalidates`) and SSR output (`TestHomepageSSR`). Optional smoke test documented in `01-VALIDATION.md` (browser CDN load) is supplementary, not a blocking must-have.

### Gaps Summary

No gaps found. Phase 1 goal is achieved in the codebase with passing build and test evidence.

**Note:** ROADMAP deliverable text mentions `ENV` env var for mode toggle; implementation correctly follows Phase 1 context decision D-01 (`buildMode` ldflags). ROADMAP wording is stale, not a code gap.

---

_Verified: 2026-05-22T06:00:00Z_
_Verifier: Claude (gsd-verifier)_
