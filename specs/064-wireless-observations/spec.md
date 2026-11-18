# Spec: Wireless observation sources

ID: 064 | Status: Draft | Owner: agent draft for maintainer review | Updated: 2026-11-11
Constitution: 1.0.0 (Proposed) | Components: both (backend + frontend)

This spec states outcomes. It does not claim ratification, reviewer approval, or
implementation authorization.

## Problem, scope, exclusions

### Problem

NodeHerder can read Zigbee2MQTT devices, ESPHome provider devices, declarative HTTP
integrations, and mm-wave spatial presence. It cannot show what radio traffic an
installed device can *observe*. The immediate motivation is the ESPHome Bluetooth proxy:
it already receives BLE advertisements, but nothing subscribes to them.

Today the only way to see that data is a separate tool, so it never reaches the hub, and
any per-device wiring would repeat the provider-coupling mistake that makes a feature
specific to one vendor instead of a capability.

### Outcome

One wireless surface in the hub that lists whatever any observing source has heard, with
operator triage so genuinely new identities can be isolated from constant background
traffic. Adding a new observing source kind (WiFi probe, USB BLE dongle) requires **no**
change to the viewer, navigation, filters, or triage UI.

### Affected users, devices, data

- Users: household operator on an authenticated browser session.
- Devices: any provider device that can observe a radio medium. Today: ESPHome with
  `bluetooth_proxy`.
- Data: BLE advertisement metadata — address, address type, RSSI, AD structures
  (names, service data, manufacturer data). Household-adjacent identifiers.

### Exclusions

- Not part of the mm-wave spatial system. No mm coordinates, zones, or workspace.
- No distance or position claim. RSSI ranging is 1–3 m indoors and is strictly inferior
  to the existing radar; the viewer presents range bands, not metres, unless a source
  declares calibration.
- No device-registry, metrics-history, MCP, or public-context exposure of observations.
- No USB BLE support in this change. The seam must not preclude it.
- No transmission, spoofing, or active scanning.
- No passive observation of third-party networks as a product feature.

### Current-code evidence

- `backend/internal/deviceproviders/esphome/session.go:24` — `Sink` has
  `Inventory`/`Ready`/`Observe`; nothing subscribes to advertisements.
- `backend/internal/deviceproviders/esphome/session.go:290-304` — the read loop only
  dispatches state messages.
- `backend/third_party/esphome_api/proto/api.proto:50,1659,1668` — the subscription RPC
  and raw advertisement messages exist in the vendored proto, unused.
- `frontend/src/domain/spatial/index.ts:188` — capability derived from data shape; the
  precedent for this spec's capability approach.
- `frontend/src/domain/devices/index.ts:78`, `frontend/src/domain/devices/page.ts:65` —
  existing provider-name gates; not a precedent to extend.
- `backend/internal/spatial/presence_engine.go:19-28` — bounded-worker idiom to reuse.

### Assumptions (to validate)

- A1: the ESPHome proxy in use reports raw advertisements (`BluetoothLERawAdvertisementsResponse`,
  message id 93) and is on a firmware that supports the subscription. Unvalidated until
  a manual check.
- A2: identity keys are household-adjacent, so local-only storage and the authenticated
  socket are sufficient. Consistent with NH-03; no external sharing.
- A3: triage decisions are global, not per-user. Current auth has no role model
  (`backend/internal/auth/tokens.go`).

## Scenarios and acceptance

### US-01 — See what is in range (priority P1)

Independent verification: unit tests for identity resolution and aggregation; a manual
check with one authorised ESPHome proxy.

- AC-01: Given an ESPHome device that implements the observation capability, when the
  session starts, then the advertisement subscription is sent and normalised
  observations reach the aggregator.
- AC-02: Given an ESPHome device that does not implement the capability, when the
  session starts, then no subscription is sent and behaviour is identical to today.
- AC-03: Given an iBeacon advertisement, when decoded, then identity is the payload
  `UUID+Major+Minor` with `stable: true`.
