# Spec: Presence signals for automations

ID: 041 | Status: Implemented (unreviewed) | Owner: unreviewed draft | Updated: 2026-11-03
Constitution: 1.0.0 (proposed) | Components: backend and frontend

## Problem, scope, exclusions

Spec 040 made the live radar trustworthy, but every derived value it shows — who is in a
zone, how many, which zone was used last, how long the room has been empty — is computed in
the browser (`frontend/src/domain/spatial/`). Automations read confirmed device state on the
hub, so none of it can drive a light. The firmware's own zone occupancy covers only its four
configured rectangles, with its own delayed-off, and only if the user configured them.

In scope: a server-side presence engine per mapped device that consumes the synchronized
frames from spec 040 and publishes one derived device with stable, read-only occupancy,
count, dwell, last-zone and empty-seconds entities; the presence tab reads zone occupancy and
counts from those entities so the screen and the automation agree.

Exclusions, each needing its own spec: polygon/zone write-back and firmware settings (P5),
calibration (P6), replay and persisted history (P7), diagnostic views (P8), Home dashboard
block, multi-sensor fusion, and any claim of person identity or people counting.

## Scenarios and acceptance

### US-01 — A zone becomes an automation trigger (priority P1)

Independent verification: backend engine tests with synthetic frames; a hub-harness
automation scenario that runs without a browser.

- AC-01: Given a mapped source device and at least one saved zone, when the hub starts or the
  mapping is saved, then one derived device exists with a stable identifier derived from the
  source and read-only entities for every zone; when the mapping or all zones are removed, the
  derived device is removed.
- AC-02: Given a track entering a zone, when occupancy turns true, then the zone's binary
  entity reports true and a configured automation triggers once; a single boundary-crossing
  sample does not toggle it (two-sample entry), and leaving for longer than the grace period
  turns it off once.
- AC-03: Given two slots reporting one target, when both are inside a zone, then that zone's
  count is one, not two.
- AC-04: Given occupancy, when it ends, then `zone_<hash>_dwell_seconds` has accumulated the
  continuous occupied time and resets on the next occupancy; `last_zone` keeps the last zone's
  name; `empty_seconds` is zero while any target is tracked and grows while the room is empty.
- AC-05: Given no browser session, when a target enters and leaves, then the derived values
  change and automations run (verified in a hub test that never opens a socket client).
- AC-06: Given the source device goes offline, when frames stop, then the derived device is
  offline and no occupancy is fabricated or frozen as current; when reports resume, values
  update again.
- AC-07: Given a hub restart, when the mapping and zones are loaded, then the derived device is
  re-registered with unknown/zero values rather than replaying pre-restart positions.
- AC-08: Given the presence tab, when a derived device exists, then zone previews show its
  occupancy and count and say so; without one they keep the geometric preview and say it is an
  overlay.
- AC-09: Given derived entities, when a user inspects them, then they are read-only (no
  command path), excluded from provider enrollment management, and their history does not grow
  at frame rate.
- AC-10: Given an engine per source, when frames arrive faster than it can process, then its
  queue is bounded, old frames are dropped, publication is limited to four reports per second
  per source (250 ms burst guard) with counter-only reports at most every 10 s, and shutdown
  stops all workers.

## Requirements and success

- FR-01: The hub MUST compute derived presence from synchronized frames for every mapped
  device with at least one zone. Acceptance: AC-01, AC-05.
- FR-02: Zone occupancy MUST use the documented hysteresis (two consecutive inside frames to
  enter, 800 ms grace to leave) and MUST be published on change. Acceptance: AC-02.
- FR-03: Counts MUST deduplicate targets by association, and dwell/empty counters MUST be
  derived from the same track state. Acceptance: AC-03, AC-04.
- FR-04: Derived entities MUST be stable across rename and restart, read-only, and removed with
  their source mapping. Acceptance: AC-01, AC-07, AC-09.
- FR-05: Availability MUST follow the source; offline MUST NOT present frozen presence as
  current. Acceptance: AC-06.
- FR-06: The presence view MUST prefer derived occupancy/count and label the source; the
  geometric preview remains the fallback. Acceptance: AC-08.
- FR-07: Engine work MUST be bounded and cancellable. Acceptance: AC-10.
- NFR-01: At most four derived publications per second per source (250 ms guard); counter-only
  reports at most every 10 s; occupancy and counts on change only.
- NFR-02: Memory bounded by sources × (zones + 3 tracks); no unbounded history.
- NFR-03: No new public route or credential; derived devices are defined locally from saved
  settings, never from network input.
- NFR-04: Derived devices must not be commandable and must not appear as enrollable providers.
- NFR-05: Changed UI stays keyboard operable and verified at 375/768/1440 px.
- SC-01: Engine tests reproduce boundary chatter, duplicate slots, count/dwell and empty-time
  behaviour with fixed clocks.
- SC-02: A hub scenario proves an automation triggered by derived zone occupancy without a
  browser.
- SC-03: Existing backend and frontend suites, lint, build, type-checks and affected browser
  specs pass, with results recorded.

## Entities and external interfaces

- **Derived device** (new, provider `presence`, connection type `derived`): identifier
  `presence-<source id>`; friendly name `<source name> (presence)`. Read-only entities per
  [ADR-001](adr-001-derived-presence-entities.md); zone entity names use the first 12 hex
  characters of the zone UUID with the name in the label and a `zoneId` attribute.
- **Frames** (spec 040): unchanged input. The engine never reads browser state.
- **Spatial settings** (existing): mapping and workspace zones; read through the settings cache.
- **Automations** (existing engine): unchanged; derived entities are ordinary device exposes.

## Questions and assumptions

- Q-01: Zone entity naming uses a hash prefix; a rename must not change names. Assumption:
  zone IDs never change once saved (true today: UUID at creation). Not blocking.
- Q-02: Should derived devices be hidden from the device list? This spec shows them, labelled,
  because they are the entities users automate on. Hiding is a separate UI decision.
- Q-03: Dwell accumulates continuously per occupancy episode; day totals are out of scope (P7
  history can aggregate them).
- Q-04: Engine filtering mirrors the browser rules but is not shared code; parity is checked
  with the same synthetic sequences, not by construction.
