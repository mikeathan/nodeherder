# Plan: Extensible IoT control center

Spec: [spec.md](spec.md) | Status: implemented; release validation remains open | Updated: 2026-10-31
Constitutions: parent 1.0.0 and backend/frontend 1.0.0, all proposed

## Design and ownership

Follow [ADR 001](adr-001-device-boundaries.md), [ADR 002](adr-002-esphome-client.md), the [provider contract](contracts/device-provider.md), and [UI design](ui-design.md). These are proposed decisions, not claims of approval. The existing device registry and `DeviceLifetimeService` remain the authoritative confirmed state; provider transport is a producer/command target, not another state store.

```text
ESPHome session ─┐
                 ├─ provider sink → ordered device lanes → state + automation → existing WS/Vuex
Zigbee adapter ──┘                                                └→ metrics sampling
UI command / automation action → shared validator/router → provider command → device report
```

### Backend boundaries

- Add a focused provider registry, session manager and consumer-owned contracts under a new `backend/internal/deviceproviders/` boundary. Statically register providers at hub startup; a provider exposes lifecycle and a command capability, and discovery is a separate optional interface. Inject clock, dialer, storage and sink; no reflection, dynamic loading, global mutable state or ESPHome conditionals in shared domain logic. Validate the precise signatures against existing tests before implementation.
- Adapt existing `backend/internal/controllers/hub-controller.go`, `internal/services/device_registrar.go`, `device_processor.go`, `device_lifetime.go` and `internal/lanes/` so provider-scoped stable IDs, metadata and ordered reports enter the existing registry. Preserve Zigbee IDs, names, topic mapping and event semantics. Distinguish state reports from physical events, publish every accepted automation input before UI/metric sampling, and never conflate command submission with confirmed state.
- Add a shared command service used by the existing WebSocket handler and `backend/internal/automations/action.go`. Resolve provider and entity from authoritative state, validate access, value type, range and options, and return structured offline/unsupported/invalid/failure results. Preserve Zigbee batch/single behavior and wire shapes. Test a fake second provider without modifying either real provider. Each change owns one reason to change; keep adapter-specific metadata and translation inside the adapter.
- ESPHome session code under a provider-specific package owns protocol mapping, connection, subscriptions, reconnect and normalized state. The patched client is isolated at `backend/third_party/esphome_api/` as specified by ADR 002. It owns no UI or automation rules. Supported entities are mapped from discovered metadata; unsupported kinds are read-only and safe. The host manager owns one session per enabled configuration, not per browser.
- Use a dedicated encrypted device-provider repository and independent startup/close lifecycle, borrowing existing repository/keyring patterns only if extracted without coupling to the optional HTTP integration service. Persist configuration, provider binding and secrets separately from public device state. Exact schema, key path, backup/rotation and rollback must be finalized before storage edits; never serialize API keys into shared DTOs.
- Protect existing public `/api/context/devices` and MCP prompt, lookup and metric paths from new private provider devices, while preserving existing Zigbee responses. Authenticated API/WS changes must use existing middleware, session/origin protections and bounded inputs. Numeric private unicast targets are the proposed default; specify IPv4/IPv6 ranges, address changes and Docker behavior before implementation. Discovery returns untrusted candidates and does not enroll automatically.

### Reliability and evidence targets

Proposed initial bounds, to be checked against device behavior and refined before implementation: 32 configured ESPHome devices; 256 entities per device; 64 KiB maximum decoded/native frame and bounded metadata; connect/handshake 10 s; 30 s maximum shutdown; exponential reconnect 1–60 s with jitter; bounded per-device queues using existing lane capacity or a smaller documented cap. Exceeding a frame, entity or queue limit produces a visible classified failure. A full ingress queue marks the session desynchronized and reconnects/resubscribes to restore state; never silently drop physical events. A stale spatial input is unknown after a configurable 5 s default, subject to firmware evidence and an explicit user-visible setting. These are design targets, not measured guarantees; validate them with fixtures and amend the spec if they change.

