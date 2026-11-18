# Plan: Capability-driven wireless observation sources

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft for maintainer review | Updated: 2026-11-11
Constitutions: parent 1.0.0 (Proposed); backend 1.0.0, frontend 1.0.0 (Proposed)
ADRs: [001](adr-001-observation-capability-mapping.md),
[002](adr-002-observation-storage.md), [003](adr-003-observation-viewer.md)

Repository baseline: `4941a09d` (branch `perf/e2e-faster-setup`).
No ratification, reviewer approval, or completed verification is claimed. Every
"Pending" result below is unrun.

Artifact set: `spec.md`, this `plan.md`, `tasks.md`, ADRs 001–003, and
`contracts/observations.md`.

## 1. Goal

Any enrolled device that can observe a radio medium — today the ESPHome Bluetooth
proxy, later a WiFi-probe or USB-BLE source — contributes to **one** shared observation
surface. Adding a source kind must require no change to the viewer, navigation, filters
or triage UI.

Explicit non-goal: replacing or extending the mm-wave spatial system. This is a
separate feature with different physics and different fidelity.

## 2. What already exists (and what to copy)

### 2.1 The capability pattern — copy this

`frontend/src/domain/spatial/index.ts:188` derives a capability from **data shape**, not
a provider whitelist:

```ts
/** Whether a device exposes enough readable entities to offer the spatial view. */
export function supportsSpatialMapping(device: Device | undefined): boolean {
  if (!device?.provider) return false;
  const entities = Object.values(device.exposes ?? {});
  return entities.filter(isSpatialCoordinate).length >= 2 && entities.some(isSpatialFlag);
}
```

consumed by `deviceTabs()` at `frontend/src/domain/devices/page.ts:36`. This is the
architectural precedent: an abstract predicate, no `provider === 'x'` branch.

### 2.2 The composition point — copy this

`frontend/src/integrations/registry.ts:17` binds every installed capability in exactly
one place (`const presenters: readonly IntegrationPresenter[]`) and every lookup falls
back safely via `presenters.find(...)`. New capability = one array entry, zero consumer
changes.

### 2.3 The optional-capability idiom — copy this

`backend/internal/deviceproviders/manager.go` uses progressively-enhanced optional
interfaces (`HealthReporter`, `IdentityBinder`, `InventoryBinder`) type-asserted at
runtime; `driver.go:67-73` shows the per-instance wiring. The observation seam reuses
this rather than widening an existing interface.

### 2.4 The bounded-worker idiom — copy this

`backend/internal/spatial/presence_engine.go:19-28` — bounded queue (`defaultQueueSize =
64`), `publishGap`, `counterInterval`, `workerTick`, and a `Dropped` callback. The
observation aggregator follows it closely.

### 2.5 The wart to avoid repeating

`isESPHomeDevice()` (`frontend/src/domain/devices/index.ts:78`) and
`deviceTabLabel(tab, isESPHome)` (`frontend/src/domain/devices/page.ts:65`) gate on
`provider === 'esphome'`. Correct today because ESPHome is the only provider, but it is
precisely what makes a feature provider-specific. **This plan adds no new provider-name
gate.** The two existing sites are out of scope and recorded as debt.

## 3. Design

### 3.1 Layers

```
ESPHome session (BLE proxy advertising RPC, proto msg id 93)
        │  optional Observer interface, type-asserted
        ▼
internal/observations          ← transport-agnostic; identity, triage, aggregation
        │                        imports only models/* + utils
        ├── ObservationSink ──→ HubController → ws.Broadcast("observations")
        ├── IdentityStore  ──→ AppConfigCache (settings.db)
        └── bounded queue (drop-oldest, counted)
        ▼
internal/deviceproviders/esphome   ← protobuf + AD-structure decoding only
```

Dependency direction is enforced: `internal/observations` must not import
`internal/deviceproviders/*`, `internal/mqtt` or `internal/api` (BE-01). Enforce with a
compile-time import test or reviewer check.

### 3.2 Transport-agnostic domain types

`backend/models/observations/observation.go`:

```go
// Medium names the physical channel an observation arrived over. It is a closed set so
// consumers can switch exhaustively; unknown media are rejected at the trust boundary.
type Medium string

const (
	MediumBLEAdvertisement Medium = "ble.adv"
	// MediumWiFiProbe is reserved for a future source. Declaring it here keeps the
	// extension path additive; no implementation exists yet.
	MediumWiFiProbe Medium = "wifi.probe"
)

// Observation is one sensed transmission, normalised away from its transport.
// Payload holds transport-specific details and is never persisted.
type Observation struct {
	Source     string         `json:"source"`     // enrolled provider device id
	Medium     Medium         `json:"medium"`
	Identity   Identity       `json:"identity"`
	RSSI       int            `json:"rssi"`       // dBm, as reported by the observing source
	Payload    map[string]any `json:"payload,omitempty"`
	ObservedAt time.Time      `json:"observedAt"` // UTC
}

// Identity is the stable handle a viewer groups sightings by. Stable is false for
// address-derived identities that rotate, so consumers never present them as arrivals.
type Identity struct {
	Key    string `json:"key"`              // "ibeacon:<uuid>:<major>:<minor>" | "addr:..." | "fp:<hash>"
	Kind   string `json:"kind"`             // ibeacon | eddystone | bthome | adv
	Vendor uint16 `json:"vendor,omitempty"` // BLE company id; 0 when absent
	Stable bool   `json:"stable"`
	Label  string `json:"label,omitempty"`
}
```

