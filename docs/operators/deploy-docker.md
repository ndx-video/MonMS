# Docker deployment

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

Optional packaging when operators prefer containers. **Image = frozen L1 engine. Volume = git-managed L2 + persistent L3.**

See example files in the site repo: `Dockerfile.example`, `docker-compose.example.yml`.

## Four layers in Docker

| Layer | In image? | At runtime |
|-------|-----------|------------|
| **L1** Engine | Yes (production ldflag build) | Same image on staging and production |
| **L2** Structure | No | Bind mount of git checkout at `/app` |
| **L3** Content | No | Named volume at `/app/.pb_data` |
| **L4** Audience | No | Reverse proxy in front |

Structure promotes via **git tag + checkout** on the mount. Content promotes via **Publish to live** — unchanged from bare-metal.

## Build

```bash
go build -ldflags "-X main.buildMode=production" -o monms .
docker build -f Dockerfile.example -t monms:production .
```

Copy `monms` into the site root for the Docker build context — do not commit the binary.

## Single instance

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

Use `docker-compose.example.yml` as a starting point. Each service needs separate `.pb_data` volumes and environment-specific `.monms/config.json`.

Shape deploy:

```bash
monms site sync --site /path/to/staging --ref v1.2.0
monms site sync --site /path/to/production --ref v1.2.0
docker compose restart staging production
```

Optional startup sync: `shapeSync.enabled` in `.monms/config.json`.

## Checklist

- Production ldflag build for fsnotify + template cache.
- Named volume for `.pb_data/`.
- Secrets at runtime only — not in the image.
- TLS at reverse proxy; MonMS HTTP inside container.

## What not to do

- Commit `monms` to the site repo.
- Bake `.pb_data/` into the image.
- One container for both staging and production.
- Rebuild the image on every structure tag.

## Related

- [Getting started](getting-started.md)
- [CLI reference](../reference/cli.md)
