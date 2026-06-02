# Shaping and agents

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)  
> **PocketBase collections API:** [official docs](https://pocketbase.io/docs/api-records/)

MonMS enables external AI agents to safely mutate a Git-tracked site during **shaping** — templates, schema, and assets (L2 only).

**Structure rail only:** Agents do not routine-push editorial copy to production. Clients publish content via `/_monms/publish`. See [Getting started](getting-started.md).

## Safety guardrails

1. **Go template dry-run** — `html/template.ParseFiles` with layout + page (same as production SSR).
2. **HTML structure check** — `golang.org/x/net/html` tokenizer for balanced tags.
3. **Pre-commit hook** — on validation failure, `git checkout -- .` and abort commit.

## Prerequisites

### Binary access

Resolution order for `monms validate`:

1. `MONMS_BIN` environment variable (recommended)
2. `monms` on `$PATH`
3. `../../monms` relative to the site directory

### PocketBase admin token

```bash
curl -s -X POST "$MONMS_URL/api/collections/_superusers/auth-with-password" \
  -H "Content-Type: application/json" \
  -d '{"identity":"admin@example.com","password":"your-admin-password"}' \
  | jq -r '.token'
```

```bash
export POCKETBASE_ADMIN_TOKEN="<token-from-above>"
```

Never hardcode tokens. See [Security](security.md).

| Variable | Purpose |
|----------|---------|
| `MONMS_URL` | Running server base URL |
| `POCKETBASE_ADMIN_TOKEN` | Admin JWT for collection management |
| `MONMS_BIN` | Path to `monms` binary (optional) |
| `MONMS_PUBLISH_TOKEN` | Production import secret (consultant setup only) |
| MonMS API key | MCP agentic access (`Authorization: Bearer monms_…`) — create at `/_monms/api-keys` |

For MCP-driven shaping helpers (validate, schema list, editorial CRUD as the key owner), prefer an **API key** over long-lived superuser JWTs. Collection management during dual-write still uses `POCKETBASE_ADMIN_TOKEN`. See [MCP and API keys](mcp-and-api-keys.md).

Use a dedicated SSH key scoped to the site directory. See [Security](security.md).

## Schema dual-write

When creating a collection:

### Step 1 — POST via PocketBase API

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

### Step 2 — Write `site/schema/{name}.json`

For client-publishable collections, add `"editorial": true` (read from JSON by MonMS export — PocketBase strips unknown keys on import).

```bash
git add schema/press_releases.json
git commit -m "agent: add press_releases schema"
```

Use the `agent:` prefix in commit messages (D-45).

## Template conventions

Page templates use `{{define "body"}}` only — see [Templates and routing](templates-and-routing.md).

Fragments live in `templates/fragments/` at `/fragments/{name}` without the base layout.

## Validation

```bash
monms validate --site .
```

Pre-commit hook validates staged `*.gohtml` only. Schema-only commits skip template validation.

**Rollback caveat:** `git checkout -- .` does not remove newly added untracked files after failed validation — clean up manually.

## press_releases walkthrough

1. Obtain admin token (above).
2. POST collection + write `schema/press_releases.json`.
3. Create `templates/press/index.gohtml` with `{{define "body"}}`.
4. `monms validate --site .`
5. `git commit -m "agent: add press_releases collection and press index template"`
6. Open `/press` — renders without restart.

If `/press` 404s, verify mirror+index path: `templates/press/index.gohtml`.

## Related

- [Templates and routing](templates-and-routing.md)
- [Security](security.md)
- [Extensibility with Sentinel](extensibility-with-sentinel.md) — no in-process plugins
