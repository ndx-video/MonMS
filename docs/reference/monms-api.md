# MonMS HTTP API

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

This reference documents **MonMS-only** routes. PocketBase collection REST (`/api/collections/...`), auth, and admin APIs are covered by the [official PocketBase documentation](https://pocketbase.io/docs/).

Canonical path constants: `internal/monmsroutes/routes.go`.

## Namespaces

| Prefix | Content type | Audience |
|--------|--------------|----------|
| `/api/monms/*` | JSON only | Machine clients, CI, Sentinel flows |
| `/_monms/*` | HTML + session helpers | Browser operators |

## `POST /api/monms/content/import`

Production editorial upsert (content rail).

**Auth:** `Authorization: Bearer <MONMS_PUBLISH_TOKEN>`

**Body:**

```json
{
  "collections": [
    {
      "name": "hero_content",
      "records": [
        { "id": "homepage", "title": "...", "body": "..." }
      ]
    }
  ]
}
```

**Response:** `200` with `{ "upserted": N, "collections": M, "warnings": [...] }`

**Rules:**

- Denies system collections (`_superusers`, `users`).
- Max 1000 records per collection per request.
- Upsert by record ID — deletions are not propagated from staging exports.
- Skips PocketBase file-type fields (warns in export path).

Used by staging **Publish to live** and `monms content publish`.

## `GET /_monms/publish`

HTML **Publish to live** console.

**Auth:** PocketBase superuser session + email in `site/.monms/config.json` `publisherEmails`.

Shows diff preview, last publish time, production URL configuration status.

## `GET /_monms/publish/diff`

JSON diff of pending editorial changes vs last publish.

**Auth:** Same as publish page.

## `POST /_monms/publish`

Executes publish: export staging editorial collections → POST to `productionUrl` + `/api/monms/content/import`.

**Auth:** Same as publish page.

## `POST /_monms/auth/sync`

Bridges PocketBase admin `localStorage` token to HttpOnly `monms_auth` cookie for inline editing on public pages.

**Auth:** `Authorization: Bearer <superuser token>`

**Response:** `204 No Content` with `Set-Cookie`.

Called automatically from the base layout script when a user visits `/` after logging in at `/_/`.

## `GET|POST /_monms/auth/logout`

Clears `monms_auth` cookie. Optional `?redirect=/`.

## Related

- [CLI reference](cli.md) — `monms content` subcommands
- [Publish to live](../user-guide/publish-to-live.md)
- [External dependencies](external-dependencies.md)
