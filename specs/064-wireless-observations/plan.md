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

- Accordion keyboard traversal and focus order at 390px/1280px in a real browser — **UNRUN**.
  FE-04 is therefore not yet satisfied.
- A live ESPHome Bluetooth proxy check — **UNRUN** and requires separate authorisation; no
  test depends on household hardware (NH-05). Assumption A1 (firmware reports raw
  advertisements, message id 93) remains unvalidated.
- `npm run type-check:e2e` — **UNRUN**; no e2e spec was added for this feature.

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
