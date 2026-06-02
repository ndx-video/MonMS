# Security

> **Bundled PocketBase:** [v0.38.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.38.1)

SSH key scope, API token storage, site Git hygiene, and operator escape hatches for the MonMS site repository.

## SSH key scope

Use a **dedicated SSH key** exclusively for agent site access.

```bash
ssh-keygen -t ed25519 -C "monms-agent-site" -f ~/.ssh/monms_agent_ed25519
```

Restrict in `authorized_keys` with `command=` / `restrict` or chroot to the site subdirectory only. Never reuse the agent key elsewhere.

**Compromise response:** revoke key, audit `git log`, rotate PocketBase admin credentials.

## PocketBase admin token

Obtained via `POST /api/collections/_superusers/auth-with-password`. Grants full collection management.

- Store in environment or secrets vault — **never commit**.
- **Never log** the token.
- Rotate on suspected compromise (change admin password at `/_/`, re-fetch token).
- Default token lifetime ~7 days — re-authenticate on `401`.

## MonMS API keys (MCP)

Per-user keys for the optional MCP HTTP listener and future machine surfaces.

- Created at `/_monms/api-keys` (MonMS Console); full secret shown **once**.
- Stored hashed in `monms_api_keys`; optional env `MONMS_API_KEY_PEPPER` — never commit.
- Superusers always manage their own keys; `users` accounts only when `mcp.allowNonSuperuserKeys` is `true`.
- Revoke promptly on compromise; keys inherit the owner's PocketBase permissions.

See [MCP and API keys](mcp-and-api-keys.md).

## Publish token (`MONMS_PUBLISH_TOKEN`)

- Same value on staging (outbound) and production (import gate).
- Authorizes **only** `POST /api/monms/content/import`.
- Never commit or log.
- Clients publish via `/_monms/publish` with superuser login + `publisherEmails` allowlist.

## Git hygiene

Never commit:

| Path / pattern | Reason |
|----------------|--------|
| `.pb_data/` | PocketBase SQLite and keys |
| `.monms/config.json` | Production URL, publisher emails |
| `.monms/publish-state.json` | Last-publish checksum |
| `content/` | Ephemeral exports (by default) |
| `.env`, `*.pem`, `*secret*` | Credentials |

Commit `.monms/config.example.json` only.

If a secret was committed: rotate immediately, use BFG or `git filter-repo`, force-push only with team coordination.

## Operator escape hatches

```bash
git -C site log --oneline
git -C site revert <sha>
git -C site checkout -- .
monms validate --site site/
```

`git commit --no-verify` skips validation — emergency only. Never routine for agents.

## Related

- [Shaping and agents](shaping-and-agents.md)
- [Publish to live](../user-guide/publish-to-live.md)
