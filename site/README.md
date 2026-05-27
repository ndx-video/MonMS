# MonMS site

Git-tracked **structure** (L2) for this deployment: templates, schema JSON, and assets. The `monms` binary (L1) stays generic.

Editorial **content** (L3) lives in `.pb_data/` at runtime — not in Git. Clients publish copy to production via `/_monms/publish`.

## Directory layout

```
site/
├── schema/                 # PocketBase collection definitions (JSON)
├── content/                # Editorial exports (optional; often gitignored)
├── templates/
│   ├── layouts/base.gohtml
│   ├── fragments/
│   ├── errors/
│   └── *.gohtml
├── assets/
├── .monms/                 # config.example.json committed; config.json gitignored
└── .pb_data/               # Runtime PocketBase data — DO NOT COMMIT
```

## Environment variables

| Variable | Purpose |
|----------|---------|
| `MONMS_URL` | Running server URL |
| `MONMS_SITE` | Override site path |
| `MONMS_BIN` | Path to `monms` (pre-commit hook) |
| `POCKETBASE_ADMIN_TOKEN` | Admin JWT for shaping (never commit) |
| `MONMS_PUBLISH_TOKEN` | Production import secret (never commit) |

## Documentation

Operator and editor guides live in the **MonMS engine repository**, not in this site repo:

**[../docs/README.md](../docs/README.md)** — user guide, operators manual, API reference.

> **Bundled PocketBase:** v0.38.1 (see engine `go.mod`)

Scaffold or refresh this tree:

```bash
monms init --site .
```