Patched transport must reject malformed/truncated frames before allocation, close sockets on all failures, synchronize shared state, support context cancellation and idempotent shutdown, and never downgrade Noise to plaintext. Test a fake peer for both transports, version negotiation, command requests, subscriptions, disconnect and replay boundaries, unknown message types, malformed input, fuzzing, race detection and resource cleanup. Physical-device checks are separately scoped and never routine tests.

### Frontend and spatial design

For ESPHome only, reuse `ExposeRow` inside a focused bounded section component with independently searchable, alphabetically sorted pages. On wide screens, stack Diagnostics beneath Readings and place Configuration in the adjacent column; keep the existing responsive single-column behavior. The existing device and settings stores remain authoritative. The management dialog compares successful enrollment results with the existing registry, reports unmatched registered devices separately, and distinguishes loading, errors and a genuinely empty result. Cancel or fence stale dialog requests on close/reopen. Do not infer IP addresses or credentials or automatically remove or re-enroll devices.

During snapshot hydration, retain a bounded number of early WebSocket reports for devices not yet known to Vuex and reconcile them only against entities present in the returned snapshot or an already-received device addition. Overflow fails the snapshot explicitly so recovery can retry; unknown IDs never create devices. Normal reporting, automation and device lifecycles do not change.

Spatial mapping drafts and entity pagination react to the stable device ID and relevant configuration rather than the live device object. Validate each nonempty mapping against current readable entity metadata before saving; report incomplete, duplicate-axis and removed-field errors in the form. Empty slots remain optional.
For conventional, unambiguous `Target N X`, `Target N Y` and `Target N Active` labels, additionally identify mixed target numbers, fields placed in a differently numbered row and reversed axes. Apply this only to recognized labels; unknown label schemes remain manually mappable.
Each selector prefers only the recognized field matching its target and axis; when no such field exists it offers unrecognized candidates as a fallback. Preserve any previously saved invalid selection visibly for correction. Mapped coordinate freshness is limited to one presentation update per 250 ms per entity, without changing automation or history rules.

Review follow-up: fence canceled manager operations and asynchronously join the old worker before restarting it, preserving a single session. Represent unsupported text patterns as read-only inventory instead of rejecting discovery. Apply an explicit fallback category by entity kind; do not override declared categories. Active sensors can be change-only; a time-based active expiry would suppress stationary targets, so retain confirmed active state while the session is online and require fresh coordinates. Keep a stable inventory-derived list in the frontend while resolving current values from the device registry. Clarify tab labels within the existing layout. Extend only native capabilities whose discovery, state and command contracts can be validated; do not invent support for unknown types.

Place an expandable spatial guide beside the view. Reuse the integration-help pattern of topic navigation and cards, with accessible inline diagrams that explain matching discovered X/Y/active entities and reading the coordinate plane. The guide is informational, responsive and does not infer or modify mappings.

Preserve existing layout, colors, typography, navigation, styles and Zigbee presentation as the user directed. Add **Add ESPHome device** to the existing Devices header; it opens ESPHome setup directly, while Zigbee continues using **Permit join**. Add ESPHome settings and discovered capabilities within existing patterns described in [ui-design.md](ui-design.md), reusing `frontend/src/components/device-list/DeviceList.vue`, `device/DevicePage.vue`, `components/entity/`, `composables/useDeviceCommand.ts`, and existing `store/modules/hub/` and `store/modules/ws/`. The historical prototype is not a redesign specification. Add narrow provider setup and spatial components only where needed. Shared contracts go in `src/types/` and `src/contracts/`; services handle transport, composables own lifecycle, components own presentation. Runtime-validate ingress and preserve Vuex as the single shared state. Test a fake provider descriptor to prove generic rendering.

Define a protocol-neutral spatial view model, mapping configuration, freshness and optional target-active evidence. It consumes validated observations in millimetres, shows a labelled grid, orientation/scale and a textual target list, and never treats absent coordinates as zero or claims simultaneous samples. Feature-specific mapping belongs to a provider/capability extension; no device-name checks in the shared renderer. Verify 1440/768/375 px, keyboard, focus, reduced motion, offline, stale, empty and pending/error states.

## Initial constitution check (historical design baseline)

