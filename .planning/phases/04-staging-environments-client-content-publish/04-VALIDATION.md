---
phase: 4
slug: staging-environments-client-content-publish
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-23
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for content export/import, publish API, and staging UI.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + `httptest` |
| **Config file** | none |
| **Quick run command** | `go test ./internal/content/... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/content/... -count=1 -short`
- **After every plan wave:** `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Req ID | Behavior | Test Type | Automated Command | Status |
|--------|----------|-----------|-------------------|--------|
| PUB-01 | Editorial flag from schema JSON | unit | `go test ./internal/schema/... -run Editorial -count=1` | Wave 0 |
| PUB-02 | Content JSON file shape | unit | `go test ./internal/content/... -run Export -count=1` | Wave 1 |
| PUB-03 | CLI export writes files | integration | `go test ./internal/content/... -run ExportCLI -count=1` | Wave 2 |
| PUB-04 | Import upserts by ID | integration | `go test ./internal/content/... -run Import -count=1` | Wave 1 |
| PUB-05 | Import API token gate | integration | `go test ./internal/content/... -run ImportAPI -count=1` | Wave 3 |
| PUB-06 | Publish page 200 | integration | `go test ./internal/content/... -run PublishUI -count=1` | Wave 4 |
| PUB-07 | Non-publisher 403 | integration | `go test ./internal/content/... -run PublisherGate -count=1` | Wave 4 |
| PUB-08 | Checksum after publish | unit | `go test ./internal/content/... -run Checksum -count=1` | Wave 1 |
| PUB-09 | Diff detects changes | unit | `go test ./internal/content/... -run Diff -count=1` | Wave 1 |
| ENV-01–03 | Docs four layers | manual | README review | Wave 5 |
| MED-01–02 | Skip file fields + MEDIA.md | unit + manual | `go test ./internal/content/... -run FileField -count=1` | Wave 1+5 |

---

## Wave 0 Gaps

- [ ] `internal/content/*_test.go`
- [ ] `internal/schema/editorial_test.go`
- [ ] `internal/testutil/content.go`
- [ ] `workspace/.monms/config.example.json`
