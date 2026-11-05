# Plan: Presence signals for automations

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: unreviewed draft | Updated: 2026-11-03
Constitutions: parent 1.0.0 (proposed); backend 1.0.0 (proposed); frontend 1.0.0 (proposed)
Decision record: [ADR-001](adr-001-derived-presence-entities.md)

## Context and design

Spec 040 emits `devices.SpatialFrame` from `DeviceLifetimeService.publishSpatialTargets` on the
device lane. Zone geometry lives in `models/settings` (`SpatialWorkspace.Zones`), reachable
through `DeviceConfigCache`. The hub already has a provider seam that creates a device
lifetime and routes reports into automation, history and WebSocket output:
`DeviceProcessor.RegisterProviderDevice`, `UpdateProviderDevice`, `ProviderOffline`,
`RemoveDevice`.

**New package `backend/internal/spatial`** owns the whole feature and depends only on models
and utils, so controllers can use it and the pure rules stay transport-free:

- `rules.go` — pure functions and state machines shared by nothing else, mirroring the browser
  rules that matter for occupancy: One Euro smoothing per axis, deadband, plausibility gate,
  track association with merge, two-sample zone entry and 800 ms exit grace, per-zone counts
  and continuous dwell, last zone, empty seconds. Constants are named and documented with the
  same values as `frontend/src/domain/spatial`.
- `entities.go` — builds the derived device and its entity map from a source device, mapping
  and zones: stable ID `presence-<sourceId>`, provider `presence`, read-only exposes, zone
  entity names `zone_<12 hex>_*` with the zone name in the label and `zoneId` in attributes.
- `presence_engine.go` — one bounded work queue and one worker; `Observe(frame)` copies the frame and
  never blocks the device lane, `Sync(sourceId)` rebuilds zones/entities after configuration
  changes, `Remove(sourceId)` retires the derived device, `Close()` cancels the worker. A
  token-bucket publisher caps derived reports at one per second per source plus transition
  flush.

**Settings** gains `DeviceConfigCache.SpatialWorkspace(id)` returning an owned copy of the
saved workspace (zones only are consumed) under the read lock.

**Wiring (`internal/controllers/hub-controller.go`)**: the existing `OnSpatialTargets`
callback broadcasts and then hands the frame to the engine; the device-config listener calls
`engine.Sync(cfg.Id)`; device removal calls `engine.Remove(id)`; device creation calls
`engine.Sync(id)`, and startup reconciles every stored device once. The engine reaches the
existing device processor seam (`RegisterProviderDevice`/`UpdateProviderDevice`/
`ProviderOffline`/`RemoveDevice`) behind its own sink lock, so derived state reaches
automation, metric and WebSocket consumers exactly like a provider report. Derived devices are
projections: direct deletion is refused, and provider re-registration deliberately does not
preserve their binary values.

**Frontend**: zone previews prefer a derived device's occupancy/count entities and state the
source; the geometric hysteresis remains the fallback. The presence tab shows last zone and
empty time when available. Resolution of derived entities is a small domain helper that scans
the device list for the `presence-<id>` device, so no new WS contract is needed.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 041 and ADR-001 record outcomes, exclusions, decisions and questions | PASS | PASS (artifacts updated; questions non-blocking) |
| NH-02 | Derived reports reach automations from accepted state through the processor seam; availability follows the source; re-registration starts from unknown values and offline never presents frozen presence | PASS (design) | PASS (automation callback, hub integration, re-registration test; see debt for the processor seam vs lane wording) |
| NH-03 | No new route, no credentials, no network input; derived devices defined from local settings | PASS (design) | PASS (review: no route, credential or network input) |
| NH-04 | New package boundaries, processor seam and ADR alternatives recorded; no storage-schema change | PASS (design) | PASS (presence package imports models/utils only; no schema change) |
| NH-05 | AC-01…AC-10 map to engine tests, a hub automation scenario and browser checks | PASS (design) | PASS (checks recorded above; real-device check unrun and stated) |
| NH-06 | Bounded queue, ≤1 report/s/source, counters every 10 s, removal with mapping, no history growth at frame rate | PASS (design) | PASS (bounded queue, 10 s counters, removal with mapping) |
| BE-01 | `presence` package imports models/utils only; controller owns wiring; no transports in rules | PASS (design) | PASS (package layout and controller adapters) |
| BE-02 | Derived reports use the existing lifetime seed/update semantics; availability and removal explicit | PASS (design) | PASS (register/update/offline/remove tests) |
| BE-03 | One worker, bounded queue, context cancellation, `Close` joins | PASS (design) | PASS (`presence_engine_test.go` Close and bounded queue; `-race` clean) |
| BE-04 | No schema change; workspace accessor returns a copy; derived device stored like other providers | PASS (design) | PASS (`TestSpatialZonesReturnsOwnedCopy`) |
| BE-05 | No route/auth change; derived entities are not commandable | PASS (design) | PASS (read-only entities, no route change) |
| FE-01 | UI resolution helper in `domain/spatial`, presentation in the panel | PASS (design) | PASS (resolver in `domain/spatial`, presentation in panel) |
| FE-02 | No new wire contract; derived devices arrive as normal `deviceAdded`/`deviceUpdated` | PASS (design) | PASS (derived devices arrive as ordinary device events; no new contract) |
| FE-03 | Offline/unknown states labelled; geometric fallback says it is an overlay | PASS (design) | PASS (source labeled, fallback labelled as overlay) |
| FE-04 | Zone preview labels carry the source; keyboard/viewport checks in the browser spec | PASS (design) | PASS for badges/labels and existing viewport checks; screen-reader behaviour inspected only |
| FE-05 | No new persistence in the browser; no household data added to logs | PASS (design) | PASS (no browser persistence, no new logging) |