**Extension rule:** a new transport adds a `Medium` constant plus a decoder. It adds no
viewer, route, navigation item or filter component.

### 3.3 Driver seam (`Observer`)

```go
// Observer receives normalised observations from a session. It is optional: a driver
// that cannot observe a medium simply does not implement it, and Session.Run sends no
// observation subscription.
type Observer interface {
	Observe(ctx context.Context, batch []observations.Observation) error
}
```

Why not widen `Sink` (`session.go:24`: `Inventory`/`Ready`/`Observe`)? Because that
forces every sink to implement an unrelated concern — an SRP violation and a breaking
change to test doubles. Capability detection becomes structural:
`if _, ok := instance.(Observer); ok`. No provider string comparison.

### 3.4 ESPHome decoding boundary

All ESPHome-specific work stays in the provider package:

- subscribe: `subscribe_bluetooth_le_advertisements` (`api.proto:50`)
- decode: `BluetoothLERawAdvertisementsResponse` (`api.proto:1668`, id 93) containing
  `BluetoothLERawAdvertisement` (`api.proto:1659`: `address`, `rssi`, `address_type`,
  fixed 62-byte `data`)
- parse AD structures from `data`: `0x01` flags, `0x09` name, `0x16` service data,
  `0xFF` manufacturer data
- normalise → `observations.Observation`, hand the batch to `Observer`

Use the **Raw** message. `BluetoothLEAdvertisementResponse` (`api.proto:1641`) is
deprecated and removed in ESPHome 2025.8.0 (noted at `api.proto:1640`), so building on
the parsed variant would create immediate debt.

### 3.5 Identity resolution — the correctness hinge

"Show only new devices" is only meaningful if keys are stable. Resolution order
(`internal/observations/identity.go`):

1. payload identity — iBeacon `UUID+Major+Minor`, Eddystone `Namespace+Instance`,
   BTHome device id → `Stable: true`
2. public / static address → `Stable: true`
3. rotating random address (BLE `address_type`, top bits `01`) → `fp:<hash>`,
   `Stable: false`

`Stable: false` identities go to a separate ephemeral lane and are never counted as
arrivals. Without this, consumer phones re-key every ~15 minutes and the triage list is
dominated by fake arrivals — the feature becomes worthless.

### 3.6 What is stored where — direct answer

| Data | Store | Rationale / precedent |
| --- | --- | --- |
| Triage decisions, labels | `settings.db` via `AppConfigCache` | user intent must survive restart; mirrors `DeviceConfigCache.HasOverride` (`store/store.go:157`) |
| Live observations + render buffer | in-memory bounded ring | high cardinality, high volume, ephemeral by nature |
| Optional capture/replay | dedicated BoltDB in `GetDataDir()` | mirrors `spatialhistory.NewStoreAt` + `Prune` |
| Anything that becomes a real entity | normal pipeline → `metrics.db` | only aggregated values, never per-identity |

**Not `metrics.db` for observations.** `metrics.db` is a per-expose time series:
`AggregateExposeValue` assumes known expose names and `DefaultMetricsCleanupTask`
prunes it as entity history. Raw observations break all three assumptions —
cardinality (every rotating address is "new"), volume (100–300 adv/s per anchor × N
anchors dwarfs home telemetry), and shape (a stream, not entity values). ADR-002
records the alternatives.

**Critical constraint:** never create a device or entity per observed identity. That
would push strangers' MACs into the device registry, into metrics, and out through
`/api/context/devices` to MCP. Aggregate only. This is the same reasoning
`docs/external-integrations-investigation.md` §1.3 applies to add-ons.

### 3.7 Aggregation and broadcast

Aggregator owned by `internal/observations`, modelled on `presence_engine.go`:

```go
// Deps are the aggregator's collaborators. Every field is injected; the service owns no
// global state and starts no work until Start is called.
type Deps struct {
	Context context.Context
	Store   IdentityRepository
	Clock   utils.Clock
	Publish func(Snapshot)
	Queue   int            // 0 uses defaultQueueSize
	Dropped func()         // observable drop counter (NH-06)
}
```

Observations **do not** traverse `HubController.Ingest` — ingress is a 1000-deep lane
and per-device lanes serialise by design (`hub-controller.go:45,647,795`). They get a
bounded queue that drops oldest and counts drops.

Composition mirrors the spatial recorder at `hub.go:137-143`:

```go
observationService := observations.NewService(observations.Deps{
	Context: ctx, Store: store.AppConfig(), Clock: utils.NewRealClock(),
	Publish: func(s observations.Snapshot) { _ = ws.Broadcast(ws.Observations, s) },
})
providerManager := startDeviceProviders(ctx, hub, observationService)
```

### 3.8 Frontend capability and composition

`frontend/src/domain/observations/capability.ts`:

