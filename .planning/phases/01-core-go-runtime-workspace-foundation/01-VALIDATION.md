---
phase: 1
slug: core-go-runtime-workspace-foundation
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-05-22
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none — Wave 0 installs test packages |
| **Quick run command** | `go test ./... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -short -count=1`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd-verify-work`:** Full suite must be green; manual smoke: `monms init && monms serve`, curl `/`, `/_/`, `/assets/main.css`
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 0 | — | — | N/A | build | `go build -o /dev/null . && go mod verify` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 0 | WRK-01 | T-01-03 | Workspace path resolution; validation before serve | unit | `go test ./internal/config/... ./internal/workspace/... -count=1 -short` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 0 | — | — | N/A | scaffold | `go test ./... -short -count=1` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | ENG-01 | — | N/A | build | `go build -o /dev/null .` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 1 | ENG-02, ENG-04 | — | N/A | unit | `go test ./internal/templates/... -run 'TestCacheFlush\|TestDevNoCache' -count=1 -short` | ❌ W0 | ⬜ pending |
| 01-02-03 | 02 | 1 | ENG-01, SEC-01 | T-01-06 | Admin route reachable | integration | `go test ./internal/router/... -run 'TestServeStarts\|TestAdminDashboard' -count=1 -timeout 120s` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 2 | WRK-04 | — | N/A | unit | `go test ./internal/templates/... -run TestResolveSlug -count=1` | ❌ W0 | ⬜ pending |
| 01-03-02 | 03 | 2 | ENG-02, ENG-04 | — | N/A | unit | `go test ./internal/templates/... -run 'TestCacheFlush\|TestDevNoCache' -count=1` | ❌ W0 | ⬜ pending |
| 01-03-03 | 03 | 2 | ENG-03 | T-01-08 | Debounced fsnotify flush | integration | `go test ./internal/templates/... -run TestWatcherInvalidates -count=1 -timeout 30s` | ❌ W0 | ⬜ pending |
| 01-04-01 | 04 | 3 | WRK-03 | T-01-01 | Asset path jail | integration | `go test ./internal/router/... -run TestAssetsHandler -count=1 -timeout 60s` | ❌ W0 | ⬜ pending |
| 01-04-02 | 04 | 3 | WRK-04, WRK-05 | T-01-02 | SSR homepage + styled 404 | integration | `go test ./internal/router/... -run 'TestHomepageSSR\|Test404NoPanic' -count=1 -timeout 60s` | ❌ W0 | ⬜ pending |
| 01-04-03 | 04 | 3 | ENG-05, ENG-06 | — | N/A | integration | `go test ./internal/router/... -run 'TestIdleMemory\|TestTTFB' -count=1 -timeout 120s` | ❌ W0 | ⬜ pending |
| 01-05-01 | 05 | 2 | DEMO-03 | T-01-13 | CDN pins in embed scaffold | static | `test "$(grep -v '^#' internal/scaffold/embed/base.gohtml \| grep -c 'htmx.org@1.9.12')" -eq 1` | ❌ W0 | ⬜ pending |
| 01-05-02 | 05 | 2 | WRK-01, WRK-02 | T-01-12 | Init writes under workspace root | unit | `go test ./internal/scaffold/... -run 'TestInitScaffold\|TestInitGit' -count=1` | ❌ W0 | ⬜ pending |
| 01-05-03 | 05 | 2 | DEMO-03 | — | N/A | unit | `go test ./internal/scaffold/... -run TestBaseLayoutCDN -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` / module initialization — no Go module exists yet
- [ ] `internal/config/config_test.go` — ResolveWorkspace acceptance cases
- [ ] `internal/workspace/validate_test.go` — ValidateWorkspace acceptance cases
- [ ] `internal/testutil/workspace.go` — temp workspace fixture
- [ ] `internal/testutil/buildmode.go` — production-mode test helper
- [ ] `internal/templates/cache_test.go` — ENG-02, ENG-04
- [ ] `internal/templates/resolver_test.go` — WRK-04, D-10 nested paths
- [ ] `internal/templates/watcher_test.go` — ENG-03
- [ ] `internal/router/handlers_test.go` — ENG-01, WRK-03, WRK-05, SEC-01, TestHomepageSSR
- [ ] `internal/scaffold/init_test.go` — WRK-01, WRK-02, DEMO-03
- [ ] `internal/router/perf_test.go` — ENG-05, ENG-06 benchmarks
- [ ] Build tag or test ldflags helper for production mode tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual base layout renders correctly | DEMO-03 | CDN assets require browser | Load `/` in browser; verify HTMX/Alpine scripts present in page source |
| fsnotify invalidation under real file write | ENG-03 | Timing-sensitive with real filesystem | Edit `.gohtml` while server running; confirm next request reflects change |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
