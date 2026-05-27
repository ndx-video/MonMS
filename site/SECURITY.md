# MonMS Workspace Security Guide

This document covers SEC-03: SSH key scope, API token storage policy, site Git hygiene, and operator escape hatches for the MonMS workspace.


## SSH Key Scope

Use a **dedicated SSH key** exclusively for agent site access. A shared or personal key grants broader access than needed and increases blast radius if compromised.

### Recommended Setup

1. Generate a dedicated key pair for the agent:
   ```bash
   ssh-keygen -t ed25519 -C "monms-agent-site" -f ~/.ssh/monms_agent_ed25519
   ```

2. Restrict the key in `~/.ssh/authorized_keys` using the `command=` and `restrict` options to limit what the key can do and where it can access:
   ```
   restrict,command="internal-sftp -d /path/to/site",from="agent-ip" ssh-ed25519 AAAA... monms-agent-site
   ```
   Or use directory-restricted access by configuring your SSH server (`Match` block with `ChrootDirectory`) to confine the agent to the site subdirectory only.

3. Never reuse the agent key for other systems or people.

### Key Compromise Response

If the agent SSH key is compromised:

1. **Immediately revoke** — remove the public key from `authorized_keys`.
2. **Audit git history** — check `git log --oneline` for unexpected agent commits.
3. **Rotate** — generate a new key pair and update the agent's configuration.
4. Review PocketBase admin token exposure (see §PocketBase Admin Token).


## PocketBase Admin Token

The PocketBase admin token (obtained via `POST /api/collections/_superusers/auth-with-password`) grants **full collection management access** — creation, modification, and deletion of all collections.

### Storage Policy

- Store the token in an **environment variable** or a **secrets vault** (e.g., HashiCorp Vault, AWS Secrets Manager, `.env` file outside the site git repository):
  ```bash
  export POCKETBASE_ADMIN_TOKEN="eyJhbGci..."
  ```
- **Never commit** a token string to git — not in templates, schema files, scripts, or comments.
- **Never log** the token to stdout/stderr or application logs.

### Token Scope

The admin token is scoped to **collection management only** (AGT-01). Do not use superuser credentials for:
- Regular content API calls (use collection-scoped API rules instead)
- File uploads or media operations
- Any operation that a non-admin API rule can handle

### Token Rotation

Rotate the admin token immediately on any suspected compromise:

1. Log into the PocketBase admin dashboard at `/_/`.
2. Navigate to Settings → Admins → change the admin password.
3. Re-obtain a fresh token and update the `POCKETBASE_ADMIN_TOKEN` environment variable wherever it is stored.

### Token Lifetime

PocketBase admin tokens expire (default 7 days). The agent should handle `401 Unauthorized` responses by re-authenticating via `/api/collections/_superusers/auth-with-password` rather than caching tokens indefinitely.


## Publish token (`MONMS_PUBLISH_TOKEN`)

Production content import uses a **scoped publish token** — not the PocketBase admin JWT (PUB-05).

### Storage policy

- Set `MONMS_PUBLISH_TOKEN` in the **host environment** on both staging (outbound publish) and production (import API gate) — same value, never commit.
- **Never log** the token in application output or CLI error messages.
- Rotate immediately if exposed; update both environments together.

### Scope

The publish token authorizes **only** `POST /api/monms/content/import` on production. It cannot manage collections, users, or structure. Clients use the publish console at `/_monms/publish` with superuser login + publisher allowlist — see [EDITING-GUIDE.md](EDITING-GUIDE.md).


## Git History Safety

The site Git repository records every mutation as an auditable commit. Maintaining Git hygiene prevents accidental secret exposure and keeps the audit trail clean.

### Runtime paths must never be committed

`monms init` scaffolds `site/.gitignore` excluding runtime data and publish secrets:

| Path | Reason |
|------|--------|
| `.pb_data/` | PocketBase database and encryption keys |
| `.monms/config.json` | Production URL + publisher allowlist (site-specific) |
| `.monms/publish-state.json` | Last-publish checksum (environment-specific) |
| `content/` | Editorial export snapshots (ephemeral by default) |

Commit `.monms/config.example.json` as the template only. Older sites without `.gitignore` should add the same patterns manually:

```bash
git -C site check-ignore -v .pb_data/ .monms/config.json
```

### Files That Must Never Be Committed

| File / Pattern | Reason |
|---|---|
| `.pb_data/` | PocketBase database and encryption keys |
| `.monms/config.json` | Staging publish config (production URL, publisher emails) |
| `.monms/publish-state.json` | Publish checksum state |
| `content/` | Editorial export snapshots |
| `.env` | Environment secrets |
| `*.key`, `*.pem`, `*.p12` | Private keys and certificates |
| `*token*`, `*secret*`, `*password*` | Credential files (use env vars instead) |

### Secret Accidentally Committed

If a secret is accidentally committed to the site repository:

1. **Rotate the secret immediately** — treat it as compromised regardless of repository visibility.
2. Remove from history using [BFG Repo-Cleaner](https://rtyley.github.io/bfg-repo-cleaner/) or `git filter-repo`:
   ```bash
   bfg --delete-files secret-file.txt site/
   git -C site reflog expire --expire=now --all && git -C site gc --prune=now --aggressive
   git -C site push --force
   ```
3. Notify all repository collaborators to re-clone (force-push rewrites history).


## Operator Escape Hatches

Use these commands to inspect, recover, or override workspace state in emergency situations.

### Inspect Commit History

```bash
git -C site log --oneline
```

### View Staged Changes Before Committing

```bash
git -C site diff --cached
```

### Revert a Bad Agent Commit

```bash
git -C site revert <commit-sha>
```

Creates a new commit that undoes the specified commit — preserves history, safe for shared repositories.

### Manual Working Tree Rollback

Discard all uncommitted changes and restore the working tree to HEAD:

```bash
git -C site checkout -- .
```

**Warning:** This is destructive and irreversible for uncommitted changes.

### Pre-Commit Hook Bypass (Emergency Only)

```bash
git -C site commit --no-verify -m "emergency: ..."
```

**WARNING:** `--no-verify` skips template validation entirely. Use only to unstick a broken hook or in a genuine emergency. Never use `--no-verify` in a normal agent workflow — it defeats the safety guardrails (AGT-03, AGT-04).

After using `--no-verify`, immediately run manual validation and fix any issues:

```bash
monms validate --site site/
```

### View Files Changed in Last Commit

```bash
git -C site show --stat HEAD
```
