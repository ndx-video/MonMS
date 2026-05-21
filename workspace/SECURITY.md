# MonMS Workspace Security Guide

This document covers SEC-03: SSH key scope, API token storage policy, workspace Git hygiene, and operator escape hatches for the MonMS workspace.


## SSH Key Scope

Use a **dedicated SSH key** exclusively for agent workspace access. A shared or personal key grants broader access than needed and increases blast radius if compromised.

### Recommended Setup

1. Generate a dedicated key pair for the agent:
   ```bash
   ssh-keygen -t ed25519 -C "monms-agent-workspace" -f ~/.ssh/monms_agent_ed25519
   ```

2. Restrict the key in `~/.ssh/authorized_keys` using the `command=` and `restrict` options to limit what the key can do and where it can access:
   ```
   restrict,command="internal-sftp -d /path/to/workspace",from="agent-ip" ssh-ed25519 AAAA... monms-agent-workspace
   ```
   Or use directory-restricted access by configuring your SSH server (`Match` block with `ChrootDirectory`) to confine the agent to the workspace subdirectory only.

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

- Store the token in an **environment variable** or a **secrets vault** (e.g., HashiCorp Vault, AWS Secrets Manager, `.env` file outside the workspace git repository):
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


## Git History Safety

The workspace Git repository records every mutation as an auditable commit. Maintaining Git hygiene prevents accidental secret exposure and keeps the audit trail clean.

### .pb_data Must Never Be Committed

`workspace/.pb_data/` contains the PocketBase SQLite database, encryption keys, and logs. It must be excluded from git:

```bash
# Add immediately after monms init
echo ".pb_data/" >> workspace/.gitignore
git -C workspace add .gitignore
git -C workspace commit -m "chore: exclude .pb_data from git"
```

Verify exclusion:
```bash
git -C workspace check-ignore -v .pb_data/
```

### Files That Must Never Be Committed

| File / Pattern | Reason |
|---|---|
| `.pb_data/` | PocketBase database and encryption keys |
| `.env` | Environment secrets |
| `*.key`, `*.pem`, `*.p12` | Private keys and certificates |
| `*token*`, `*secret*`, `*password*` | Credential files |

### Secret Accidentally Committed

If a secret is accidentally committed to the workspace repository:

1. **Rotate the secret immediately** — treat it as compromised regardless of repository visibility.
2. Remove from history using [BFG Repo-Cleaner](https://rtyley.github.io/bfg-repo-cleaner/) or `git filter-repo`:
   ```bash
   bfg --delete-files secret-file.txt workspace/
   git -C workspace reflog expire --expire=now --all && git -C workspace gc --prune=now --aggressive
   git -C workspace push --force
   ```
3. Notify all repository collaborators to re-clone (force-push rewrites history).


## Operator Escape Hatches

Use these commands to inspect, recover, or override workspace state in emergency situations.

### Inspect Commit History

```bash
git -C workspace log --oneline
```

### View Staged Changes Before Committing

```bash
git -C workspace diff --cached
```

### Revert a Bad Agent Commit

```bash
git -C workspace revert <commit-sha>
```

Creates a new commit that undoes the specified commit — preserves history, safe for shared repositories.

### Manual Working Tree Rollback

Discard all uncommitted changes and restore the working tree to HEAD:

```bash
git -C workspace checkout -- .
```

**Warning:** This is destructive and irreversible for uncommitted changes.

### Pre-Commit Hook Bypass (Emergency Only)

```bash
git -C workspace commit --no-verify -m "emergency: ..."
```

**WARNING:** `--no-verify` skips template validation entirely. Use only to unstick a broken hook or in a genuine emergency. Never use `--no-verify` in a normal agent workflow — it defeats the safety guardrails (AGT-03, AGT-04).

After using `--no-verify`, immediately run manual validation and fix any issues:

```bash
monms validate --workspace workspace/
```

### View Files Changed in Last Commit

```bash
git -C workspace show --stat HEAD
```
