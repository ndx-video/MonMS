# MonMS Agent Mutation Guide

## Overview

MonMS enables external AI agents to safely mutate a Git-tracked workspace. Agents can:

- Create PocketBase collections via the PocketBase REST API — changes take effect immediately, no binary restart required (AGT-01).
- Edit workspace templates (`*.gohtml`) — changes are visible on the next request via in-memory cache invalidation driven by fsnotify (AGT-02).
- Schema JSON files serve as a Git audit trail and are re-imported on next server start for bootstrap self-healing (D-32, D-33).

**Structure rail only:** Agents and consultants mutate **L2 structure** (templates, schema, assets). Editorial **content publish** (L3) is client-driven via `/api/monms/publish` — agents do not routine-push copy to production. See [README.md](README.md) § content publish and [MEDIA.md](MEDIA.md).

**Phase 2 safety guardrails** protect the workspace from invalid mutations:

1. **Go template dry-run** — `html/template.ParseFiles` with the layout, matching the production parse path (D-37, AGT-03).
2. **HTML structure check** — `golang.org/x/net/html` tokenizer verifies balanced tags (D-38, AGT-04).
3. **Pre-commit hook with automatic rollback** — on any validation failure, `git checkout -- .` restores the workspace and the commit is aborted (D-43, AGT-05).


## Prerequisites

Before mutating the workspace, ensure the following are in place.

### Binary Access

The `monms` binary is resolved in this priority order (D-42):

1. **Recommended:** Set `MONMS_BIN` to the absolute path of the binary:
   ```bash
   export MONMS_BIN=/usr/local/bin/monms
   ```
2. Add `monms` to `$PATH`:
   ```bash
   export PATH="$PATH:/path/to/monms/directory"
   ```
3. Place the binary at `../../monms` relative to the workspace directory.

The pre-commit hook uses the same resolution order. If the binary is not found, the hook fails loudly.

### PocketBase Admin Token

Obtain a short-lived admin token via:

```bash
curl -s -X POST "$MONMS_URL/api/collections/_superusers/auth-with-password" \
  -H "Content-Type: application/json" \
  -d '{"identity":"admin@example.com","password":"your-admin-password"}' \
  | jq -r '.token'
```

Store the token as an environment variable — **never hardcode it**:

```bash
export POCKETBASE_ADMIN_TOKEN="<token-from-above>"
```

See `SECURITY.md` for token rotation and storage policy.

### Environment Variables Summary

| Variable | Purpose | Example |
|---|---|---|
| `MONMS_URL` | Running server base URL | `http://localhost:8090` |
| `POCKETBASE_ADMIN_TOKEN` | PocketBase admin JWT for collection management | `eyJhbGci...` |
| `MONMS_BIN` | Path to the `monms` binary (optional) | `/usr/local/bin/monms` |
| `MONMS_PUBLISH_TOKEN` | Shared secret for production content import (consultant setup only; never commit) | — |

Content CLI (operator fallback): `monms content export|import|diff|publish --workspace .` — see [README.md](README.md).

### SSH Key

Use a dedicated SSH key restricted to the workspace subdirectory only. See `SECURITY.md` §SSH Key Scope.


## Schema Dual-Write Workflow (D-33, AGT-01)

When creating a new collection, always perform a **dual-write**: create the collection live via the API, then persist the schema JSON to the workspace for audit and bootstrap sync.

### Step 1 — POST collection via API

```bash
curl -s -X POST "$MONMS_URL/api/collections" \
  -H "Authorization: Bearer $POCKETBASE_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "press_releases",
    "type": "base",
    "fields": [
      {"name": "title", "type": "text"},
      {"name": "body",  "type": "text"}
    ]
  }'
```

On success, PocketBase returns the created collection object with HTTP 200.

### Step 2 — Write matching JSON to workspace/schema/

```bash
cat > workspace/schema/press_releases.json << 'EOF'
{
  "name": "press_releases",
  "type": "base",
  "fields": [
    {"name": "title", "type": "text"},
    {"name": "body",  "type": "text"}
  ]
}
EOF
```

For collections clients will **publish to production**, add `"editorial": true` after `"type": "base"` (PUB-01, D-54). MonMS reads this flag from raw schema JSON — PocketBase strips unknown keys on import.

```json
{
  "name": "hero_content",
  "type": "base",
  "editorial": true,
  ...
}
```

Do not mark system or auth collections as editorial.

The file must be valid JSON (no comments, no trailing commas). The format matches `internal/schema/sync.go` — a single collection object with `name`, `type`, and `fields` keys. On next server start, this file is automatically imported for self-healing if the collection was deleted from PocketBase.

### Step 3 — Commit with agent: prefix (AGT-06)

```bash
git add schema/press_releases.json
git commit -m "agent: add press_releases schema"
```

Always use the `agent:` prefix in commit messages so mutations are traceable in git history (D-45).

The pre-commit hook runs `monms validate` automatically. Schema-only commits (no staged `*.gohtml` files) skip template validation.


## Template Editing Conventions (D-47)

### Required Structure

All page templates must use the `{{define "body"}}` wrapper:

```html
{{define "body"}}
<section class="press-releases">
  <h1>Press Releases</h1>
</section>
{{end}}
```

