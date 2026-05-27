# Architecture overview

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

MonMS (**Monolithic Management System**) is an agent-malleable, single-binary CMS built on Go and PocketBase. The Go binary stays generic; site structure lives in a Git-tracked `site/` directory.

This page summarizes product intent verified against the current codebase. For step-by-step operations, see [Getting started](getting-started.md).

## Core pillars

| Pillar | Implementation |
|--------|----------------|
| **Single-binary monolith** | PocketBase embedded in `monms`; target **< 30 MB** idle heap (production build) |
| **Zero-compilation malleability** | Templates and schema change on disk; production uses fsnotify cache invalidation |
| **Git-managed structure** | L2 in `site/` Git; tagged releases for shape deploy |
| **Client-driven content publish** | L3 in `.pb_data/`; JSON upsert via `/_monms/publish` |
| **Inline contextual editing** | HTMX blur-save on authenticated pages; PocketBase admin at `/_/` as fallback |

## Runtime model

- **Production vs development mode** is compile-time via `main.buildMode` ldflags — not an `ENV` variable.
- PocketBase data directory: `{site}/.pb_data/`.
- Template validation mirrors production SSR: `html/template.ParseFiles(layout, page)`.
- Slug resolution uses **mirror + index**: `/` → `index.gohtml`, `/press` → `press/index.gohtml` or `press.gohtml`.

## What MonMS deliberately omits

MonMS does **not** support plugins or in-process extensions. Server-side customization belongs in the **MonMS JSON API** or in **Sentinel** (external orchestration). Rationale: [Extensibility with Sentinel](extensibility-with-sentinel.md).

## Out of scope (core model)

- React/Next.js or Node build pipelines
- External databases (PostgreSQL, MySQL)
- Multi-site Host-header routing (v2 backlog)
- Kubernetes-style orchestration (optional single-container Docker only)

Backlog items live in `.planning/ROADMAP.md` in the engine repository.

## Legacy specs

Older narrative specs in [`specs/monms-prd.md`](../../specs/monms-prd.md) and [`specs/staging.md`](../../specs/staging.md) may contain stale detail. **The codebase and this documentation tree are authoritative.**