```ts
/** Whether a device can observe at least one radio medium. */
export function supportsObservation(device: Device | undefined): boolean;

/** Whether any enrolled device offers an observation source; gates navigation. */
export function hasObservationSources(devices: readonly Device[]): boolean;
```

Both read a declared capability field, never a provider name. Navigation stays in
`frontend/src/components/shell/navigation.ts` (single source for sidebar, header titles
and keyboard order), and the existing ordering test
(`frontend/src/__tests__/components/shell/navigation.test.ts`) is extended.

Facets are **declared by the source**, not hardcoded in the component:
`ObservationFilters.vue` renders whatever facet descriptors the snapshot supplies, so a
future medium adds a facet without touching the component. This is the frontend mirror
of §3.2's extension rule.

### 3.9 Placement

| Surface | Level | Gate |
| --- | --- | --- |
| Nav "Wireless" (after Devices) | hub | `hasObservationSources` |
| `/wireless` viewer (radar + table + triage) | hub | same |
| Device **Settings** tab → "Emit observations" | device | `supportsObservation` |

Navigation rather than Settings: this is a live instrument like the spatial live view,
not configuration. Per-device emission sits beside the existing spatial controls on the
device page, mirroring how the Presence tab is per-device while the workspace is global.

### 3.10 Module layout

```
backend/models/observations/{observation.go,snapshot.go}
backend/internal/observations/{service.go,identity.go,aggregate.go,store.go,service_test.go,identity_test.go}
backend/internal/deviceproviders/esphome/{advertisement.go,advertisement_test.go}
backend/internal/ws/eventhub.go                      (Observations event + actions)
backend/models/settings/config.go                    (additive identity map)
frontend/src/contracts/observations.ts
frontend/src/types/observations.type.ts
frontend/src/store/modules/observations/{index.ts,state.ts}
frontend/src/domain/observations/{capability.ts,identity.ts,facets.ts,projection.ts}
frontend/src/components/wireless/{WirelessViewer.vue,ObservationRadar.vue,ObservationTable.vue,ObservationFilters.vue,IdentityTriage.vue}
frontend/src/components/shell/navigation.ts
frontend/src/router/index.ts
specs/064-wireless-observations/**
```

### 3.11 Reuse vs deliberate divergence

Share: `spatial/view.ts` view-id pattern; `useSpatialView.ts` composable shape;
`radarViewBox`/`rotatePoint` (`domain/spatial/radar.ts`); `utils/units.ts` for units;
the cadence discipline from `presence_engine.go`.

Do **not** share `SpatialCanvas.vue`. It is mm-Cartesian with zones, workspace geometry,
sensor cones and heat cells; the wireless view is polar RSSI rings with no workspace.
Forcing one into the other corrupts both (ADR-003).

Honest physics note for the viewer: RSSI→distance is log-distance with 1–3 m accuracy
indoors, and 10 dB ≈ 3× distance, so per-packet jitter of ±6 dB moves a blip 2×. The
canvas therefore renders filtered, aggregated positions; raw packets live in the table.

## 4. Spec requirements (to be committed as spec.md)

### User stories

- US-01 (P1) — As the operator, I see one wireless surface listing every identity any
  enrolled source has heard, so I can tell what is actually in range.
- US-02 (P1) — I mark identities as known or ignored so the list isolates genuinely new
  things.
- US-03 (P2) — I filter by declared facets (kind, vendor, source) without recapturing.
- US-04 (P3) — A new observing source kind appears in the same view with no UI change.

### Acceptance criteria

- AC-01: Given an ESPHome device that implements `Observer`, when the session starts,
  then the advertising subscription is sent and normalised observations reach the
  aggregator.
- AC-02: Given an ESPHome device that does not implement `Observer`, when the session
  starts, then no subscription is sent and behaviour is identical to today.
- AC-03: Given an iBeacon advertisement, when decoded, then identity is the payload
  `UUID+Major+Minor` with `stable: true`.
- AC-04: Given a rotating random address, when decoded, then identity is `fp:<hash>`
  with `stable: false` and it is excluded from the arrivals count.
- AC-05: Given sustained observations beyond queue capacity, when the queue is full,
  then the oldest are dropped, the drop count increments, and memory stays bounded.
- AC-06: Given no new observations, when the publish interval elapses, then no snapshot
  is broadcast.
- AC-07: Given any hub session with ≥1 observing source, when a snapshot is produced,
  then one `observations` WS event carries source descriptors and identities.
- AC-08: Given an identity marked ignored, when it is observed again, then it is
  excluded from the broadcast snapshot and the decision survives restart.
- AC-09: Given a device declaring no observation capability, then no observation UI is
  offered for it.
- AC-10: Given no observing source is enrolled, then the "Wireless" navigation item is
  absent.
- AC-11: Given a malformed snapshot payload, when it reaches the browser, then it is
  rejected at ingress and application state is unchanged.
- AC-12: Given `observedAt` values, then all timestamps are UTC with second-or-better
  resolution and the browser renders in local time, distinguishing "no observation"
  from a zero value.

### Non-functional

- NFR-01: queue and snapshot cardinality bounded; drop observable (NH-06).
- NFR-02: snapshot cadence ≤ 4 Hz regardless of inbound rate.
- NFR-03: no new runtime dependency on either stack.

### Open questions

All three are resolved (spec.md §Questions and assumptions).

