# Spec: Configurable external integrations

ID: 032 | Status: M1 implemented and verified locally; later milestones deferred | Owner: NodeHerder contributors | Updated: 2026-10-25
Constitution: proposed 1.0.0 | Components: backend and frontend

## Problem, scope, exclusions

Household users need to connect external services by answering ordinary questions, without
understanding HTTP, response mappings or provider identifiers. Bin collections are the first
example; weather and energy prices are realistic subsequent uses. The pre-feature baseline had
no general integration catalog, configuration flow, credential store or refresh lifecycle.

This spec records the requested outcomes, independently of the execution architecture. The
[investigation and implementation plan](plan.md) reviews the existing
[DeepSeek report](../../docs/external-integrations-investigation.md), compares alternatives,
and proposes a staged design. [Tasks](tasks.md), the
[wire contract](contracts/waste_schedule.v1.md), and the
[execution/security contract](contracts/execution-and-security.md), and
[host YAML configuration contract](contracts/configuration.md) make the remaining work concrete.
The [operations/help contract](contracts/operations-and-help.md) and small typed definition format
describe reusable operations, request references and the in-app guide.
The [Connections navigation and page-boundary contract](contracts/connections-navigation.md) defines
the new top-level category, future type structure and preservation of existing pages.
The [audit](review.md), [runtime ownership](contracts/runtime-lifecycle.md) and
[API/recovery contract](contracts/api-and-recovery.md) close the remaining execution paths.
The user authorized M1 implementation on 2026-10-25. This does not authorize later roadmap
phases, deployment, pushing commits, or changes to other views. Actual execution and limitations are recorded in [implementation review](implementation-review.md) and tasks.md.

Scope: integration discovery, dynamic setup, persistence, bounded external requests, credential
isolation, automatic refresh, health, stable read contracts and a path to agents/automations.
Confirmed M1 is a manual read-only bin collection integration with shared YAML definition editing,
for personal use on Docker. Automatic refresh remains a later M2 roadmap phase, outside the first release.
The user has now explicitly required equal signed-in access, simple desktop/mobile use, UI-editable
configuration and favours host-saved versionable YAML. Those preferences are recorded directly.
The overall roadmap includes automatic refresh; the user confirmed manual refresh for the first
release on 2026-10-25. No recurring provider work or scheduled continuation ships in M1. Provider requests may use POST to retrieve
data; this does not authorize external purchases, messages or device commands.

Exclusions: marketplace, arbitrary scripts, visual workflow editor, automatic API repair by an
LLM, silent provider switching, a replacement device/automation system, and packaging NodeHerder
as a Home Assistant app. The initial UI lives only within Connections: Home, Devices, Overview
and Panel content/layout are unchanged; no cross-page promotion, pins or room tiles. Push transports and device control are future capability-specific work.

## Scenarios and acceptance

### US-01 — Configure a useful integration (M1)

Independent verification: fixture-backed API tests and browser tests against a fake provider.

- AC-01: Selecting Bin Collections starts an authenticated setup flow; entering a postcode
  produces address choices when supported, or a clear unsupported-journey outcome when a different input recipe is required.
- AC-02: Only a candidate belonging to the current shared setup session can be selected.
  Changing the postcode invalidates previous candidates. No first address is silently selected. M1 accepts only the exact-property choice journey;
  unsupported required_input branches never become an empty list or guessed selection.
- AC-03: Saving an exact property produces a typed waste schedule, preserving source,
  confidence, coverage, subscription qualifications/caveats and local collection dates. Unknown/missing data is distinguishable
  from an explicitly empty schedule. Provisional neighbour data cannot drive automations.
- AC-04: A slow provider displays pending progress; Check status continues the same lookup after
  its retry deadline. Disconnect/restart preserves its saved identity until expiry. In M1, no
  recurring polling is promised; M2 automates continuation. Ambiguous submission recovery reuses
  the saved key and identical body only under the verified provider contract/local TTL. Repeating
  local Save after response loss returns the already claimed instance/operation, not a new lookup.

### US-02 — Refresh safely, then automatically (M1/M2)

Independent verification: fake-clock refresh and restart tests with temporary storage.

- AC-05: An enabled saved instance refreshes without the browser open; manual refresh joins
  or coalesces existing work and observes the same quota/backoff policy.
- AC-06: Timeouts, rate limiting, outages and invalid responses preserve the last validated
  snapshot and display its age/error. An unchanged cached response does not change source age.
