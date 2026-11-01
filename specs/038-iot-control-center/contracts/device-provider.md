# Draft contract: physical device providers

Status: design proposal, not an implemented wire contract | See [ADR 001](../adr-001-device-boundaries.md)

## Ownership and lifecycle

The application owns one statically registered provider instance per protocol and one session per configured device, independent of browser count. The provider owns its connections and emits discovered descriptors, ordered observations, connection state and classified failures to a core-owned sink. The existing registry owns device identity and confirmed values; the event pipeline owns automation ordering and publication. Provider-specific configuration and credentials stay outside the public device aggregate.

The lifecycle is `disabled → connecting → online → reconnecting/offline → stopping → stopped`. A device becomes online only after an authenticated session and inventory/subscription initialization; on disconnect all live values are unavailable, retaining last known values only as explicitly stale. The manager owns reconnect with capped exponential backoff plus bounded jitter. Repeated start/stop are idempotent; cancellation closes sockets and joins work. The implementation plan must specify exact time, count and memory limits, including how lanes recover after a full queue.

An ESPHome session accepts at most 256 known entities and has a 512-message ingress
queue; discovery has a 10-second deadline. The protocol reader never blocks on a full
queue: it signals desynchronization and the session closes. The future supervisor must
mark the device unavailable and reconnect with a fresh inventory/subscription. The sink
owns delivery into the existing device pipeline and must return promptly on cancellation;
session callbacks do not own confirmed state. A single session attempt does not retry
commands or reconnect itself; those belong to the supervisor and command router.
The supervisor allows one active attempt per configured device, marks it offline as soon as
an attempt ends, and retries after 1, 2, 4, 8, 16, 32 and then 60 seconds with positive
jitter capped at 60 seconds. A connection must remain ready for 60 seconds to reset the
backoff. Cancellation interrupts the wait and ends the current attempt; no command is replayed.
Each session accepts at most 32 pending commands. It rejects commands before readiness
or after disconnection, checks cancellation before sending, validates against that
session's discovered descriptor, and sends once. Cancellation after transmission begins
is ambiguous and must never trigger automatic replay. Command submission never updates
confirmed state; the next report remains the only confirmation.

## Identity and discovery

The installation assigns a stable opaque device ID on first enrollment. Provider identity and the provider's stable physical identifier form a unique binding, separate from display name or IP. ESPHome entity keys are scoped to that binding; type and native metadata are recorded with a generation so removed or changed entities become unavailable rather than silently rebinding user automations. Reconnecting or changing an address must not create a second device. Discovery offers untrusted candidates, never connects or enrolls without an explicit signed-in action.

The first successful ESPHome DeviceInfo response binds its normalized unicast MAC address
to the encrypted enrollment before inventory or commands are enabled. Missing, malformed,
zero and multicast addresses are rejected; a duplicate binding or a changed address on
reconnect is rejected. A MAC address is an identifier, not proof of identity: the Native
API encryption key is required to authenticate an encrypted connection.
If the user did not supply a name, a bounded and valid reported friendly name (or
device name) may be used for display; it is never used as the physical identity.
Discovered entity identifiers are also bound to their provider-reported kind in the
encrypted enrollment, including retired identifiers. A previously used identifier
cannot silently acquire a different kind after a firmware update or temporary removal.
The binding history is bounded to 1024 identifiers per enrollment.

Each entity descriptor carries an opaque ID, display name, kind, category, value type, read/write/event permissions, unit and optional numeric range/step or finite options. Unknown kinds are read-only safe fallbacks. Bound metadata, validate UTF-8 and finite numeric values, and prevent device-provided strings from becoming HTML, routes, log format strings or commands. Exact compatibility mapping into current Zigbee-oriented `Entity` must be proven before freezing this contract.

The ESPHome adapter's first translation boundary combines the native subdevice ID and
32-bit entity key into a device-scoped identifier of two eight-digit lower-case hexadecimal
values separated by a hyphen; the main device uses subdevice ID zero.
It retains the native kind separately, validates human-readable strings as UTF-8 with
a 256-byte maximum, bounds selects to 128 options, and rejects non-finite or inverted
numeric ranges and non-positive steps. Missing state remains unavailable; it must not
be converted to zero. These descriptors are intermediate provider data and are not a
second state store; conversion to the existing entity model and duplicate/key-generation
handling remain required before integration.