- **Q1 — resolved:** additive `Capabilities []string` on `devices.Device`. Declaration
  beats observation for determinism. Blocks nothing further.
- **Q2 — resolved:** triage decisions are global, matching the current no-role auth model.
- **Q3 — resolved:** optional capture retention is 7 days with `Prune`, matching
  `spatialhistory.Retention`.

## 5. ADRs

- **ADR-001 — Capability mapping (accepted).** Options: provider-name gate; device
  capability field; provider registry; snapshot-derived. Decision: additive
  `Capabilities []string` field plus an optional `Observer` interface.
- **ADR-002 — Observation storage (accepted).** Options: `metrics.db` per identity;
  settings-only with memory ring; dedicated BoltDB capture. Decision: triage in
  settings, live data in memory, optional 7-day capture.
- **ADR-003 — Viewer geometry (proposed).** Options: reuse `SpatialCanvas.vue`; shared
  primitive module; separate polar component. Recommendation: separate component sharing
  only primitives. Awaiting review.

## 6. Constitution Check

Planned checks are not passed checks; "PASS (design)" means supported by design
evidence only.

| Rule | Evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | spec scenarios AC-01…AC-12 precede code | PASS (design) | PENDING |
| NH-02 | No device state, command or acknowledgement added; observations never enter `DeviceProcessor` | PASS (design) | PENDING |
| NH-03 | New event on the existing authenticated socket; no new public route; household-adjacent data stays local, never in public device context/MCP | PASS (design) | PENDING |
| NH-04 | Optional-interface seam; no new dependency; ADRs 001–003 | PASS (design) | PENDING |
| NH-05 | Traceability §7 records actual commands and results | PASS | PARTIAL: suites pass; browser keyboard/viewport checks unrun |
| NH-06 | Bounded queue, drop counter, retention, shutdown cancels workers | PASS | PASS |
| BE-01 | `internal/observations` imports only `models/*` + `utils`; ESPHome decode stays in provider | PASS (design) | PENDING |
| BE-02 | `DeviceLifetimeService` untouched | N/A (no device-state change) | PENDING |
| BE-03 | Worker ownership, bounded queue, context propagation, `-race` | PASS (design) | PENDING |
| BE-04 | Additive settings keys, existing-data tolerance, no metrics schema change | PASS (design) | PENDING |
| BE-05 | No route change; WS contract documented; no MCP surface | PASS (design) | PENDING |
| FE-01 | New Vuex module for snapshot + triage; components presentational | PASS (design) | PENDING |
| FE-02 | Runtime validation in `src/contracts/observations.ts`; mocks updated | PASS (design) | PENDING |
| FE-03 | Empty/loading/offline/reconnect specified; no optimistic command state | PASS (design) | PENDING |
| FE-04 | Keyboard operation, accessible names, status not colour-only, 390/1280px | PASS (design) | PENDING (browser checks unrun) |
| FE-05 | No secrets; household data in memory only; no `VITE_*`; no browser persistence | PASS (design) | PENDING |

## 7. Traceability and verification

Actual results recorded 2026-11-11 on branch `feat/wireless-observations`. Commands were run
from `backend/` and `frontend/` respectively.

| IDs | Paths | Procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, AC-02 | `esphome/advertisement.go`, `_test.go` | structural capability: plain sink fails the `Observer` assertion; observing sink passes; subscription message sent | **PASS** |
| FR-02, AC-03, AC-04 | `internal/observations/ble.go`, `ble_test.go` | payload precedence, address-type classification, unusable and zero addresses, fingerprint stability, malformed AD | **PASS** (7 tests) |
| FR-03, AC-05 | `internal/observations/service_test.go` | sustained load beyond capacity drops and counts | **PASS** |
| FR-04, AC-06 | `service_test.go` with `mocks.NewMockClock` | registration publishes descriptors; unchanged state publishes nothing | **PASS** |
| FR-05, AC-07 | `internal/ws/eventhub.go`, `eventhub_test.go` | event and action registration | **PASS** |
| FR-06, AC-08 | `models/settings/observation_cache.go`, `_test.go` | round-trip, restart, unknown→delete normalisation, invalid input, storage failure, public-snapshot stripping | **PASS** (7 tests) |
| FR-07, AC-09 | `domain/observations/capability.ts` + test; `models/devices/device.go`; `esphome/driver.go` | capability predicate true/false; provider-agnostic proof | **PASS** (8 tests) |
| FR-08, AC-10 | `shell/navigation.ts` + test | gating visible/hidden and ordering | **PASS** (5 tests) |
| FR-09, AC-11 | `contracts/observations.ts` + test | malformed payload rejected; unknown facet tolerated | **PASS** (11 tests) |
| FR-10, AC-12 | `domain/observations/presentation.ts` + test | absent RSSI is null not zero; UTC parse | **PASS** (15 tests) |
| NFR-01, NFR-02 | `service.go` | bounded queue and capped cadence | **PASS** |

Command results:

