# Integration execution and security contract

Status: M1 implementation contract; later-milestone provisions remain deferred. See [implementation review](../implementation-review.md).
Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## Delivery and repair boundary

User-confirmed M1 is manual setup, refresh and Check status for personal Docker use, including repairs to a supported provider
through a shared authenticated UI. M2 introduces automatic provider refresh and continuation. The same
execution service handles both; M1 does not silently start recurring provider jobs. A bounded
request worker or expired-session janitor is not a recurring provider scheduler.

Endpoint repair does not require installing new provider code. M1 uses generic operations and a
bounded search_select recipe, with arbitrary request IDs wired in YAML. The
[operations/help contract](operations-and-help.md) and [small typed format](definition-format.md)
define the grammar. The shared waste capability validator, not a WhenIsBins Go adapter, owns
output semantics. Unsupported new operation/transport/auth mechanics still need reviewed code.

### Definition v1: exact permitted changes

Users may change HTTPS destinations, request paths/methods/inputs, response selectors/captures,
choice presentation, request references in the supported setup/refresh recipe, and typed field
mappings. Credentials/destinations retain the explicit reconnection and egress rules below.
M1 supports GET/POST and one initial recipe, not arbitrary branches/scripts/graph execution.

Reference roots are input and state. Input names are declared in inputs; named captures declare
ordinary state fields; transport validators are reserved and request/revision scoped, never a
single instance-wide ETag. The compiler rejects unknown
or uninitialized references, not just malformed syntax. Secret slots are isolated from these roots.
Raw response captures/status pointers are independent of the selected result mapping subtree.

Use the small set of capture types and prefix removal in the typed format. No general coercion,
expression or transform framework is needed in M1. The common capability code validates/sorts dates
and normalizes known categories; unknown labels retain their original value. Collection mapping
traverses one bounded list and applies the same fields to every bin. Optional absent captures never
erase existing state; complete-only captures wait for terminal success and valid output.

The execution response envelope is `{status_code, headers, body}`. Header names are normalized
case-insensitively; `ETag` and `Retry-After` are explicit sources, not JSON Pointer expressions.
The runtime handles 304 before body parsing, parses Retry-After as seconds or HTTP date, and
preserves status/header/body separation. Auth is `none`, `bearer` or `optional_bearer` in M1;
the latter injects a Bearer token only when its declared private slot is configured. Credential
placement, Host, Cookie, Authorization, proxy, TLS and hop-by-hop headers are not free-form edits.

For example, moving collections from `/collections` to `/data/collections` or renaming a request
parameter is repairable without recompilation. A new auth protocol, new transform, new setup
branch, non-JSON format, transport, semantic output or wholly new provider adapter may still need
a release. Request ID renaming and supported recipe wiring do not require a release. This is
the explicit limit of the M1 repair promise.

### YAML definition saving and activation

The [configuration contract](configuration.md) is authoritative for the host YAML source of truth,
Form/YAML editors, disk hashes, atomic writes, DB activation/journal recovery and Git/backup handling.
All signed-in users can save/test/reload/restore definitions. There is one file per definition; the
DB holds immutable compiled runtime revisions, not a competing editable configuration store.

Validation checks known inputs/outputs, pointers, types, resource bounds and outbound destinations.
Offline sample-response validation and an explicit live Test action use redacted previews. Tests
creating provider work use the same idempotency machinery as setup. Pending operations pin their
original compiled revision; saves affect new work. A failed validation changes neither file nor
active state. A saved-but-not-applied failure is visible and recoverable, never reported as success.

A change of scheme/host/port is highlighted in the normal Save review. Every signed-in user can
accept that change; no separate privileged account is required. Stored credentials are not silently
forwarded to a new origin: the same user can explicitly reconnect/reassociate them. Grants remain
separate from raw mapping data so an imported/pasted definition cannot authorize its own secret
access. This is a data-flow safeguard, not a role-based restriction. No secret values or private
setup inputs are written to provider YAML or definition downloads.

Proposed hard bounds: 64 KiB definition, up to eight named requests, 64 extraction fields, 256-byte pointers,
1,000 collection/service elements and 366 dates per service, additionally subject to 2 MiB decoded
response, request deadlines and global/provider admission bounds. Unsupported definitions fail
closed. These are proposed limits to validate against captured provider fixtures before coding.

## Access: existing login, equal controls for every user

Use NodeHerder's existing authenticated session/bearer middleware. Every successfully authenticated
principal has the same integration permissions in this shared-household deployment. There is no
additional role, feature allowlist, per-user setup privacy or privilege migration. Existing signed-in
users can use the feature immediately when enabled; unauthenticated requests remain rejected.

| Operation | Access |
| --- | --- |
| Catalog, instance/resource/health/configuration reads | Any authenticated user |
| Add/configure/edit/test/refresh/Check status/reload/restore/remove | Any authenticated user |
| Continue/edit/cancel a setup started by another household user | Any authenticated user, subject to current session/definition revision |
| Change destination and reconnect/replace credentials | Any authenticated user; destination/credential change remains explicit |
| Read raw secret values or use public device/MCP routes for these resources | No secret-read API and no new public integration access |

