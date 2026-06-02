# MCP and API keys

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

MonMS can expose a **Model Context Protocol (MCP)** HTTP listener for agentic clients (Cursor, Claude Code, custom automations). Access is gated by **per-user API keys** stored in PocketBase, not by sharing superuser passwords.

## Overview

| Piece | Location |
|-------|----------|
| API key UI | `/_monms/api-keys` (MonMS Console) |
| MCP settings | `/_monms/mcp` (superusers only) |
| Key storage | `monms_api_keys` collection (engine bootstrap) |
| Owner accounts | `_superusers` or `users` (auth collection, engine bootstrap) |
| Listener config | `site/.monms/config.json` → `mcp` block |

The MCP server binds **separately** from PocketBase `--http` (default `127.0.0.1:8091`). Changing host or port requires **`monms restart`**.

## Configuration

Copy from [site/.monms/config.example.json](../../site/.monms/config.example.json) and set:

```json
"mcp": {
  "enabled": false,
  "host": "127.0.0.1",
  "port": "8091",
  "allowNonSuperuserKeys": false
}
```

| Field | Meaning |
|-------|---------|
| `enabled` | Start MCP HTTP listener on serve |
| `host` | Bind address (default loopback) |
| `port` | TCP port (separate from main site port) |
| `allowNonSuperuserKeys` | Let `users` collection accounts create keys in `/_monms/api-keys` |

Superusers configure this at **MCP** in the console (`/_monms/mcp`). Keys are managed at **API Keys** (`/_monms/api-keys`).

## API keys

### Creating keys

1. Sign in at `/_/` (superuser) or as a `users` account when allowed.
2. Open **MonMS Console** → **API Keys**.
3. Enter a label and **Generate key**.
4. Copy the secret immediately — it is shown **once**. Only a prefix is stored in PocketBase.

Key format: `monms_` + random hex (see `internal/apikeys`).

### Revoking keys

Use **Revoke** on the API Keys page. Keys are deleted from `monms_api_keys` (hard delete in v1).

### Who can manage keys

| Account | Can create keys when |
|---------|----------------------|
| `_superusers` | Always |
| `users` | `mcp.allowNonSuperuserKeys` is `true` |

Keys are **not** exposed via public PocketBase collection APIs; dashboard handlers perform ownership checks.

### Hashing pepper

Optional env `MONMS_API_KEY_PEPPER` overrides the site-derived default used to hash secrets at rest. Set the same value on every instance that must validate the same keys. Never commit the pepper.

## MCP client connection

When `mcp.enabled` is `true`:

| Item | Value |
|------|--------|
| URL | `http://<mcp.host>:<mcp.port>/mcp` |
| Transport | Streamable HTTP + SSE (`mcp-go`) |
| Auth header | `Authorization: Bearer <full_api_key>` |

Permissions follow the **key owner**: PocketBase collection rules run as that user via a short-lived auth token minted per request. MonMS-specific tools add extra gates (for example `monms_content_diff` requires publisher allowlist or superuser).

### Tools

| Tool | Purpose |
|------|---------|
| `monms_list_collections` | Editorial + markdown collection names from schema |
| `monms_schema_list` | Basenames of `site/schema/*.json` |
| `monms_list_records` | List records in a PB-native editorial collection |
| `monms_get_record` | Get one record by id |
| `monms_update_record` | Patch fields on one record |
| `monms_content_diff` | Pending publish diff (publisher or superuser owner) |
| `monms_validate` | Validate all site `.gohtml` templates |

Content **import** and production publish still use `MONMS_PUBLISH_TOKEN` on `POST /api/monms/content/import` — not MCP keys in v1.

## Agents vs superuser JWT

| Method | Best for |
|--------|----------|
| `POCKETBASE_ADMIN_TOKEN` (superuser JWT) | Shaping: schema dual-write, collection CRUD via PocketBase REST |
| **MonMS API key** | MCP tools: editorial reads/writes, validate, schema listing — scoped to key owner |
| `MONMS_PUBLISH_TOKEN` | Production content import only |

See [Shaping and agents](shaping-and-agents.md) for structure-rail workflows.

## Security notes

- Default MCP bind is **loopback**. Exposing MCP on a tailnet or LAN is an explicit operator choice (reverse proxy, firewall).
- Do not log full API keys. Dashboard shows the secret once.
- Rotate keys by revoking and creating a new one; there is no per-key scope UI in v1.
- Compromise: revoke affected keys, rotate `MONMS_API_KEY_PEPPER` if used, audit `monms_api_keys` in PocketBase admin.

## Related

- [MonMS HTTP API](../reference/monms-api.md) — route reference
- [Security](security.md) — tokens and git hygiene
- [Getting started](getting-started.md) — layers and namespaces
- [Extensibility with Sentinel](extensibility-with-sentinel.md) — HTTP integration patterns