| Rule | Before implementation | Before merge |
| --- | --- | --- |
| NH-01 | BLOCKED: finalize versioned acceptance, contracts and remaining Q-02–05 | Pending actual acceptance |
| NH-02 | BLOCKED: prove state/event, command, ordering and retry semantics | Pending regression evidence |
| NH-03 | BLOCKED: finalize target, public MCP/context and secret-storage policy | Pending security tests |
| NH-04 | Proposed boundaries and ADRs; schemas, migration and dependency verification pending | Pending review |
| NH-05 | Planned tests below; no application checks run | Pending results |
| NH-06 | BLOCKED: finalize repository schema, key lifecycle, resource bounds and recovery | Pending recovery tests |
| BE-01/02/03/04/05 | Existing ownership, device semantics, lifecycle, storage and route checks planned | Pending checks |
| FE-01/02/03/04/05 | Vuex, contracts, confirmed state, accessibility and secret checks planned | Pending checks |

## Contracts, rollout and recovery

Before production code, finish wire schemas in `contracts/`: authenticated onboarding/configuration/credential mutations and errors; device/health/entity DTOs; incremental WS payloads and hydration/resync; existing MQTT consumers and MCP/public filters; persisted schema, key rotation, deletion, backup, migration and downgrade. Preserve existing API and WS shapes where possible; document additive fields and frontend mock updates. If older versions cannot interpret new provider records, require a backup and specify disabling the provider before rollback; do not claim data recovery without testing it. Classify errors without logging secrets or raw untrusted payloads.

## Traceability and verification

| Acceptance | Owners and evidence | Result |
| --- | --- | --- |
| AC-01, 04, 08, 10 | patched module and ESPHome provider: fake Noise/plaintext peers, bounded framing, key/persistence and lifecycle tests | Pending; upstream v1.4.0 probes failed as in [client evaluation](client-evaluation.md) |
| AC-02, 03, 05, 11 | controllers/services/automations: fake second provider, repeated events, ordered state, Zigbee compatibility and command failure tests | Pending |
| AC-06, 12 | existing UI and provider setup: contract tests, mocked browser interactions, responsive/accessibility checks and before/after Zigbee visual comparison | Pending |
| AC-07 | spatial domain and view: fixtures for actual verified units, missing/stale/partial targets and orientation | Pending firmware evidence |
| AC-09 | all suites use fakes and temporary storage; any real-device procedure separately scoped | Pending |

From `backend/`: `go test ./...`, `go build ./...`, `go vet ./...` and race tests for changed packages; separately run patched nested-module tests, race and bounded fuzzing. From `frontend/`: `npm test -- --runInBand`, `npm run lint`, `npm run build` and applicable Playwright UI checks. Record exact commands, results, failures and unrun checks. Review the final full diff and public GoDoc/JSDoc against engineering rules; record unresolved findings.

## Risks, debt and completion

The selected upstream release demonstrably leaks a socket, races and panics on malformed input; it is not suitable unchanged. A local patch entails ongoing protocol, security, license and dependency maintenance and does not inherit maturity automatically. The patched Go module has passed `GOTOOLCHAIN=go1.24.8 go test -race -count=1 ./pkg/...` and `GOTOOLCHAIN=go1.24.8 go test ./...`. Five-second `FuzzPlaintextRead` and `FuzzDecodeEncryptedMessage` runs also passed; these checks do not establish full encrypted framing or independent Noise interoperability. Public MCP/context exposure, collision-free identity, key recovery and firmware-specific radar semantics are mandatory gates, not approved exceptions. Resolve Q-02–05 before dependent implementation. The initial design baseline above predates implementation; see tasks and the review verification below for current evidence. Broader release acceptance remains open.

## Review correction implementation