- `go build ./...` — **PASS**
- `go vet ./...` — **PASS**
- `go test ./...` — **PASS**, 52 packages
- `go test -race ./internal/observations/... ./internal/deviceproviders/... ./models/settings/... ./internal/ws/... ./internal/` — **PASS**, no data races
- `npm test -- --runInBand` — **PASS**, 149 suites / 859 tests
- `npm run lint` — **PASS**, 0 errors (23 pre-existing warnings in untouched files)
- `npm run build` — **PASS**
- `npx vue-tsc --noEmit` — **PASS**

Unrun checks, stated as gaps:

- Browser keyboard traversal and focus order at 390px/1280px — **UNRUN**. FE-04 is not
  satisfied.
- A live ESPHome Bluetooth proxy check — **UNRUN** and requires separate authorisation.
  Assumption A1 (firmware reports raw advertisements, message id 93) remains unvalidated.
- `npm run type-check:e2e` — **UNRUN**; no e2e spec was added for this feature.

## 7a. Reassessment round (2026-11-11)

Recorded after deciding the hardware path. Full reasoning in the revised plan.

**Corrections to earlier statements in this plan:**

- Adding `esp32_ble_tracker` requires **one USB flash**, because the code partition must
  be resized. OTA alone is insufficient. The earlier "reflash over OTA" claim was wrong.
- `esp32_ble_tracker` defaults `continuous: true`, so the radio scans permanently.
  ESPHome documents an `on_client_connected`/`on_client_disconnected` pattern that scans
  only while a consumer is connected; that is the recommended configuration for a shared
  device.

**Defects found by re-reading the implementation:**

| ID | Severity | Defect | Resolution |
| --- | --- | --- | --- |
| D-1 | blocking | `Service.RegisterSource` had no production caller, so `snapshot.sources` was always empty and filters never rendered | Fixed: `sourceRegistry` interface, announce on subscribe, withdraw on exit |
| D-2 | high | Five distinct failure modes produced identical empty state; `Dropped` was never wired | Fixed: `SourceDiagnostics` counters per source, viewer states status and explains silence |
| D-3 | medium | No per-device observation affordance | Open as T019 |
| D-4 | high | Diagnostics were absent from `renderSignature`, so counters changed without a publish. A source receiving advertisements that decode to nothing would freeze the viewer at `0 advertisements` — defeating D-2 entirely | Fixed: counters participate in change detection; regression test asserts a diagnostics-only change publishes |
| D-5 | high | Identity RSSI was a lifetime maximum, so a device that moved away stayed frozen at its closest-ever reading, making the range estimate permanently wrong | Fixed: RSSI tracks the newest reading; test asserts -40 then -90 reports -90 |
| D-6 | medium | The advertised local name was decoded into `Payload` and never read, so the viewer could only show an opaque key | Fixed: name becomes the identity label, with the operator label still winning at snapshot time |
| D-7 | high | Diagnostics counted twice: the session reported raw counts and `ObserveAdvertisements` counted the forwarded batch again, so 10 advertisements reported 18 — corrupting the counter used to judge whether hardware works | Fixed: one owner for counting; `TestForwardingDoesNotDoubleCount` |
| D-8 | blocking | `NavSidebar` and `WirelessViewer` called vuex's bare `useStore()`, but `main.ts` installs the store under a `Symbol` injection key, so the store was `undefined` and the sidebar crashed on mount | Fixed: import the `@/store` singleton as every other component does. Confirmed by `layout.spec.ts`, which failed before and passes after |
| D-9 | medium | The observations module carried a `stale` flag and a `markStale` action that nothing dispatched, so the "Disconnected" notice was dead | Fixed: derive from the canonical `ws/getConnectionStatus` getter |
| D-10 | blocking | `Supervisor.Run` wraps the sink in `*supervisedSink`, which embeds the `SupervisorSink` interface. Embedding promotes only that interface's methods, so `Observer` was not promoted. The session's assertion failed and, because a failed assertion is a supported state, observation was silently dead: the device connected, states flowed, no advertisement was ever subscribed. Found only on real hardware | Fixed: `Unwrap()` plus a generic `capability[T]` helper; the same latent bug in the `Identify` assertion was fixed too |
| D-11 | usability | The viewer was a census of MAC addresses and not usable for anything. Found by using it | Fixed in the proximity round: isolation, distance with band, trend, Focus view |

**Lesson recorded:** D-8 was invisible to every unit suite — 867 tests passed while the
application crashed on first render. Only the browser spec caught it. That is the concrete
cost of FE-04 still being open, and the reason the browser check must not stay deferred.

**Verification after the reassessment round:**

- `go build ./...`, `go vet ./...` — PASS
- `go test ./...` — PASS, 52 packages
- `go test -race` on affected packages — PASS, no races
- `npm test -- --runInBand` — PASS, 149 suites / 867 tests (8 new)
- `npm run lint` — PASS, 0 errors
- `npm run build` — PASS
- Still unrun: browser keyboard/viewport checks (T020), live hardware validation (T022)

No new dependencies, no schema migration. `SourceDiagnostics` is additive and optional at
browser ingress, so an older backend payload still parses.

## 8. Task sequence

Preparation

- T001 Resolve Q1 (capability field) in spec; Q1 blocks T006–T009. Done: decision recorded.
- T002 Commit `specs/064-wireless-observations/` with spec, plan, ADRs, contract. Done: files linked.

Coverage and implementation