No rule exception is requested.

## Contracts, data, rollout

- New entities are ordinary device exposes; no WS/MQTT/MCP schema changes. Consumers see a new
  device with provider `presence`.
- Persistence: the derived device is stored like any provider device (source of truth remains
  the mapping/workspace); on restart it is re-registered and starts from unknown/zero rather
  than replaying positions (AC-07).
- Rollout: engine starts with the hub; disabling or deleting the mapping removes the derived
  device. Reverting the feature leaves ordinary devices untouched; derived devices are
  identifiable by provider and can be removed manually if the feature is reverted.
- Recovery: engine restart re-syncs; source offline marks the derived device offline.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, FR-04, AC-07, AC-09 | `backend/internal/spatial/entities.go`, `presence_engine.go`, settings accessor | `entities_test.go` identity/rename/read-only; `presence_engine_test.go` register/remove/offline/concurrency; `device_processor_test.go` no binary replay on re-registration; `hub-controller_test.go` registers on device sync, retires on mapping clear and broadcasts the new device list. AC-07 restart reconciliation is implemented (startup sync) but has no automated restart test | pass (AC-07 code-reviewed) |
| FR-02, AC-02, SC-01 | `backend/internal/spatial/rules.go` | `rules_test.go`: two-sample entry, 800 ms exit grace, boundary chatter toggles at most once | pass |
| FR-03, AC-03, AC-04 | `rules.go`, `presence_engine.go` | `rules_test.go`: duplicate slots count once, dwell accumulates and resets per episode, last zone retained, empty seconds grow; `presence_engine_test.go` counter flush ages them | pass |
| FR-05, AC-06 | `presence_engine.go`, controller wiring | `presence_engine_test.go` offline Sync calls Offline and frames do not publish; availability changes re-sync from the controller | pass |
| FR-01, AC-05, SC-02 | controller wiring + `backend/testing/hubharness` | `hub-controller_test.go` registers a provider source, lets the real lifetime emit frames, and asserts derived occupancy (no browser); `device_processor_test.go` asserts a derived occupancy update reaches the automation callback | pass |
| FR-07, AC-10, NFR-01–NFR-02 | `presence_engine.go` | `presence_engine_test.go`: queue drops while the worker is busy, change-only publication, one-second flush, Close stops the worker | pass |
| FR-06, AC-08, NFR-05 | `frontend/src/domain/spatial`, `SpatialPanel.vue` | `derived.test.ts` resolver tests; `e2e/devices.spec.ts` injects a derived device and asserts the tab reports its values as the authority; geometric fallback still rendered without one | pass |
| NFR-03, NFR-04 | ADR, entities, controller | review: no new route; every derived entity is read-only; derived IDs are ignored by the engine as sources and never enrollable | pass |

Checks actually run (2026-11-03): `go build ./...`, `go vet ./...`, `go test ./... -count=1`
pass; `go test -race ./internal/spatial/... ./internal/controllers/... ./internal/services/...
./models/settings/...` pass; frontend `npm test` — 133 suites / 718 tests (branch total); `npm run build`,
`npm run lint` (0 errors) and `npm run type-check:e2e` pass; `npx playwright test
e2e/devices.spec.ts` — 18 passed across desktop and mobile including the derived-presence
assertion; full suites — desktop 140 passed / 1 failed / 9 skipped, mobile 149 passed / 1 failed. The
single failure is the weather Connections spec, which reproduces unchanged on `main` and is
recorded as pre-existing/environmental.

One intermediate frontend run failed a history performance benchmark while a Playwright suite
held the machine; the same suite passed in isolation and in the clean final run, so it is
recorded as contention rather than a regression.

## Debt, risks, exceptions

- **Debt (display lag):** dwell, last zone and empty time are diagnostic entities, so the
  default observation policy throttles their `deviceUpdated` broadcasts to 30 s even though the
  engine reports every 10 s. The presence tab can therefore show counters up to 30 s old; this
  is recorded rather than hidden.
- **Debt (serialization):** the engine calls the processor seam from its own worker rather
  than the derived device's lane. Derived devices have no other writer (direct deletion is
  refused), the engine holds a sink lock across lifecycle and delivery, and a report is
  dropped if its state was removed, so reports cannot land after removal. This differs from
  the "normal lane" wording of earlier drafts.
- **Risk:** Go and TypeScript rules can drift. Mitigation: the same synthetic sequences and
  documented constants; the browser keeps smoothing for the dot while the hub owns occupancy.
- **Risk:** derived devices clutter the device list. Mitigation: provider/connection labels and
  the ADR note; hiding is a separate UI decision (Q-02).
- **Risk:** persistence of derived devices after a downgrade. Mitigation: documented removal,
  provider-tagged.
- **Debt (pre-existing):** ESLint does not cover `.ts`; TypeScript gates are `vue-tsc`.
- No exception requested.

## Completion

Implemented and checked after an independent review round; the blockers it raised were fixed:
derived values are not replayed on re-registration, startup reconciles stored devices, direct
deletion of a derived device is refused, the engine serializes lifecycle and reports with one
lock, and offline derived values fall back to the geometric preview. Evidence is recorded
above, with AC-07 and the processor-seam serialization called out as code-reviewed rather than
automated.

Known limits: derived devices appear in the device list labelled `presence`/`derived` (Q-02),
with delete hidden and direct deletion refused; dwell and empty time are diagnostic and only
recorded if a user opts in, and their display can lag up to the 30 s diagnostic broadcast
policy; Go rules mirror the browser by equivalent tests, not shared code (Q-04).
