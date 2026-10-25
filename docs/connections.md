# Connections: personal Docker setup

Connections adds external information independently of Devices. The first release supports UK
bin collections through a supplied WhenIsBins definition. Updates are manual: Refresh retrieves a
schedule, and Check status continues a pending lookup. No provider work runs on a recurring timer.
All signed-in users have the same controls. Nothing is added to Home, Overview, Devices or Panel.

## Enable private persistence

Existing backend Compose mounts already persist `./backend/configs` and `./backend/data`. Provider
YAML lives in `backend/configs/integrations/waste_collection/whenisbins.yaml`; private state lives
in `backend/data/integrations.db` (0600). No temporary-storage fallback is permitted.

Create a private directory **outside the checkout**, owned by the user running the container, and
provision a keyring explicitly. For example, after creating `/srv/nodeherder-secrets` with mode 0700:

```sh
node scripts/create-integration-keyring.mjs /srv/nodeherder-secrets/integration-keyring.json
```

Back up that file separately. Set `NODEHERDER_SECRETS_DIR=/srv/nodeherder-secrets` in the root local
`.env`. With the stack manager use `scripts/manage-stack.sh start all --connections`, and keep
`--connections` on later backend/all build/rebuild commands. For direct Compose commands, add
`-f docker-compose.connections.yml` alongside your usual backend/frontend files. The overlay mounts the external secrets directory read-only;
it never generates a new key inside a container. Keep existing MQTT/Zigbee deployment arguments.
Do not run a second backend against the same data directory: one process owns the bbolt database.

Without a readable keyring, Connections reports unavailable inside its own page. The base stack
still runs device features; the optional overlay is not required for those. A locked/corrupt/newer
DB or unwritable mount also degrades this feature rather than replacing private data. Check keyring
mode 0600, directory permissions and the selected environment paths before retrying/restarting.
Never fix lost encryption keys by generating an unrelated keyring: restore the matching backup.

Non-Compose installs use `INTEGRATION_KEYRING_FILE`, `INTEGRATIONS_CONFIG_DIR` and `DATA_DIR` directly.
Defaults are `/run/nodeherder-secrets/integration-keyring.json`, `./configs/integrations`, `./data`.
Use a persistent writable config/data path. Configure `FRONTEND_ALLOWED_ORIGINS` with the exact
browser origin, including scheme/port; cookie mutations require that Origin and the feature header.

## Use and repair

Open Connections in the sidebar. Add a connection, enter a postcode, select an explicit address and
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
YAML file and loaded with Reload providers in Connections (or on restart). Invalid files retain the last active revision and show a repair error.

Destination changes require explicit review outside YAML before saved inputs/tokens can be sent to
the new origin. This is shared configuration, not an admin role. Only public HTTPS on port 443 is
supported; redirects, local/private/metadata destinations, proxy environment and executable templates
are blocked. Tests/previews do not overwrite saved results. Last checked and source verification
are distinct; an unchanged response does not make stale provider data fresh.

## Back up, restore and rotate

For a simple consistent backup, stop the backend using your normal deployment procedure, then copy
the config tree and `integrations.db`. Copy the external keyring to a **separately protected backup**.
Do not commit the DB or keyring. Start the backend after the copy. The repository also provides a
transactional encrypted `Backup(io.Writer)` for future administrative tooling; there is no public
backup-download API. Test a restore into isolated directories before relying on a backup.

Restore matching YAML, DB and keyring with their permissions. Startup retains last-good compiled
revisions when a host file is invalid; use Configure operations to compare and Reload. Unknown DB
schemas fail closed. This is the initial schema; there is no older Connections schema to migrate.

To rotate, first take/verify the backup, then run the provisioning script with the same path and
`--rotate`. It adds a fresh active key and retains all older keys. Restart the backend to load the
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