- T003 Red test: identity resolution table (`observations/identity_test.go`). IDs AC-03/04.
- T004 Implement `models/observations` + `identity.go`. Depends T003. Done: tests pass.
- T005 Red test: bounded queue drop + cadence with fake clock. IDs AC-05/06.
- T006 Implement `observations/service.go` + `aggregate.go` + `store.go`. Depends T004/T005.
- T007 Implement settings persistence + WS actions and event. IDs AC-07/08. Depends T006.
- T008 Implement ESPHome `Observer` seam + AD parser. IDs AC-01/02/03. Depends T004.
- T009 Wire composition in `hub.go` via functional option. Depends T006/T008.
- T010 Frontend contracts/types + capability domain + tests. IDs AC-09/11/12.
- T011 Vuex `observations` module + WS event mapping. Depends T010.
- T012 Viewer components (radar, table, filters, triage) + route + nav gate. IDs AC-10.
- T013 Per-device "Emit observations" settings row. ID AC-09.

Handoff

- T014 Run planned checks; record actual results. Gate NH-05.
- T015 Failure/recovery/accessibility checks. IDs AC-06/10/11; gate FE-03/04.
- T016 Docs, contract publication, final rule audit. Gates NH-01/NH-04/NH-05.

## 9. Debt, risks, exceptions

- Existing provider-name gates left in place (out of scope; recorded so they are not
  taken as precedent).
- `deviceproviders.Enrollment` requires a private numeric address and port
  (`endpoint.go`), so a future USB-BLE source needs an enrollment schema change; the
  `Observer` seam deliberately does not assume IP.
- No distance is claimed. RSSI ranging is 1–3 m indoors and inferior to the existing
  mm-wave system; the viewer labels it range-only and gain-aware.
- Privacy: identity keys are household-adjacent identifiers. Local storage and the
  authenticated socket only; never in public device context, MCP or browser
  persistence.
- No constitutional exceptions requested.

## 10. Completion

Acceptance evidence per §7, `contracts/observations.md` published, navigation and
capability tests green, reviewer approval recorded. Draft until reviewed.

## 11. Review round 2026-11-18 — the integration page, judged by using it

Four reports, all from one sitting with a live EPL and 571 advertisements. They are not
four cosmetic complaints; they have two root causes and one missing design idea.

### 11.1 What was reported, and what actually produces it

| Report | Actual cause |
| --- | --- |
| "0 devices heard" beside "571 heard · 571 decoded" | The census is captured once at `AddLocal` and never republished (ADR-005). It was captured while the `Active` gate was still closed, so it recorded zero identities. The counters are written by a different code path (`RecordSourceActivity`) that the gate does not touch, so they kept climbing. |
| "Automatic refresh? for a wireless sensor? how do I disable/enable it?" | `ConnectionSettings` decides `local` from `definition?.local`, but `useConnectionDetail` never fetches a definition for a local provider — `hasDocument` returns false precisely so it does not 404. `definition` is always `null`, so `local` is always `false`: the schedule fieldset renders, "Change address" renders, the address copy renders, and **Save schedule can only ever fail** with `unsupported_schedule`. |
| "What's the point of the results page?" | There isn't one, as built. It shows a frozen snapshot of a live source, so it is stale by construction and redundant with the radar by design. Its only raison d'être was that the page template has tabs. |
| "Home card has a white border" | `HomeObservationCard` styles with `--border-subtle`, `--surface-raised`, `--accent-strong` and `--text-muted`. **None of the four exists in this project.** The hardcoded fallbacks win on every theme, and `#dee2e6` reads as a white outline on a dark surface. |

### 11.2 Defects

| ID | Severity | Defect |
| --- | --- | --- |
| D-12 | blocking | A local provider's resource is stored, so it is captured once and freezes. Nothing republishes it: `Refresh` is hidden in the UI and the scheduler skips local instances. The Results page permanently disagrees with the Home tile, which reads the live socket. |
| D-13 | blocking | The `Active` gate is honoured by `ObserveAdvertisements` but not by `RecordSourceActivity`, so counters advance while identities are discarded. "571 heard, 0 devices" is a state the UI can present and a person cannot act on. |
| D-14 | high | `local` is unreachable on the client (`ConnectionSettings.vue:32`), so a continuous integration is offered controls that cannot act: Automatic refresh, Save schedule (always fails), Change address, and copy about an address and about "saved dates". |
| D-15 | medium | Enabling and disabling is one of four peer buttons labelled generically. The operator asked how to do it, which is the evidence that it is not findable. `UiSettingRow` + `UiToggle` already exist for exactly this. |
| D-16 | medium | The provider is displayed as the raw key `wireless_observation/hub`, because the definition is absent by design, not by accident. |
| D-17 | medium | Lifetime packet counters are presented in the same visual register as the device count. They answer "is the radio working", not "what is here", and pairing them with a device count invites reading them as results. |
| D-18 | high | The Home tile paints four undefined custom properties, giving an untokened near-white border on every theme. It also has no `.is-observation` layout rule, unlike `.is-presence*`, and carries a primary "View radar" button that no neighbouring tile has. |
| D-19 | medium | The Help tab has only the generic topics: `observationPresenter` declares no `help`, so nothing explains the `bluetooth_proxy` requirement, what a device count means, or why counters reset on restart. |
| D-20 | medium | Removing or disabling the integration stops the gate but the websocket snapshot in the store is never cleared, so the Home tile keeps showing the last census indefinitely. The "Not observing — add it as an integration to start" branch can only be reached before the first publish. |