The existing store uses friendly-name hashes for non-Zigbee devices. Provider devices
instead persist under their enrollment UUID, retaining a friendly-name-to-UUID mapping
for existing lookups and metrics. A write must reject conflicts with a different device
or provider. Bridge-map refresh must rebuild provider mappings from persisted device
records as well as Zigbee metadata; it must not change legacy ID resolution.
Inventory is installed as a new immutable device generation on its existing per-device
lane only after subscription succeeds. Reconnection retires the previous lifetime before
publishing a replacement with the same ID; no exposed map is mutated under readers.
Reports and offline transitions use the same lane. A disconnect clears pending sampled
output and broadcasts offline; retained entity values are stale while the device is offline.
ESPHome entity keys remain the expose names used for state and commands; their discovered
display labels are stored in the generic `attributes.label` field. Numeric sensors and
numbers map to numeric exposes, binary sensors and switches to binary, selects to enum,
and text sensors and buttons to read-only text and write-only button kinds respectively.
Number bounds/step and select options come only from discovered metadata. Light power is
supported; brightness, colour and effects still require capability-specific mapping.
Other identifiable discovery messages appear as read-only unsupported entities. Unavailable
observations pass `null` into the existing confirmed-state pipeline rather than measured zero.

State messages must match the discovered kind, subdevice ID and key. Missing states and
non-finite numeric states are unavailable, with a nil value; valid text states are bounded
to 256 UTF-8 bytes. Commands require a writable descriptor and a matching JSON value:
switch boolean, number finite number within discovered range and step, select an exact
advertised option, or button `null`. Encoding a command never changes confirmed state.
Non-password native text entities use bounded UTF-8 text and their advertised length and
RE2-compatible pattern; unsupported patterns are rejected rather than ignored. Password-mode
text is not exposed until a write-only, secret-safe flow is implemented.

## State and commands

Spatial mappings and display options belong to the existing per-device settings
record, not browser-local storage. They contain at most three explicitly selected
entity triplets, a bounded display scale and a right-angle orientation. Mappings
do not change the underlying entity values or automation behavior.
Repeated mapped coordinate reports also produce presentation-only freshness updates, at most
four times per entity per second. Ordinary changed-value updates follow their existing policy. These updates use the existing
authenticated device WebSocket and do not run automations or write metrics or storage.

An observation identifies the device and entity and contains a typed value or explicit unavailable state plus receive time. Apply accepted observations in device order before any sampling/debounce; do not collapse equal events. Provider transport failure cannot be represented as a measured zero. A lost queue item must cause an observable desynchronization and a full inventory/state resubscription rather than silent success. A provider may coalesce only state presentations after the automation boundary, never physical events.

One command dispatch accepts authenticated user or trusted automation context, device ID, entity ID and a typed requested value. It validates current capability and constraints and routes by the persisted provider binding. Offline, stale-generation, unauthorized and unsupported requests fail without sending. Transport write success means submitted, not confirmed: only a later observation confirms state. Do not replay an ambiguous button or other physical command after reconnect. Existing Zigbee batch/single automation semantics must remain independently regression tested.

The shared registrar command boundary checks the current device and writable entity,
rejects offline/disabled devices, then delegates provider-specific value validation to
the active session. WebSocket and automation requests use this same boundary; legacy
MQTT commands retain their existing wire behavior. Provider automation commands are sent
in deterministic entity-ID order and do not optimistically alter automation context.

## Privacy and persistence

Provider configuration and optional encryption secrets are encrypted in a provider-owned repository independent of optional HTTP integrations. API writes accept a key, but all reads return only configured/unconfigured status; logs and metric labels contain neither keys nor raw provider payloads. New devices are excluded from the existing public `/api/context/devices` and unauthenticated MCP projections and query lookups by default. Authenticated device UI uses existing session-protected routes and origin checks. Test all public consumers, including MCP prompt and metrics lookup, rather than filtering only list responses.

Preserve current device records and Zigbee commands without migration. The additive provider binding and encrypted records require a versioned schema, backup/restore and downgrade behavior before implementation. Deleting an enrollment must stop its session, fence stale events and remove key material; specify whether historical metrics remain and how they are recovered.