- AC-07: Restart restores pending operations for manual continuation in M1, and automatically
  resumes enabled pending/due work with bounded concurrency and jitter in M2.
  Disable/delete prevents an old in-flight response from republishing data. Shutdown cancels
  requests, joins owned workers and closes storage in order.
- AC-08: Expired collection horizons show missing future coverage; the system never invents
  dates by repeating a weekly pattern. A local date rollover updates derived values even if
  provider content has not changed: M1 recomputes projections on local reads; M2/M3 schedule
  notification-related projections through the one coordinator.

### US-03 — Consume and evolve integrations safely (M1/M3/M4)

Independent verification: contract, authorization, migration and replacement tests.

- AC-09: Authenticated consumers read stable NodeHerder resource IDs and capability schemas;
  provider-private configuration and credentials are absent. Existing public device/MCP
  routes expose no new integration data.
- AC-10: A read-only assistant consumer and an explicitly scoped automation consumer can use
  the same validated waste resource, including freshness/quality checks. Integration refresh
  itself cannot issue a device command. This is required for the complete feature, although
  the automation UI may follow the first production slice.
- AC-11: Adding a provider for an existing capability does not require changes to its core
  lifecycle, consumers or setup renderer. The supported packaging mechanism and rebuild/restart
  requirements are disclosed; arbitrary new capabilities are not promised to be configuration-only.
- AC-12: An incompatible provider/configuration upgrade fails without destroying the previous
  usable instance. A confirmed provider replacement preserves resource IDs where semantics
  genuinely match, and identifies changes requiring automation review.

### US-04 — Protect the household and host (M1)

Independent verification: isolated egress/security tests and UI privacy inspection.

- AC-13: Cloud requests cannot reach unauthorized origins, local/metadata addresses or unsafe
  redirect targets. Secrets remain bound to their integration and permitted origin. Responses,
  setup steps, parsing and concurrent work have enforced limits.
- AC-14: Credentials are write-only through the UI and encrypted at rest with a separately
  supplied key. Missing keys fail closed. Snapshots/provider notes are also encrypted because text may contain
  private information. Tests cover backup/restore, key rotation and redaction.
- AC-15: Every authenticated user can view, add, edit, test, refresh, reload, restore and remove
  integrations/configuration and continue another user's setup. There are no additional roles,
  feature allowlists or creator-only restrictions. Revision checks prevent lost edits; existing
  authentication, cookie Origin/header checks and public-route exclusions remain enforced.
- AC-16: Desktop/mobile UI provides simple one-column setup/edit forms, accessible names/focus,
  44px touch targets, clear pending/error/offline states and safe deletion feedback at 390px/1280px.
  Form and editable YAML views operate on the same file document; phones do not require raw YAML
  for normal setup or supported connection repairs. YAML errors preserve the user's edit buffer.
- AC-17: Any signed-in user can repair supported endpoint/path, request-field or response-mapping
  changes via Form or YAML without a Go release. Save validates and atomically writes the host
  definition, activates its compiled revision and reports file/active status honestly. Changed
  credential destinations are reviewed/reconnected by that user, with no separate privileged role.
  Unsupported mechanics are rejected explicitly rather than claimed configuration-only.
- AC-18: The shipped M1 flow never submits representative opt-in or displays neighbour data as a household answer.
  Unsupported setup/provisional/non-exact replies fail clearly without publishing a snapshot.
  A future representative journey must add an explicit supported recipe and scoped, encrypted
  consent before transient display; it may never enter durable schedules/reminders/automation.
- AC-19: Encrypted setup/pending-body records follow the explicit TTL; completion/cancel/expiry
  removes transient payloads, candidates, consent and orphaned secrets. The minimal encrypted
  saved binding remains only until instance deletion. Key rotation/restoration and logical
  deletion versus backup retention are documented and verified before production release.

- AC-20: Provider definitions are ordinary host YAML files, visible/editable/reloadable in the UI
  and suitable for intentional Git commits. Credentials, household selections and fetched state
  stay outside shareable YAML. UI/disk conflicts, invalid host edits, read-only storage and
  saved-not-applied recovery preserve last-good operation. The app never commits/pushes automatically.