AC-13: Skip report-age availability monitors for provider devices; their supervisor owns availability. Move ping waiting to one session-owned worker, close the client on worker shutdown and join it. Keep bounded ingress and command queues.
AC-14: Add explicit button presentation and null-only automation validation. Use a short bounded submission guard for stateless UI actions, clearing it on reset/disconnect; feedback says submitted, never executed.
AC-15: Keep polling and mutation errors separate. Capture dialog generations for mutation completions; retain loaded rows on refresh failure and serialize conflicting edits/removals.
AC-16: Store bounded allowlisted public device information in existing properties and device_class/entity_kind in entity attributes; use shared presentation helpers, retaining legacy fallbacks.
Verification: owned Go session/lifetime/sink tests, frontend domain/command tests and mocked Playwright management/automation/device tests; full test/build/vet/lint plus affected race checks. No physical-device traffic. NH-01 specification updated before edits; NH-02 confirmed state preserved; NH-03 metadata allowlisted and bounded, secrets excluded; NH-04 additive properties/attributes preserve consumers; NH-05 regressions below; NH-06 workers bounded/joined. BE-01–05 and FE-01–05 retain existing boundaries/auth/storage and require lifecycle, contract, failure and UI evidence. No exceptions claimed.

## Review verification result — 2026-10-31

