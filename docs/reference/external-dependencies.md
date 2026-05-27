# External dependencies

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1) — pinned in [`go.mod`](../../go.mod)

MonMS embeds PocketBase and Go's `html/template` engine. This page lists authoritative external references. MonMS docs do **not** duplicate PocketBase or Go template manuals.

## PocketBase

| Topic | Reference |
|-------|-----------|
| Version bundled with MonMS | [PocketBase v0.38.1 release](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1) |
| REST API, auth, collections | [PocketBase documentation](https://pocketbase.io/docs/) |
| Collection records API | [API records](https://pocketbase.io/docs/api-records/) |
| Admin dashboard | Served at `/_/` on every MonMS instance |

When upgrading MonMS, check `go.mod` for the PocketBase version — API behavior follows upstream release notes.

## Go HTML templates

MonMS site pages use Go's `html/template` (not Jinja, not Handlebars).

| Topic | Reference |
|-------|-----------|
| Template syntax, actions, pipelines | [`html/template` package docs](https://pkg.go.dev/html/template) |
| Auto-escaping, context rules | Same package docs — use `.` field access, `{{if}}`, `{{range}}`, `{{define}}` / `{{template}}` |

MonMS-specific conventions: [Templates and routing](../operators/templates-and-routing.md).

## Sentinel (extensions)

For server-side automation beyond MonMS routes, see [Extensibility with Sentinel](../operators/extensibility-with-sentinel.md).

- Product site (planned): [https://ndx-sentinel.com](https://ndx-sentinel.com)
- Documentation (planned): [https://docs.ndx-sentinel.com](https://docs.ndx-sentinel.com)

## Related

- [MonMS HTTP API](monms-api.md) — routes MonMS owns
- [Architecture overview](../operators/architecture-overview.md)
