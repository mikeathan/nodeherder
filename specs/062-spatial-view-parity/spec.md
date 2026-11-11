# Spec: Spatial view parity and presence card states

ID: 062 | Status: Draft | Owner: agent draft | Updated: 2026-11-10
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

Three related presentation gaps in the spatial surfaces:

- The Home presence card shows **Unknown** whenever the derived presence device exists
  but has no `occupied` value. That is the initial state before the hub's first spatial
  frame, not an unknown occupancy. The device header chip can say **Clear** from the last
  frame, so the card and chip disagree.
- History → Positions uses the shared viewer and canvas but never opts into the Distance
  control; the live view does.
- The Positions base view fits the recorded content and centers on its bounds, so the
  sensor is off-center. The live view centers on the sensor. The difference is obvious on
  a phone.

Facts observed in the current tree:

- `HomePresenceCard.vue` maps `values.occupied === null` to `Unknown`;
  `useDerivedPresence` returns null until the hub publishes `occupied` after a spatial
  frame. `PresenceRailTile.vue` does the same. `SpatialPanel.vue`'s `serverPresence`
  renders a null `occupied` as **clear**, which is less honest.
- `SpatialPanel.vue` passes `distance-option` / `show-distances` to `SpatialViewer` and
  `SpatialCanvas`. `SpatialPositions.vue` does neither, and its browser test explicitly
  asserts that the Distance checkbox is absent.
- `SpatialCanvas` already computes a target's distance from the same origin it draws the
  marker from, so replay distances are correct once enabled.
- `SpatialPositions.vue` `baseViewBox` centers on `(bounds.min + bounds.max) / 2`, not the
  sensor. The live panel uses `radarViewBox(scale)`, which is sensor-centered.

Scope: frontend presentation plus one pure geometry helper. No backend or protocol change.

Exclusions: changing the derived presence engine's initial state, adding a card fallback
to firmware occupancy, changing live tracking/filtering, persisting new view options, and
adding a range selector to Positions.

## Scenarios and acceptance

### US-01 — Honest presence card state (priority P1)

Independent verification: `src/__tests__/domain/spatial/derived.test.ts` and the Home
presence-card browser case.

- AC-01: Given the derived device exists but `occupied` is null, the Home card explains
  that presence has not been derived yet and does not claim Clear or Occupied.
- AC-02: Given `occupied` is true or false, the card still shows Occupied or Clear with
  count, last zone and empty seconds when present.
- AC-03: The live presence panel does not render a null `occupied` as "clear"; it reports
  that presence has not been derived yet.
- AC-04: The compact rail tile does not label missing derived data as a measured clear.

### US-02 — Replay distance parity (priority P1)

Independent verification: the Positions browser case in `e2e/devices.spec.ts`.

- AC-05: The Positions player offers the same Distance checkbox as the live view.
- AC-06: Enabling it appends each recorded target's distance from the sensor to its label,
  using the same drawn origin as the marker.
- AC-07: Disabling it restores the label; the Heatmap view does not show an inert Distance
  checkbox.

### US-03 — Sensor-centered replay (priority P1)

Independent verification: `src/__tests__/domain/spatial/radar.test.ts` and the Positions
browser case.

- AC-08: The Positions fitted view is centered on the sensor like the live view, at
  desktop and mobile widths.
- AC-09: The fitted viewBox keeps the canvas aspect and still contains the recorded
  content and the sensor.
- AC-10: Zoom, pan, reset-view and focus behavior are unchanged; the sensor remains
  visible at every zoom level.
- AC-11: No horizontal overflow at 375, 768 or 1440 px.

### US-04 — Compact presence overview tile (priority P1)

Independent verification: the Home presence overview browser case.

- AC-12: The overview tile uses the same Home tile type scale as its neighbours
  (0.875 rem/600 name, 0.82 rem state, 0.7 rem metadata) and no oversized bold header.
