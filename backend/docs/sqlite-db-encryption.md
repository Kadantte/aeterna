# SQLite Database Encryption

This document explains how full-database encryption works in Aeterna, how automatic migration behaves, and how to operate it safely in production.

## Scope

This feature applies to **SQLite only**.

Aeterna already encrypts sensitive payloads at application level (messages, farewell content, attachments). This layer adds encryption for the SQLite file itself.

## Configuration

Use these variables in `.env`:

```env
DB_ENCRYPTION_ENABLED=false
DB_ENCRYPTION_AUTO_MIGRATE=true
# Fixed path, expected by install.sh and backend runtime:
DB_ENCRYPTION_KDF_CONTEXT_FILE=./secrets/db_kdf_context
```

### Meaning

- `DB_ENCRYPTION_ENABLED`
  - `false`: database runs in plain SQLite mode
  - `true`: database runs in encrypted mode

- `DB_ENCRYPTION_AUTO_MIGRATE`
  - `true`: if the on-disk format does not match the requested mode, Aeterna migrates automatically at startup
  - `false`: startup fails on mismatch

- `DB_ENCRYPTION_KDF_CONTEXT_FILE`
  - fixed path to the persisted context used to derive the SQLite encryption key
  - must stay `./secrets/db_kdf_context`
  - generated once if missing (when encryption is enabled), reused afterwards

## Key Derivation Model

Aeterna does not reuse the master key directly as SQLCipher passphrase.

Instead, it derives a DB key using:

1. master key (`secrets/encryption_key`)
2. persisted KDF context (`secrets/db_kdf_context`)
3. HKDF-SHA256 with internal salt

This keeps key domains separated and avoids coupling DB encryption directly to raw master key material.

## Docker Requirements

The backend needs both:

- Docker secret: `/run/secrets/encryption_key`
- bind-mounted secrets directory: `/app/secrets` (host `./secrets`)

The context file (`db_kdf_context`) is stored in that mounted `secrets` directory so it survives container recreation.

## Startup Behavior

At startup, Aeterna attempts to open the database according to `DB_ENCRYPTION_ENABLED`.

### If `DB_ENCRYPTION_ENABLED=true`

1. open as encrypted
2. if open fails and `AUTO_MIGRATE=true`, probe plain
3. if plain opens, migrate plain -> encrypted
4. reopen encrypted

### If `DB_ENCRYPTION_ENABLED=false`

1. open as plain
2. if open fails and `AUTO_MIGRATE=true`, probe encrypted
3. if encrypted opens, migrate encrypted -> plain
4. reopen plain

If probes fail in both modes, startup fails.

## Migration Mechanics

Migration uses `ATTACH DATABASE ...` + `sqlcipher_export('target')`, then swaps files atomically.

### Plain -> Encrypted

- source: current `aeterna.db` (plain)
- target temp: `aeterna.db.enc.tmp`
- temporary backup created before swap:
  - `aeterna.db.plain-to-encrypted.<UTC_TIMESTAMP>.bak`

### Encrypted -> Plain

- source: current `aeterna.db` (encrypted)
- target temp: `aeterna.db.plain.tmp`
- temporary backup created before swap:
  - `aeterna.db.encrypted-to-plain.<UTC_TIMESTAMP>.bak`

WAL/SHM sidecars are cleaned for original, backup, and temp targets.

After migration succeeds and the database is reopened in the target mode, Aeterna removes the `.bak` file automatically.  
If startup fails before final reopen (or backup deletion fails), the `.bak` may remain for recovery.

## Operational Playbook

### Migrate current plain DB to encrypted

1. Set:

```env
DB_ENCRYPTION_ENABLED=true
DB_ENCRYPTION_AUTO_MIGRATE=true
```

2. Restart backend:

```bash
docker compose up -d backend
```

3. Check logs:

```bash
docker compose logs backend --since=5m
```

Expected result:

- migration log activity
- final startup with `Database connection successfully opened (encrypted): ...`

4. Verify plain sqlite can no longer read it:

```bash
sqlite3 data/aeterna.db '.tables'
```

Expected:

- `Error: file is not a database`

### Roll back to plain DB

Option A (automatic reverse migration):

1. Set:

```env
DB_ENCRYPTION_ENABLED=false
DB_ENCRYPTION_AUTO_MIGRATE=true
```

2. Restart backend:

```bash
docker compose up -d backend
```

Option B (manual backup restore, only if a `.bak` still exists after a failed/partial startup):

1. Stop backend
2. Restore the `.bak` file to `aeterna.db`
3. Start backend

## Important Safety Notes

1. **Do not rotate** `secrets/db_kdf_context` automatically.
   - Changing it changes derived DB key.
   - If changed unintentionally, encrypted DB may become unreadable.

2. Backup files produced during migration may contain plaintext data.
   - Handle them as sensitive artifacts.
   - Remove them after validation if your policy requires it.

3. Keep `secrets/encryption_key` and `secrets/db_kdf_context` together in backup strategy.
   - Losing either can block decryption paths.

## Troubleshooting

### `file is not a database`

- expected if you open encrypted `aeterna.db` with plain sqlite client
- expected during migration probes
- not necessarily an error by itself

### `hmac check failed for pgno=1`

- can appear while probing wrong mode before migration
- if backend finishes startup in expected mode, this is transitional noise

### Startup fails after enabling encryption

Check:

1. `secrets/encryption_key` exists and is valid
2. `secrets/db_kdf_context` exists or can be created
3. `./secrets` is mounted to `/app/secrets` in compose
4. backend logs for final fatal line

### Startup fails with `DB_ENCRYPTION_AUTO_MIGRATE=false`

This is expected on mode mismatch. Either:

- switch `AUTO_MIGRATE=true`, or
- restore DB file matching requested mode.

## Recommended Production Defaults

For fresh installations that require full DB encryption:

```env
DB_ENCRYPTION_ENABLED=true
DB_ENCRYPTION_AUTO_MIGRATE=true
# Fixed path, do not change:
DB_ENCRYPTION_KDF_CONTEXT_FILE=./secrets/db_kdf_context
```

For installations that want to stay plain for now:

```env
DB_ENCRYPTION_ENABLED=false
DB_ENCRYPTION_AUTO_MIGRATE=true
```

This keeps a reversible path available if DB format changes later.