### 11.3 Design decisions

**The census is read, not stored (ADR-005).** `instanceDTO` resolves a local provider's
resource through `Normalize` on every read; nothing writes `Instance.Snapshot` for a local
provider. Rejected: a refresh schedule (reintroduces the arbitrary cadence ADR-004
rejected); republishing on each observation snapshot (writes to disk at up to 4 Hz and
bumps `Revision`, which is the ETag behind every settings write, so renames and enable
toggles would race).

Two facts make this safe, and both are constraints on the implementation rather than
observations about it:

- `instanceDTO` is called inside `repo.View` (`service.go:485,509`), so normalization runs
  in a **read** transaction. That is only acceptable because the observation adapter never
  allocates an identity — it uses `context.ResourceID` and returns one reservation, which
  only the publish path commits (`publication_result.go:48`). T036 must not call
  `reserveResourceIdentities` from the DTO path.
- The lock order integrations→observations already holds: `publishLocal` calls
  `s.localSource()` under `s.mu` today. The reverse must stay impossible — `LocalEnabled`
  takes `s.mu`, and it is reached from `ObserveAdvertisements` at a point where no
  observation lock is held.

**`continuous` becomes data, not a client guess.** The backend puts one bool on
`InstanceDTO` derived from `Definition.Local`. This replaces three client-side rules that
currently disagree: `useConnectionDetail.continuous` (catalog origin),
`IntegrationEntry.continuous` (hardcoded `typeId === 'wireless_observation'`, with a
comment admitting it must move), and `navigation`'s consuming a passed-in flag. The
`hasDocument` guard stays, because it answers a different question (is there a definition
endpoint to read), but it stops being the definition of "continuous".

**`ConnectionSettings` takes `continuous`, not `definition`.** One prop decides the
address control, the schedule fieldset and both explanatory paragraphs, so they cannot
drift apart. For a continuous provider the settings pane states what it *is* rather than
silently omitting: "Reads every enrolled sensor continuously. Nothing to fetch, nothing to
schedule." A hidden control with no explanation is the same failure as a control that
does nothing.

**Enable/disable becomes the primary control.** A `UiSettingRow` with a `UiToggle`
labelled "Observing", help text saying what turning it off does (the sensor stops being
read; saved settings are unaffected). Save name stays; Remove stays at the bottom. No new
component and no new primitive.

**The Results tab becomes the live "now" view, and loses the counters.** Same component,
same contract, different content because the payload is now live: the device count as the
headline, how many are shown versus filtered, then source health moved into a `<details>`
named "Packets" — visible when it explains something (a silent source, or drops) and
collapsed otherwise. One primary action to the radar. Renaming the tab to "Now" was
considered and rejected: it would churn the navigation contract and the e2e labels for a
word, and "Results" is still the accurate name for the section once it shows the truth.

**The Home tile joins the tile family.** `--nh-border`, `--nh-surface`, `--nh-accent`,
`--nh-text-muted`, `--nh-radius-l` and `--nh-shadow-1`, matching `IntegrationTile`; the
whole tile becomes the link through an overlay `::after` like `IntegrationTile`'s, so the
shadowed button disappears. This is not a restyle for taste: the four undefined names are
a bug that shipped a white border.

**The gate closes visibly.** The observation service publishes one clearing snapshot when
the gate goes from open to closed, and `Snapshot()` returns the empty view while closed.
The Home tile then says "Not observing" because that is true, instead of showing the last
census forever. `RecordSourceActivity` also stops counting while the gate is closed, so
"heard" means "heard while observing" — which is what a person reading the number assumes.

### 11.4 Requirements added to spec

Recorded as AC-23…AC-27 in [spec.md](spec.md).

### 11.5 Tasks

Sequenced in [tasks.md](tasks.md) as T035…T045. Backend first (the payload everyone
reads), then the settings pane, then the surfaces, then the token guard.

### 11.6 Rule checks

- BE-01: `internal/observations` still imports only `models/*` + `utils`; the new gate
  handling adds no reverse dependency. PASS (design).
- BE-03: read-time resolution runs under the existing service lock; the lock order
  integrations→observations is one-way and must be asserted in a comment and a test.
  PASS (design).
- BE-04: no schema change. A record written earlier by this branch keeps a snapshot that
  is now ignored for local providers. PASS (design).
- FE-01: presentational components keep taking props; the new `continuous` field is
  validated at ingress with the rest of the DTO. PASS (design).
- FE-02 / FE-04: existing e2e for this page is updated, not bypassed; the changed surfaces
  are re-checked at 390px and 1280px. PENDING (unrun).
- NH-05: every task below records its actual command and result when it is done. PENDING.
- NH-06: the clearing snapshot is bounded like every other publish. PASS (design).

### 11.7 Deliberately out of scope

- The tab set itself (Results / Setup / Help) for a continuous provider. Discussed in
  §11.3; left alone.
