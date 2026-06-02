---
name: monms-architecture
description: >-
  MonMS four-layer model, dual promotion rails, HTTP namespaces, and skill routing.
  Use for any MonMS task, cold start, or when terminology (site directory vs ./site
  default, staging vs buildMode) is ambiguous.
---

# MonMS Architecture

Load this skill first for any MonMS work. For full command lists and D-* decisions, see [PROJECT.md](../../../PROJECT.md).

## Four layers

| Layer | Artifact | In Git? | Promoted how |
|-------|----------|---------|--------------|
| **L1** Engine | `monms` binary | No | Semver release |
| **L2** Structure | Site directory: `templates/`, `schema/`, `assets`, `documents/` (markdown) | Yes | Git tag → `monms site sync` |
| **L3** Content | `.pb_data/` records | No | PB-native: `/_monms/publish`; Markdown: Git structure rail |
| **L4** Audience | Production URL | No | Read-only |

**Structure rail** and **content rail** are independent for PB-native collections. Markdown collections (`monms.source=markdown` in schema JSON) are Git-canonical under `{siteDir}/documents/` and sync into `.pb_data/` on bootstrap — see [markdown-content.md](../../../docs/operators/markdown-content.md).

## Site directory (do not assume `./site`)

MonMS resolves **one site directory per process** — the L2 deployable tree containing `templates/`, `schema/`, `assets/`, `.monms/`, and `.pb_data/`.

| Resolution | Rule |
|------------|------|
| `-s` / `--site PATH` | Wins over env (D-26) |
| `MONMS_SITE` | Used when CLI flag omitted |
| Default | `./site` if neither is set |

**The folder name is not fixed.** The same engine binary and parent repo may run `./site-stage` on one host and `./site-prod` on another. Docs and this repo often use `./site` as the demo default — always confirm which path the running instance uses before editing files or running CLI commands.

Paths below are **relative to the configured site directory** (written `{siteDir}/…`). The site tree is often its **own Git repo**; the engine repo is the parent.

## Terminology traps

| Term | Meaning |
|------|---------|
| **Site directory** | Configured L2 root (`-s` / `MONMS_SITE` / default `./site`). Legacy docs may say "workspace" — use **site** for the concept, not a hardcoded path. |
| **Shaping phase** | Consultants/agents edit L2 in the site directory's Git repo |
| **Staging instance** | A deployed MonMS with its own `.pb_data/` for client editorial work |
| **`buildMode=development`** | Compile-time ldflag on the binary — **not** the product "staging" phase |
| **`buildMode=production`** | Enables template cache + fsnotify watcher on entire site tree |

## HTTP namespaces

| Prefix | Type | Audience |
|--------|------|----------|
| `/api/monms/*` | JSON REST | Machine clients, Bearer tokens |
| `/_monms/*` | HTML + session | Browser operators (publish console, auth bridge) |
| `/api/collections/...` | PocketBase REST | Collection CRUD — see [PocketBase docs](https://pocketbase.io/docs/) |
| `/_/` | PocketBase admin SPA | Full management fallback |

Canonical path constants: `internal/monmsroutes/routes.go`.

## What to change vs leave alone

| Change type | Where |
|-------------|-------|
| New page/route | `{siteDir}/templates/{slug}.gohtml` |
| Global layout/HTMX | `{siteDir}/templates/layouts/base.gohtml` + `internal/scaffold/embed/base.gohtml` |
| New collection | PocketBase API + `{siteDir}/schema/{name}.json` |
| Markdown documents | `{siteDir}/documents/{type}/**/*.md` + `monms documents sync` — package `internal/documents/` |
| Markdown page SSR | `{siteDir}/templates/doc.gohtml` + slug lookup in `internal/router/documents.go` |
| SSR behavior | `internal/router/ssr.go` |
| Cache/watcher | `internal/templates/` |
| Validation rules | `internal/validate/validate.go` |
| Content publish rail | `internal/content/` — clients use `/_monms/publish` |
| Operator console UI | `internal/monmsdash/` — see [monms-dash](../monms-dash/SKILL.md) |
| Planning artifacts | `.planning/` — only when user asks |

## Skill routing

| Task | Load skill |
|------|------------|
| Any MonMS task (start here) | `monms-architecture` |
| Edit L2 templates, schema, or assets in the configured site directory | [monms-site-shaping](../monms-site-shaping/SKILL.md) |
| Edit `internal/*`, `main.go`, tests, CLI | [monms-engine-development](../monms-engine-development/SKILL.md) |
| Docker, `site sync`, `.monms/config.json`, logging | [monms-operators-deploy](../monms-operators-deploy/SKILL.md) |
| `/_monms/` console, notifications, dashboard pages | [monms-dash](../monms-dash/SKILL.md) |
| Plugins, in-process extensions, Sentinel | [monms-extensibility](../monms-extensibility/SKILL.md) |

## Universal boundaries (always apply)

- Self-hosted, sovereign infrastructure over managed cloud defaults
- Never hardcode secrets; reverse proxy handles external access
- Use `vi` for terminal file edits — not `nano`
- Never commit: `{siteDir}/.pb_data/`, `{siteDir}/.monms/config.json`, `{siteDir}/.monms/publish-state.json`, `{siteDir}/content/`, tokens

## Related docs

- [PROJECT.md](../../../PROJECT.md) — cold-start index, commands, D-* table, pitfalls
- [docs/operators/getting-started.md](../../../docs/operators/getting-started.md) — layers and rails in depth
- [docs/operators/markdown-content.md](../../../docs/operators/markdown-content.md) — dual-rail markdown CMS
