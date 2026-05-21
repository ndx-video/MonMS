# Phase 2: Agent Mutation Engine & Safety Guardrails - Context

**Gathered:** 2026-05-22
**Status:** Ready for planning

<domain>
## Phase Boundary

Enable an external AI agent to safely mutate the Git-tracked workspace — creating PocketBase collections via REST API, editing templates and schema JSON — with pre-commit validation (Go template dry-run + HTML structure check) and automatic `git checkout -- .` rollback on failure. Phase 2 delivers operator/agent documentation, validation tooling, and git hook wiring. It does not include inline contextual editing (Phase 3), chatbox/API agent orchestration, or REST webhook rollback (v2 EXT-01).

</domain>

<decisions>
## Implementation Decisions

### Schema Mutation Workflow
- **D-33:** Agent schema mutations use **dual-write**: POST to `/api/collections` for immediate live effect (AGT-01, no restart), then write matching `workspace/schema/{collection_name}.json` in the same atomic mutation unit for Git audit trail and D-32 bootstrap sync compatibility.
- **D-34:** Schema JSON files must match the format consumed by `internal/schema/sync.go` — single collection object or array batch per file; invalid JSON is log-and-continue on serve (Phase 1 behavior), but pre-commit validation for schema files is out of scope unless the agent modifies them in the same commit as templates.
- **D-35:** Collection creation sample walkthrough uses `press_releases` with `title` (text) and `body` (text) fields — canonical example in `workspace/agent-guide.md` and integration test fixture.

### Validation Toolchain
- **D-36:** Validation is an integrated **`monms validate`** CLI subcommand in the engine binary — not a separate script or npm dependency. Zero Node.js pipeline required.
- **D-37:** Template dry-run reuses production parse path: for each staged/modified `*.gohtml`, parse with `templates/layouts/base.gohtml` via `html/template.ParseFiles` matching `internal/router/ssr.go` loader — no PocketBase server required (AGT-03).
- **D-38:** HTML structure check uses Go `golang.org/x/net/html` tokenizer/parser for well-formedness (balanced tags, valid nesting) on modified `.gohtml` files — no external `html-validator` npm package (AGT-04). Fail on parse errors with file path and line context in stderr.
- **D-39:** `monms validate` accepts `--workspace` flag (same resolution as serve/init per D-26) and optionally `--files` for explicit paths; default reads git-staged `.gohtml` paths from workspace repo when invoked from pre-commit hook.

### Pre-Commit Hook Lifecycle
- **D-40:** `monms init` installs an executable `workspace/.git/hooks/pre-commit` when git is initialized (extends D-08) — idempotent: skip if hook exists and contains `monms-validate-hook` marker, otherwise write/update.
- **D-41:** Pre-commit hook runs `monms validate` against staged `*.gohtml` files only — schema-only or CSS-only commits skip template validation unless `.gohtml` staged.
- **D-42:** Hook discovers `monms` binary via `$MONMS_BIN` env var if set, else `command -v monms`, else relative `../../monms` from workspace (document all three in agent-guide). Fail loudly if binary not found.

### Rollback & Failure Handling
- **D-43:** On any validation failure, pre-commit hook runs `git checkout -- .` in workspace root to restore entire working tree (AGT-05), then exits non-zero with validation errors printed — matches PRD instant revert semantics.
- **D-44:** Rollback is all-or-nothing — no partial file restore or staged-only reset. Agent must re-apply mutation after fixing validation errors.
- **D-45:** Successful validation allows commit to proceed; agent uses descriptive commit messages with `agent:` prefix, e.g. `agent: add press_releases collection and press index template` (AGT-06).

### Agent Documentation & Security
- **D-46:** `workspace/agent-guide.md` documents end-to-end mutation workflow: read current state → dual-write schema → edit templates following D-10/D-11 routing conventions → run `monms validate` manually before commit → git commit → verify in browser.
- **D-47:** Template editing conventions in agent-guide: always use `{{define "body"}}` in page templates, extend `base.gohtml` layout, HTMX fetch patterns for collection lists, mirror+index slug rules from Phase 1.
- **D-48:** `workspace/SECURITY.md` documents SEC-03: dedicated SSH key restricted to workspace subdirectory only, PocketBase admin API token stored outside git (env/vault), token scoped to collection management, rotate on compromise, never commit `.pb_data/` or credentials.
- **D-49:** Phase 2 validates the sample `press_releases` operation via integration test — POST collection (or schema import), commit template, pre-commit passes, page renders without restart (AGT-02).

