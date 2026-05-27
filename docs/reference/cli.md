# CLI reference

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

Global flag: `-s` / `--site PATH` (default `./site` or `MONMS_SITE`).

## `monms init`

Scaffold site templates, schema, assets, `.monms/config`, pre-commit hook. Idempotent — skips existing files.

```bash
monms init --site ./my-site
```

## `monms validate`

Lint `*.gohtml` with the same parse path as production SSR.

```bash
monms validate --site ./site templates/index.gohtml
```

Pre-commit mode: no file args → validates staged `.gohtml` from git.

## `monms content`

Editorial content rail (collections with `"editorial": true` in schema JSON).

| Subcommand | Purpose |
|------------|---------|
| `export` | Write `site/content/{collection}.json` |
| `import` | Upsert `content/*.json` into local `.pb_data/` |
| `diff` | Field-level changes since last publish; exit 1 if pending |
| `publish` | Export + POST to production (`--to URL`, `MONMS_PUBLISH_TOKEN`) |

```bash
monms content export --site ./site
monms content publish --site ./site --to https://production.example.com
```

Clients normally use `/_monms/publish` instead.

## `monms site sync`

Fetch tags and checkout a shape ref in the site Git repo.

```bash
monms site sync --site ./site --ref v1.2.0
monms site sync --ref main --remote origin --force
```

Optional startup: `shapeSync` in `.monms/config.json`.

## `monms serve` (default)

Start PocketBase + MonMS routes. PocketBase subcommands (`superuser`, etc.) pass through.

```bash
monms serve --http=127.0.0.1:8090
```

Production binary: `go build -ldflags "-X main.buildMode=production" -o monms .`

## `monms stop`

SIGTERM running `monms serve` processes for this binary.

## Related

- [MonMS HTTP API](monms-api.md)
- [Getting started](../operators/getting-started.md)