Use verified Claims.UserID only for created_by/updated_by audit where needed. Ignore browser-supplied
ownership/role assertions. The household installation ID is local, not a client-selected tenant.
All configuration sessions are shared; If-Match/revision checks return conflicts instead of silently
replacing another user's edits. No creator-only 404 rule. Missing/invalid auth is 401; an unsafe
origin/destination can still be rejected for every user. Service-token calls follow the existing
middleware with the same feature access; no new service-principal allowlist is introduced.

M1 publishes no integration data or references on the global/settings WS hub; protected REST
polling is sufficient. The [navigation contract](connections-navigation.md) confines the UI to
Integrations and introduces no dashboard setting or device snapshot extension.
Future WS may use the same shared-household access with session validity checked at delivery.
Private multiple-household tenancy would require a separate design; equal authenticated access is
intentional here. No changes to anonymous/public existing routes are implied.

### Cookie mutation and origin policy

For integration POST/PUT/DELETE endpoints, if the session cookie is present, require both an exact
configured frontend Origin and `X-NodeHerder-Request: 1`. Missing/null/unapproved Origin is 403.
The frontend helper adds this header for these mutations. Do not treat Referer, user-supplied Host
or forwarded headers as an origin grant. If a cookie and bearer are both sent, cookie protection
still applies. Bearer-only non-browser requests may omit Origin; a supplied Origin must be approved.
All browser responses/PII-bearing reads use `Cache-Control: no-store`.

Integration CORS preflight allows this header plus If-Match/If-None-Match for approved origins, but authorization/origin checks also run
on the actual request: current CORS code suppresses response headers rather than rejecting the
underlying request. Use the router's existing PUT method for configuration updates, not PATCH.
No other route's CORS/authentication behaviour is implicitly broadened.

## Credential and private-state custody

Native startup with no explicit INTEGRATION_KEYRING_FILE uses the OS user-config directory's
nodeherder/integration-keyring.json. For a first installation only (no existing integrations.db),
it creates a private directory/key atomically and reuses it on restart. Existing DB + missing key,
invalid key, unsafe permissions or filesystem failures remain fail-closed. This local first-run
correction supersedes the previous universal manual-provisioning requirement (AC-31).

An explicitly configured INTEGRATION_KEYRING_FILE must already exist as a private regular file.
Docker Compose declares that explicit mounted path; it never provisions an ephemeral container key.
Format remains `{active_key_id: string, keys: {key_id: base64_32_byte_key}}`. Keys are never replaced
as a recovery shortcut, reused from JWT_SECRET_KEY, exposed through the browser or added to YAML.

Use AES-256-GCM with fresh random nonce per encryption. Store ciphertext version/key ID; authenticate
installation ID, record type, record ID and schema version as associated data. Encrypt credentials,
setup session contents, exact pending request payloads, private provider bindings and all snapshots.
Provider notes may contain private text even after matched_address is removed; encrypted storage
and protected/no-store APIs apply to them too.

Missing/invalid keyring prevents integration execution/private reads and yields a user-visible
`key_unavailable` condition; independent device management still starts. With an old key missing,
affected records fail closed and report `key_version_missing`. Never overwrite unreadable records.
Recover by restoring the keyring backup or explicitly deleting/reconfiguring the affected integration;
credential re-entry alone may not recover an encrypted property binding.

Rotation: atomically provision a file containing old and new keys with a new active key ID, restart,
then re-encrypt owned private records in bounded transactional batches. Readers accept retained key
versions; new writes use the active key. Resume unfinished batches after restart. Do not retire an
old key until active records and all retained backups requiring it have expired or been rewrapped.
Record counts/key IDs only. Test mixed-version databases, interruption and missing-old-key recovery.

Backup: include host YAML definitions, a consistent DB snapshot and separately protected keyring backup; document their
matching key versions. Never bundle keyring into exported profiles or browser diagnostics. Restore
all three into an isolated deployment before declaring the procedure tested. Proposed backup retention
is seven days; a user-managed longer retention must disclose that deleted data remains there.
Record deletion is logical: bbolt free pages and existing backups are not promised to be securely erased.

## Setup retention, consent and provisional answers

Setup expiry is 30 minutes idle and 24 hours absolute. Reads/polling do not renew idle expiry;
valid user submissions may. The attempt record stores immutable canonical JSON body bytes, a
separate idempotency key and credential reference; no Authorization/Cookie header or secret value.
Encryption and retention apply even when an anonymous provider requires no credential.

Retain exact submission body/key until its operation is terminal or the setup's absolute expiry,
then delete it. A saved instance's unresolved attempt also expires after 24 hours. Reuse after an
ambiguous response is allowed only inside a verified provider idempotency window; once uncertain
or expired, show reconciliation/re-setup rather than automatically submitting a new operation.
The fetched OpenAPI advertises 30 days or lookup removal, while local retention remains 24 hours;
this is conditional provider behavior, not a promise of exactly-once effects after remote data loss.