The existing device JSON gains an optional `provider` string. Empty or absent means a
legacy device and preserves existing public behavior; every device created by a new
provider must persist a non-empty provider name before entering the shared registry.
The public context and MCP projections omit all devices with a provider name, including
ID-based metrics access. Authenticated hub and metrics consumers retain the full store.
The field must survive persistence/restart, and malformed provider onboarding cannot
create an unmarked device. Public APIs continue to use their existing wire shapes.

### Enrollment repository schema and recovery

Use a separate bbolt database at `DATA_DIR/device-providers/integrations.db` and a separate
device-provider keyring. Reuse the existing encrypted-record repository machinery directly,
not the optional HTTP integration service: its schema 1, installation ID, per-record nonce,
authenticated encryption and identity-bound AAD remain unchanged. The `instances` bucket
contains up to 32 version-1 enrollment records, keyed by installation-generated UUID. Each
record holds provider name, numeric address, TCP port and provider credential; no state or
entity observations are stored there. Enrollment writes and deletes commit atomically.
On the first successful native API device-info response, the provider binds the reported
unicast MAC address to the enrollment inside the encrypted record. A different address on
reconnect or a MAC already bound to another enrollment is rejected before inventory or
commands are enabled. This prevents accidental identity merging, but a MAC on plaintext
connections is not cryptographic authentication. Noise must be configured when peer
authentication is needed. Missing/invalid MAC addresses fail closed with a diagnostic.
An optional 256-byte UTF-8 display name is stored with the enrollment; an empty name uses
an automatically generated non-sensitive display label. Each provider driver validates
its own configuration and creates one instance per enrollment. The manager saves new
enrollments before starting them, rejects duplicates and owns cancellation and worker joins.
It exposes redacted enrollment snapshots; failed startup of provider storage does not prevent
existing Zigbee startup. Configuration mutation and shutdown serialize with one another.
The service validates addresses, credentials and duplicate physical bindings before writes;
the repository rejects duplicate endpoint bindings and corrupted records. Deleting a record
does not secure-erase old bbolt pages or backups; historical metrics remain under their existing
retention policy, so deletion UI must state that scope. The service stops and fences the session
before deleting its record; failures leave the enrollment intact.

Back up the encrypted database and its matching keyring together using the repository's
consistent snapshot API. Retain old keys while old records or backups require them; rotate in
bounded batches and verify restored backups before removing any key. A missing key for an
existing database fails closed; never generate a replacement key. Unsupported schema fails
closed. On rollback, stop the provider first; older versions ignore this separate database.
Restoring a database without its matching key cannot recover credentials. Key provisioning
uses `DEVICE_PROVIDER_KEYRING_FILE` for an existing externally managed file, or
`DEVICE_PROVIDER_KEYRING_DIR` for a managed absolute directory and the filename
`device-provider-keyring.json`. With neither set, the managed location is the NodeHerder
subdirectory of the operating system's user configuration directory. Managed files are created
only on a first install with no provider database; Docker must provide a writable persistent
private key directory. Startup wiring and Docker volume documentation remain required before
enabling the provider.

### ESPHome onboarding HTTP contract

`GET /api/esphome/devices` and `POST /api/esphome/devices` are authenticated and
return `{ "data": ..., "error": ... }` with `Cache-Control: private, no-store`.
GET returns redacted enrollment records (`id`, `provider`, optional `name`, numeric
`address`, `port`); credentials never appear. POST accepts one JSON object containing
`name` (optional), `address`, `port`, and `encryption_key` (optional). It assigns the ID,
validates and persists configuration before starting a background connection; 201 means
enrolled, not online. Invalid input returns 400 without echoing secrets, unauthenticated
requests 401, CSRF failures 403, unavailable storage 503. Cookie mutations require the
existing allowed-origin policy and `X-NodeHerder-Request: 1`. Requests are limited to
4 KiB and unknown fields or extra JSON values are rejected. The endpoint never probes
the network before enrollment; availability and discovery arrive via the device stream.
`POST /api/esphome/discover` performs a bounded optional mDNS scan under the same
authentication and cookie-origin policy. It returns at most 64 validated numeric private
address/port candidates; the user must explicitly select and enroll one. It does not
connect or persist discovery results. Docker networks may not receive local multicast.

