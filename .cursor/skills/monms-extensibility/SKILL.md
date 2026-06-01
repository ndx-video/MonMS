---
name: monms-extensibility
description: >-
  MonMS extensibility guardrails — no in-process plugins, customization boundaries,
  and Sentinel HTTP integration patterns. Use when asked to add plugins, webhooks,
  WordPress-style extensions, or Sentinel integration.
---

# MonMS Extensibility

MonMS has **no plugin or extension system**. That is intentional.

## Why no in-process plugins

WordPress-style extensions would:

- Undermine **performance** — extra runtime surface in the same process as the public site
- Undermine **security** — third-party code inherits your attack surface
- Undermine **stability** — operators get blamed for user-installed extensions

Embedding Sentinel's Lua VM inside MonMS was considered and **rejected**. MonMS is a single, auditable binary with Git-tracked structure — not a plugin host.

## Where customization belongs

| Need | Use |
|------|-----|
| Templates, schema, assets | Configured site directory (Git) + [monms-site-shaping](../monms-site-shaping/SKILL.md) |
| Editorial copy | PocketBase collections + inline edit + publish rail |
| Extra server-side logic | MonMS JSON API (`/api/monms/*`) or **Sentinel** (external) |
| Custom operator dashboards | Sentinel kit pages/forms on a subdomain |
| Deep PocketBase behavior | [PocketBase REST API](https://pocketbase.io/docs/) — not duplicated in MonMS docs |

## MonMS JSON API surface

Machine clients call `/api/monms/*` with Bearer tokens:

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `POST /api/monms/content/import` | `MONMS_PUBLISH_TOKEN` | Production editorial upsert |

Full reference: [docs/reference/monms-api.md](../../../docs/reference/monms-api.md).

## Sentinel integration (external)

Sentinel is a **separate product** — git-native orchestration (Go + Lua kits). It does not load into the `monms` process.

Integration patterns:

1. **External processor** — Sentinel flows call MonMS over HTTP (`POST /api/monms/content/import`, PocketBase collection APIs)
2. **Co-located data** — operator-defined; out of band from MonMS binary
3. **Parallel dashboard** — Sentinel kit UI on subdomain; MonMS serves public site + `/_/` admin

A `monms` object in Sentinel Lua is planned — until documented, use HTTP calls.

When building Sentinel kits that call MonMS, use the **`sn-kit-build`** skill (sibling repo).

## Anti-patterns — do not implement

- Go hooks/plugins loaded at runtime into `monms`
- Lua VM embedded in MonMS
- WordPress-style `mu-plugin` or filter system
- Duplicating PocketBase collection API in MonMS docs/code
- Routing custom logic through SSR slug resolution instead of `/api/monms/*`

## Related docs

- [docs/operators/extensibility-with-sentinel.md](../../../docs/operators/extensibility-with-sentinel.md)
- [monms-architecture](../monms-architecture/SKILL.md) — HTTP namespace split
- [monms-engine-development](../monms-engine-development/SKILL.md) — where to add MonMS-native routes