AC-13–16 are implemented with owned regressions and passing backend/full frontend/targeted browser checks, detailed in [tasks](tasks.md#review-correction-evidence--2026-10-31). The earlier design gate table is historical, not current verification evidence. No unresolved code finding from this review remains; the broader dependency-security, physical deployment connectivity and firmware-interoperability release gates recorded in tasks remain open. No production-readiness or exception approval is claimed.

## Feature-folder grouping — 2026-10-31

Move the new frontend files out of flat `components/`, `composables/`, `domain/` and
`services/` lists into feature subfolders: `device/esphome`, `device/provider`,
`device/spatial`, `device-list/esphome`, `composables/esphome`, `composables/spatial`,
`composables/device`, `domain/esphome`, `domain/spatial`, `domain/provider`,
`domain/devices` and `services/esphome`. Mirror the source paths under `__tests__/`.
Imports and the generated component registry were updated; behavior and contracts are
unchanged. Reused the extracted ESPHome header actions in both device list and card views.

## Frontend responsibility refactor — 2026-10-31

Preserve existing behavior/contracts while separating snapshot transport orchestration, request lifetime and bounded mutation replay. Audit all changed frontend files: separate ESPHome response decoding from transport, list polling from dialog interactions, spatial validation from presentation and freshness subscription from drafts, and entity ordering from search/pagination. Keep small focused existing components intact; avoid generic frameworks. Existing snapshot overflow/reconnect/logout/config replay tests, service/domain tests, frontend full tests/build/lint/e2e type checks and mocked desktop/mobile ESPHome/spatial browser checks verify equivalence. NH-02/03/06: preserve confirmed state, ingress validation, queue bounds, cancellation and cleanup. FE-01–05: retain Vuex/auth/shared controls and UI behavior. Backend unchanged; backend suite remains a merge check.

## Spatial workspace design — 2026-10-31

Extend the existing Spatial tab, preserving all other device/navigation layouts. Use
button navigation inside the card: Live dashboard, Zone editor, Room builder,
Entry / exit, Heatmap & analytics. Keep the live canvas and settings shared; place
help inside its relevant view, plus a separate “How to create zones” expander.

Ownership: `types/spatial.ts` wire types; `domain/spatial/workspace.ts` geometry,
validation and occupancy previews; `domain/spatial/analytics.ts` bounded session
aggregation; spatial composables own drafts/timer cleanup; spatial Vue components own
presentation. `backend/models/settings/spatial.go` owns equivalent validation and
cloning. Existing settings persistence carries optional `spatial.workspace` version 1:
zones (typed polygons), room outline/objects, sensor placement/FOV and explicit entry
entity mappings. Old settings without workspace load defaults. No automatic migration,
network exposure or physical commands. Rollback ignores the optional workspace; back
up existing settings before a downgrade that re-saves them. No new storage boundary,
dependency or ADR is required for this additive existing-config extension.

Session analytics: one-second frames, at most 3600 frames / three positions per frame,
60-minute retention, 250mm grid cells, no position persistence. Only frames with fresh
targets contribute heat; coverage is reported separately. Clear confirms scope. Device
identity, mapping changes and disconnect clear trails to avoid invented paths. Offline
invalidates occupancy previews and pauses collection. Entry controls reuse discovered
ExposeRow/command semantics and never infer crossings or assumed state from geometry.

Design gate: NH-01 PASS (AC-17–22 before edits); NH-02 PASS (previews separate from
firmware, no automation changes); NH-03 PASS (existing config boundary, validation,
in-memory household positions only); NH-04 PASS (existing state/service/transport);
NH-05 PASS design (owned geometry/aggregation/settings tests and mocked UI planned);
NH-06 PASS design (bounded timer/history, additive settings, undo/reset before save).
BE-01/03/04 PASS design (pure settings validation/deep clone, no goroutines, old data);
BE-02 N/A (no lifetime changes); BE-05 N/A (no route/assistant changes).
FE-01–05 PASS design (Vuex settings, typed optional contract, honest requested saves,
shared controls/keyboard/numeric editing/theme tokens, no browser persistence).
Execution gates remain pending: full backend tests/build/vet, settings race checks,
frontend tests/lint/build and mocked keyboard/offline/375/768/1440 checks. Review the
final task diff against all applicable rules. Physical-device tests are excluded.

AC-23: proceed with both explicit bindings and overlays as the stated default.
Optional `zone.entities` maps beginX/beginY/endX/endY to discovered entity names;
`domain/spatial/device-zones.ts` plans native-unit commands after complete preflight.
A focused zone-device component uses the existing `useDeviceCommand` pending/report
lifecycle. Four coordinate writes are not transactional: explain temporary intermediate
bounds and partial failure, stop if refused, never retry automatically. Backend command
validation remains authoritative. Only explicit Apply touches firmware; layout saving,
removing overlays and undo never send commands. Polygon serialization and mode switching
are withheld because actual native text contracts have not been demonstrated.

### Spatial workspace verification — 2026-10-31

AC-17–23 are implemented in the files listed above plus `SpatialZoneDevice.vue`.
Zone-device bindings also optionally select a readable binary occupancy entity;
reported occupancy is displayed separately from the geometric preview. Native unit,
range, step, duplicate and unsupported-control checks run before any coordinate send.
No household-device commands or tests were performed.

- Backend: `go test ./...`, `go build ./...`, `go vet ./...` and
  `go test -race ./models/settings` passed. New `spatial_test.go` covers legacy JSON,
  round-trip, validation, crossing/degenerate geometry, bounds and independent clones.
- Frontend: `npm test -- --runInBand` passed (102 suites, 565 tests).
  `npm run build`, `npm run lint` (28 existing warnings, zero errors) and
  `npm run type-check:e2e` passed. Owned domain tests cover inclusive geometry,
  transforms, exclusions, missing/offline states, retention, heat aggregation, trail
  gaps, unit conversion and atomic preflight rejection.
- Browser: `npx playwright test e2e/devices.spec.ts e2e/ExposeControl.spec.ts`
  passed all 26 desktop/mobile checks. The expanded spatial flow covers keyboard
  help, live-update draft preservation, zone creation/rename/dimensions, saved settings
  payload, four explicit mocked coordinate requests, undo, room objects/outline,
  contextual entry help, clear confirmation and offline invalidation. Viewports:
  1440/768/375 and Pixel 7. Desktop 1440 and 375 screenshots inspected. This found
  and corrected selected-zone loss on save and overlapping zone/furniture labels.
- Final source/diff review: existing engineering boundaries and command pending truth
  retained; settings snapshots own nested maps/slices; geometry bounded and validated
  on both sides; lifecycle cleanup and bounded undo/history present; status text does
  not depend on colors; exported APIs documented; no dependency or route additions.
  NH-01–06, BE-01/03/04 and FE-01–05 pass for this scope with the evidence above;
  BE-02/05 remain N/A for unchanged lifetime/route behavior. No exception claimed.
- Remaining limitations: session analytics have no backend history/backfill; polygon
  firmware writes/mode switches are unsupported without a verified native serialization
  contract. Room geometry is visual context rather than an RF simulation. Firmware
  persistence/interoperability is not demonstrated by mocked submission tests.

Final follow-up: the two spatial browser checks passed again after label placement and
reported-occupancy additions; final build/lint/e2e type checks passed. Existing frontend
and backend code graphs were refreshed and merged, and the combined graph report/HTML
was regenerated (10,800 nodes). This is a code refresh: changed README/spec prose was
not semantically reindexed, so documentation graph currency remains pending. No semantic
refresh is claimed. `git diff --check` passed; unrelated staged work was preserved.

## Spatial review correction plan — 2026-10-31

AC-24–26 reopen verification. Add an optional correlated settings-save branch in
`backend/internal/ws/eventhub.go`, retaining the existing persistence callback and
legacy behavior. Validate correlation identifiers before saving. Return bounded,
correlated success/failure metadata through the existing authenticated WS broadcast;
all clients may receive metadata, but only the matching pending UUID/device accepts
it. Existing WS request size/auth controls remain authoritative.

Frontend `contracts/device-settings.ts` decodes result metadata; the WS module owns
routing/state, and `services/device-settings.ts` owns request subscription, timeout,
abort/disconnect cleanup and confirmed Vuex publication. One pending spatial save
per view, ten-second timeout, no automatic retries. Failure leaves drafts/history
intact; timeout means outcome unknown. Older backends ignore the extra requestId and
cannot confirm: show the bounded timeout instead of optimistic success. Update mock
hub support and both producer/consumer tests. No new storage/dependency boundary.

`useSpatialView` distinguishes hydrated settings from unsaved drafts and fences save
completion by device identity/unmount. Same-value acknowledgements do not replace
workspace refs. `domain/spatial/workspace.ts` owns rectangle bounds/corner resizing;
`useWorkspaceEditor` owns a gesture's immutable anchor and applies vertex updates.
Canvas emits gesture start/end; parent presentation no longer owns resize rules.
Known mm coordinates convert to metres for `formatQuantity(...,'distance')` with
precision preserving sub-metre readings. Zone theme aliases use existing info/danger/
ok colors; heatmap palette remains a data color scale. Explain both geometry copies.

Design check: NH-01 PASS (AC-24–26); NH-02 PASS (no device command changes);
NH-03 PASS (existing socket authorization, validated metadata); NH-04 PASS (typed
additive contract, existing cache persistence, pure domain geometry); NH-05 PASS design
(owned save/failure/lifecycle/geometry regressions planned); NH-06 PASS design
(bounded timer, explicit unknown timeout outcome, preserved drafts/undo).
BE-01/03/04/05 and FE-01–05 PASS design; BE-02 N/A (no lifetime semantics change).
Execution pending: frontend full tests/lint/build/e2e type check, targeted mocked
spatial browser flows, backend tests/build/vet and affected race checks; final diff
review. No household-device traffic. Update existing source/merged code graphs.

### Spatial review correction evidence — 2026-10-31

- AC-24: confirmed-save service regressions cover exact request/device matching,
  malformed replies, copied submissions, rejection, timeout, abort, disconnect and
  hub reset. Spatial-view regressions cover preserved Undo/drafts on failure/retry,
  edits during a pending save, remote refresh, device switching, unmount cleanup and
  semantically equivalent JSON with different Go/browser key ordering. A matching
  acknowledgement does not replace the workspace ref. No optimistic save is used
  by the spatial view. Legacy settings callers retain their original wire behavior.
- Protocol tests in `internal/ws/eventhub_test.go` cover successful/rejected correlated
  replies, no leaked repository paths, stripped transport metadata and rejection of
  invalid correlation tokens. Existing legacy save tests still pass. The normal Node
  mock hub supports the new optional reply, and the inbound WS companion test checks
  runtime decoding/routing. Responses remain on the existing authenticated socket;
  only the matching request/device publishes confirmed local configuration.
- AC-25: owned domain/editor regressions retain the initial anchor after vertex
  reordering and test Undo. The mocked browser flow now performs an actual corner
  drag through both axes, verifies fixed-anchor dimensions and undoes it on desktop
  and mobile. Keyboard/numeric geometry uses the shared bounds/rectangle helpers.
- AC-26: spatial distance output uses the known-distance formatter, converts mm to
  metres and retains millimetre precision; tests include negative/missing/nonfinite
  quantities. Zone type colors are central aliases of info/danger/ok theme tokens.
  Both polygon validators now explain collinear intersections, legal shared adjacent
  endpoints and the twice-area threshold (rejecting area <= 0.5 square millimetres).
- `npm test -- --runInBand`: PASS (106 suites, 582 tests).
- `npm run build`, `npm run lint` (28 existing warnings, zero errors), and
  `npm run type-check:e2e`: PASS.
- `npx playwright test e2e/devices.spec.ts --grep 'spatial mappings'`: PASS (desktop
  and mobile, 2 tests). Includes rejected-save recovery, pending controls, applied
  mock coordinate requests, keyboard help, fixed-anchor drag/Undo, offline states and
  1440/768/375/390px width checks. Inspected the final 375px capture.
- Backend `go test ./...`, `go build ./...`, `go vet ./...`, and
  `go test -race ./internal/ws ./models/settings`: PASS.
- Final rule/diff review: existing state/transport/cache boundaries preserved;
  correlation input and reply metadata validated; confirmed-only publication and
  scoped cleanup tested; exported APIs documented; physical commands unchanged;
  theme/units/geometry findings corrected. `git diff --check` PASS. NH-01–06,
  BE-01/03/04/05 and FE-01–05 pass for the correction scope; BE-02 remains N/A.
  No exception or maintainer approval is asserted. No unresolved finding from the
  five-item spatial review remains. No household-device traffic was generated.
- Both code source graphs were refreshed and merged; combined report/HTML regenerated.
  README/spec semantic indexing remains pending, as previously recorded; a code graph
  refresh does not establish documentation currency.

### Spatial code ownership cleanup — 2026-10-31

- Moved projection/snapping, workspace validation/defaults, view metadata, entry
  filters, analytics constants and device/zone command formatting into
  `domain/spatial/`; components now bind state and emit events. Removed unused
  `SpatialRadar`/`SpatialRadarPlane` now that the shared canvas renders live data.
- Device icon selection is a semantic domain function. It ignores config/diagnostic
  support entities and prioritizes presence/occupancy/motion over an auxiliary light;
  provider-scoped hex IDs use discovered labels. The UI registry only maps semantic
  families to glyphs.
- Provider entity labels fall back to the discovered description/device class when the
  label attribute is absent. A 74-entity regression asserts one inventory registers one
  device; provider pages split readings from writable controls to keep long inventories
  navigable.
- History entity options and history/chart/entity-dialog labels use the shared display
  label helper while retaining stable expose keys for queries and preferences.
- Occupancy investigation: provider reconnect now carries matching binary states into
  the replacement inventory, preserving change-only occupancy until a fresh report.
  Numeric readings are not carried. A backend process restart cannot preserve in-memory
  state; firmware initial-state behavior and explicit firmware zone changes remain
  separate causes to verify.
- Spatial mapping selectors reuse a stable inventory snapshot across live readings;
  otherwise every coordinate report could replace option DOM nodes and reset an open
  or newly changed select. The browser flow covers Target 2 select/clear with a saved
  partial mapping and a concurrent live report.
- The spatial view lists inactive target status below the canvas (not active, stale,
  mismatch, incomplete, offline). Inactive targets are not drawn; mapping stays
  editable so a target can be configured before the radar detects it. Live status is
  kept out of the mapping fieldset so selectors receive no live-reactive props.
- Zone creation and correlated settings saves use one canonical UUID helper with a
  `crypto.getRandomValues` fallback, because Firefox omits `crypto.randomUUID` on
  non-secure LAN HTTP origins.
- Added inline SVG help illustrations for zones, room builder, entry and heatmap.
- Heatmap keeps room outline/furniture/sensor as optional context through one
  **Room context** toggle; zones, targets, trails, FOV and handles are suppressed so
  activity cells remain readable. The grid keeps its independent toggle.
- `npm test -- --runInBand`: PASS (109 suites, 595 tests). `npm run build`,
  `npm run lint` (28 existing warnings, zero errors): PASS.
- `npx playwright test e2e/devices.spec.ts --grep 'spatial mappings|ESPHome entity sections'`:
  PASS (4 desktop/mobile checks), including Room context and the split
  Readings/Controls sections. `npx playwright test e2e/DeviceMetrics.spec.ts e2e/EntityViewDialog.spec.ts`:
  PASS (95 passed, 9 skipped) for history selectors/charts and dialog labels.
  `npx playwright test e2e/ActivityFeed.spec.ts e2e/SensorReportingSettings.spec.ts`:
  PASS (14) for activity/settings shared-label regressions.
- Backend `go test ./...`, `go build ./...`, `go vet ./...`: PASS.
  No household-device traffic was generated.

## Mapping interaction stability follow-up (2026-11-01)

Investigate native selection between input and change events, which synchronous
Playwright selectOption coverage does not exercise. Keep draft ownership in
useSpatialView and presentation in SpatialMappingEditor. Use a stable metadata-only
view for warnings/labels as well as options so callbacks cannot subscribe the editor
to live device replacement; commit native input promptly if needed by reproduction.
Do not change persisted mapping shapes, optional empty rows, validation or save
acknowledgement semantics. Add meaningful coverage to existing source companion and
device browser tests, including saved full Target 2, live updates, clearing, save
failure and reopening. No dependencies or ADR required.

Pre-implementation rule review: NH-01 PASS (AC-22/23); NH-02 PASS (UI drafts only);
NH-03 PASS (no ingress/exposure/data logging changes); NH-04 PASS (existing ownership
and contracts); NH-05 pending regression and frontend suite/lint/build;
NH-06 PASS (empty rows retained, existing confirmed saves); FE-01/02/03/05 PASS by
unchanged state/transport/security boundaries; FE-04 pending keyboard, failure and
1440/768/375 px browser checks. Backend rules N/A: no backend edits planned.
Verification uses an owned Node mock on 4210 and Vite on 4200 because standard
4110/4100 ports were occupied. No requests or commands to household devices.
Existing Chromium partial/full mapping checks pass; reported native failure is still
under investigation. Final evidence and rule review pending.

### Final evidence for AC-22/23

Reproduction: changing the native select value and dispatching input, then delivering
a coordinate report before change, restored target2_x on the original source.
The original selectOption test passed because it dispatches change immediately.
Warning/label callbacks subscribed editor rendering to device replacement despite
stable option arrays. They now read the same stable inventory; input commits the
draft immediately, with change retained for compatibility. Actual inventory/label
changes still refresh options and warnings. The help text explains clearing all
three roles before saving.

- Frontend Jest: PASS, 109 suites / 597 tests, including new companion tests for
  metadata refresh and unmapping across config refresh, failed save, retry, reopen.
- Frontend lint: PASS, zero errors / 28 existing warnings; TypeScript test files are
  not covered by this ESLint configuration. Production build: PASS.
- Isolated existing device browser scenario (temporary port-remapped copy, owned
  Node mock/Vite on 4210/4200): PASS, desktop and mobile. Covers pending input/live
  report, keyboard typeahead + Tab clearing, unchanged Target 1, empty-row save and
  reopen, save rejection/retry, and no horizontal overflow at 1440/768/375/390 px.
  Desktop/mobile screenshots inspected. Early keyboard attempts using Home or
  ArrowUp failed in macOS headless Chromium; standalone native-select checks
  confirmed typeahead is the supported method used in the final test.
- A Firefox attempt was interrupted during worker warm-up; Firefox/Safari behavior
  has no completed browser verification. No real-device checks performed.
- Final scoped diff reviewed against clean-code and Vue engineering rules:
  existing state ownership/contracts retained, new callbacks documented, no
  unrelated production changes or unresolved findings in this fix.
- Final gates: NH-01–06 and FE-01–05 PASS for this scoped fix using the evidence
  above; this does not assert completion/approval of the wider branch.

- E2E TypeScript check: PASS. Existing frontend source graph refreshed with AST
  extraction, then source/merged graph reports and HTML regenerated. Spec documents
  are outside the indexed frontend/backend roots.
