---
name: monms-operators-deploy
description: >-
  MonMS operator deployment — staging/production topology, {siteDir}/.monms/config.json,
  monms site sync, shapeSync, Docker layout, and file logging. Use for Docker,
  site sync, config.json, logging, or multi-instance setup.
---

# MonMS Operators & Deploy

Foundation: [monms-architecture](../monms-architecture/SKILL.md). Shaping workflows: [monms-site-shaping](../monms-site-shaping/SKILL.md).

## Staging vs production

Typical deployment: **two MonMS instances** with the same tagged shape, independent content. Each instance points at its **own site directory** (`-s` / `MONMS_SITE`); names like `./site-stage` and `./site-prod` are common — do not assume `./site`.

| | Staging | Production |
|---|---------|------------|
| Purpose | Clients edit and preview copy | Audience reads live site |
| `.pb_data/` | Staging SQLite (never wholesale-synced) | Production SQLite |
| Structure | Same Git tag (e.g. `v1.2.0`) | Same Git tag |
| Content | Edited inline here | Updated via **Publish to live** only |

Consultants **shape** structure and tag releases. Clients **stage** content and publish — consultants are not in the routine content loop.

Operator entry: sign in at `/_/` (PocketBase admin), then open **MonMS Console** in the admin header → `/_monms/`. Console UI patterns: [monms-dash](../monms-dash/SKILL.md).

## Site directory layout

Whatever path MonMS is started with (`-s ./site-stage`, `-s ./site-prod`, default `./site`, etc.), that directory contains:

```
{siteDir}/
├── schema/           # L2 — collection definitions
├── content/          # L3 — editorial exports (optional, often gitignored)
├── templates/        # L2 — Go HTML templates
├── assets/           # L2 — CSS, static files
├── .monms/           # Operator config + logs
│   └── logs/         # Rotated log files (gitignored)
└── .pb_data/         # L3 runtime — DO NOT COMMIT
```

The site directory is often its **own Git repository** nested under the engine repo. The folder name is operator-defined; `./site` is only the CLI default.

## `.monms/config.json`

Live config is **gitignored**. Commit `{siteDir}/.monms/config.example.json` only (path relative to the configured site directory).

| Field | Purpose |
|-------|---------|
| `siteUrl` | This instance URL (startup banner links) |
| `productionUrl` | Publish-to-live target (not this instance) |
| `publisherEmails` | Emails allowed at `/_monms/publish` |
| `allowedHosts` | Injects `monms serve --origins` when CLI flag omitted |
| `bind` | Injects `--http=host:port` when CLI `--http` omitted |
| `logging` | File log levels under `.monms/logs/` |
| `loggingRotation` | Lumberjack rotation settings |
| `shapeSync` | Optional fetch + checkout at serve startup |

**CLI always wins:** `--http`, `--origins` override config.json values.

Logging defaults when `logging` omitted: production binary → error/warn/schema; development binary → all levels. See [docs/reference/logging.md](../../../docs/reference/logging.md).

## One-time consultant setup

1. Copy `config.example.json` → `config.json`; set `siteUrl`, `productionUrl`, `publisherEmails`
2. Set `MONMS_PUBLISH_TOKEN` on **both** staging (outbound) and production (import gate) — same secret, never commit

## Structure promotion

```bash
monms site sync --ref v1.2.0 -s /path/to/site-stage
# same engine repo, different directory for prod:
monms site sync --ref v1.2.0 -s /path/to/site-prod
```

Fetches and checks out the Git tag in the site repo. Operator-defined CI (GitHub Actions, cron) is also valid.

Optional automatic sync at serve startup via `shapeSync` in config.json.

## Docker layout

From [docs/operators/deploy-docker.md](../../../docs/operators/deploy-docker.md):

| Layer | Docker approach |
|-------|-----------------|
| L1 Engine | `monms` binary in image |
| L2 Structure | Mount the site directory (templates, schema, assets) — path is deployment-specific |
| L3 Content | Volume for `.pb_data/` inside that directory |

Examples in scaffold: `Dockerfile.example`, `docker-compose.example.yml` under `internal/scaffold/embed/` (often copied into a demo `./site` checkout).

After shape deploy: restart instance or rely on production watcher for template changes.

## Security hygiene

- Dedicated SSH key scoped to site directory for agents
- Never commit `.pb_data/`, `config.json`, `publish-state.json`, `content/`, tokens
- `git commit --no-verify` — emergency only

Full guide: [docs/operators/security.md](../../../docs/operators/security.md).

## Operator escape hatches

```bash
export SITE=./site-stage   # match the instance you are fixing
git -C "$SITE" log --oneline
git -C "$SITE" revert <sha>
git -C "$SITE" checkout -- .
monms validate -s "$SITE"
```

## Related docs

- [docs/operators/getting-started.md](../../../docs/operators/getting-started.md)
- [docs/operators/deploy-docker.md](../../../docs/operators/deploy-docker.md)
- [docs/user-guide/publish-to-live.md](../../../docs/user-guide/publish-to-live.md)
- [docs/user-guide/media-urls.md](../../../docs/user-guide/media-urls.md) — CDN URLs, no blob copy between envs
