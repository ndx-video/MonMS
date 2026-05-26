# Docker deployment (git-on-volume)

Optional packaging for MonMS when operators prefer containers over a bare binary on the host. This model keeps **Git as the structure rail (L2)** and uses Docker only for the **engine (L1)** plus **persistent runtime state (L3)**.

> **Image = frozen L1 engine. Volume = git-managed L2 + persistent L3.**

Full lifecycle spec: [../specs/staging.md](../specs/staging.md)

## Four layers in Docker

| Layer | Artifact | In the image? | At runtime |
|-------|----------|---------------|------------|
| **L1** Engine | `monms` binary (production ldflag) | Yes — baked at `docker build` | Same pinned image on staging and production |
| **L2** Structure | `templates/`, `schema/`, `assets/` | No | Bind mount of git checkout at `/app` |
| **L3** Content | `.pb_data/` SQLite | No | Named volume at `/app/.pb_data` |
| **L4** Audience | Public URL | No | Reverse proxy in front of production service |

Structure promotes via **git tag + checkout on the mounted workspace**, not by rebuilding the image. Content promotes via **Publish to live** (`/_monms/publish`) — unchanged from non-Docker deploys.

## Prerequisites

1. **Production binary** — build in the MonMS engine repo:

   ```bash
   go build -ldflags "-X main.buildMode=production" -o monms .
   ```

   Copy `monms` into this workspace root (gitignored). Do not commit it.

2. **Build the engine image**:

   ```bash
   docker build -f Dockerfile.example -t monms:production .
   ```

3. **Site config** — copy `.monms/config.example.json` to `.monms/config.json` on each environment. Keep `productionUrl` and `publisherEmails` environment-specific.

4. **Publish token** — set the same `MONMS_PUBLISH_TOKEN` on staging (outbound publish) and production (import API gate). Never commit or `COPY` into the image.

## Single instance (development or one environment)

```bash
docker run --rm \
  -p 8090:8090 \
  -v "$(pwd):/app" \
  -v monms_pb_data:/app/.pb_data \
  -v "$(pwd)/.monms/config.json:/app/.monms/config.json:ro" \
  -e MONMS_PUBLISH_TOKEN="${MONMS_PUBLISH_TOKEN}" \
  monms:production
```

## Staging + production

Use [docker-compose.example.yml](docker-compose.example.yml) as a starting point:

```bash
cp docker-compose.example.yml docker-compose.yml
# Edit paths: STAGING_WORKSPACE, PRODUCTION_WORKSPACE, config mounts
echo 'MONMS_PUBLISH_TOKEN=your-long-random-secret' > .env
docker compose up -d
```

Each service needs:

| Concern | Staging | Production |
|---------|---------|------------|
| Workspace mount | Active branch (e.g. `main`) | Tagged release (e.g. `v1.2.0`) |
| `.pb_data/` volume | `staging_pb_data` | `production_pb_data` |
| `.monms/config.json` | `productionUrl` → live site | `productionUrl` empty; import API only |
| `MONMS_PUBLISH_TOKEN` | Same secret as production | Same secret as staging |

Consultants typically maintain **two workspace clones** (or two host paths) so staging and production can run different git refs simultaneously.

## Structure deploy (L2 rail)

When a structure release is tagged:

```bash
git -C /path/to/production-workspace fetch --tags
git -C /path/to/production-workspace checkout v1.2.0
docker compose restart production
```

Production builds enable fsnotify on the workspace tree — template changes after checkout may reload without restart, but restarting the container is the safe default.

**Do not** rebuild the Docker image for structure-only changes unless you also upgrade the engine (L1).

## Content deploy (L3 rail)

Unchanged — clients use **Publish to live** on staging. No container steps required.

## Production checklist

- **Binary**: `-ldflags "-X main.buildMode=production"` — dev binaries skip template cache and fsnotify (D-01).
- **Persistence**: Named volume for `/app/.pb_data/` — data is lost if omitted.
- **Secrets**: `MONMS_PUBLISH_TOKEN` and `.monms/config.json` at runtime only.
- **TLS**: Terminate HTTPS at a reverse proxy (Caddy, nginx, Traefik); MonMS listens on HTTP inside the container.
- **CORS**: Set `allowedHosts` in staging `config.json` when using a public staging hostname.
- **Backups**: Snapshot the `.pb_data` volume — the content publish rail is not a full database backup.

## What not to do

- Commit `monms` to this git repo.
- Bake `.pb_data/` into the image.
- Use one container for both staging and production.
- Rebuild the image on every structure tag — Git checkout on the volume is the promotion mechanism.
- Multi-stage-build the MonMS engine source inside this workspace repo — keep engine and workspace repos separate.

## Files

| File | Purpose |
|------|---------|
| [Dockerfile.example](Dockerfile.example) | Thin L1 image; expects `monms` binary in build context |
| [docker-compose.example.yml](docker-compose.example.yml) | Staging + production with separate volumes |
| [.monms/config.example.json](.monms/config.example.json) | Committed config template |