The `base.gohtml` layout is extended automatically by the SSR handler — do **not** re-declare `<!DOCTYPE html>`, `<html>`, `<head>`, or `<body>` in page templates.

### Slug → Template Path Mapping (D-10, D-11)

URL slugs map to template files by the **mirror + index** rule:

| URL | Template path |
|---|---|
| `/` | `templates/index.gohtml` |
| `/press` | `templates/press/index.gohtml` |
| `/press/2024` | `templates/press/2024.gohtml` |
| `/about` | `templates/about/index.gohtml` or `templates/about.gohtml` |

If a slug returns 404 after mutation, verify the template path mirrors the URL slug exactly.

### HTMX Collection Fetch Pattern

To render a live PocketBase collection list in a template:

```html
{{define "body"}}
<section class="press-releases">
  <h1>Press Releases</h1>
  <ul hx-get="/api/collections/press_releases/records"
      hx-trigger="load"
      hx-target="this"
      hx-swap="innerHTML">
    <li>Loading…</li>
  </ul>
</section>
{{end}}
```

### Cache Invalidation

Template changes are visible on the next HTTP request without restarting the server. In production builds, fsnotify watches the **entire workspace tree** for `.gohtml` changes and flushes the in-memory cache (D-30, AGT-02).

### Fragment Partials

Fragment templates live in `templates/fragments/` and are served at `/fragments/<name>` without the base layout — suitable for HTMX partial swaps. Fragment files must not use `{{define "body"}}`.


## Pre-Commit Validation Lifecycle

### Automatic Validation on Commit

The pre-commit hook installed at `workspace/.git/hooks/pre-commit` runs automatically on every `git commit`. It:

1. Collects staged `*.gohtml` files.
2. Skips validation entirely if no `.gohtml` files are staged (CSS/schema-only commits pass through).
3. Runs `monms validate --workspace .` against staged files only.
4. On failure: runs `git checkout -- .` to restore the working tree, then exits non-zero with validation errors printed to stderr.

### Manual Validation

Run validation manually before committing to catch errors early:

```bash
MONMS_BIN=/path/to/monms monms validate --workspace .
```

Or with explicit file paths:

```bash
monms validate --workspace . --files templates/press/index.gohtml
```

### Validation Checks

| Check | Tool | What it catches |
|---|---|---|
| Go template dry-run | `html/template.ParseFiles` | Syntax errors, undefined actions, missing blocks |
| HTML structure | `golang.org/x/net/html` tokenizer | Unbalanced tags, invalid nesting |

### Rollback Caveat for New Files

When a mutation adds a new file (`git add new-file.gohtml`) and the pre-commit hook rolls back with `git checkout -- .`, the new untracked/staged file is **not** removed by `git checkout`. After a rollback involving new files, also clean up:

```bash
git reset HEAD <new-file.gohtml>
rm <new-file.gohtml>
```

Then fix the validation error and re-commit.


## press_releases Sample Walkthrough (D-35)

End-to-end example of the full mutation workflow.

### 1. Obtain Admin Token

```bash
export MONMS_URL="http://localhost:8090"
export POCKETBASE_ADMIN_TOKEN=$(curl -s -X POST "$MONMS_URL/api/collections/_superusers/auth-with-password" \
  -H "Content-Type: application/json" \
  -d '{"identity":"admin@example.com","password":"your-admin-password"}' \
  | jq -r '.token')
```

### 2. POST Collection

```bash
curl -s -X POST "$MONMS_URL/api/collections" \
  -H "Authorization: Bearer $POCKETBASE_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"press_releases","type":"base","fields":[{"name":"title","type":"text"},{"name":"body","type":"text"}]}'
```

### 3. Write Schema JSON

```bash
cat > schema/press_releases.json << 'EOF'
{
  "name": "press_releases",
  "type": "base",
  "fields": [
    {"name": "title", "type": "text"},
    {"name": "body",  "type": "text"}
  ]
}
EOF
```

### 4. Create Template

```bash
mkdir -p templates/press
cat > templates/press/index.gohtml << 'EOF'
{{define "body"}}
<section class="press-releases">
  <h1>Press Releases</h1>
  <ul hx-get="/api/collections/press_releases/records"
      hx-trigger="load"
      hx-target="this"
      hx-swap="innerHTML">
    <li>Loading…</li>
  </ul>
</section>
{{end}}
EOF
```

### 5. Validate Manually

```bash
monms validate --workspace .
```

Expected output: no errors (exit 0).

### 6. Commit

```bash
git add schema/press_releases.json templates/press/index.gohtml
git commit -m "agent: add press_releases collection and press index template"
```

The pre-commit hook runs automatically. On success, the commit proceeds.

### 7. Verify in Browser

Open `http://localhost:8090/press` — the page renders immediately without restart.

**Debugging:** If `/press` returns 404, verify the template path is `templates/press/index.gohtml` (mirror+index rule, D-10). The error page displays the attempted template path for diagnosis (D-19).

## Related guides

| Guide | Purpose |
|-------|---------|
| [README.md](README.md) | Four layers, dual rails, content publish setup |
| [MEDIA.md](MEDIA.md) | CDN URL fields for publishable assets |
| [EDITING-GUIDE.md](EDITING-GUIDE.md) | Human inline editing + Publish to live (clients) |
| [SECURITY.md](SECURITY.md) | SSH scope, tokens, publish secrets |
