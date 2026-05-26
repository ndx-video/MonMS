---
phase: 04-staging-environments-client-content-publish
verified: 2026-05-26T22:00:00Z
status: passed
score: 14/14 must-haves verified
overrides_applied: 0
re_verification: false
---

# Phase 4: Staging Environments & Client Content Publish — Verification Report

**Phase Goal:** Clients publish editorial content from staging to production via admin UI; structure continues to promote via Git tags; media uses shared CDN URLs.

**Verified:** 2026-05-26T22:00:00Z  
**Status:** passed  
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Four layers (L1–L4) documented with distinct promotion rails | ✓ VERIFIED | `workspace/README.md` L1–L4 table; root `README.md` cross-links `specs/staging.md` |
| 2 | Structure promotes via Git tags; content via JSON upsert independently | ✓ VERIFIED | `workspace/README.md` dual-rail table; `specs/staging.md` referenced |
| 3 | Staging and production are separate instances with separate `.pb_data/` | ✓ VERIFIED | `workspace/README.md` § environments; `main.go` `DefaultDataDir: filepath.Join(abs, ".pb_data")` |
| 4 | Editorial collections marked `"editorial": true` in schema JSON | ✓ VERIFIED | `workspace/schema/hero_content.json`; `internal/schema/editorial.go` parses raw JSON (D-54) |
| 5 | `workspace/content/*.json` holds exported editorial records with stable IDs | ✓ VERIFIED | `internal/content/export.go` `CollectionFile` shape; `TestExportAllWritesHeroContent` |
| 6 | `monms content export` writes editorial snapshots | ✓ VERIFIED | `internal/content/cmd.go` `runExport`; `TestContentExportCLI` |
| 7 | `monms content import` upserts idempotently by ID | ✓ VERIFIED | `internal/content/import.go` `UpsertRecord`; `TestImportFilesIdempotent`, `TestContentImportCLIIdempotent` |
| 8 | `POST /api/monms/content/import` requires valid Bearer publish token | ✓ VERIFIED | `internal/content/auth.go` `RequirePublishToken`; `TestImportAPIUnauthorized`, `TestImportAPIFailClosedEmptyToken` |
| 9 | Import rejects non-editorial collection names | ✓ VERIFIED | `ImportPayload` allowlist; `TestImportAPIUnauthorized/non editorial`, `TestImportPayloadRejectsNonEditorial` |
| 10 | Staging admin UI at `/api/monms/publish` with diff preview | ✓ VERIFIED | `internal/content/publish_handlers.go`, `embed/publish.gohtml`; `TestPublishUIReturns200` |
| 11 | Publisher role gates publish; editors edit without publishing | ✓ VERIFIED | `RequirePublisher` + `IsPublisher`; `TestPublisherGate` (403 editor, 200 publisher) |
| 12 | Editor badge links **Publish to live** for publishers only | ✓ VERIFIED | `workspace/templates/layouts/base.gohtml` `{{if .IsPublisher}}`; `internal/router/ssr.go` sets `IsPublisher`; `TestEditorBadge_*` |
| 13 | Publish-state checksum tracks unpublished changes (PUB-08/09) | ✓ VERIFIED | `internal/content/state.go`, `checksum.go`, `diff.go`; `TestDiffExportDetectsTitleChange`; `TestPublisherGate` updates `publish-state.json` |
| 14 | Publishable media uses CDN URLs; file blobs not copied (MED-01/02) | ✓ VERIFIED | `export.go` skips `FieldTypeFile` with `slog.Warn`; `TestExportSkipsFileFields`; `workspace/MEDIA.md` |

