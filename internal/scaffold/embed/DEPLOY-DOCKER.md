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

Structure promotes via **git tag + checkout on the mounted site** on **both** staging and production hosts (operator policy). Content promotes via **Publish to live** (`/_monms/publish`) — JSON outside Git, unchanged from non-Docker deploys.

## Prerequisites

1. **Production binary** — build in the MonMS engine repo:

   ```bash
   go build -ldflags "-X main.buildMode=production" -o monms .
   ```

   Copy `monms` into this site root (gitignored). Do not commit it.

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
| Workspace mount | Tagged shape (e.g. `v1.2.0`) — same tag as production | Tagged shape (e.g. `v1.2.0`) |
| `.pb_data/` volume | `staging_pb_data` | `production_pb_data` |
| `.monms/config.json` | `productionUrl` → live site | `productionUrl` empty; import API only |
| `MONMS_PUBLISH_TOKEN` | Same secret as production | Same secret as staging |

Consultants **shape** on a site checkout and tag releases. Operators maintain **two workspace clones** (or two host paths) so staging and production can run as separate instances; a deploy policy pulls the same tag to both when shape changes ship.

## Structure deploy (L2 rail — shaping)

When a shape release is tagged, operator policy updates **both** site checkouts. Examples:

**Built-in CLI (cron/CI):**

```bash
monms site sync --site /path/to/staging --ref v1.2.0
monms site sync --site /path/to/production --ref v1.2.0
docker compose restart staging production
```

**Optional startup sync** — set in each instance's `.monms/config.json`:

```json
"shapeSync": {
  "enabled": true,
  "ref": "v1.2.0",
  "remote": "origin",
  "force": false,
  "failOnError": false
}
```

When `failOnError` is false (default), serve continues with the current checkout if sync fails.

**Manual git (equivalent):**

```bash
git -C /path/to/staging-site fetch --tags && git -C /path/to/staging-site checkout v1.2.0
git -C /path/to/production-site fetch --tags && git -C /path/to/production-site checkout v1.2.0
docker compose restart staging production
```

Production builds enable fsnotify on the site tree — template changes after checkout may reload without restart, but restarting the container is the safe default.

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
- Multi-stage-build the MonMS engine source inside this site repo — keep engine and site repos separate.

## Files

| File | Purpose |
|------|---------|
| [Dockerfile.example](Dockerfile.example) | Thin L1 image; expects `monms` binary in build context |
| [docker-compose.example.yml](docker-compose.example.yml) | Staging + production with separate volumes |
| [.monms/config.example.json](.monms/config.example.json) | Committed config template |