- Any change to classifier behaviour, radar geometry or triage. Spec 065 owns those.
- The per-device observation row (T019), still open and still not blocking.
- Migrating the remaining undefined-token sites elsewhere in the tree; T045 proposes a
  guard that will report them without fixing them.

## 12. Review round 2026-11-18 (second) — navigation and a rendering bug

Reported after using the first round's result:

- "we should be able to open the radar by clicking the integration card as well — else I
  need to click the card, go to the integration page, then click open radar."
- "when disabled if you are in radar page we should have some link to the integration page
  to setup, eg enable/disable. like shortcuts."

Reviewing those screenshots turned up two further defects, both rendering bugs the report
did not name:

| ID | Severity | Defect |
| --- | --- | --- |
| D-21 | high | The board entry showed `Devices heard  18/11/2026, 10:48:34` — a timestamp where the device count belongs. `IntegrationEntry` renders `row.date`; this presenter wrote the count into `row.value`, so the count was never displayed and the same timestamp appeared twice (once here, once in "Checked"). Every other presenter puts the value in `date`, which is why no other card has this. |
| D-22 | medium | Results rendered "23devices in range". The `<strong>` and its following text node were separated in the template, and Vue condenses that whitespace, so the unit ran into the number. |

Design decisions:

**The capability names its own destination.** `IntegrationWidget` gains an optional
`destination`, so a presenter can point its board entry at its live surface. The radar is
the thing a person wants; the integration page is a summary of it, and requiring two
navigations to reach the obvious destination was the complaint. `IntegrationEntry` gained
no type branch — it renders whatever the capability declared, and falls back to the
integration page.

**The card keeps the settings reachable.** If the entry leads elsewhere, the integration
page would become unreachable from the board. A settings action appears on the entry
exactly when the capability has redirected it, so the exchange is one click for the radar
and one click for settings, rather than two for the radar and one for settings.

**The whole row is the target.** The name link gets an overlay (`::after`, as the Home
tiles do) so the summary, the icon and the empty space all open the same destination. The
action buttons sit above it, or Refresh would be unreachable.

**The radar explains "off" and links to the fix.** `observing === false` renders a notice
with a link to the instance's Setup section, and the header always carries an
"Integration settings" action. Naming it that rather than "Settings" is deliberate: the
sidebar already has a Settings destination, and two identically-named links is an
accessibility problem as well as a confusing one.

Ownership: `spec.md` US-05 AC-28…AC-30, `tasks.md` T046…T049.

### 12.1 Verification (actual results)

- `backend/`: unchanged this round; `go build ./...` and `go test ./...` still pass.
- `frontend/`: `npm test -- --runInBand` PASS 153 suites / 1001 tests; `npm run lint`
  PASS 0 errors (23 pre-existing warnings); `npm run build` PASS;
  `npm run type-check:e2e` PASS.
- Playwright desktop: `ConnectionsPage` + `wireless` + `home` PASS 74/74. Mobile:
  `ConnectionsPage` + `wireless` PASS 60/60.
- The new disabled-state browser test initially leaked state: the mock is one long-lived
  process, so turning observation off left every later spec looking at an empty radar. It
  now restores the flag in `afterEach`, and the fixture rebuilds each broadcast so the
  toggle takes effect rather than replaying a snapshot captured at connect time.

## 13. Review round 2026-11-18 (third) — tense and a phantom definition

| ID | Severity | Defect |
| --- | --- | --- |
| D-23 | high | Results read "23 devices in range" directly above "No sensor is reporting". Both lines were true of a different moment: identities outlive their source by up to ten minutes by design (`UnregisterSource` deliberately keeps aggregates so a proxy restart does not blank the view), so the count was describing the recent past in the present tense. |
| D-24 | high | "Provider not found" appeared on a local provider. `hasDocument` consulted the catalog, and `provider?.origin !== 'local'` is TRUE when the catalog has not loaded, so any route that skipped the catalog — a Home card, a deep link, or the post-disable reload — requested a definition that cannot exist. |

Design decisions:

**Tense follows the data, and both sentences are computed together.** The headline and the
status note moved into `censusSummary` in `domain/integrations/observation.ts`. Building
them in one place is the fix, not a tidiness preference: the contradiction came from two
independent expressions, each of which had a locally reasonable wording. One function that
sees both facts cannot emit "in range" and "nothing is reporting" together. The count is
still shown — it is true, just not current — and says "heard earlier", with one note
explaining the ten-minute drop-off.

**One status note, not a stack of conditionals.** The old markup had a silent-source
paragraph, a no-source paragraph and a `v-else`, and a silent source alongside identities
from a departed one fell between them. A single computed note with explicit cases cannot
leave a combination undescribed.

**A provider's own flag answers "does it have a document", not the catalog.** `continuous`
comes from the compiled definition, so it is present in every instance payload and needs no
second request. The catalog is kept only as a fallback for a backend that predates the
field, which by definition has no local providers.

Verification: `spec.md` AC-31…AC-32, `tasks.md` T050…T051, with the actual results recorded
there. The fixture was also corrected: it described one identity on a source with no
delivery counters, which is exactly the contradictory shape the page now refuses to
describe, so it could no longer assert "in range" for a sensor that had never delivered.