Completion removes candidates, full matched address and transient request history; retain only
the encrypted minimal binding needed for subsequent lookup recovery. Disable keeps the binding;
delete removes it and instance-scoped credentials. Cancel/expiry removes setup inputs, request/key,
transient consent and unreferenced credentials. Already submitted upstream work may continue if the
provider has no cancellation API; no new local polling or resubmission follows cancellation.
One context-owned janitor purges expired rows on startup and every minute; it issues no provider I/O.
Run summaries contain only sanitized outcome/timing/revision, at most 20 per instance/seven days.

M1 supports exact property-choice only, never submits representative opt-in and never displays
neighbour results. Unsupported journeys stop clearly. The following consent contract applies only
when a future supported recipe explicitly adds representative setup; it is not hidden M1 work.

Default is no opt-in to a representative property. If the provider offers `opt_in`, ask a separate,
unchecked consent question explaining neighbouring-property dates. Store `{accepted, accepted_at,
actor_id, wording_version, setup_revision}` encrypted in that setup only. A changed property or
new setup requires new consent and a distinct submission key. Send allow_postcode_representative
only after acceptance. For automatic representative responses, display the limitation explicitly;
do not falsely claim the user opted in. Never offer the option when the provider says unavailable.

| Result | Display | Persist as schedule | Reminders/automation |
| --- | --- | --- | --- |
| Exact non-provisional final | Normal, with source/coverage | Yes, validated | Later consumer applies freshness/date/subscription rules |
| Exact non-provisional partial | Label incomplete; continue same lookup | Only separately qualified exact snapshot, preserving incomplete coverage | Never infer missing dates; later consumer policy required |
| Provisional neighbour | Current authorized response only; prominent “neighbouring property, your address unconfirmed” | No durable snapshot/cache, no feed, no logs/browser persistence | Never |
| Final representative property | Label representative; M1 refuses to promote it to an exact household resource | Not accepted as configured exact resource in M1 | Never in M1 |

For provisional results persist only lookup identity/next permitted check time, not the neighbour
payload. Do not replace last-good exact data with provisional data. Expiry/failure must not promote
it. This follows the provider policy linked in the investigation; it does not forbid transient display.

## Local-date derivation

`capabilities/waste.go` owns a pure projection from validated snapshot + injected clock to next date,
days-until and freshness. Local REST reads recompute it using Europe/London. The M1 UI reads local
state at most every 60 seconds while visible and reloads on focus, so it crosses midnight without
provider refresh or a second backend scheduler. Snapshot revision changes only for stored content;
the response's evaluated_at describes time-dependent projections. UI polling never calls upstream.

When M2/M3 need scheduled notifications, the single coordinator owns the next local-midnight or
configured reminder deadline in the same work table, tagged `projection` versus `provider`.
Projection work performs no network call. Recalculate zoned deadlines after each run and restart;
do not use a fixed 24-hour timer across DST. No timer per output field and no device-lane injection.


The [runtime lifecycle](runtime-lifecycle.md) and [API/recovery contract](api-and-recovery.md) fix
composition/joins, strict DB paths, scoped validators, completion receipts, credential/review routes,
renewal/error classification and exact-only M1 setup. They resolve the remaining execution gaps.


## Personal Docker deployment (first release)

The user confirmed personal Docker use on 2026-10-25. Reuse the current backend mounts:
backend/configs → /backend/configs and backend/data → /backend/data. Set the integration config
root to /backend/configs/integrations and read DATA_DIR=/backend/data strictly; the owned DB is
/backend/data/integrations.db. Do not create a second deployment service or use temporary storage.

Add one read-only directory mount in the implementation's backend Compose configuration:

```yaml
environment:
  DATA_DIR: /backend/data
  INTEGRATIONS_CONFIG_DIR: /backend/configs/integrations
  INTEGRATION_KEYRING_FILE: /run/nodeherder-secrets/integration-keyring.json
volumes:
  - ${NODEHERDER_SECRETS_DIR:?Set an existing secrets directory outside the checkout}:/run/nodeherder-secrets:ro
```

This is the additional deployment contract, not a replacement for the existing config/data mounts.
NODEHERDER_SECRETS_DIR points to a real private host directory outside the checkout; its keyring is
0600 and readable by the backend container's runtime UID. Document a deliberate one-time host key
provisioning step using deployment tooling; never generate/replace keys silently during app startup,
put them in VITE variables/Git or require a new UI/admin role. Preflight verifies the host path/file
exists and container readability without printing key contents. No key is generated by this plan.

Mount the directory rather than a single key file so an atomic key-file replacement is visible
through the mount. The backend reloads the configured keyring on restart; rotation keeps old keys
until records/backups are migrated or expired. Back up YAML and a consistent runtime DB plus the
keyring separately outside the checkout; restore all three into an isolated Docker test before
production acceptance. A missing mount/key fails the Integrations feature closed and leaves device
management running. Tests use generated throwaway keys/containers and never the household stack.

Occasional manual use may use the provider's anonymous mode within its limits. The existing optional
private token slot supports a user-supplied free token for regular use; an invalid token is not dropped
to anonymous automatically. Do not bake a token into the image or YAML, request one by email on the
user's behalf, or infer commercial/distributed usage. Later automatic refresh is a separate phase.