- AC-21: Configure operations includes a Help page/tab for every signed-in user, accessible on
  desktop/mobile without losing the unsaved draft. It explains requests, inputs/state references,
  setup/refresh wiring, schemas/mappings, validation and URL/field-repair examples. Field/error links
  target relevant help sections; bundled content works without external network/AI access. Help
  leads with practical tasks and concrete examples; internal/YAML notation is optional reference.
  Guided repair uses Provider field and Use this value controls with readable choices, preserving
  source/key/fixed-value edits in Form/YAML. Automatically saved values are explanatory text,
  not read-only text boxes that look editable.
- AC-22: Ordinary typed backend configuration plus small explicit checks validate YAML; no
  runtime JSON Schema engine or schema-generated UI is required. Validation rejects
  unknown request/input/state references, conflicting status sets and unsupported operations.
  Request IDs are arbitrary and explicitly referenced by setup/refresh configuration. Renaming a
  request and its references or adding another bin colour requires no provider-specific Go change.
  Provider mappings validate consumed fields and produce the shared waste output contract.

- AC-23: Connections is a separate top-level section after Devices on desktop and mobile, preserving
  all existing sections and device counts. Its catalog distinguishes types, providers and configured
  instances; only supported types can be installed. Normal setup requires no operation editing.
  Authenticated deep links, active navigation and Help work without a configured device. The
  approved name is Connections; this is one shared home with no household switcher/extra household
  heading or breadcrumb. Connections is reached from its shared main-navigation entry.
- AC-26: Connection setup, provider editing, results and health are presented only in Connections.
  Home areas/tiles/counts, Panel rooms/device controls/status, Overview content/counts/activity and
  Devices content remain unchanged. No connection pin/card, room assignment, dashboard-setting
  field, extra greeting/date/header, explanatory banner or management button is added there.
- AC-27: The shared Connections main-navigation entry is the discovery/entry route. Do not add
  cross-promotion buttons, links, notices or connection-specific footer copy to Devices, Home,
  Overview or Panel. Links between views inside Connections and authenticated direct routes work.
  Desktop/mobile negative UI assertions verify the page boundaries and unchanged baseline counts.

Withdrawn scope: AC-24/FR-14 (Home resource tiles) and AC-25/FR-15 (Panel room tiles/details) are
retired following the user's Connections-only decision. The earlier placement samples were
exploration, not approved product scope. Their IDs are not reused; T029/T030 are withdrawn too.
Any future dashboard/panel presentation requires a separately scoped design and user decision.

## Requirements and success

| ID | Outcome | Acceptance |
| --- | --- | --- |
| FR-01 | Discover/configure integrations using ordinary user inputs | AC-01–04, AC-16 |
| FR-02 | Persist configuration, operation identity and validated data | AC-04, AC-06–07 |
| FR-03 | Refresh automatically with cache, recovery and health | AC-05–08 |
| FR-04 | Offer stable typed resources to authenticated consumers | AC-03, AC-09–10 |
| FR-05 | Add providers without changing common runtime/consumers | AC-11 |
| FR-06 | Support explicit upgrades/reconfiguration/replacement | AC-12 |
| FR-07 | Enforce egress, credential, authorization and privacy boundaries | AC-13–15 |
| FR-08 | Repair supported HTTP endpoint/request/response mappings through the shared UI without a release | AC-17 |
| FR-09 | Separate representative consent/display from durable household data and enforce private-state retention | AC-18–19 |
| FR-10 | Use host YAML as definition source of truth with equal-access Form/YAML editing and intentional Git versioning | AC-15–17/20 |
| FR-11 | Explain operation configuration through contextual and full-page help | AC-21 |
| FR-12 | Reuse generic operations with typed validated configuration and explicit flow references | AC-22 |
| FR-13 | Expose an extensible top-level Connections category beside Devices | AC-23 |
| FR-16 | Keep initial connection UI contained and preserve existing pages | AC-26 |
| FR-17 | Use main navigation without cross-page promotion or extra prose | AC-27 |

Proposed measurable targets, to confirm before implementation:

- NFR-01: Initial household capacity: 50 configured instances, four global external requests,
  one operation per instance, and a separate per-provider quota gate. Excess requests coalesce
  or receive a bounded busy response; they do not create unbounded goroutines/queues.
- NFR-02: Default request deadline 15 seconds, 30-second total work attempt and 2 MiB decoded
  JSON response limit. Provider-specific exceptions require recorded limits. Provider waiting
  persists a continuation record rather than holding a browser request for two minutes; M1
  continues on user action, M2 through scheduling.
- NFR-03: Under an available fake provider, work begins within 60 seconds of its persisted due
  time, excluding an explicitly recorded provider quota delay. A status/read API performs no
  synchronous upstream call. The due-time guarantee begins in M2. Proposed shutdown target:
  ten seconds with cancellable fake I/O.
