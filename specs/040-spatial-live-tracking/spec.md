# Spec: Spatial live tracking

ID: 040 | Status: Implemented (unreviewed) | Owner: unreviewed draft | Updated: 2026-11-03
Constitution: 1.0.0 (proposed) | Components: backend and frontend

## Problem, scope, exclusions

The Spatial tab already renders live target positions, room outlines, furniture, trails,
zones and heat cells (spec 038, AC-17–AC-26; `frontend/src/domain/spatial/`,
`frontend/src/components/device/spatial/`). It is visually complete but hard to trust:
the dot jumps while a person stands still, a walking person zig-zags or swaps labels,
and there is no way to tell a fresh measurement from a stale one at a glance.

Observed facts in current code:

- `DeviceLifetimeService.publishSpatialFreshness` forwards only mapped X/Y values, per
  entity, at most one update per entity per 250 ms
  (`backend/internal/services/device_lifetime.go:173-206`). X and Y therefore arrive in
  separate messages with up to 250 ms skew; a point is composed from two different
  moments. Emitted under the general `deviceUpdated` event
  (`backend/internal/controllers/hub-controller.go:846`).
- The view composes a target only when the mapped axis entities were seen within the
  last 5 s and the mapped active flag is exactly `true`
  (`frontend/src/domain/spatial/index.ts:95-120`). Nothing is filtered or predicted.
- `SpatialCanvas.vue` draws the raw coordinate for each mapped slot (`:187`), and slots
  are fixed rows 1–3 in the saved mapping, not tracked objects.
