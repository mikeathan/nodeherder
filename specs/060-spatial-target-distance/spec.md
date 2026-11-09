# Spec: Distance-to-target labels in the spatial viewer

ID: 060 | Status: Draft | Owner: agent draft | Updated: 2026-11-09
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

The spatial viewer draws each target as a dot with a label (`Track 1`) but never says how
far away it is. Reading distance off the grid requires counting grid squares, which is
inaccurate at a glance and impossible on a phone.

Facts observed in the current tree:

- `SpatialFrameTarget` (`src/types/spatial.ts`) carries only `id`, `active`, `x`, `y` —
  the hub frame contract has **no distance field** (`backend/models/devices/spatial.go`).
- The presence device separately exposes `target_distance` (`unit: "m"`, numeric,
  read-only). It is one scalar for the firmware's own current target, is not one of the
  mapped coordinate entities, and does not correspond to targets 2 and 3.
- `SpatialTarget` in `src/domain/spatial/index.ts:6` already includes
  `distance: Math.hypot(x, y)`, but that type feeds the mapping-status panel, not the
  viewer's track pipeline (`useSpatialTracking` → `useSpatialPresence`).
- Viewer positions are filtered by a One Euro filter with a 35 mm stationary dead band and
  an implausible-jump gate (`src/domain/spatial/filter.ts`), so the drawn dot is already
  stable.
- `SpatialViewer.vue` already offers checkboxes for Grid and Field of view next to
  `#extra` display controls; those options are per-view-session state, not persisted.
- Target labels are rendered by `SpatialCanvas.vue` as `<text>` beside each dot, gated by
  `showTargetLabels`.

Scope: a new viewer checkbox that adds the target's distance to its existing label, in the
live spatial view. This is presentation only.

Exclusions: no change to the hub frame contract, the `target_distance` expose mapping, the
positions replay viewer, automation inputs, or persisted workspace settings.

## Scenarios and acceptance

### US-01 — Read the distance without counting grid squares (priority P1)

Independent verification: `src/__tests__/domain/spatial/radar.test.ts` and the spatial
case in `e2e/devices.spec.ts`.

- AC-01: Given the spatial view with a live target, when "Distance" is unchecked (the
  default), then each target label is unchanged (`Track 1`).
- AC-02: Given the same view, when "Distance" is ticked, then each target label also shows
  its distance from the sensor (for example `Track 1 · 2.4 m`).
- AC-03: Given a target at mapped coordinates (x, y), then the shown distance equals
  `hypot(x, y)` converted to metres — the same geometry as the drawn dot, not the
  `target_distance` expose.
- AC-04: Given a target whose position is filtered, then the label is derived from the
  filtered position, so ticking "Distance" does not reintroduce raw jitter.
- AC-05: Given a stationary target with sub-centimetre filter movement, then the displayed
  text does not change every frame (values are rounded to 0.1 m).
- AC-06: Given the "Raw" option is enabled, then the distance still matches the position
  being drawn, because both come from the same chosen point.
- AC-07: Given a target whose coordinate is unavailable, then no distance is invented; the
  label keeps its existing form and no placeholder unit is shown.
- AC-08: Given the checkbox, then it is reachable by keyboard, has an accessible name, and
  its state is announced.

## Requirements and success

- FR-01: MUST add a "Distance" checkbox to the spatial viewer View panel beside Grid and
  Field of view; acceptance: AC-01, AC-02, AC-08.
- FR-02: MUST compute the distance from the same point the canvas draws (filtered unless
  "Raw" is selected) rather than from the `target_distance` expose; acceptance: AC-03,
  AC-04, AC-06.
- FR-03: MUST round the displayed distance to one decimal place in metres so the label is
  readable and does not flicker; acceptance: AC-05.
- FR-04: MUST omit the distance rather than render a placeholder when the position is
  unknown; acceptance: AC-07.
- FR-05: MUST format the value through `src/utils/units.ts` so no call site concatenates a
  unit; acceptance: AC-03.
- NFR-01: MUST add no per-frame allocation or extra observable; the value is derived in the
  existing `renderedTracks` computed pass.
- NFR-02: MUST not add a dependency, persist new settings, or change any wire contract.

## Entities and external interfaces

- `SpatialDisplayTrack` (`x`, `y`, `rawX`, `rawY`, `state`, `label`, …) gains no new field;
  the canvas derives the label text from the point it already selects per frame.
- `SpatialFrame` / `SpatialFrameTarget` — unchanged.
- `target_distance` expose — deliberately unused; the label must agree with the drawing.
- Workspace persistence and viewer option lifetime — unchanged (the checkbox is view state,
  like Grid and Field of view).

## Questions and assumptions

- Q1: Should the label use the firmware's `target_distance` instead? Consequence: it can
  contradict the drawn dot and cannot describe more than one target. Resolution: no —
  compute from the mapped coordinates.
- Q2: Should distance persist per device like a workspace setting? Consequence: it would
  need persistence and migration for a display preference. Resolution: out of scope; it
  behaves like Grid/Field of view.
- Assumption A1: 0.1 m is a useful step for an indoor sensor whose range is metres.
- Assumption A2: The label already tolerates a longer string; it is SVG text positioned
  beside the dot.
