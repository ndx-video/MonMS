---
phase: "02"
plan: "03"
subsystem: workspace-docs-integration-test
tags: [documentation, security, schema, integration-test, agent-mutation]
dependency_graph:
  requires: [02-01, 02-02]
  provides: [workspace-docs, press-releases-schema, press-releases-integration-test]
  affects: [workspace, internal/router, internal/validate]
tech_stack:
  added: []
  patterns: [httptest-integration, pocketbase-schema-bootstrap, template-validate-dry-run]
key_files:
  created:
    - workspace/agent-guide.md
    - workspace/SECURITY.md
    - workspace/schema/press_releases.json
    - internal/router/press_releases_test.go
  modified: []
decisions:
  - "Integration test uses package router (not router_test) to access unexported startTestServer"
  - "Schema written before startTestServer so RegisterBootstrapHook imports it via Bootstrap()"
  - "cache.Flush() used to simulate fsnotify cache invalidation in test (same production pattern)"
  - "TestValidateRejectsInvalidTemplate checks for 'template parse error' matching validate.go error format"
metrics:
  duration: "~2 minutes"
  completed: "2026-05-21T16:37:49Z"
  tasks_completed: 2
  files_created: 4
  files_modified: 0
---

# Phase 02 Plan 03: Workspace Docs, Schema Fixture, and Integration Test Summary

## One-liner

Workspace agent-guide and SECURITY docs (D-46/D-47/D-48), press_releases schema fixture (D-35), and end-to-end integration test (D-49) covering AGT-01 through AGT-05 — closes Phase 2.

## What Was Built

### Task 1: Workspace Documentation and Schema Fixture

**`workspace/agent-guide.md`** — Six-section mutation workflow guide (D-46, D-47, AGT-06):
1. Overview — dual-write schema, cache invalidation, Phase 2 safety guardrails
2. Prerequisites — MONMS_BIN binary resolution (three paths per D-42), admin token curl example, env var table
3. Schema Dual-Write Workflow — full press_releases curl example, schema JSON write, `agent:` prefix commit (AGT-06)
4. Template Editing Conventions — `{{define "body"}}` requirement, mirror+index slug rules (D-10/D-11), HTMX fetch pattern with press/index.gohtml sample
5. Pre-Commit Validation Lifecycle — hook behavior, manual validate command, new-file rollback caveat (D-43/D-44)
6. press_releases Sample Walkthrough — complete end-to-end steps with debugging note (D-35, D-19)

**`workspace/SECURITY.md`** — Four-section security guide (D-48, SEC-03):
1. SSH Key Scope — dedicated key, `restrict,command=` authorized_keys pattern, compromise response
2. PocketBase Admin Token — `POCKETBASE_ADMIN_TOKEN` env var storage, token scope, rotation procedure
3. Git History Safety — `.pb_data/` gitignore requirement, list of files never to commit, secret purge procedure (BFG)
4. Operator Escape Hatches — `git revert`, `git checkout -- .`, `--no-verify` warning, diff commands

**`workspace/schema/press_releases.json`** — Valid JSON with `name=press_releases`, `type=base`, `fields=[title, body]`, matching `sync.go ImportCollectionsByMarshaledJSON` format.

### Task 2: Integration Test

**`internal/router/press_releases_test.go`** — package `router` (not `router_test`) for access to `startTestServer`:

- **`TestPressReleasesOperation`**: Five-step end-to-end test:
  1. Writes `schema/press_releases.json` to temp workspace before `startTestServer`
  2. Starts server — `RegisterBootstrapHook` imports press_releases collection via `ImportCollectionsByMarshaledJSON` (AGT-01)
  3. Writes `templates/press/index.gohtml` and calls `cache.Flush()` (AGT-02)
  4. Calls `validate.ValidateTemplate` and `validate.ValidateHTML` — both return nil (AGT-03, AGT-04)
  5. GET `/press` returns 200 with "Press Releases" in body (AGT-02 confirmed)

- **`TestValidateRejectsInvalidTemplate`**: Verifies `ValidateTemplate` returns error containing "template parse error" for `{{if}}` with missing condition arg (AGT-05 signal).

## Test Results

```
--- PASS: TestPressReleasesOperation (0.03s)
--- PASS: TestValidateRejectsInvalidTemplate (0.00s)
PASS
ok  github.com/monms/monms/internal/router  0.033s
```

Full suite:
```
ok  github.com/monms/monms/internal/config     0.006s
ok  github.com/monms/monms/internal/router     0.256s
ok  github.com/monms/monms/internal/scaffold   0.014s
ok  github.com/monms/monms/internal/templates  0.508s
ok  github.com/monms/monms/internal/validate   0.003s
ok  github.com/monms/monms/internal/workspace  0.002s
```

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 + 2 | `7865a53` | feat(02-03): workspace docs, schema fixture, and press_releases integration test |

## Deviations from Plan

None — plan executed exactly as written.

## Requirements Satisfied

| Requirement | Evidence |
|---|---|
| AGT-01 | `TestPressReleasesOperation` step 1–2: press_releases collection created via schema bootstrap |
| AGT-02 | `TestPressReleasesOperation` step 3+5: template mutated, `cache.Flush()`, GET `/press` returns 200 |
| AGT-03 | `TestPressReleasesOperation` step 4: `ValidateTemplate` returns nil |
| AGT-04 | `TestPressReleasesOperation` step 4: `ValidateHTML` returns nil |
| AGT-05 | `TestValidateRejectsInvalidTemplate`: bad syntax detected; pre-commit hook documents rollback behavior |
| AGT-06 | `workspace/agent-guide.md` §Schema Dual-Write: `agent:` commit prefix documented |
| SEC-03 | `workspace/SECURITY.md`: SSH key scope, POCKETBASE_ADMIN_TOKEN env storage, .pb_data gitignore |

## Known Stubs

None — all files contain complete, functional content.

## Threat Flags

No new security surface introduced beyond what the plan's threat model covers. The `workspace/agent-guide.md` curl examples use `$POCKETBASE_ADMIN_TOKEN` placeholder (T-02-08 mitigated). The `press_releases.json` schema passes through `ImportCollectionsByMarshaledJSON` validation (T-02-09 accepted). The `agent:` prefix convention is documented but not tooling-enforced in Phase 2 (T-02-10 accepted per plan).

## Self-Check: PASSED

- `workspace/agent-guide.md` ✓
- `workspace/SECURITY.md` ✓
- `workspace/schema/press_releases.json` ✓ (valid JSON, name=press_releases)
- `internal/router/press_releases_test.go` ✓
- Commit `7865a53` ✓
- All tests green ✓