### Claude's Discretion
- Exact stderr message format and exit codes for `monms validate`.
- Whether to add `--fix` or `--verbose` flags on validate subcommand.
- Pre-commit hook shell vs embedded script content (bash assumed available).
- Integration test harness: httptest against running PocketBase vs mock — prefer real embedded PocketBase in testutil pattern from Phase 1.
- Agent-guide tone and section ordering beyond required walkthrough content.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `specs/monms-prd.md` §4 — Agent Mutation Engine: schema/UI mutation flow, safety guardrails (validation, HTML lint, git rollback)
- `specs/monms-prd.md` §7.2–§7.3 — Version control audit trail, agent SSH/API token permissions
- `.planning/PROJECT.md` — Safety guardrails requirement, Git-managed state, zero-compilation constraint
- `.planning/REQUIREMENTS.md` — AGT-01 through AGT-06, SEC-03 traceability for Phase 2
- `.planning/ROADMAP.md` — Phase 2 deliverables list (agent-guide, validators, pre-commit hook, SECURITY.md, press_releases sample)

### Phase 1 Foundation (locked — do not re-litigate)
- `.planning/phases/01-core-go-runtime-workspace-foundation/01-CONTEXT.md` — D-01 through D-32 (routing, workspace layout, schema sync, fsnotify, init command)
- `internal/schema/sync.go` — Declarative schema JSON import format and bootstrap hook
- `internal/templates/resolver.go` — Mirror+index slug resolution for template validation context
- `internal/router/ssr.go` — Template parse/execute pattern validate must mirror
- `internal/scaffold/init.go` — Init command extension point for hook installation

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/templates/cache.go` + `resolver.go` — Template parse/load logic for dry-run validator
- `internal/router/ssr.go` — Canonical `ParseFiles(layoutPath, pagePath)` + `ExecuteTemplate(w, "base", data)` pattern
- `internal/schema/sync.go` — Schema JSON format and `ImportCollectionsByMarshaledJSON` for dual-write verification
- `internal/scaffold/init.go` — `maybeGitInit()` hook point; extend to install pre-commit after git init
- `internal/testutil/workspace.go` — Temp workspace fixture for validation and hook integration tests

### Established Patterns
- Early CLI dispatch in `main.go` (`init` before serve) — add `validate` subcommand same pattern
- Workspace path resolution via `internal/config/config.go` — reuse for validate subcommand
- Phase 1 log-and-continue for bad schema JSON on serve — agent mutations should produce valid JSON proactively
- fsnotify debounced cache flush already handles agent commits in production (01-03) — no Phase 2 watcher changes needed

### Integration Points
- Pre-commit hook lives in `workspace/.git/hooks/` — separate from engine repo hooks
- Agent operates on workspace Git repo, not engine source repo
- PocketBase `/api/collections` admin endpoint for AGT-01 — requires admin auth token documented in SECURITY.md
- Template changes visible on next request via existing cache invalidation — AGT-02 satisfied by Phase 1 infrastructure

</code_context>

<specifics>
## Specific Ideas

- PRD §4.3 specifies three guardrail stages: Go template dry-run, HTML lint, git checkout rollback — Phase 2 implements all three without external npm toolchain.
- Sample operation from ROADMAP: create `press_releases` collection via POST `/api/collections` — use as integration test and agent-guide centerpiece.
- Agent-guide should reference Phase 1 error page path display (D-19) as debugging aid when slug/template mismatch occurs after mutation.
- SECURITY.md should warn that `.pb_data/` must stay out of workspace git (D-28 operator responsibility).

</specifics>

<deferred>
## Deferred Ideas

- **EXT-01 REST webhook / chatbox git rollback** — v2 backlog; Phase 2 only documents manual `git log` / `git revert` in SECURITY.md as operator escape hatch
- **Automatic git commits on human inline edits** — PRD NFR §7.2 mentions; belongs to Phase 3 inline editing workflow
- **Chatbox or in-app agent API hook** — PRD §4 describes future interaction layer; Phase 2 documents external agent workflow only
- **Schema JSON pre-commit validation** — useful but not AGT requirement; defer unless planner finds low-cost addition
- **Agent-managed CSS variable theming (EXT-03)** — v2 backlog

</deferred>

---

*Phase: 2-Agent Mutation Engine & Safety Guardrails*
*Context gathered: 2026-05-22*
