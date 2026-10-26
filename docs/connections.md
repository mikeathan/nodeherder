# Integrations: local and personal Docker setup

Integrations adds external information independently of Devices. The first release supports UK
bin collections through a supplied WhenIsBins definition. Updates are manual: Refresh retrieves a
schedule, and Check status continues a pending lookup. No provider work runs on a recurring timer.
All signed-in users have the same controls. Nothing is added to Home, Overview, Devices or Panel.

## Local run: automatic first-time setup

A native backend run now creates a storage encryption key automatically on its first startup.
No Docker mount, provider API token or manual key command is needed. The key is stored outside
this checkout in the OS user-configuration directory:

- macOS: `~/Library/Application Support/nodeherder/integration-keyring.json`
- Linux: `$XDG_CONFIG_HOME/nodeherder/integration-keyring.json`, or `~/.config/nodeherder/integration-keyring.json`

Restart the local backend after updating to load this behavior. Page reads do not restart or
reinitialise private storage. The key directory is private (0700), and the key file is 0600.
The same key is reused on later starts. Back it up separately from `backend/data/integrations.db`.
If a DB already exists but its key is missing, startup reports unavailable and requires the
matching backup; it never invents a replacement key. Existing invalid keys also stay untouched.

Setting `INTEGRATION_KEYRING_FILE` selects an explicit key file instead: it must already exist.
It takes precedence over all managed defaults and never creates a missing explicit file. If that
setting is empty, `INTEGRATION_KEYRING_DIR` can select an absolute managed directory; otherwise
native runs use the OS path above. The default storage key is not a WhenIsBins API credential.

## Docker: automatic first-time setup

The production and development backend Compose files now work with integrations by default.
No extra flag, host Node installation or manual key generation is required. Existing bind mounts
still persist `./backend/configs` and `./backend/data`. A separate named Docker volume,
`integration-secrets`, is mounted at `/run/nodeherder-secrets`; Docker prefixes the actual volume
name with the Compose project name. The key is stored inside a private child directory:

```text
/run/nodeherder-secrets/managed/integration-keyring.json
```

The managed directory is 0700 and the key is 0600. Using a child is intentional: a newly created
Docker volume root may be 0755. First startup provisions the key only when no integrations DB
exists; later starts reuse it. Config/data/key mounts survive normal container replacement and
`manage-stack.sh down` (which does not delete volumes). Keep the same Compose project name.

Provider YAML lives in `backend/configs/integrations/waste_collection/whenisbins.yaml`; private
state lives in `backend/data/integrations.db` (0600). The key is kept separately from both trees.
There is no temporary-storage fallback and installation performs no provider lookup.

### Updating an existing installation

Use the updated backend image **and** Compose files, then recreate the backend. Restarting an
old container alone does not add mounts or environment settings. The usual manager command is:

```sh
./scripts/manage-stack.sh rebuild backend
```

`--connections` is not needed for the default. If the previous broken installation never created
an integrations DB, it will initialize automatically. An existing DB without its original key
remains unavailable and requires the matching key backup; the fix does not invent a replacement.
If `INTEGRATION_KEYRING_FILE` was explicitly set in your root `.env`, that override still wins:
check its actual mount rather than silently switching an existing database to a different key.

Do not delete the secrets volume, use `docker compose down -v`, or prune it while retaining the DB
unless you have a verified matching key backup. A different Compose project name selects a different
volume and may require restoring/reattaching the original key. Do not run two backends against the
same data directory: one process owns the bbolt DB.

### Optional externally managed key

For an existing externally held key or a deliberately read-only host keyring, retain the existing
`--connections` override. This is an advanced option, not a requirement for ordinary startup.
Create a private directory outside the checkout, owned by the container user. For a **new DB only**,
a fresh host key can be provisioned after creating `/srv/nodeherder-secrets` with mode 0700:

```sh
node scripts/create-integration-keyring.mjs /srv/nodeherder-secrets/integration-keyring.json
```

For an existing DB, copy its matching keyring there instead of generating a new one. Set the file
mode to 0600 and set `NODEHERDER_SECRETS_DIR=/srv/nodeherder-secrets` in the root local `.env`.
Then use `scripts/manage-stack.sh start all --connections`, retaining that flag on later backend/
all commands. Direct Compose users add `-f docker-compose.connections.yml` alongside their usual
files. The overlay replaces the default named-volume mount with the supplied host directory,
read-only, and selects `/run/nodeherder-secrets/integration-keyring.json` as an existing explicit file.

### Diagnosing startup

The authenticated Integrations page distinguishes storage-key, saved-data and provider-directory
failures. Startup logs include `Integrations startup unavailable [code]` and a safe reason, without
raw key/config values. Use `docker logs nodeherder-backend` and `docker inspect` to check mounts.
Missing keys, corrupt/locked storage and unwritable directories disable only integrations; existing
device features continue running. Restore the key backup rather than replacing encrypted data.