- LD2450 target slots are reused, and zone target counts have been reported to include
  ghost targets from untracked slots
  ([esphome#15025](https://github.com/esphome/esphome/issues/15025)); the firmware offers
  no persistent target identity.

Assumptions (validate during implementation): household installation is wall-mounted at
roughly 1–2 m, LD2450 speed values are not mapped in `SpatialMapping`, and no consumer
outside the browser consumes spatial coordinates.

In scope, one increment:

1. Complete, synchronized target frames from the backend, replacing per-axis deltas.
2. Client-side normalisation: adaptive smoothing, deadband, plausibility gate, held and
   predicted presentation states.
3. Stable track identity with duplicate suppression, so labels and colours stay with a
   person instead of a slot.
4. Presence as a first-class device tab with a header status chip, so one radar serves the
   page instead of an always-on card duplicating the workspace.

Exclusions and deferred work (each needs its own spec):

- Derived zone occupancy/count/dwell signals for automations (highest next value; needs
  a decision on where derived presence entities live in the automation engine).
- Polygon zone write-back through the firmware's `Polygon Zone N` text entities, and
  further firmware settings beyond the existing rectangle path.
- Calibration wizard (freeze-and-drag transform, installation-angle write-back).
- Replay and persisted history (today analytics are session-only by design).
- Interference finder, jitter scatter, hold/off-delay tests.
- Home dashboard embedding: Home cards are built from `HomeBlock`/`HomeSourcePresentation`
  (`frontend/src/components/home/`), not arbitrary components; adding a radar block is a
  separate settings/UI contract. The device-page chip is the intended entry point.
- Multi-sensor fusion.
- Any claim of person identity, people counting, or improved measurement accuracy.

## Scenarios and acceptance

### US-01 — Positions that hold still and stay attached to the same person (priority P1)

Independent verification: unit tests over recorded frame sequences with fixed timestamps;
`frontend/e2e/devices.spec.ts` for the rendered states.

- AC-01: Given a saved mapping whose device reports at least 2 Hz, when a spatial surface
  is open, then composed points use one frame whose X and Y were read from the same
  device state, and no rendered point mixes axes more than 150 ms apart.
- AC-02: Given a person stationary for at least 10 s, when raw and filtered positions are
  compared over that window, then filtered root-mean-square horizontal movement is at
  least 50% lower than raw and the displayed point does not oscillate between two
  clusters more than 200 mm apart.
- AC-03: Given a person walking at about 1 m/s, when filtering is enabled, then the
  displayed position trails the filtered estimate by no more than 150 ms and the path
  contains no jump exceeding 0.5 m between consecutive frames.
- AC-04: Given a coordinate change implying motion faster than 3 m/s, when it disagrees
  with the previous sample, then it is not rendered as movement; the target is shown as
  predicted for at most 1 s and recovers only when consecutive samples agree, and a
  sustained disagreement is accepted rather than held forever.
- AC-05: Given the active flag turns false, when no further samples arrive, then the last
  position is shown as held and visibly distinct for at most 2 s, then removed; a held or
  predicted position is never labelled as a current measurement.
- AC-06: Given no frame has been received yet (older backend, or a device whose server does
  not send frames), when the sampled `deviceUpdated` path supplies positions, then the view
  works from sampled data and states that it is sampled. Once frames have arrived the view
  stays on frames: silence is normal for a still target, so it ages to `idle` rather than
  switching sources or dropping the position; a disconnect or offline state clears it.
- AC-07: Given two mapped slots within 350 mm reporting similar movement for three
  consecutive frames, when both would render, then one track is shown, the duplicate is
  reported in the diagnostics row, and the merge is reversible on separation.
- AC-08: Given a track disappears and reappears within 8 s, when it is re-acquired, then
  it keeps its label and colour; after 8 s or after the device goes offline, a new track
  label is issued.

### US-02 — Zone previews that do not flicker (priority P2)

- AC-09: Given a target hovering on a zone boundary, when consecutive frames alternate
  inside/outside, then the zone preview changes state at most once per second and the
  preview states which source (geometric preview or device report) it used.

### US-03 — Presence is a first-class part of the device page (priority P2)

- AC-10: Given a device that exposes mappable coordinates and an active flag, when the page
  loads, then a Presence tab is offered; a device without them has no such tab, and a link
  to the withheld tab falls back to Controls instead of an empty panel.
- AC-11: Given any spatial-capable device, when the Controls tab is shown, then no radar is
  drawn there; a header chip states offline, not set up, clear or the number of reporting
  targets, opens the Presence tab, and is operable by keyboard with a name that includes its
  state. A page never renders two live canvases.
- AC-12: Given the Presence tab at 375, 768 and 1440 px, when it renders, then status and
  save stay reachable in a sticky header, the view switcher and display controls sit above
  the canvas, target mapping is the last section, no horizontal overflow occurs, and status
  is not conveyed by colour alone.

### US-04 — Evidence that normalisation helps (priority P2)

- AC-13: Given a spatial surface, when the user switches raw/filtered or changes smoothing,
  then both are visualised from the same frames, a level change recomputes the position from
  recent samples immediately, and a readout states measured jitter.
- AC-14: Given the help content, when opened, then it explains slot-versus-track,
  held/predicted states, that smoothing trades lag for calm, and that coordinates are not
  measurements of a person's identity or a guarantee of accuracy.

## Requirements and success

- FR-01: The backend MUST emit complete target frames (per mapped target: active flag and
  both axis values with units) from accepted device state on the device's lane, at most
  one frame per 100 ms per device, with a trailing flush so the final suppressed change is
  not lost, and no frame when nothing changed. Acceptance: AC-01, AC-06.
- FR-02: The frontend MUST prefer frames for as long as they arrive and MUST use sampled
  coordinates only as a fallback, labelling which source is in use. Acceptance: AC-01, AC-06.
- FR-03: The frontend MUST filter position per track with adaptive smoothing, a deadband,
  a plausibility gate, gap reset, and bounded hold, exposing `live`, `predicted`, `held`
  and `idle` (aged) states. Acceptance: AC-02–AC-05.
- FR-04: The frontend MUST associate points to stable tracks using slot stickiness and a
  distance gate, MUST suppress duplicate tracks reversibly, and MUST issue new identities
  only after retention expires. Acceptance: AC-07, AC-08.
- FR-05: Zone previews MUST apply enter/exit hysteresis over filtered tracks and MUST
  identify their source. Acceptance: AC-09.
- FR-06: Presence MUST be a device tab, offered only when the device can map coordinates,
  with a header chip as the entry point and exactly one canvas per page. Acceptance:
  AC-10–AC-12.
- FR-07: The UI MUST show both raw and filtered positions on request with a jitter
  readout. Acceptance: AC-13.
- FR-08: Help text MUST describe states, limits and non-claims. Acceptance: AC-14.
- NFR-01: With one mapped device, the backend MUST send no more than 10 frames per second
  and no more than 1 KB per frame; frame emission MUST stop when the device is offline or
  removed, and pending flush timers MUST be cancelled on shutdown.
- NFR-02: Filtering and association MUST run in bounded time per frame for at most 3
  targets, and MUST NOT allocate unbounded history (retention only, no full session log).
- NFR-03: Positions MUST remain in browser memory only; no new persistence, no household
  position data in logs or metrics, no new public route (NH-03, FE-05).
- NFR-04: Presence surfaces MUST do no work while the document is hidden and MUST NOT
  animate when nothing changes.
- NFR-05: Changed surfaces MUST remain keyboard operable and verified at 375/768/1440 px
  with the recorded checks in the frontend constitution.
- SC-01: Synthetic-sequence unit tests demonstrate AC-02 (≥50% jitter reduction) and
  AC-03 (≤150 ms lag, no >0.5 m step) with fixed clocks and no device access.
- SC-02: Frame contract tests demonstrate bounded rate, trailing flush, no-change
  suppression, offline stop and shutdown cancellation.
- SC-03: Existing backend and frontend suites, lint, build, `type-check:e2e` and the
  affected Playwright desktop specs pass, with actual results recorded in the plan.

## Entities and external interfaces

- **Spatial target frame** (new WS event `spatialTargets`): producer
  `DeviceLifetimeService`; consumer browser `hub/updateSpatialTargets`; payload documented
  in [contracts/spatial-targets.md](contracts/spatial-targets.md). Additive: no existing
  event, field or persisted setting changes. Older browsers ignore it; newer browsers fall
  back to `deviceUpdated` when no frames arrive.
- **Spatial mapping** (existing, unchanged): rows of `x`, `y`, `active` entity names per
  target. Frames reuse it; no unit or identity is inferred from entity names.
- **Spatial workspace** (existing, unchanged): zones, outline, objects, sensor placement.
  The presence view and zone hysteresis read it; no new fields.
- **Track** (new, browser-only): display identity derived from frames. Never persisted and
  never presented as a person.

## Questions and assumptions

- Q-01: Is 100 ms the right frame cap, or should it follow the firmware update-speed
  setting? Default: constant 100 ms, documented; further tuning changes one constant.
  Owner: repository owner. Not blocking.
- Q-02: Should the smoothing preference persist in `SpatialSettings`? This spec keeps it
  session-local to avoid a settings contract change; persisting it is additive later.
  Blocking only for the persistence expectation, which is out of scope.
- Q-03: Home dashboard embedding is blocked by the Home block/source model; it needs a
  separate spec. Not blocking this increment.
- Q-04: Assumption — mapped X and Y entities declare a length unit, and the two units may
  differ; frames carry per-axis units so mixed units are handled without guessing.
- Q-05: Assumption — a track may hide a second person within 350 mm; the diagnostics row
  and the reversible merge must make that visible rather than silent.