Browser preflight requests from configured frontend origins must allow the required
`X-NodeHerder-Request` header for ESPHome discovery and enrollment routes. Disallowed
origins and unrelated routes retain their existing CORS policy; this does not replace
authentication or the mutation origin check.

`DELETE /api/esphome/devices/:id` requires the same authentication and mutation origin
checks. It stops and joins the session, removes the device from the shared registry,
then removes the encrypted enrollment and broadcasts the refreshed device list. It leaves
historical metrics and existing automations intact. If removal fails, the saved enrollment
remains and its session is restarted. A timeout while stopping leaves it stopped until a
subsequent retry or process restart; it must not allow an overlapping session.
`PUT /api/esphome/devices/:id` replaces name, address and port. An omitted encryption_key
retains the existing secret; an explicit empty string selects plaintext, and a non-empty
value replaces it. The server validates the replacement before stopping the old session,
then joins it before committing and starting the new one. A failed write restarts the old
configuration; a stop timeout leaves the saved configuration intact for retry or restart.

### Network target and key input

The onboarding address is a numeric IP literal without a zone or a port; a separate integer port
must be in 1–65535. Normalize IPv4-mapped IPv6 before policy checks. Accept only RFC 1918 IPv4
and RFC 4193 IPv6 unique-local unicast addresses; reject hostnames, public, loopback, multicast,
unspecified and link-local addresses. Dial the validated literal directly, without DNS resolution.
Containers must have a route to the selected address; host networking is a deployment choice and
does not relax this policy. Discovery results are candidates subject to exactly the same checks.
The optional encryption key is empty for plaintext, or standard padded Base64 encoding of exactly
32 bytes for Noise; a malformed key fails validation without echoing it. Selecting encryption must
never fall back to plaintext. The transport must accept an injected dialer for isolated loopback
tests rather than making loopback valid in production.

## Spatial data

Use an explicitly configured or verified capability mapping from entity IDs to target identity, X/Y, optional distance and an independent validity signal. Units must be known and convertible to millimetres. Each coordinate has its own observation timestamp; do not claim a paired frame when the protocol sends separate entity updates. A target with invalid, stale or missing required inputs is hidden or explicitly unknown; an offline sensor never presents targets as live. The graphical plane has a textual equivalent. Firmware 1.5.0 exposes optional active flags that may be disabled by its tracking setting, so false alone is not reliable proof of no target without considering that setting.

The first spatial view requires the user to select each X, Y and active entity explicitly;
no device names or assumed entity IDs are used. Coordinates are accepted only for mm, cm
or m units and converted to mm. Each axis must have a finite, separately received UI update
within five seconds and the selected active flag must be true. Initial snapshots have no
freshness evidence and are shown as awaiting updates. Device or WebSocket disconnection
suppresses targets. Because existing live reporting can suppress unchanged readings, a
stationary target may become unknown; the UI must say it is awaiting fresh updates rather
than asserting that the target left. Do not infer a simultaneous frame. Scale and orientation
are local view controls; the mapping is not persisted in browser storage.

### Review additions: availability and descriptive metadata

Provider session status owns device availability; report age is only reading freshness and must not disable a healthy quiet device. The session owns and joins a separate keepalive waiter so state delivery stays ordered and responsive.

Device JSON adds optional `properties` with bounded public strings: `manufacturer`, `model`, `firmware_version`, `project_name`, `project_version`. ESPHome DeviceInfo produces these fields; the existing authenticated device stream/snapshot and About view consume them. Missing/invalid fields are omitted, and old records need no migration. Older readers ignore these additive fields; no credentials or network addresses are included. Entity attributes optionally add `device_class` and `entity_kind` for semantic presentation without changing stable IDs or command keys.

Button automation values are exactly JSON null. Non-null values are rejected at configuration and command encoding. Browser stateless submissions use a one-second duplicate guard, canceled on disconnect/reset, with requested feedback that never claims execution. Polling errors retain last-loaded rows; operation errors survive successful polls. Closing management retires UI completions, without retrying or rolling back an already submitted server mutation.
