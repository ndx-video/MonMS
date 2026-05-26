---
phase: 04-staging-environments-client-content-publish
reviewed: 2026-05-26T23:45:00Z
depth: standard
files_reviewed: 28
files_reviewed_list:
  - internal/schema/editorial.go
  - internal/testutil/content.go
  - workspace/.monms/config.example.json
  - workspace/schema/hero_content.json
  - internal/scaffold/embed/hero_content.json
  - internal/content/export.go
  - internal/content/import.go
  - internal/content/checksum.go
  - internal/content/diff.go
  - internal/content/state.go
  - internal/content/schema.go
  - internal/content/path.go
  - internal/content/cmd.go
  - internal/content/publish.go
  - internal/content/auth.go
  - internal/content/routes.go
  - internal/content/publish_handlers.go
  - internal/content/embed/publish.gohtml
  - main.go
  - internal/router/ssr.go
  - workspace/templates/layouts/base.gohtml
  - internal/scaffold/embed/base.gohtml
  - workspace/MEDIA.md
  - workspace/.gitignore
  - README.md
  - workspace/README.md
  - CLAUDE.md
  - workspace/EDITING-GUIDE.md
findings:
  critical: 0
  warning: 6
  info: 2
  total: 8
status: issues_found
---

# Phase 04: Code Review Report

**Reviewed:** 2026-05-26T23:45:00Z
**Depth:** standard
**Files Reviewed:** 28
**Status:** issues_found

## Summary

Phase 04 implements the editorial content rail end-to-end: schema-driven allowlists, export/import/checksum/diff, CLI and HTTP surfaces, publisher-gated staging UI, and production import with Bearer token auth. Auth boundaries are generally sound — publish token middleware fails closed, import rejects system/non-editorial collections, and publish UI requires superuser session plus publisher allowlist. SameSite=Lax on the auth cookie mitigates cross-site POST CSRF on `/api/monms/publish`.

The main gaps are behavioral: diff omits deleted records, publish is upsert-only (deletions never reach production), the CLI publish path skips publish-state updates, and failed UI publishes return HTTP 200. None are auth bypasses, but operators should understand deletion and diff limitations before relying on the console for full staging parity.

## Warnings

### WR-01: Diff does not detect deleted editorial records

**File:** `internal/content/diff.go:49-79`
**Issue:** `diffSnapshots` iterates only over records in the current export. Records present in the baseline (`workspace/content/*.json`) but removed from the live DB are never reported. After a deletion, `HasChanges` may still be true (checksum differs) but field-level changes will be incomplete unless the generic fallback message appears.
**Fix:** After the current-record loop, scan baseline records and emit `{collection}/{id}: record deleted` for IDs missing from the current snapshot:

```go
for coll, baseRecords := range baseByColl {
    curRecords := curByColl[coll] // build curByColl similarly
    for id := range baseRecords {
        if curRecords[id] == nil {
            changes = append(changes, fmt.Sprintf("%s/%s: record deleted", coll, id))
        }
    }
}
```

### WR-02: Publish is upsert-only — deletions never propagate to production

**File:** `internal/content/publish_handlers.go:91-96`, `internal/content/import.go:57-81`
**Issue:** Publish exports the current editorial snapshot and upserts records by ID. Records deleted on staging remain on production indefinitely. Clients who delete content and click **Publish now** may believe production was updated when stale records persist.
**Fix:** Document explicitly in operator guides (partially missing), and/or add an optional tombstone/delete pass in the import protocol if deletion sync is required. Minimum fix: surface a UI warning that publish does not remove production records.

### WR-03: CLI `content publish` does not update publish-state.json

**File:** `internal/content/cmd.go:111-148`
**Issue:** `runPublishCLI` POSTs to production via `PublishToProduction` but never calls `WritePublishState`. After a successful CLI publish from staging, `DiffExport` and the publish UI still report unpublished changes (PUB-08 gap for the operator fallback path).
**Fix:** Mirror the UI handler — after a successful POST, compute checksum and write publish state:

```go
if err := PublishToProduction(toURL, token, payloads); err != nil {
    return err
}
checksum, err := ChecksumExport(payloads)
if err != nil {
    return err
}
return WritePublishState(wsAbs, PublishState{
    Checksum:    checksum,
    PublishedAt: time.Now().UTC().Format(time.RFC3339),
    Collections: collectionNames(payloads),
})
```

### WR-04: Failed publish POST returns HTTP 200

**File:** `internal/content/publish_handlers.go:97-105`
**Issue:** When `PublishToProduction` fails, the handler renders an HTML error page but does not set a non-2xx status code. Clients, monitors, or HTMX consumers may treat the response as success despite `MessageError` content.
**Fix:** Set an appropriate error status before rendering:

```go
e.Response.WriteHeader(http.StatusBadGateway)
e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
return tmpl.Execute(e.Response, data)
```

### WR-05: Import is not transactional — partial upserts on mid-batch failure

**File:** `internal/content/import.go:57-81`
**Issue:** `ImportPayload` upserts collections and records sequentially and returns on first error. A failure after some collections succeed leaves production in a partially updated state with no rollback or compensating report beyond whatever was already saved.
**Fix:** Wrap import in a PocketBase transaction if available, or return a partial-failure report listing which collections succeeded. At minimum, document that operators should retry idempotently and verify via diff.

### WR-06: Publisher email matching is case-sensitive

**File:** `internal/content/auth.go:62-69`
**Issue:** `IsPublisher` compares emails with plain `==`. PocketBase normalizes emails to lowercase on registration; allowlist entries with different casing (e.g. `Publisher@Client.com` vs `publisher@client.com`) silently deny legitimate publishers.
**Fix:** Normalize both sides before compare:

```go
func IsPublisher(email string, allowed []string) bool {
    email = strings.ToLower(strings.TrimSpace(email))
    for _, a := range allowed {
        if email == strings.ToLower(strings.TrimSpace(a)) {
            return true
        }
    }
    return false
}
```

## Info

### IN-01: Dead code in checksum canonicalization

**File:** `internal/content/checksum.go:53-59`
**Issue:** The loop sorts record keys but never rebuilds maps; stability relies on `encoding/json` map key sorting. The loop is misleading dead code.
**Fix:** Remove the unused sort loop, or rebuild each record as an ordered structure if explicit canonicalization is desired.

### IN-02: Empty import request returns 200 with zero upserts

**File:** `internal/content/routes.go:70-72`, `internal/content/routes.go:46-67`
**Issue:** `POST /api/monms/content/import` with `"collections": []` succeeds with `upserted: 0`. Harmless but allows token holders to probe the endpoint without side effects.
**Fix:** Optional — reject empty payloads with `400 Bad Request` if zero-record imports are never valid.

---

_Reviewed: 2026-05-26T23:45:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