**Score:** 14/14 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/schema/editorial.go` | Parse editorial flag from schema JSON | ✓ VERIFIED | Substantive parser; wired via `internal/content/schema.go` |
| `internal/content/export.go` | Export to `workspace/content/` | ✓ VERIFIED | `ExportAll`, `ExportCollection`; used by CLI and publish handler |
| `internal/content/import.go` | Idempotent upsert by ID | ✓ VERIFIED | `UpsertRecord`, `ImportPayload`, `ImportFiles` |
| `internal/content/state.go` | Publish-state checksum tracking | ✓ VERIFIED | `ReadPublishState`, `WritePublishState` with path guard |
| `internal/content/diff.go` | Diff vs publish-state | ✓ VERIFIED | `DiffExport` compares checksum + field changes |
| `internal/content/cmd.go` | `monms content` CLI | ✓ VERIFIED | export/import/diff/publish; wired in `main.go` line 41–46 |
| `internal/content/auth.go` | Publish token middleware | ✓ VERIFIED | Constant-time compare; fail-closed on empty token |
| `internal/content/routes.go` | Import + publish routes | ✓ VERIFIED | `RegisterRoutes` in `main.go` OnServe before SSR |
| `internal/content/embed/publish.gohtml` | Publish console UI | ✓ VERIFIED | Full HTML with status, diff, Publish now form |
| `workspace/.monms/config.example.json` | Staging config template | ✓ VERIFIED | `productionUrl`, `publisherEmails` |
| `workspace/MEDIA.md` | CDN policy (MED-02) | ✓ VERIFIED | Warns against PocketBase-local file storage |
| `workspace/.gitignore` | Ignore secrets and publish state | ✓ VERIFIED | `.monms/config.json`, `publish-state.json`, `.pb_data/`, `content/` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `export.go` | `editorial.go` | `LoadEditorialCollectionNames` | ✓ WIRED | Filters editorial collections only |
| `import.go` | PocketBase `app.Save` | `UpsertRecord` | ✓ WIRED | FindRecordById → NewRecord → Save |
| `main.go` | `content.RegisterRoutes` | OnServe bind | ✓ WIRED | Before `router.RegisterRoutes` (D-14) |
| `auth.go` | `POST /api/monms/content/import` | `RequirePublishToken` Bind | ✓ WIRED | Middleware on import route |
| `POST /api/monms/publish` | production import API | `PublishToProduction` | ✓ WIRED | Export snapshot → HTTP POST → `WritePublishState` on success |
| `GET /api/monms/publish` | `diff.go` | `DiffExport` | ✓ WIRED | Page data + `/api/monms/publish/diff` JSON |
| `ssr.go` | `IsPublisher` | `LoadMonmsConfig` | ✓ WIRED | Badge link gated in `base.gohtml` |
| `workspace/README.md` | `specs/staging.md` | cross-link | ✓ WIRED | Authoritative spec linked |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|-------------------|--------|
| Publish console | `HasChanges`, `ChangeGroups` | `DiffExport` → live PB records vs `publish-state.json` checksum | Yes — `TestDiffExportDetectsTitleChange` | ✓ FLOWING |
| Import API | `ImportReport.Upserted` | Request body → `ImportPayload` → `UpsertRecord` | Yes — `TestImportAPIUnauthorized/valid token` | ✓ FLOWING |
| Editor badge | `IsPublisher` | `LoadMonmsConfig` + auth email | Yes — publisher vs editor tests | ✓ FLOWING |
| Export | `CollectionFile.Records` | `app.FindAllRecords` + `PublicExport` | Yes — stable `homepage` id in tests | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Content package tests | `go test ./internal/content/... -count=1 -short` | ok (0.578s) | ✓ PASS |
| Schema editorial tests | `go test ./internal/schema/... -count=1` | ok | ✓ PASS |
| Editor badge tests | `go test ./internal/router/... -run TestEditorBadge -count=1` | ok | ✓ PASS |
| CLI dispatch | `/tmp/monms-verify content` (no subcommand) | usage message | ✓ PASS |
| Binary builds | `go build -o /tmp/monms-verify .` | exit 0 | ✓ PASS |

### Probe Execution

Step 7c: SKIPPED — no phase-declared probes or `scripts/*/tests/probe-*.sh` for this phase.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| ENV-01 | 04-06 | Four layers documented | ✓ SATISFIED | `workspace/README.md`, root `README.md` |
| ENV-02 | 04-06 | Dual promotion rails | ✓ SATISFIED | Structure Git tag vs content JSON upsert docs |
| ENV-03 | 04-06 | Separate staging/production `.pb_data/` | ✓ SATISFIED | Environment table in `workspace/README.md` |
| PUB-01 | 04-01 | Editorial flag in schema JSON | ✓ SATISFIED | `hero_content.json`, `editorial.go` |
| PUB-02 | 04-02 | `workspace/content/*.json` exports | ✓ SATISFIED | `export.go`, export tests |
| PUB-03 | 04-03 | `monms content export` CLI | ✓ SATISFIED | `cmd.go`, `TestContentExportCLI` |
| PUB-04 | 04-02 | Idempotent import by ID | ✓ SATISFIED | `import.go`, idempotent tests |
| PUB-05 | 04-04 | Production import API + token | ✓ SATISFIED | `routes.go`, `auth.go`, import API tests |
| PUB-06 | 04-05 | Publish UI with diff preview | ✓ SATISFIED | `publish_handlers.go`, `publish.gohtml` |
| PUB-07 | 04-05 | Publisher role gate | ✓ SATISFIED | `RequirePublisher`, `TestPublisherGate` |
| PUB-08 | 04-02 | Last-published checksum state | ✓ SATISFIED | `state.go`, `checksum.go` |
| PUB-09 | 04-02/03 | `monms content diff` | ✓ SATISFIED | `diff.go`, `TestContentDiffCLI` |
| MED-01 | 04-02 | CDN URLs; blobs not copied | ✓ SATISFIED | File field skip in export; `MEDIA.md` |
| MED-02 | 04-06 | Doc warns against local file storage | ✓ SATISFIED | `workspace/MEDIA.md` |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | None in phase-modified files | — | No TBD/FIXME/stub handlers detected |

Scanned: `internal/content/**`, `internal/schema/editorial.go`, phase doc files. No unreferenced debt markers.

### Human Verification Required

None blocking. Integration tests cover publish UI HTML, publisher gate, import token auth, and mock production forward. Optional operator smoke test (two live MonMS instances with matching `MONMS_PUBLISH_TOKEN`) recommended before first client handoff but not required for phase goal achievement.

### Gaps Summary

No gaps found. All 14 requirement-mapped truths verified in code with substantive implementations, wiring, and passing tests. ROADMAP metadata still shows "Plans: 4/6 plans executed" but all six plan summaries (04-01 through 04-06) exist with committed artifacts — documentation drift only, not an implementation gap.

---

_Verified: 2026-05-26T22:00:00Z_  
_Verifier: Claude (gsd-verifier)_