- NFR-04: UK collections use date-only values with `Europe/London`; operational timestamps use
  UTC. Missing values do not become zero. Derived countdowns use local dates, including DST.
- NFR-05: Routine checks use fixtures, `httptest`, controlled clocks and temporary databases;
  they neither create live provider jobs nor command household devices.
- NFR-06: Proposed retention: latest snapshot, 20 sanitized run summaries per instance capped
  at seven days, setup expiry after 30 minutes idle/24 hours absolute, and transient address
  candidates/exact request bodies removed on completion/cancellation/expiry; keep a separate
  encrypted completion receipt for 24 hours without inputs/secret/body. Saved unresolved
  attempts expire after 24 hours. Credential/private-state encryption, backups and cleanup follow
  the execution/security contract; the fetched OpenAPI documents conditional 30-day-or-removal retention, while live recovery remains untested.

M1 acceptance: AC-01–04, AC-06, AC-08–09, AC-13–23, AC-26–27 and the M1 portion of AC-07.
M2 adds AC-05 and automatic AC-07.
M3 delivers AC-10; M4 delivers AC-11–12 with a real second-provider/replacement demonstration.
Manual M1 is a useful milestone, not completion of the overall automatic-integration feature.
All numerical limits are proposals, not existing application measurements.

## Entities and external interfaces

- Integration type: discoverable capability and provider choices.
- Integration instance: durable local UUID, shared household, user configuration,
  selected provider binding, schema versions and secret references.
- Setup operation: shared revision, audit actor, step/revision, selected candidate, expiry, idempotency identity,
  provider lookup identity and continuation time.
- Resource snapshot: stable local resource/service IDs, capability version, typed data,
  provenance, quality, coverage and separate source/check timestamps.
- Provider binding: opaque provider property IDs and optional verified cross-provider
  identifiers; these are not NodeHerder resource identity.

Producers: reusable HTTP/search-select operations using bounded host YAML definitions and a shared
typed waste capability validator. New recipes/transports remain later work. Consumers: the
authenticated Vue UI, domain read service, future authorized assistant and automation adapters.
No changes are proposed to MQTT semantics or existing public device contracts.

## Confirmed decisions and remaining verification

1. **D-01 delivery — confirmed 2026-10-25:** Manual refresh for the first release. Setup, Refresh
   and Check status are user-triggered; no recurring provider jobs. Automatic refresh remains later
   roadmap work and is not bundled into M1 or treated as already implemented.
2. **D-02 configuration — user decision recorded:** All signed-in users can view/edit/modify
   integrations; API/schema repairs must be available through the UI. Host-saved YAML is the
   recommended realization of the user's stated preference, with Form and YAML editing. New
   provider definitions can reuse installed recipes/capabilities; arbitrary new mechanics still
   need code. No separate feature permission tier is required or planned.
3. **D-03 access — user decision recorded:** Use the existing authentication boundary and equal
   integration access. The current one-shared-household deployment remains the design scope;
   no per-user/private-tenant policy or additional role migration is introduced.
4. **D-05 deployment/use — confirmed 2026-10-25:** Personal use on the user's Docker host.
   Distributed/commercial deployment is outside the first release. Use the existing configs/data
   volumes and a separate read-only host secrets-directory mount; no new deployment service.
5. **Consumer scope:** Confirm the first assistant and automation use cases; keep the production
   refresh/read service useful while dependent engine/MCP authorization work is planned.
6. **D-04 key custody — technical design selected for Docker:** A host-provisioned 0600 keyring
   outside the checkout is mounted read-only at /run/nodeherder-secrets. See the Docker contract
   for paths/rotation/restore. It is not an outstanding user choice; fixture-backed deployment and
   restore verification remain required before production use. No real key is generated by planning.
7. **Provider evidence:** OpenAPI 1.1.0 and its idempotency declaration were fetched/parsed in the
   audit; intended usage/access, council coverage and live behavior still need confirmation. Unsupported council journeys must fail clearly rather than be guessed.

Assumptions: first provider is read-only, there is one active backend writer, no household
device operations are needed for investigation, and provider changes are reviewed by humans.
Questions block only their dependent implementation, not this investigation.

M1 delivery boundary: later assistant/automation/replacement milestones are roadmap proposals,
not authorization to change those views now. Resolve their own scenarios before dependent work.