- AC-04: Given an advertisement from a rotating random address, when decoded, then
  identity is `fp:<hash>` with `stable: false` and it is excluded from the arrivals
  count.

### US-02 — Isolate genuinely new identities (priority P1)

Independent verification: store round-trip tests; UI triage manual check.

- AC-08: Given an identity marked ignored, when it is observed again, then it is
  excluded from the broadcast snapshot and the decision survives restart.
- AC-09: Given a device declaring no observation capability, then no observation UI is
  offered for it.

### US-03 — Filter without recapturing (priority P2)

- AC-11: Given a malformed snapshot payload, when it reaches the browser, then it is
  rejected at ingress and application state is unchanged.
- AC-12: Given `observedAt` values, then all timestamps are UTC with second-or-better
  resolution, the browser renders in local time, and an absent observation is
  distinguishable from a zero reading.

### US-04 — Add a source kind without UI change (priority P3)

- AC-10: Given no observing source is enrolled, then the "Wireless" navigation item is
  absent. Given one is enrolled, it appears and lists only what that source provides.

### Reliability and boundaries

- AC-05: Given sustained observations beyond queue capacity, when the queue is full,
  then the oldest are dropped, the drop count increments, and memory stays bounded.
- AC-06: Given no new observations, when the publish interval elapses, then no snapshot
  is broadcast.
- AC-07: Given a hub with at least one observing source, when a snapshot is produced,
  then one `observations` WebSocket event carries source descriptors and identities.

### US-05 — The integration page tells the truth about a live source (priority P1)

Added 2026-11-18 after using the page against a live EPL with 571 advertisements. Amends
the presentation decided in [ADR-004](adr-004-local-provider.md); see
[ADR-005](adr-005-census-is-read-not-stored.md).

- AC-23: Given a local integration and a running sensor, when any surface reads it, then
  the device count matches the hub's current observation snapshot. No surface may show a
  count captured earlier, and the integration page, the board row and a Home tile must not
  disagree about the same moment.
- AC-24: Given no enabled local integration, when advertisements arrive, then neither the
  identity list nor the per-source counters advance, and the source is reported as not
  being observed rather than as silent.
- AC-25: Given a continuous provider, then the settings pane offers no address control and
  no refresh schedule, and states why where those controls would be. It offers one
  prominently labelled control for enabling and disabling.
- AC-26: Given a device count and a packet count, then they are not presented as a pair;
  packet counts appear under a labelled disclosure that states they are lifetime totals
  for the current run.
- AC-27: Given a Home observation tile, then it uses only custom properties the theme
  layer defines, presents as a link rather than as a card containing a button, and has the
  same container styling as the integration tiles beside it.
- AC-28: Given a wireless observation entry on the Integrations board, then opening it goes
  directly to the radar, and the whole entry is the target rather than its name alone.
  The integration page remains reachable from that entry in one action.
- AC-29: Given the radar page, then a way to the integration's settings is always present,
  and when observation is turned off the page says so explicitly and links to the control
  that turns it back on.
- AC-30: Given a board entry, then a value shown beside its label is that value, not a
  timestamp: counts, dates and totals must not be substituted for one another, and a
  timestamp appears once per entry.
- AC-31: Given a census holding devices whose sources have stopped reporting, then no surface
  states both that devices are present and that nothing is reporting, about the same moment.
  The count is presented as what was heard earlier, and the reason it is no longer growing is
  stated once.
- AC-32: Given a provider with no definition document, then no surface reports a missing
  definition as an error, whether the catalog has been read or not.

## Requirements and success

- FR-01: MUST allow a provider device to contribute observations without changing its
  existing inventory/state behaviour; acceptance: AC-01, AC-02.
- FR-02: MUST resolve an identity from payload data first, public address second, and a
  rotating address last, marking stability; acceptance: AC-03, AC-04.
- FR-03: MUST bound observation ingress with drop-oldest and an observable counter;
  acceptance: AC-05.
