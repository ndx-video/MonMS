# Extensibility with Sentinel

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

MonMS has **no plugin or extension system**. That is intentional.

## Why MonMS stays monolithic

MonMS takes a monolithic approach to the CMS problem. Adding WordPress-style plugins would:

- Undermine **performance** — every extension adds runtime surface area in the same process as your public site.
- Undermine **security** — you harden the core, then inherit risk from third-party code you did not write.
- Undermine **stability** — operators get blamed when a user-installed extension misbehaves.

We considered embedding Sentinel's Lua VM inside MonMS and rejected it. MonMS is not trying to replicate WordPress. Extensions are antithetical to the strategy of a single, auditable binary with Git-tracked structure.

## Where customization belongs

| Need | Use |
|------|-----|
| Shape templates, schema, assets | `site/` Git + MonMS shaping workflow |
| Editorial copy | PocketBase collections + inline edit + publish rail |
| **Extra server-side logic** | MonMS JSON API (`/api/monms/*`) or **Sentinel** |
| Custom operator dashboards | Sentinel kit pages/forms, optionally on a subdomain |
| Deep PocketBase / collection behavior | [PocketBase REST API](https://pocketbase.io/docs/) — not duplicated here |

Filing extra functionality under **API** or **Sentinel** keeps a clean separation of concerns: MonMS remains the frozen engine and site host; Sentinel remains the git-native orchestration harness for workflows, integrations, and kit UIs.

## Sentinel (sibling product)

**Sentinel** is a separate product — a git-native orchestration harness (Go binary + Lua kits). It is developed alongside MonMS and can augment a MonMS deployment without loading code into the `monms` process.

**Official URLs (not live yet):**

| Resource | URL | Status |
|----------|-----|--------|
| Sentinel product site | [https://ndx-sentinel.com](https://ndx-sentinel.com) | Planned — not built yet |
| Sentinel documentation | [https://docs.ndx-sentinel.com](https://docs.ndx-sentinel.com) | Planned — not built yet |

Until those sites ship, Sentinel source and docs live in the sibling repository next to MonMS (`../sentinel/` in a typical checkout).

## How Sentinel can work with MonMS

Sentinel kits can interact with MonMS in several ways (conceptual — integration docs will expand when the official sites launch):

1. **External processor** — Sentinel flows call MonMS over HTTP (`POST /api/monms/content/import`, PocketBase collection APIs, etc.).
2. **Co-located data** — advanced deployments may allow Sentinel to write directly to MonMS-related storage; treat this as operator-defined and out of band from the MonMS binary.
3. **Parallel dashboard** — a Sentinel kit renders its own forms and admin UI (e.g. on a subdomain) while MonMS serves the public site and PocketBase admin at `/_/`.

A **`monms` object in Sentinel Lua** is planned so kit authors can access MonMS from flows without ad-hoc HTTP boilerplate. That API will be documented on [docs.ndx-sentinel.com](https://docs.ndx-sentinel.com) when available.

## MonMS ↔ Sentinel in official docs

Both products will cross-reference each other in their official documentation once published:

- MonMS docs (this tree) point operators to Sentinel for extensions.
- Sentinel docs will describe MonMS as a supported integration target.

## Related

- [MonMS HTTP API](../reference/monms-api.md)
- [Shaping and agents](shaping-and-agents.md)
- [Getting started](getting-started.md)
