# Phase 2: Agent Mutation Engine & Safety Guardrails - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-22
**Phase:** 2-Agent Mutation Engine & Safety Guardrails
**Mode:** `--auto` (autonomous selection, no user prompts)
**Areas discussed:** Schema mutation workflow, Validation toolchain, Pre-commit hook lifecycle, Rollback & failure handling

---

## Schema Mutation Workflow

| Option | Description | Selected |
|--------|-------------|----------|
| API-only | POST `/api/collections` only; schema JSON optional/manual export | |
| JSON-only | Write `schema/*.json`; rely on restart/bootstrap sync | |
| Dual-write | POST for live effect + write matching `schema/{name}.json` for git audit | ✓ |

**Auto-selected choice:** Dual-write (recommended default)
**Notes:** `[auto] Schema mutation workflow — Q: "How should agents create collections?" → Selected: "Dual-write" (recommended default). Aligns with D-32 declarative sync, AGT-01 no-restart requirement, and WRK-02 git tracking. Single atomic mutation unit in agent-guide.

---

## Validation Toolchain

| Option | Description | Selected |
|--------|-------------|----------|
| External npm html-validator | Invoke `html-validator` CLI per ROADMAP example | |
| Integrated `monms validate` | Go template parse + `x/net/html` well-formedness in engine binary | ✓ |
| Standalone Go test only | Validation as `go test ./...` without CLI subcommand | |

**Auto-selected choice:** Integrated `monms validate` subcommand
**Notes:** `[auto] Validation toolchain — Q: "Where do validators live?" → Selected: "Integrated monms validate" (recommended default). Zero Node.js dependency matches PROJECT.md constraints. Reuses SSR parse path from `internal/router/ssr.go`.

---

## Pre-Commit Hook Lifecycle

| Option | Description | Selected |
|--------|-------------|----------|
| Manual hook setup | Agent/operator copies hook script documented in agent-guide | |
| `monms init` installs hook | Idempotent hook write during workspace bootstrap | ✓ |
| Separate `monms install-hooks` command | Dedicated command for hook management | |

**Auto-selected choice:** `monms init` installs hook
**Notes:** `[auto] Pre-commit hook lifecycle — Q: "When is the hook installed?" → Selected: "monms init installs hook" (recommended default). Extends existing D-08 git init flow. Marker-based idempotency prevents overwriting operator customizations blindly.

---

## Rollback & Failure Handling

| Option | Description | Selected |
|--------|-------------|----------|
| Full `git checkout -- .` | Restore entire working tree on any validation failure | ✓ |
| Staged-only reset | `git restore --staged` + selective file checkout | |
| Abort without rollback | Fail commit but leave working tree dirty for agent inspection | |

**Auto-selected choice:** Full `git checkout -- .`
**Notes:** `[auto] Rollback & failure handling — Q: "What happens on validation failure?" → Selected: "Full git checkout -- ." (recommended default). Matches AGT-05 and PRD §4.3 instant revert semantics.

---

## Claude's Discretion

- Validate subcommand stderr format and exit codes
- Optional `--verbose` / `--files` flags
- Integration test harness details (httptest + embedded PocketBase)
- Agent-guide section ordering beyond required walkthrough

## Deferred Ideas

- EXT-01 REST webhook rollback — v2
- Auto git commits on human inline edits — Phase 3
- In-app agent chatbox/API hook — future phase
- Schema JSON pre-commit validation — deferred unless low-cost during planning
