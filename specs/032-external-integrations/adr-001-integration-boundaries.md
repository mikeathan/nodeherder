# ADR-001: Shared UI editing of host YAML integration definitions

Status: Proposed technical design; user access/editability decisions recorded | Date: 2026-10-24
Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Owner/review: maintainer technical review pending
Rules: NH-03/04/06, BE-01/03/04/05, FE-01/02/04/05

## Context and options

The user explicitly wants all signed-in users to access/edit/modify integrations, simple desktop/
mobile use, and repairable API/response configuration through the UI. They favour host-saved
configurations similar to YAML that can be committed and extended. A separate feature permission
hierarchy is rejected. Existing application authentication remains the trust boundary.

Options considered:

1. Code-only adapters: typed behaviour, but common endpoint/mapping repairs require releases.
2. DB-only editable profiles: UI repair is possible, but configuration is less transparent on the
   host and cannot be versioned as ordinary human-readable files without an export step.
3. Host YAML definitions + Form/YAML/Help UI + bounded generic operations: readable/committable definitions,
   common repairs without release, normal users need no endpoint details, consistent edit source.
4. Universal workflow/script runtime: wider flexibility but unnecessary language/security scope.
5. Existing HA/Node-RED bridge: useful if already installed, adds separate deployment dependency.

Device projection remains unsuitable for these resource semantics/public exposure. One-household
shared authenticated access matches the request; private multiple-household tenancy is not implied.

## Decision and consequences

Recommend option 3. [Configuration contract](contracts/configuration.md) fixes host paths, shared
controls, form/raw YAML parity, mobile behaviour, strict parsing, atomic file write, content hashes,
last-good compiled activation, reload/recovery and manual Git workflow. No automatic commit/push.
Default YAML persists through the existing configs volume; never silently fall back to temporary
storage or overwrite a user's definition during an app update.

One YAML document is the editable definition source of truth. One bbolt runtime repository stores
private instances, cache/health, encrypted credentials/bindings and compiled definition revisions.
The filesystem and DB are distinct persistence boundaries: saved_not_applied is visible and a
journal/reload reconciles failures. Typed waste resources retain local UUID identities; active/
pending operations pin compiled revisions. YAML explicitly wires arbitrary request IDs into a reusable
search/select recipe, and a shared waste validator handles output meaning. No WhenIsBins-named Go
adapter or hard-coded request IDs. New auth/transports/operation types still need reviewed code.

All authenticated users can configure/edit/test/reload/restore/remove and continue shared setups.
No extra role/feature allowlist or creator-only checks. Audit actors do not grant privileges.
Revision conflicts prevent silent lost edits. Retain actual-request cookie Origin/header checks,
bounded egress, safe host paths and scoped credentials. A changed credential destination requires
a visible reconnect action available to that same user, never a separate privileged account.
No new public integration/MCP exposure or integration-data/settings-reference WS broadcasts.

Use Connections beside Devices as the single feature entry. Discovery, setup, results, provider
editing and Help live within it. The [navigation/page contract](contracts/connections-navigation.md)
keeps Home, Devices, Overview and Panel content/layout/counts unchanged: no pins, room tiles,
cross-promotion links/notices, dashboard schema extension or extra global footer copy. One
installation is one shared home. The earlier dashboard/panel experiments are withdrawn; future
presentation requires separate scope. Protected resource APIs own connection data independently
of device and dashboard snapshots.

Keep secrets/household selections/runtime history out of committable provider YAML. Back up definitions,
DB and separate keyring with consistent hashes. [Execution/security](contracts/execution-and-security.md)
defines encrypted private-state retention and recovery. [Waste contract](contracts/waste_schedule.v1.md)
defines wire identity, coverage, quality, null/empty and explicit later replacement.

Manual M1 is user-confirmed; automatic M2 is later roadmap work; no implementation authorization is inferred.
The same bounded execution service supports both. Read-time projections need no extra provider
requests. Later coordinator owns recurring provider/projection work; janitor only owns local cleanup.

Cost: role/allowlist work is removed, while strict YAML editing and filesystem/DB recovery add work.
Re-estimate after provider fixtures and file/editor contracts are validated. Use an existing/stable
YAML parser where suitable; no unreleased parser, universal scripting engine or new UI framework.

## Validation and reconsideration

Require two-user equal-access/shared-edit tests, anonymous/CSRF/public-route negatives, strict YAML
and Form/YAML parity, disk/hash/rename/activation failure recovery, host reload, no automatic Git
actions, responsive keyboard/touch acceptance, credential-origin isolation, mixed-key restore,
and atomic private-state failures. The [help/schema contract](contracts/operations-and-help.md)
adds bundled configuration help and a small typed definition model; cross-reference validation and
request renaming must be tested. The user asked to simplify schema complexity: no runtime JSON
Schema engine or schema-generated UI is selected; ordinary typed decoding and focused checks suffice. All application checks remain unrun.

Reconsider only if requirements change to private tenants, arbitrary executable providers or new
transports. Preserve stable capability identities and user files. This proposed ADR supersedes the
earlier draft access/storage proposals in this same planning set; no earlier accepted decision,
constitutional ratification or technical approval is invented.


The [audit](review.md) adds concrete [runtime ownership](contracts/runtime-lifecycle.md) and
[API/recovery](contracts/api-and-recovery.md): optional degraded feature startup, joined cleanup,
strict storage, scoped validators, local Save receipts, credential/precondition/review paths and
declarative renewal. All snapshots are encrypted because provider notes may contain PII. M1 is
exact-choice only; representative branches need a later supported recipe. These refine the proposal,
not approval or implementation evidence. User decisions recorded 2026-10-25: manual first release and personal Docker use. Runtime/key
mount/restore verification remains implementation work; this is not approval to start production edits.

## Implementation evidence, 2026-10-25

Option 3 is implemented for M1; [implementation-review.md](implementation-review.md) records
agent self-review and isolated checks. This does not claim maintainer adoption, ratification or
approval of an exception. Later automatic scheduling/consumers/provider replacement are deferred.

## Native first-run correction (AC-31)

The user clarified the current run is native and requested removal of the manual key setup barrier.
The default now creates a private OS-user-config key only before any Connections DB exists.
Explicit paths (including Compose) still require provisioning. Key loss/corruption never triggers
replacement. Encryption/schema/rotation remain unchanged; no new dependency or JWT reuse.