- AC-13: The tile shows a small preview thumbnail, or a compact placeholder icon when there
  is no frame, instead of a full-width empty preview; the tile stays below 100 px tall at
  375 and 1440 px.
- AC-14: Device name, status, state/count/last zone and the Open live/History links align in
  one compact row, with a separator between sensors.
- AC-15: The state label remains readable without colour, and both links remain keyboard
  accessible.
- AC-16: With one sensor the card renders a single row containing the device name and
  status, with no separate card header; editing keeps a header for the remove control and
  the row omits its name. With multiple sensors the card has one section header and each
  row names its device.
- AC-17: The mini preview renders whenever a spatial frame exists, even when no target is
  currently reported; the icon placeholder appears only when there is no frame at all.

## Requirements and success

- FR-01: Provide a pure display-state helper for derived presence values where null means
  not-derived, never clear; acceptance: AC-01…AC-04.
- FR-02: Use that helper in the Home card and rail tile, and fix the live panel's null
  text; acceptance: AC-01…AC-04.
- FR-03: Opt the Positions player into the existing Distance prop and pass it to its
  canvas; leave Heatmap without it; acceptance: AC-05…AC-07.
- FR-04: Compute the Positions base view with a sensor-centered, aspect-correct helper and
  use the same view height for symbol and label scaling; acceptance: AC-08…AC-10.
- FR-10: The presence overview rail MUST use the shared Home tile type scale, a compact
  preview/placeholder and a header that adapts to the sensor count; acceptance:
  AC-12…AC-17.
- NFR-01: Add no network call, dependency or persisted setting.
- NFR-02: Preserve live view, zoom, pan, reset and playback behavior.
- NFR-03: Keep controls keyboard operable and verify desktop/mobile widths.
- NFR-04: The compact rail MUST change presentation only; derived values, links and device
  commands stay unchanged.
- NFR-05: Presence copy, state/detail labels, zone chips, the overview model and the server
  readout MUST live in `domain/spatial/presence.ts`; Positions fit/symbol math MUST live in
  `domain/spatial/positions.ts`; touched components MUST only render those models.
- NFR-06: Shared viewport visibility and compact-tile state MUST use composables, not
  component lifecycle code.

## Entities and external interfaces

- `DerivedPresenceDisplay = 'not-derived' | 'occupied' | 'clear'` plus a pure resolver in
  `src/domain/spatial/derived.ts`.
- `centeredSpatialView(bounds, center, aspect): { x, y, width, height }` in
  `src/domain/spatial/radar.ts`; no new wire type.
- `showDistances` view state in `SpatialPositions.vue`; not persisted.
- No HTTP/WS/MQTT/MCP or storage change.

## Questions and assumptions

- Q1: Should the Home card fall back to the source device's firmware occupancy when the
  derived value is missing? Resolution: no. Derived presence is the authority (spec 041);
  the card explains the gap instead of inventing a state. A future firmware-occupancy
  fallback must be explicitly labelled.
- Q2: Should Positions use the live range selector? Resolution: not for now. A
  sensor-centered fit keeps the content visible without adding another control.
- Q3: Is the replay distance wrong? Resolution: no. `SpatialCanvas` derives it from the
  same recorded origin as the marker, so enabling the existing option is correct.
- Assumption A1: The hub's derived device exists once a mapping and at least one zone are
  saved; before the first frame its `occupied` entity has no value.
- Assumption A2: The workspace sensor is the coordinate origin; the live view and the
  replay fit should share that center.

Success criteria:

- SC-01: No presence surface converts missing derived data into Clear.
- SC-02: Positions offers Distance and is sensor-centered on desktop and mobile.
- SC-03: Existing and new unit/browser checks pass with no unrelated change.
- SC-04: The presence overview tile reads as a compact Home row, matches neighbouring type
  sizes, and stays under 100 px tall at both tested widths.
- SC-05: No presence/spatial label or presence derivation remains in the touched components;
  the domain and composable modules are covered by unit tests.