- FR-04: MUST publish only on change at a bounded cadence; acceptance: AC-06, NFR-02.
- FR-05: MUST deliver observations over the existing authenticated WebSocket;
  acceptance: AC-07.
- FR-06: MUST persist operator triage decisions across restart; acceptance: AC-08.
- FR-07: MUST derive source capability from declared capability, never provider name;
  acceptance: AC-09, AC-10.
- FR-08: MUST validate snapshot payloads at browser ingress; acceptance: AC-11.
- FR-09: MUST NOT create a device, entity, or metric series per observed identity;
  acceptance: AC-05, AC-11.
- NFR-01: Queue and snapshot cardinality MUST be bounded and drops observable (NH-06).
- NFR-02: Snapshot cadence MUST NOT exceed 4 Hz regardless of inbound rate.
- NFR-03: No new runtime dependency on either stack.
- NFR-04: Observation UI MUST meet FE-04 (accessible names, keyboard operation, status
  not conveyed by colour alone) at 390px and 1280px.

Units and time: RSSI in dBm; distances, if ever declared, in metres with the declaring
source's own accuracy; all timestamps UTC. Retention: live data in memory only; triage
decisions indefinite; any optional capture file 7 days (see ADR-002).

## Entities and external interfaces

| Entity | Owner | Meaning |
| --- | --- | --- |
| Observation | `models/observations` | One sensed transmission, normalised away from transport |
| Identity | `models/observations` | Stable grouping key plus stability flag and facets |
| Snapshot | `models/observations` | Bounded, change-driven broadcast payload |
| Identity decision | `models/settings` | Operator triage state: unknown, known, or ignored |

External interfaces: one additive WebSocket event (`observations`) on the existing
authenticated socket, plus one additive WS action pair for saving and deleting identity
decisions. Contract: [contracts/observations.md](contracts/observations.md).

Compatibility: additive only. No route change, no public surface change, no device
payload change beyond one optional capability field.

## Questions and assumptions

| ID | Question | Consequence | Owner | Resolution |
| --- | --- | --- | --- | --- |
| Q1 | Entry-point capability: optional field on `devices.Device`, or provider-side registry? | Affects the device model and T006–T009 | Maintainer | **Resolved 2026-11-11: additive `Capabilities []string` field** (see below). |
| Q2 | Triage decisions global or per-user? | Affects storage and future auth | Maintainer | **Resolved 2026-11-11: global.** Recorded as assumption A3. |
| Q3 | Optional capture retention: 7 days or unbounded with manual prune? | Affects storage growth | Maintainer | **Resolved 2026-11-11: 7 days plus `Prune`.** |

### Q1 resolution

`devices.Device` (`backend/models/devices/device.go:163`) gains one optional field:

```go
// Capabilities lists abstract features this device provides, e.g. "wireless.observe".
// It is declared by the provider at registration and never inferred from a provider
// name, so consumers stay provider-agnostic. Omitted when empty.
Capabilities []string `json:"capabilities,omitempty"`
```

The provider driver sets it during registration. Consumers read the field; no consumer
compares provider names.

Why not observe the capability from the live `observations` snapshot instead (nav gated
on `sources.length > 0`)? That needs no model change and auto-covers future sources, but
it answers "has this device observed?" rather than "can it observe?". An enrolled but
silent proxy would show no UI, which is wrong for a settings row and for an empty state.
Declaration is deterministic; observation is not.

Why not a provider-side capability registry? It reintroduces name coupling at every call
site and puts presentation concerns in the provider layer.

Compatibility: additive with `omitempty`, so existing device payloads are unchanged and
the browser type stays optional. `publicDevices` already filters `Provider != ""`
(`backend/internal/public_devices.go:28`), so provider devices and their capabilities
never enter the unauthenticated context or MCP surface.

An empty `Capabilities` field MUST NOT be read as "definitely cannot observe"; unknown
is treated as absent capability by consumers, consistent with progressive enhancement.
