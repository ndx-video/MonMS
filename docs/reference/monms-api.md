# MonMS HTTP API

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

This reference documents **MonMS-only** routes. PocketBase collection REST (`/api/collections/...`), auth, and admin APIs are covered by the [official PocketBase documentation](https://pocketbase.io/docs/).

Canonical path constants: `internal/monmsroutes/routes.go`.

## Namespaces

| Prefix | Content type | Audience |
|--------|--------------|----------|
| `/api/monms/*` | JSON only | Machine clients, CI, Sentinel flows |
| `/_monms/*` | HTML + session helpers | Browser operators (dashboard, publish console, auth bridge) |

Static assets for the operator dashboard are served at `/_monms/static/*` (embedded, offline-capable — no CDN).

## `GET /_monms/` and `GET /_monms`

Operator **dashboard home**. Requires PocketBase superuser session (`monms_auth` cookie or Bearer via auth middleware).

Role-aware navigation:

- **All signed-in superusers:** view site, PocketBase admin link
- **`publisherEmails` allowlist:** publish console
- **Consultants (superusers):** structure/admin tools, MCP settings
- **`CanManageAPIKeys`:** API Keys nav (superusers, or `users` when `mcp.allowNonSuperuserKeys`)

Unauthenticated requests redirect to PocketBase admin at `/_/` (`monmsroutes.AdminPath`).

## `POST /api/monms/content/import`

Production editorial upsert (content rail).

**Auth:** `Authorization: Bearer <MONMS_PUBLISH_TOKEN>`

**Body:**

```json
{
  "collections": [
    {
      "name": "hero_content",
      "records": [
        { "id": "homepage", "title": "...", "body": "..." }
      ]
    }
  ]
}
```

**Response:** `200` with `{ "upserted": N, "collections": M, "warnings": [...] }`

**Rules:**

- Denies system collections (`_superusers`, `users`).
- Max 1000 records per collection per request.
- Upsert by record ID — deletions are not propagated from staging exports.
- Skips PocketBase file-type fields (warns in export path).

Used by staging **Publish to live** and `monms content publish`.

## `GET /_monms/publish`

HTML **Publish to live** console (rendered inside the dashboard shell).

**Auth:** PocketBase superuser session + email in `site/.monms/config.json` `publisherEmails`.

Shows diff preview, last publish time, production URL configuration status.

## `GET /_monms/publish/diff`

JSON diff of pending editorial changes vs last publish.

**Auth:** Same as publish page.

## `POST /_monms/publish`

Executes publish: export staging editorial collections → POST to `productionUrl` + `/api/monms/content/import`.

**Auth:** Same as publish page.

## `POST /_monms/auth/sync`

Bridges PocketBase admin `localStorage` token to HttpOnly `monms_auth` cookie for inline editing on public pages.

**Auth:** `Authorization: Bearer <superuser token>`

**Response:** `204 No Content` with `Set-Cookie`.

Called automatically from the base layout script when a user visits `/` after logging in at `/_/`.

## `GET|POST /_monms/auth/logout`

Clears `monms_auth` cookie. Optional `?redirect=/`.

## `GET|POST /_monms/api-keys`

Operator **API key management** (dashboard HTML).

**Auth:** PocketBase session (`monms_auth` cookie). Superusers always; `users` collection accounts only when `mcp.allowNonSuperuserKeys` is `true` in `site/.monms/config.json`.

**POST** creates a key (form field `name`). The full secret is shown once in the page flash area.

## `POST /_monms/api-keys/revoke`

Revokes a key owned by the signed-in account. Form field `id` is the `monms_api_keys` record id.

## `GET|POST /_monms/mcp`

Superuser-only **MCP settings** form: enable listener, host, port, and `allowNonSuperuserKeys`. Persists to `site/.monms/config.json`. Restart `monms` after bind changes.

## `GET /_monms/system`

Superuser-only **System** console: runtime status (build mode, listen address, URLs, MCP, editorial flags, logging) and lifecycle actions.

**Auth:** PocketBase superuser session (`monms_auth` cookie). Non-superusers receive `403`; unauthenticated requests redirect to `/_/`.

## `POST /_monms/system/restart`

Superuser-only. Stops other `monms` processes sharing this executable, then restarts (re-exec current argv, or fork detached when the running instance is a daemon child). Matches `monms restart` scope — **binary-wide**, not site-scoped.

Returns HTML with navbar flash `Restarting MonMS…`. The connection may drop when the process restarts.

## `POST /_monms/system/shutdown`

Superuser-only. Stops other instances, then sends SIGTERM to the current process. Matches `monms stop` plus self-termination from the dashboard. **Binary-wide.**

Returns HTML with navbar flash `Shutting down MonMS…`.

Lifecycle buttons are disabled on platforms where process enumeration is unsupported (same limitation as `monms stop` on non-Linux).

## MCP HTTP server

When `mcp.enabled` is `true`, MonMS listens separately from PocketBase `--http` (default `127.0.0.1:8091`).

| Item | Value |
|------|--------|
| Endpoint | `http://<mcp.host>:<mcp.port>/mcp` (Streamable HTTP + SSE) |
| Auth | `Authorization: Bearer <monms_api_key>` |
| Key format | `monms_` + hex secret; only a prefix is stored in PocketBase |
| Permissions | PocketBase rules and MonMS gates evaluated as the **key owner** (`_superusers` or `users` record) |

**Tools:** Editorial — `monms_list_collections`, `monms_schema_list`, `monms_list_records`, `monms_get_record`, `monms_update_record`, `monms_content_diff` (publisher/superuser), `monms_validate`. **Doctree (markdown rail)** — `monms_doctree_bindings`, `monms_doctree_forest`, `monms_doctree_list`, `monms_doctree_get`, `monms_doctree_write`, `monms_doctree_delete`, `monms_doctree_sync`, `monms_doctree_diff`, `monms_doctree_sections`.

**Env:** `MONMS_API_KEY_PEPPER` optional; overrides site-derived hashing pepper.

## Related

- [MCP and API keys](../operators/mcp-and-api-keys.md) — operator setup and security
- [CLI reference](cli.md) — `monms content` subcommands
- [Publish to live](../user-guide/publish-to-live.md)
- [External dependencies](external-dependencies.md)
