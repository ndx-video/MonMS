# Getting started

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

MonMS separates **engine**, **structure**, **content**, and **audience** into four layers. Structure and content promote on **independent rails**.

## Phases of work

| Phase | Who | Focus |
|-------|-----|-------|
| **Development** | MonMS engine developers | Building this Go repository |
| **Shaping** | Consultants, agents | Templates, schema, assets in `site/` Git |
| **Staging** | Client editors/publishers | Editorial copy on the staging instance |
| **Production** | Audience; clients publish into | Live site |

*Development* (product phase) is not the same as compile-time `buildMode=development` when running a locally built binary.

## Four layers

| Layer | Artifact | In Git? | Promoted how |
|-------|----------|---------|--------------|
| **L1** Engine | `monms` binary | No (semver release) | Same pinned binary on staging and production |
| **L2** Structure | `site/` templates, schema, assets, `documents/` (markdown) | Yes | **Git tag** → operator pulls into both instances |
| **L3** Content | `.pb_data/` records | No | **Publish to live** (PB-native) or **Git** (markdown) |
| **L4** Audience | Production URL | No | Read-only |

## Two promotion rails

| Rail | What moves | Mechanism | Who |
|------|------------|-----------|-----|
| **Structure** | Templates, schema JSON, CSS/assets, markdown `documents/` | Git tag + `monms site sync` (or operator CI) | Consultants, agents |
| **Content (PB-native)** | Inline-edited PocketBase records | `POST /api/monms/content/import` / `/_monms/publish` | Clients (publishers) |
| **Content (markdown)** | Git-tracked `.md` under `documents/` | Same structure rail — sync on bootstrap | Consultants, developers |
| **Media** | CDN URL strings in text fields | Upserted with content — **no blob copy** | Clients |

Git tags carry **structure** (including markdown documents). PB-native editorial copy stays out of Git. See [Markdown document rail](markdown-content.md).

## Staging vs production

Typical deployment runs **two MonMS instances**:

| | Staging | Production |
|---|---------|------------|
| **Purpose** | Clients edit and preview copy | Audience reads live site |
| **`.pb_data/`** | Staging SQLite (never wholesale-synced) | Production SQLite |
| **Structure** | Same tagged shape as production | Same tagged shape (e.g. `v1.2.0`) |
| **Content** | Edited inline here | Updated via **Publish to live** only |

Consultants **shape** structure and tag releases. Clients **stage** content and publish — consultants are not in the routine content loop.

## One-time consultant setup

1. Edit `site/.monms/config.json`: `siteUrl` (this instance), `productionUrl` (publish target), `publisherEmails`, optional `allowedHosts` (CORS when `--origins` omitted), optional `logging` (see [File logging](../reference/logging.md)).
2. Set `MONMS_PUBLISH_TOKEN` on **both** staging (outbound publish) and production (import API gate) — same secret, never commit.
3. Optional: enable [MCP and API keys](mcp-and-api-keys.md) for agentic clients — configure `mcp` in `config.json`, create keys at `/_monms/api-keys`, restart after bind changes.

## Site directory layout

```
site/
├── schema/           # L2 — collection definitions (JSON)
├── documents/        # L2 — Git markdown (monms.source=markdown collections)
├── content/          # L3 — PB-native editorial exports (optional, often gitignored)
├── templates/        # L2 — Go HTML templates (includes doc.gohtml for markdown pages)
├── assets/           # L2 — CSS, static files
├── .monms/           # Publish config + file logs (example committed; config.json gitignored)
│   └── logs/         # Rotated log files (gitignored)
└── .pb_data/         # L3 runtime — DO NOT COMMIT
```

See [Architecture overview](architecture-overview.md), [Markdown document rail](markdown-content.md), [Shaping and agents](shaping-and-agents.md), and the slim [site README](../../site/README.md).

## HTTP namespaces

| Prefix | Purpose |
|--------|---------|
| `/api/monms/*` | MonMS JSON REST (machine clients) |
| `/_monms/*` | Operator HTML tools (dashboard, publish, API keys, MCP settings, auth bridge) |
| MCP listener | Optional agentic HTTP (`mcp.enabled` in config) — see [MCP and API keys](mcp-and-api-keys.md) |
| `/api/collections/...` | PocketBase collection REST — see [PocketBase docs](https://pocketbase.io/docs/) |
| `/_/` | PocketBase admin SPA |

Full MonMS surface: [MonMS HTTP API](../reference/monms-api.md).