Non-Compose installs use `INTEGRATION_KEYRING_FILE`, `INTEGRATION_KEYRING_DIR`,
`INTEGRATIONS_CONFIG_DIR` and `DATA_DIR` directly. With no key override, native startup retains the
OS config path. Config/data defaults are `./configs/integrations` and `./data`; Docker explicitly
selects `/run/nodeherder-secrets/managed`. Use persistent writable config/data/key paths.
Cookie mutations still require the exact configured `FRONTEND_ALLOWED_ORIGINS` origin and the
feature request header; no authentication behavior changes in this startup fix.

## Use and repair

Open Integrations in the sidebar. Add an integration, enter a postcode, select an explicit address and
save. A provider token is optional for the supplied personal-use flow; provider access/quotas may
change. Supply your own token if required. No live lookup is created by installation or page reads.
Unsupported address-input branches and approximate/neighbour results are rejected clearly.

Configure operations offers Form, YAML and Help. URL/endpoint/JSON field-location repairs within
the supported GET/POST/search-select/waste primitives need no rebuild. Other authentication,
transports or capabilities require application support. Map collections once, not per bin colour.
Request names are arbitrary identifiers linked by the setup/refresh sections.

Save changes writes the host file and applies it; it does not commit or push Git. Review fixed
constants/comments for private values before committing a definition. Tokens and selected addresses
belong in Setup, never in shareable YAML. Form edits can normalize YAML formatting. Reload reads
host changes, Restore previous creates a new revision, and stale edits return conflicts. Avoid
simultaneous host-editor/Git checkout and UI writes. Upgrades never overwrite an existing host file.
A new provider for the supported recipe can be copied through the UI or added as a matching host
YAML file and loaded with Reload providers in Integrations (or on restart). Invalid files retain
the last active revision and show a repair error.

Destination changes require explicit review outside YAML before saved inputs/tokens can be sent to
the new origin. This is shared configuration, not an admin role. Only public HTTPS on port 443 is
supported; redirects, local/private/metadata destinations, proxy environment and executable templates
are blocked. Tests/previews do not overwrite saved results. Last checked and source verification
are distinct; an unchanged response does not make stale provider data fresh.

## Back up, restore and rotate

For a simple consistent backup, stop the backend using your normal deployment procedure, then copy
the config tree and `integrations.db`. Back up the matching keyring **separately**. For the default
Docker key, the existing stopped container retains its volume mounts and supports `docker cp`:

```sh
install -d -m 700 "$HOME/nodeherder-key-backup"
docker cp nodeherder-backend:/run/nodeherder-secrets/managed/integration-keyring.json \
  "$HOME/nodeherder-key-backup/integration-keyring.json"
chmod 600 "$HOME/nodeherder-key-backup/integration-keyring.json"
```

Use a separately protected backup location outside the checkout. For external custody, back up its
host key file instead. Protect the actual key bytes, not only the Docker volume name.
Do not commit the DB or keyring. Start the backend after the copy. The repository also provides a
transactional encrypted `Backup(io.Writer)` for future administrative tooling; there is no public
backup-download API. Test a restore into isolated directories before relying on a backup.

Restore matching YAML, DB and keyring with their permissions. Restore the original named key
volume using your Docker backup tooling, or use the optional external-key overlay with the
exported matching keyring. A new empty volume cannot decrypt an existing database. Startup retains
last-good compiled revisions when a host file is invalid; use Configure operations to compare and Reload. Unknown DB
schemas fail closed. This is the initial schema; there is no older integrations schema to migrate.

To rotate an externally managed key, first take/verify the backup, then run the provisioning
script with the same host path and `--rotate`. Managed-volume users can export their matching
keyring and switch to the external-key option before using this explicit rotation procedure.
The script adds a fresh active key and retains all older keys. Restart the backend to load the
new keyring. New writes use the new key; the local janitor re-encrypts up to 32 older records per
minute without provider calls. Interrupted rotation is safe with the mixed keyring. **Keep old keys
for every retained backup**; this first release deliberately provides no key-pruning command.
Removing a required old key fails closed and requires restoring the matching mixed keyring.

Setup candidates expire after 30 idle minutes (24-hour absolute ceiling). Pending attempts expire
within 24 hours. Completed receipts keep only minimal completion identity for 24 hours; expired
setup tombstones explain expiry for 24 hours without retaining addresses/tokens. Cancel removes
unfinished private setup/attempt records. One previous good provider definition is retained for up
to seven days; pending operations pin their original definition until terminal/expiry.

## Verification boundaries

Automated tests use temporary encrypted databases and synthetic HTTP providers. Browser fixtures
must prove their identity before mutations; no live council/WhenIsBins POST or household device
command is part of routine verification. Real provider availability, account limits and deployment
mount permissions remain operational checks on your installation, not claims made by fake tests.

Deletion removes live records. Encrypted bytes may remain in bbolt free pages and retained backups;
this is logical deletion, not a secure-erasure guarantee. Apply your backup-retention policy and
protect all associated keys accordingly.
